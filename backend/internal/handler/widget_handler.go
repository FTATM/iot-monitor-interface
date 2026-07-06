package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type WidgetHandler struct {
	repo model.WidgetRepository
}

func NewWidgetHandler(repo model.WidgetRepository) *WidgetHandler {
	return &WidgetHandler{repo: repo}
}

func (h *WidgetHandler) GetWidgetById(w http.ResponseWriter, r *http.Request) {
	// 1. API Logic: Parse the HTTP request
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid widget ID", http.StatusBadRequest)
		return
	}

	// 2. Business Logic: This used to be in the Service layer
	if id <= 0 {
		http.Error(w, "ID must be a positive number", http.StatusBadRequest)
		return
	}

	// 3. Data Logic: Call the Repository directly
	widget, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// 4. API Logic: Format the HTTP response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(widget)
}
