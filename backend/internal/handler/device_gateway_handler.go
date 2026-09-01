package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type DeviceGatewayHandler struct {
	service        model.DeviceGatewayService
	sessionService model.SessionManagerService
	cacheService   model.CacheService
}

func NewDeviceGatewayHandler(gatewayService model.DeviceGatewayService, sessionService model.SessionManagerService, cacheService model.CacheService) *DeviceGatewayHandler {
	return &DeviceGatewayHandler{
		service:        gatewayService,
		sessionService: sessionService,
		cacheService:   cacheService,
	}
}

// HTTPTelemetry receives JSON telemetry over HTTP POST
func (h *DeviceGatewayHandler) HTTPTelemetry(w http.ResponseWriter, r *http.Request) {
	var res Response
	var data []model.DeviceDataPayloadReq // รองรับข้อมูลแบบ Array

	// 1. ถอดรหัส JSON Body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	ctx := r.Context()
	successCount := 0

	// 2. วนลูปประมวลผลข้อมูลแต่ละอุปกรณ์
	for _, reqData := range data {
		if reqData.DeviceName == "" {
			continue // ข้ามหากไม่มีการระบุชื่ออุปกรณ์
		}

		// 3. ค้นหา Device ID จากชื่อ (⚡ UPDATED: Now using cacheService)
		deviceId, err := h.cacheService.GetDeviceIdByName(ctx, reqData.DeviceName)
		if err != nil || deviceId <= 0 {
			slog.WarnContext(ctx, "Unknown device name in HTTP Telemetry", slog.String("name", reqData.DeviceName))
			continue // ข้ามหากไม่พบอุปกรณ์ในระบบ
		}

		// 4. สร้างโครงสร้าง DeviceData และแปลงค่า Scale
		deviceData := model.DeviceData{
			DeviceId:  deviceId,
			ValueData: int(math.Round(reqData.ValueData * model.DeviceScale)),
		}

		// 5. แจ้งสถานะ Active และส่งต่อให้ Service จัดการ Batch
		h.sessionService.MarkDeviceActive(deviceId)
		h.service.Add(deviceData)
		successCount++
	}

	// ตรวจสอบว่ามีข้อมูลถูกประมวลผลสำเร็จหรือไม่
	if successCount == 0 && len(data) > 0 {
		res.Message = "No valid devices found in payload"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	res.Message = "Success"
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) Command(w http.ResponseWriter, r *http.Request) {
	var res Response
	var req model.GatewayCommand

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	hasDevice := req.DeviceId > 0
	hasGroup := req.GroupId > 0

	// ⚡ FIX: Changed len(req.Payload) != 0 to == 0
	// It ensures you provide either a DeviceId OR a GroupId (not both),
	// ensures the payload has items, and ensures a protocol exists.
	if (hasDevice == hasGroup) || len(req.Payload) == 0 || req.Protocol == "" {
		res.Message = "Missing required fields (Requires DeviceId XOR GroupId, valid Payload, and Protocol)"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	detachedCtx := context.WithoutCancel(r.Context())

	// Delegate the actual routing to the SessionManager Service
	err := h.sessionService.RouteCommand(detachedCtx, &req)
	if err != nil {
		res.Message = "Error routing command"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
			slog.Int("deviceId", req.DeviceId),
			slog.Int("groupId", req.GroupId),
			slog.String("protocol", req.Protocol),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Message = "Command routed successfully"
	respondJson(w, http.StatusOK, &res)
}

// HTTPCommandPolling allows HTTP physical devices to fetch queued commands via GET
func (h *DeviceGatewayHandler) HTTPCommandPolling(w http.ResponseWriter, r *http.Request) {
	var res Response

	deviceIdStr := r.URL.Query().Get("deviceId")
	deviceId, err := strconv.Atoi(deviceIdStr)
	if err != nil || deviceId <= 0 {
		res.Message = "Invalid deviceId parameter"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	h.sessionService.MarkDeviceActive(deviceId)

	payload, exists := h.sessionService.PopHTTPCommand(deviceId)
	if !exists {
		respondJson(w, http.StatusNoContent, nil)
		return
	}

	res.Data = payload
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) DeviceStatus(w http.ResponseWriter, r *http.Request) {
	var res Response // Using your standard response struct

	deviceIdStr := r.URL.Query().Get("deviceId")
	deviceId, err := strconv.Atoi(deviceIdStr)
	if err != nil || deviceId <= 0 {
		res.Message = "Invalid deviceId parameter"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// Ask the Session Manager!
	isOnline := h.sessionService.IsDeviceOnline(deviceId)
	if isOnline {
		if err = h.service.UpdateDeviceLastSeen(r.Context(), deviceId); err != nil {
			slog.Error("Error",
				slog.String("track", err.Error()),
			)
		}
	}

	res.Message = "Success"
	res.Data = map[string]any{
		"deviceId": deviceId,
		"isOnline": isOnline,
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	cacheType := r.URL.Query().Get("type") // expecting "device" or "group"

	if name == "" {
		http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
		return
	}

	switch cacheType {
	case "device":
		// ⚡ Use the injected Cache Service instead of the listener package
		h.cacheService.InvalidateDevice(name)
		slog.Info("Invalidated device cache", slog.String("deviceName", name))
	case "group":
		// ⚡ Use the injected Cache Service
		h.cacheService.InvalidateGroup(name)
		slog.Info("Invalidated group cache", slog.String("groupName", name))
	default:
		http.Error(w, "Invalid 'type' parameter. Must be 'device' or 'group'", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
