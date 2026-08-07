package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/auth"
	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type DeviceHandler struct {
	service     model.DeviceService
	roleService model.RoleService
}

func NewDeviceHandler(service model.DeviceService, rs model.RoleService) *DeviceHandler {
	return &DeviceHandler{service: service, roleService: rs}
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
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
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
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
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
