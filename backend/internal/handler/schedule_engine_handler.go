package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/FTATM/iot-monitor-interface/internal/model"
)

type ScheduleEngineHandler struct {
	service model.ScheduleEngineService
}

func NewScheduleEngineHandler(service model.ScheduleEngineService) *ScheduleEngineHandler {
	return &ScheduleEngineHandler{service: service}
}

// Your main app calls this AFTER doing an INSERT or UPDATE in Postgres
func (h *ScheduleEngineHandler) SyncSchedule(w http.ResponseWriter, r *http.Request) {
	// r.PathValue is a new feature in Go 1.22
	schedID := r.PathValue("id")
	if len(schedID) < 32 {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	err := h.service.SyncJob(context.Background(), schedID)
	if err != nil {
		http.Error(w, "Failed to sync job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ScheduleEngineHandler) UnsyncSchedule(w http.ResponseWriter, r *http.Request) {
	schedID := r.PathValue("id")
	if len(schedID) < 32 {
		http.Error(w, "Invalid UUID", http.StatusBadRequest)
		return
	}

	// Capture the boolean to check if it was actively running in memory[cite: 4]
	wasRunning := h.service.CancelJob(schedID)

	if wasRunning {
		log.Printf("[HANDLER] Successfully removed active job %s from memory", schedID)
	} else {
		log.Printf("[HANDLER] Job %s was not found in memory (it may have already finished or was previously unsynced)", schedID)
	}

	w.WriteHeader(http.StatusOK)
}
