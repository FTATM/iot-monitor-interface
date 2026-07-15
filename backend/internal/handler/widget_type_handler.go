package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type WidgetTypeHandler struct {
	service model.WidgetTypeService
}

func NewWidgetTypeHandler(service model.WidgetTypeService) *WidgetTypeHandler {
	return &WidgetTypeHandler{service: service}
}

func (h *WidgetTypeHandler) GetById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid widget ID", http.StatusBadRequest)
		return
	}

	widgetType, err := h.service.GetWidgetTypeById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(widgetType)
}

func (h *WidgetTypeHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	widgetTypes, err := h.service.GetWidgetTypeAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(widgetTypes)
}
