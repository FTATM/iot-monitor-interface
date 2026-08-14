package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type DeviceGatewayHandler struct {
	service        model.DeviceGatewayService
	sessionService model.SessionManagerService
}

func NewDeviceGatewayHandler(gatewayService model.DeviceGatewayService, sessionService model.SessionManagerService) *DeviceGatewayHandler {
	return &DeviceGatewayHandler{
		service:        gatewayService,
		sessionService: sessionService,
	}
}

// HTTPTelemetry receives JSON telemetry over HTTP POST
func (h *DeviceGatewayHandler) HTTPTelemetry(w http.ResponseWriter, r *http.Request) {
	var res Response
	var data model.DeviceDataRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if data.DeviceId <= 0 {
		res.Message = "Missing or invalid DeviceId"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	h.sessionService.MarkDeviceActive(data.DeviceId)
	h.service.Add(data)

	res.Message = "Success"
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceGatewayHandler) Command(w http.ResponseWriter, r *http.Request) {
	var res Response
	var req model.CommandRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if req.DeviceId <= 0 || req.Command == "" || req.Protocol == "" {
		res.Message = "Missing required fields (deviceId, command, protocol)"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// Delegate the actual routing to the SessionManager Service
	err := h.sessionService.RouteCommand(r.Context(), &req)
	if err != nil {
		res.Message = "Error routing command"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
			slog.Int("deviceId", req.DeviceId),
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

	command, exists := h.sessionService.PopHTTPCommand(deviceId)
	if !exists {
		respondJson(w, http.StatusNoContent, nil)
		return
	}

	res.Data = map[string]string{"command": command}
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

	res.Message = "Success"
	res.Data = map[string]any{
		"deviceId": deviceId,
		"isOnline": isOnline,
	}

	respondJson(w, http.StatusOK, &res)
}
