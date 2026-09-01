package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const activeThreshold = 5 * time.Minute
const staleThreshold = 3 * time.Hour // How long before a device is considered permanently dead
const sweepInterval = 1 * time.Hour  // How often the sweeper checks memory

type sessionManagerService struct {
	repo model.DeviceGatewayRepository

	// Protocol Clients
	mqttClient mqtt.Client
	udpServer  *net.UDPConn

	// Thread-safe registries mapped by Device ID
	mu          sync.RWMutex
	tcpConns    map[int]net.Conn
	udpAddrs    map[int]*net.UDPAddr
	httpQueues  map[int][]model.DeviceCommand
	lastSeenMap map[int]time.Time

	cacheSvc model.CacheService
}

// NewSessionManagerService satisfies the injection signature in app.go
func NewSessionManagerService(repo model.DeviceGatewayRepository, cacheSvc model.CacheService) model.SessionManagerService {
	return &sessionManagerService{
		repo:        repo,
		tcpConns:    make(map[int]net.Conn),
		udpAddrs:    make(map[int]*net.UDPAddr),
		httpQueues:  make(map[int][]model.DeviceCommand),
		lastSeenMap: make(map[int]time.Time),
		cacheSvc:    cacheSvc,
	}
}

// --- SETUP PROTOCOLS ---

func (s *sessionManagerService) SetMQTTClient(client any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := client.(mqtt.Client); ok {
		s.mqttClient = c
		slog.Info("SessionManager: MQTT client registered")
	}
}

func (s *sessionManagerService) SetUDPServer(conn *net.UDPConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.udpServer = conn
	slog.Info("SessionManager: UDP server registered")
}

// --- DEVICE REGISTRATIONS ---

func (s *sessionManagerService) RegisterTCP(deviceId int, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Close any existing stale connection
	if oldConn, exists := s.tcpConns[deviceId]; exists && oldConn != conn {
		oldConn.Close()
	}
	s.tcpConns[deviceId] = conn
}

func (s *sessionManagerService) UnregisterTCP(deviceId int, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, exists := s.tcpConns[deviceId]; exists && existing == conn {
		delete(s.tcpConns, deviceId)
	}
}

func (s *sessionManagerService) RegisterUDP(deviceId int, addr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.udpAddrs[deviceId] = addr
}

func (s *sessionManagerService) MarkDeviceActive(deviceId int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSeenMap[deviceId] = time.Now()
}

func (s *sessionManagerService) IsDeviceOnline(deviceId int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastSeen, exists := s.lastSeenMap[deviceId]
	if !exists {
		return false
	}

	// Consider online if seen within threshold, OR if it has an active TCP socket
	isRecent := time.Since(lastSeen) <= activeThreshold
	_, hasTCP := s.tcpConns[deviceId]

	return isRecent || hasTCP
}

// --- COMMAND ROUTING ---

func (s *sessionManagerService) RouteCommand(ctx context.Context, req *model.GatewayCommand) error {
	switch req.Protocol {
	case "MQTT":
		return s.routeMQTT(ctx, req)
	case "TCP":
		return s.routeTCP(ctx, req)
	case "UDP":
		return s.routeUDP(ctx, req)
	case "HTTP":
		return s.routeHTTP(ctx, req)
	default:
		return fmt.Errorf("unsupported protocol: %s", req.Protocol)
	}
}

func (s *sessionManagerService) routeMQTT(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.RLock()
	client := s.mqttClient
	s.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("MQTT client not initialized")
	}

	payloadBytes, err := json.Marshal(req.Payload)
	if err != nil {
		return err
	}

	var topic string
	if req.GroupId > 0 {
		groupName, err := s.repo.GetDeviceGroupNameById(ctx, req.GroupId)
		if err != nil || groupName == "" {
			return fmt.Errorf("failed to fetch group name for id %d: %w", req.GroupId, err)
		}
		topic = fmt.Sprintf("device-group/%s/cmd", groupName)
	} else if req.DeviceId > 0 {
		deviceName, err := s.repo.GetDeviceNameById(ctx, req.DeviceId)
		if err != nil || deviceName == "" {
			return fmt.Errorf("failed to fetch device name for id %d: %w", req.DeviceId, err)
		}
		topic = fmt.Sprintf("device/%s/cmd", deviceName)
	} else {
		return fmt.Errorf("missing target ID for MQTT command")
	}

	token := client.Publish(topic, 1, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		return token.Error()
	}

	slog.InfoContext(ctx, "Routed command via MQTT", slog.String("topic", topic))
	return nil
}

