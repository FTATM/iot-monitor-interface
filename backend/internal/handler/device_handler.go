package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/auth"
	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type DeviceHandler struct {
	service             model.DeviceService
	roleService         model.RoleService
	deviceGatewayClient model.DeviceGatewayClient
}

type ChartData struct {
	Ingress int `json:"ingress"`
	Egress  int `json:"egress"`
}

func NewDeviceHandler(service model.DeviceService, rs model.RoleService, deviceGatewayClient model.DeviceGatewayClient) *DeviceHandler {
	return &DeviceHandler{service: service, roleService: rs, deviceGatewayClient: deviceGatewayClient}
}

func (h *DeviceHandler) GetAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	deviceDetails, err := h.service.GetAllDeviceDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = deviceDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var devices []model.DeviceCreate

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
		ActionName: "Create",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "No Access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&devices); err != nil {
		res.Message = "Invalid Body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.CreateDevice(r.Context(), devices, authUserId)
	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			res.Message = "Device Duplicate"
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var device model.DeviceUpdate

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
		ActionName: "Update",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "No Access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&device); err != nil {
		res.Message = "Invalid Body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	userId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	err = h.service.UpdateDevice(r.Context(), &device, userId)

	if err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			res.Message = "Device Duplicate"
			respondJson(w, http.StatusBadRequest, &res)
		} else {
			res.Message = "Error"
			slog.ErrorContext(r.Context(), res.Message,
				slog.String("track", err.Error()),
			)
			respondJson(w, http.StatusInternalServerError, &res)
		}
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Device",
		ActionName: "Delete",
	}

	hasAccess, err := h.roleService.Access(r.Context(), acc)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if !hasAccess {
		res.Message = "No Access"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	idStr := r.PathValue("id")
	deleteDeviceId, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = "Invalid user Id"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.DeleteDevice(r.Context(), deleteDeviceId, authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) TriggerManualCommand(w http.ResponseWriter, r *http.Request) {
	var res Response
	var req model.CommandRequest

	// 1. Verify User Auth & Access
	_, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Unauthorized"
		respondJson(w, http.StatusUnauthorized, &res)
		return
	}

	// 2. Decode frontend payload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// 3. Optional: Verify in your database that this user owns this DeviceId

	// 4. Send the command to the Gateway via the client
	if err := h.deviceGatewayClient.ExecuteManualCommand(r.Context(), &req); err != nil {
		slog.ErrorContext(r.Context(), "Failed to execute manual command via Gateway", slog.String("error", err.Error()))
		res.Message = "Failed to communicate with device"
		respondJson(w, http.StatusServiceUnavailable, &res)
		return
	}

	res.Message = "Command executed"
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) GetProtocolType(w http.ResponseWriter, r *http.Request) {
	var res Response
	protocol, err := h.service.GetProtocolType(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = protocol
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) ChartStream(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("deviceId")
	if deviceID == "" {
		http.Error(w, "Missing deviceId", http.StatusBadRequest)
		return
	}

	// Set standard SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	// Create a buffered channel and register it
	clientChan := make(chan model.ChartData, 10)
	h.service.AddClient(deviceID, clientChan)

	// Ensure cleanup when the client disconnects
	defer h.service.RemoveClient(deviceID, clientChan)

	slog.DebugContext(r.Context(), "New Vue client connected to SSE stream!")

	// Listen for data or disconnects
	ctx := r.Context() // This context cancels when the client closes the browser
	for {
		select {
		case <-ctx.Done():
			slog.DebugContext(r.Context(), "Client disconnected from SSE stream")
			return // Exiting the function closes the HTTP connection

		case payload := <-clientChan:
			// SSE format requires data to start with "data: " and end with "\n\n"
			dataBytes, _ := json.Marshal(payload)
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))

			// Push the data to the browser immediately
			err := rc.Flush()
			if err != nil {
				// If flushing fails, the client disconnected or network dropped
				slog.DebugContext(r.Context(), "Stream connection lost")
				return
			}
		}
	}
}

func (h *DeviceHandler) GetAllDeviceName(w http.ResponseWriter, r *http.Request) {
	var res Response
	deviceDetails, err := h.service.GetAllDeviceName(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = deviceDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *DeviceHandler) ChartHistory(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	maxPointsStr := r.URL.Query().Get("maxPoints")
	deviceIdStr := r.URL.Query().Get("deviceIds")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// layout := "2006-01-02T15:04:05"
	fromTime, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		res.Message = "Missing From Date"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	toTime, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		res.Message = "Missing To Date"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	deviceIds := []int{}
	for idStr := range strings.SplitSeq(deviceIdStr, ",") {
		if idInt, err := strconv.Atoi(idStr); err == nil {
			deviceIds = append(deviceIds, idInt)
		}
	}
	maxPoints, err := strconv.Atoi(maxPointsStr)
	if err != nil || maxPoints <= 0 {
		maxPoints = 100 // Default fallback
	}

	historyData, err := h.service.GetChartHistory(r.Context(), deviceIds, maxPoints, fromTime, toTime)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = historyData
	respondJson(w, http.StatusOK, &res)
}
