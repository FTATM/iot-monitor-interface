package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/auth"
	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type CanvasHandler struct {
	service     model.CanvasService
	roleService model.RoleService
}

func NewCanvasHandler(service model.CanvasService, roleService model.RoleService) *CanvasHandler {
	return &CanvasHandler{service: service, roleService: roleService}
}

func (h *CanvasHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var res Response

	canvas, err := h.service.GetAllCanvas(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = canvas
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) GetAllCanvasRoleDetail(w http.ResponseWriter, r *http.Request) {
	var res Response

	canvas, err := h.service.GetAllCanvasRoleDetail(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if canvas == nil {
		canvas = make([]model.CanvasRoleDetail, 0)
	}
	res.Data = canvas
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) GetDetailById(w http.ResponseWriter, r *http.Request) {
	var res Response
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = "invalid widget Id"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	canvasDetail, err := h.service.GetCanvasDetailById(r.Context(), id)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = canvasDetail
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) GetAllDetailByUser(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	canvasDetails, err := h.service.GetAllCanvasDetailByUserRole(r.Context(), authUserId)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = canvasDetails
	respondJson(w, http.StatusOK, &res)
}

func (h *CanvasHandler) UpsertCanvasRole(w http.ResponseWriter, r *http.Request) {
	var err error
	var res Response
	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get user id"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:     authUserId,
		MenuName:   "Canvas Access",
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

	upsertCanvasRole := model.UpsertCanvasRole{}
	if err = json.NewDecoder(r.Body).Decode(&upsertCanvasRole); err != nil {
		res.Message = "Invalid request body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	err = h.service.UpsertCanvasRole(r.Context(), &upsertCanvasRole, authUserId)
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
