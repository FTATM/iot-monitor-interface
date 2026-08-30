package listener

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

func StartUDPServer(ctx context.Context, port string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to resolve UDP address", slog.String("error", err.Error()))
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to start UDP listener", slog.String("error", err.Error()))
		return
	}
	slog.InfoContext(ctx, "UDP Listener started", slog.String("port", port))

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down UDP listener...")
		conn.Close()
	}()

	buffer := make([]byte, 2048) // Buffer size for incoming packets

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.WarnContext(ctx, "UDP read error", slog.String("error", err.Error()))
			continue
		}

		packetData := buffer[:n]

		//! fixed to data payload new
		var data model.DeviceData
		if err := json.Unmarshal(packetData, &data); err != nil {
			slog.Debug("Invalid JSON received over UDP",
				slog.String("ip", remoteAddr.String()),
			)
			continue
		}

		if data.DeviceId > 0 {
			sessionSvc.MarkDeviceActive(data.DeviceId)
			sessionSvc.RegisterUDP(data.DeviceId, remoteAddr)
		}

		// Pass to the Batcher Service!
		svc.Add(data)
	}
}
