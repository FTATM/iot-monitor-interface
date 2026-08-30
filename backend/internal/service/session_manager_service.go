package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type sessionManagerService struct {
	mu                sync.RWMutex
	tcpConns          map[int]net.Conn
	udpAddrs          map[int]UDPSession
	httpQueue         map[int]QueuedCommand
	mqttClient        mqtt.Client
	udpServer         *net.UDPConn
	lastActive        map[int]time.Time
	deviceGatewayRepo model.DeviceGatewayRepository
	prefixError       string
}

type QueuedCommand struct {
	Payload   []model.DeviceCommand
	ExpiresAt time.Time
}

type UDPSession struct {
	Addr     *net.UDPAddr
	LastSeen time.Time
}

func NewSessionManagerService(deviceGatewayRepo model.DeviceGatewayRepository) model.SessionManagerService {
	service := &sessionManagerService{
		tcpConns:          make(map[int]net.Conn),
		udpAddrs:          make(map[int]UDPSession),
		httpQueue:         make(map[int]QueuedCommand),
		lastActive:        make(map[int]time.Time),
		deviceGatewayRepo: deviceGatewayRepo,
		prefixError:       "sessionManagerService",
	}

	// Start the background cleanup process
	service.startSweeper()

	return service
}

func (s *sessionManagerService) SetMQTTClient(client any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Type assertion back to mqtt.Client
	if mqttClient, ok := client.(mqtt.Client); ok {
		s.mqttClient = mqttClient
	}
}

func (s *sessionManagerService) SetUDPServer(conn *net.UDPConn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.udpServer = conn
}

func (s *sessionManagerService) RegisterTCP(deviceId int, newConn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// THE TAKEOVER: If an old connection exists for this ID, forcefully close it.
	if oldConn, exists := s.tcpConns[deviceId]; exists && oldConn != newConn {
		slog.Warn("Device Takeover: Closing old TCP socket for new connection", slog.Int("deviceId", deviceId))
		oldConn.Close()
	}

	// Save the new connection
	s.tcpConns[deviceId] = newConn
}

func (s *sessionManagerService) UnregisterTCP(deviceId int, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// THE SAFE DELETE: Only remove the device if the map holds THIS exact connection
	if existingConn, exists := s.tcpConns[deviceId]; exists && existingConn == conn {
		delete(s.tcpConns, deviceId)
	}
}

func (s *sessionManagerService) RegisterUDP(deviceId int, addr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.udpAddrs[deviceId] = UDPSession{
		Addr:     addr,
		LastSeen: time.Now(),
	}
}

func (s *sessionManagerService) RouteCommand(ctx context.Context, req *model.GatewayCommand) error {
	const fname = "RouteCommand"
	s.mu.Lock()
	defer s.mu.Unlock()

	switch req.Protocol {
	case "MQTT":
		if s.mqttClient != nil {
			// ⚡ FIX 1: Convert the Go slice into a JSON byte array
			jsonBytes, err := json.Marshal(req.Payload)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				return err
			}

			if req.DeviceId > 0 {
				if len(req.Payload) == 0 {
					return fmt.Errorf("empty payload for device %d", req.DeviceId)
				}

				// Using the optimized approach we discussed!
				name := req.Payload[0].DeviceName
				topic := fmt.Sprintf("device/%s/cmd", name)

				// ⚡ FIX 2: Pass 'jsonBytes' here, NOT req.Payload
				token := s.mqttClient.Publish(topic, 1, false, jsonBytes)
				token.Wait()
				return token.Error()

			} else if req.GroupId > 0 {
				name, err := s.deviceGatewayRepo.GetDeviceGroupNameById(ctx, req.GroupId)
				if err != nil {
					return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
				}
				topic := fmt.Sprintf("device-group/%s/cmd", name)

				// ⚡ FIX 3: Pass 'jsonBytes' here, NOT req.Payload
				token := s.mqttClient.Publish(topic, 1, false, jsonBytes)
				token.Wait()
				return token.Error()
			}
		}
		return fmt.Errorf("MQTT client not available")

	case "TCP":
		if conn, exists := s.tcpConns[req.DeviceId]; exists {
			jsonBytes, err := json.Marshal(req.Payload)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				return err
			}
			_, err = conn.Write(jsonBytes)
			return err
		}
		return fmt.Errorf("device %d TCP connection offline", req.DeviceId)

	case "UDP":
		if session, exists := s.udpAddrs[req.DeviceId]; exists && s.udpServer != nil {
			jsonBytes, err := json.Marshal(req.Payload)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				return err
			}
			_, err = s.udpServer.WriteToUDP(jsonBytes, session.Addr)
			return err
		}
		return fmt.Errorf("device %d UDP address unknown or server offline", req.DeviceId)

	case "HTTP":
		s.httpQueue[req.DeviceId] = QueuedCommand{
			Payload:   req.Payload,
			ExpiresAt: time.Now().Add(60 * time.Second),
		}
		return nil
	}

	return fmt.Errorf("unsupported protocol: %s", req.Protocol)
}

func (s *sessionManagerService) PopHTTPCommand(deviceId int) ([]model.DeviceCommand, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	queuedCmd, exists := s.httpQueue[deviceId]
	if !exists {
		return queuedCmd.Payload, false
	}

	// Always delete it so it never gets fetched twice
	delete(s.httpQueue, deviceId)

	// Check if the command has expired (Stale Data Protection)
	if time.Now().After(queuedCmd.ExpiresAt) {
		// It expired, so we return false as if it was empty
		return queuedCmd.Payload, false
	}

	// It is valid and fresh, send it to the device!
	return queuedCmd.Payload, true
}

func (s *sessionManagerService) MarkDeviceActive(deviceId int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActive[deviceId] = time.Now()
}

func (s *sessionManagerService) IsDeviceOnline(deviceId int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. If it has an active TCP socket, it is 100% online
	if _, exists := s.tcpConns[deviceId]; exists {
		return true
	}

	if lastTime, exists := s.lastActive[deviceId]; exists {
		// If the last active time is within the last 3 minutes, consider it online
		if time.Since(lastTime) < 3*time.Minute {
			return true
		}
	}

	return false
}

func (s *sessionManagerService) startSweeper() {
	ticker := time.NewTicker(1 * time.Hour)

	go func() {
		for range ticker.C {
			s.mu.Lock()
			now := time.Now()

			// 1. Clean up HTTP Queue
			for deviceId, cmd := range s.httpQueue {
				if now.After(cmd.ExpiresAt) {
					delete(s.httpQueue, deviceId)
				}
			}

			// 2. Clean up Dead UDP Devices (e.g., no packets in 24 hours)
			udpTimeout := now.Add(-24 * time.Hour)
			for deviceId, session := range s.udpAddrs {
				if session.LastSeen.Before(udpTimeout) {
					delete(s.udpAddrs, deviceId)
				}
			}

			s.mu.Unlock()
		}
	}()
}
