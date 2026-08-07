package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/FTATM/iot-monitor-interface/internal/auth"
	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type ScheduleHandler struct {
	service     model.ScheduleService
	roleService model.RoleService
}

func NewScheduleHandler(service model.ScheduleService, roleService model.RoleService) *ScheduleHandler {
	return &ScheduleHandler{service: service, roleService: roleService}
}

func (h *ScheduleHandler) GetAllDetail(w http.ResponseWriter, r *http.Request) {
	var res Response
	scheduleDetails, err := h.service.GetAllDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}
	res.Data = scheduleDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *ScheduleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var res Response
	var createScheduleReq model.CreateScheduleReq
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Scheduler",
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

	if err = json.NewDecoder(r.Body).Decode(&createScheduleReq); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.CreateSchedule(r.Context(), &createScheduleReq, authUserId)
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

func (h *ScheduleHandler) Update(w http.ResponseWriter, r *http.Request) {
	var res Response
	var updateScheduleReq model.UpdateScheduleReq
	var err error

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Scheduler",
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

	if err = json.NewDecoder(r.Body).Decode(&updateScheduleReq); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	userId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	err = h.service.UpdateSchedule(r.Context(), &updateScheduleReq, userId)
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
