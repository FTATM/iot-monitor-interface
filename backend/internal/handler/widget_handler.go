package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type WidgetHandler struct {
	service model.WidgetService
}

func NewWidgetHandler(service model.WidgetService) *WidgetHandler {
	return &WidgetHandler{service: service}
}

func (h *WidgetHandler) GetById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid widget ID", http.StatusBadRequest)
		return
	}

	widget, err := h.service.GetWidgetDetailById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(widget)
}

func (h *WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var err error
	var widgets []model.Widget
	if err = json.NewDecoder(r.Body).Decode(&widgets); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.CreateWidgets(r.Context(), widgets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(widgets)

}

func (h *WidgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	var err error

	var updateWidget []model.Widget
	if err = json.NewDecoder(r.Body).Decode(&updateWidget); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.UpdateWidget(r.Context(), updateWidget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateWidget)
}

func (h *WidgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	var err error

	var idsWidget []int
	if err = json.NewDecoder(r.Body).Decode(&idsWidget); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.service.DeleteWidgets(r.Context(), idsWidget)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
