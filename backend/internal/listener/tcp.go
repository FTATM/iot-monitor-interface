package listener

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"net"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

func StartTCPServer(ctx context.Context, port string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService) {
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
		go handleTCPConnection(ctx, conn, svc, sessionSvc)
	}
}

func handleTCPConnection(ctx context.Context, conn net.Conn, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService) {
	remoteAddr := conn.RemoteAddr().String()
	slog.InfoContext(ctx, "Device connected via TCP", slog.String("ip", remoteAddr))

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
		rawLine := scanner.Bytes()

		//! fixed to data payload new
		var data model.DeviceData
		if err := json.Unmarshal(rawLine, &data); err != nil {
			slog.Debug("Invalid JSON received over TCP",
				slog.String("ip", remoteAddr),
			)
			continue
		}

		if data.DeviceId > 0 {
			connectedDeviceID = data.DeviceId // Save the ID
			sessionSvc.MarkDeviceActive(data.DeviceId)
			sessionSvc.RegisterTCP(data.DeviceId, conn)
		}

		// Pass to the Batcher Service!
		svc.Add(data)
	}

	// FIX: Check why the scanner loop stopped (Error vs normal EOF)
	if err := scanner.Err(); err != nil {
		slog.ErrorContext(ctx, "TCP stream read error",
			slog.String("error", err.Error()),
			slog.String("ip", remoteAddr),
		)
	}
}
