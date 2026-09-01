package listener

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

// เพิ่ม cacheSvc เข้ามาในพารามิเตอร์
func StartUDPServer(ctx context.Context, port string, svc model.DeviceGatewayService, sessionSvc model.SessionManagerService, cacheSvc model.CacheService) {
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

	// ลงทะเบียน UDP Connection เข้า SessionManager สำหรับส่งคำสั่ง (Routing)
	sessionSvc.SetUDPServer(conn)

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

		packetData := bytes.TrimSpace(buffer[:n])
		if len(packetData) == 0 {
			continue
		}

		var incomingData []model.DeviceDataPayloadReq

		// ตรวจสอบรูปแบบ Payload ว่าเป็น Array หรือ Object
		switch packetData[0] {
		case '[':
			if err := json.Unmarshal(packetData, &incomingData); err != nil {
				slog.Debug("Invalid JSON array received over UDP", slog.String("ip", remoteAddr.String()))
				continue
			}
		case '{':
			var single model.DeviceDataPayloadReq
			if err := json.Unmarshal(packetData, &single); err != nil {
				slog.Debug("Invalid JSON object received over UDP", slog.String("ip", remoteAddr.String()))
				continue
			}
			incomingData = append(incomingData, single)
		default:
			continue
		}

		// ประมวลผลข้อมูลแต่ละรายการใน Payload
		for _, req := range incomingData {
			if req.DeviceName == "" {
				continue
			}

			// ค้นหา Device ID จาก Cache Service ตัวใหม่
			deviceId, err := cacheSvc.GetDeviceIdByName(ctx, req.DeviceName)
			if err != nil || deviceId <= 0 {
				slog.Warn("Unknown device in UDP payload", slog.String("device", req.DeviceName))
				continue
			}

			// ลงทะเบียน IP/Port เพื่อใช้ส่ง Command กลับ และอัปเดตสถานะ Online
			sessionSvc.RegisterUDP(deviceId, remoteAddr)
			sessionSvc.MarkDeviceActive(deviceId)

			// สร้างข้อมูลสำหรับ Database พร้อมปรับ Scale
			deviceData := model.DeviceData{
				DeviceId:  deviceId,
				ValueData: int(math.Round(req.ValueData * model.DeviceScale)),
			}

			// ส่งเข้าระบบ Batcher
			svc.Add(deviceData)
		}
	}
}
