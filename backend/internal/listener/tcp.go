package listener

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

func StartTCPServer(ctx context.Context, port string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, cacheSvc model.CacheService) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to start TCP listener", slog.String("error", err.Error()))
		return
	}
	slog.InfoContext(ctx, "TCP Listener started", slog.String("port", port))

	// Wait for shutdown signal to close the listener
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down TCP listener...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return // Context canceled, safe exit
			}
			slog.WarnContext(ctx, "Failed to accept TCP connection", slog.String("error", err.Error()))
			continue
		}

		// Handle each device in its own goroutine so they don't block each other
		go handleTCPConnection(ctx, conn, svc, sessionSvc, cacheSvc)
	}
}

func handleTCPConnection(ctx context.Context, conn net.Conn, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, cacheSvc model.CacheService) {
	remoteAddr := conn.RemoteAddr().String()
	slog.InfoContext(ctx, "Device connected via TCP", slog.String("ip", remoteAddr))

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		// Send a keep-alive probe every 3 minutes.
		// If the device is physically gone, this connection will fail and close itself.
		tcpConn.SetKeepAlivePeriod(3 * time.Minute)
	}

	var connectedDeviceID int

	defer func() {
		conn.Close()
		// Pass the exact connection object so it doesn't accidentally delete a new takeover connection
		if connectedDeviceID > 0 {
			sessionSvc.UnregisterTCP(connectedDeviceID, conn)
		}
		slog.InfoContext(ctx, "Device disconnected from TCP", slog.String("ip", remoteAddr))
	}()

	// Read data line-by-line (assuming devices send JSON separated by \n)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		rawLine := bytes.TrimSpace(scanner.Bytes())
		if len(rawLine) == 0 {
			continue
		}

		var incomingData []model.DeviceDataPayloadReq

		// 1. Handle both Array and Object JSON payloads
		switch rawLine[0] {
		case '[':
			if err := json.Unmarshal(rawLine, &incomingData); err != nil {
				slog.Debug("Invalid JSON array received over TCP", slog.String("ip", remoteAddr))
				continue
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(rawLine, &single); err != nil {
				slog.Debug("Invalid JSON object received over TCP", slog.String("ip", remoteAddr))
				continue
			}
			incomingData = append(incomingData, single)
		default:
			continue
		}

		// 2. Process the payload and resolve the ID
		for _, req := range incomingData {
			if req.DeviceName == "" {
				continue
			}

			// Look up the Device ID using our Cache Service
			deviceId, err := cacheSvc.GetDeviceIdByName(ctx, req.DeviceName)
			if err != nil || deviceId <= 0 {
				slog.Warn("Unknown device in TCP payload", slog.String("device", req.DeviceName))
				continue
			}

			// 3. Register the TCP connection for the very first valid device ID we see on this socket
			if connectedDeviceID == 0 {
				connectedDeviceID = deviceId
				sessionSvc.RegisterTCP(deviceId, conn)
			}

			// Mark device as active
			sessionSvc.MarkDeviceActive(deviceId)

			// Transform to database model with scaled value
			deviceData := model.DeviceData{
				DeviceId:  deviceId,
				ValueData: int(math.Round(req.ValueData * model.DeviceScale)),
			}

			// Pass to the Batcher Service
			svc.Add(deviceData)
		}
	}

	// FIX: Check why the scanner loop stopped (Error vs normal EOF)
	if err := scanner.Err(); err != nil {
		slog.ErrorContext(ctx, "TCP stream read error",
			slog.String("error", err.Error()),
			slog.String("ip", remoteAddr),
		)
	}
}