func (s *sessionManagerService) routeTCP(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cmd := range req.Payload {
		id, err := s.cacheSvc.GetDeviceIdByName(ctx, cmd.DeviceName)
		if err != nil || id <= 0 {
			slog.Warn("Unknown device in TCP payload routing", slog.String("device", cmd.DeviceName))
			continue
		}

		conn, exists := s.tcpConns[id]
		if !exists {
			slog.Warn("Device offline or TCP connection missing", slog.Int("deviceId", id))
			continue
		}

		cmdBytes, _ := json.Marshal(cmd)
		cmdBytes = append(cmdBytes, '\n')

		if _, err := conn.Write(cmdBytes); err != nil {
			slog.Error("Failed to write to TCP connection", slog.Int("deviceId", id), slog.String("err", err.Error()))
		}
	}
	return nil
}

func (s *sessionManagerService) routeUDP(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.RLock()
	server := s.udpServer
	s.mu.RUnlock()

	if server == nil {
		return fmt.Errorf("UDP server not initialized")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cmd := range req.Payload {
		id, err := s.cacheSvc.GetDeviceIdByName(ctx, cmd.DeviceName)
		if err != nil || id <= 0 {
			continue
		}

		addr, exists := s.udpAddrs[id]
		if !exists {
			continue
		}

		cmdBytes, _ := json.Marshal(cmd)
		if _, err := server.WriteToUDP(cmdBytes, addr); err != nil {
			slog.Error("Failed to write to UDP", slog.Int("deviceId", id), slog.String("err", err.Error()))
		}
	}
	return nil
}

func (s *sessionManagerService) routeHTTP(ctx context.Context, req *model.GatewayCommand) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cmd := range req.Payload {
		id, err := s.cacheSvc.GetDeviceIdByName(ctx, cmd.DeviceName)
		if err != nil || id <= 0 {
			continue
		}
		s.httpQueues[id] = append(s.httpQueues[id], cmd)
	}
	return nil
}

// --- HTTP POLLING QUEUE ---

func (s *sessionManagerService) PopHTTPCommand(deviceId int) ([]model.DeviceCommand, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	commands, exists := s.httpQueues[deviceId]
	if !exists || len(commands) == 0 {
		return nil, false
	}

	// Empty the queue after popping
	delete(s.httpQueues, deviceId)
	return commands, true
}

func (s *sessionManagerService) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(sweepInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.sweepStaleConnections()
			case <-ctx.Done():
				ticker.Stop()
				slog.Info("Shutting down session manager sweeper...")
				return
			}
		}
	}()
}

func (s *sessionManagerService) sweepStaleConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	clearedCount := 0

	// Iterate through the lastSeenMap to find devices that haven't reported in
	for deviceId, lastSeen := range s.lastSeenMap {
		if now.Sub(lastSeen) > staleThreshold {
			// Device is stale, clear all its memory footprints
			delete(s.lastSeenMap, deviceId)
			delete(s.httpQueues, deviceId)
			delete(s.udpAddrs, deviceId)

			// If a TCP connection somehow got orphaned without triggering UnregisterTCP, forcefully close and delete it
			if conn, exists := s.tcpConns[deviceId]; exists {
				conn.Close()
				delete(s.tcpConns, deviceId)
			}

			clearedCount++
		}
	}

	if clearedCount > 0 {
		slog.Debug("Session manager swept dead connections", slog.Int("cleared", clearedCount))
	}
}
