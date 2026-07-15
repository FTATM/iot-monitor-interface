package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type CanvasHandler struct {
	service model.CanvasService
}

func NewCanvasHandler(service model.CanvasService) *CanvasHandler {
	return &CanvasHandler{service: service}
}

func (h *CanvasHandler) GetDetailById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid widget ID", http.StatusBadRequest)
		return
	}

	canvasDetail, err := h.service.GetCanvasDetailById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(canvasDetail)
}

func (h *CanvasHandler) GetAllByUserId(w http.ResponseWriter, r *http.Request) {
	var err error
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)

	canvas, err := h.service.GetAllCanvasByUserId(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(canvas)

}
