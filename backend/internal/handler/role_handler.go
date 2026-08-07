package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/auth"
	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type RoleHandler struct {
	service model.RoleService
}

func NewRoleHandler(service model.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var res Response
	var err error
	var upsertRole model.UpsertRole

	authUserId, ok := r.Context().Value(auth.AuthUserIdKey).(int)
	if !ok {
		res.Message = "Can't get userId"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	acc := &model.Access{
		UserId:   authUserId,
		MenuName: "Role",
	}

	if err = json.NewDecoder(r.Body).Decode(&upsertRole); err != nil {
		res.Message = "Invalid body"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	if upsertRole.RoleId == 0 {
		acc.ActionName = "Create"
	} else {
		acc.ActionName = "Update"
	}

	hasAccess, err := h.service.Access(r.Context(), acc)
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

	err = h.service.UpsertRole(r.Context(), &upsertRole)
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

func (h *RoleHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var res Response
	roles, err := h.service.GetAll(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = roles
	respondJson(w, http.StatusOK, &res)
}

func (h *RoleHandler) GetMenuAvailable(w http.ResponseWriter, r *http.Request) {
	var res Response
	menuAvailable, err := h.service.GetMenuActionAvailable(r.Context())
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = menuAvailable
	respondJson(w, http.StatusOK, &res)
}

func (h *RoleHandler) GetDetailById(w http.ResponseWriter, r *http.Request) {
	var res Response
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		res.Message = "Invalid widget ID"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	detail, err := h.service.GetDetailById(r.Context(), id)
	if err != nil {
		res.Message = "Error"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	res.Data = detail
	respondJson(w, http.StatusOK, &res)
}
