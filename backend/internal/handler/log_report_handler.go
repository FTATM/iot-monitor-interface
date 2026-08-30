package handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/xuri/excelize/v2"
)

type LogReportHandler struct {
	service model.LogReportService
}

func NewLogReportHandler(service model.LogReportService) *LogReportHandler {
	return &LogReportHandler{service: service}
}

func (h *LogReportHandler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")       // "system" or "device"
	format := r.URL.Query().Get("format") // "json", "csv", or "excel"

	// Parse Date Filters (Default to last 24 hours if missing)
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	toTime := time.Now()
	fromTime := toTime.Add(-24 * time.Hour)

	if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
		fromTime = t
	}
	if t, err := time.Parse(time.RFC3339, toStr); err == nil {
		toTime = t
	}

	entityTypesStr := r.URL.Query().Get("entityTypes")
	var entityTypes []string

	for idStr := range strings.SplitSeq(entityTypesStr, ",") {
		entityTypes = append(entityTypes, idStr)
	}

	filter := model.LogFilter{
		From:        fromTime,
		To:          toTime,
		Keyword:     r.URL.Query().Get("keyword"),
		EntityTypes: entityTypes,
	}

	// Route to correct exporter based on the active tab
	if tab == "system" {
		h.exportSystemLogs(w, r, format, filter)
	} else if tab == "device" {
		h.exportDeviceLogs(w, r, format, filter)
	} else {
		http.Error(w, "Invalid tab parameter", http.StatusBadRequest)
	}
}

func (h *LogReportHandler) SearchLogs(w http.ResponseWriter, r *http.Request) {
	var res Response
	tab := r.URL.Query().Get("tab") // "system" or "device"

	// Parse Pagination
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50 // Default table limit
	}

	offset := (page - 1) * limit

	// Parse Date Filters (Default to last 24 hours)
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	toTime := time.Now()
	fromTime := toTime.Add(-24 * time.Hour)

	if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
		fromTime = t
	}
	if t, err := time.Parse(time.RFC3339, toStr); err == nil {
		toTime = t
	}
	entityTypesStr := r.URL.Query().Get("entityTypes")
	var entityTypes []string

	for idStr := range strings.SplitSeq(entityTypesStr, ",") {
		entityTypes = append(entityTypes, idStr)
	}

	sortDesc := r.URL.Query().Get("sortDesc") == "true"
	sortBy := r.URL.Query().Get("sortBy")

	filter := model.LogFilter{
		From:        fromTime,
		To:          toTime,
		Keyword:     r.URL.Query().Get("keyword"),
		EntityTypes: entityTypes,
		Limit:       limit,
		Offset:      offset,
		SortBy:      sortBy,
		SortDesc:    sortDesc,
	}

	type PaginatedResponse struct {
		Logs       any `json:"logs"`
		TotalCount int `json:"totalCount"`
	}

	// Fetch Data based on Tab
	switch tab {
	case "system":
		logs, err := h.service.SearchSystemLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to search system logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		count, err := h.service.CountSystemLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to count system logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		if logs == nil {
			logs = []model.AuditLogReport{}
		}

		// ⚡ Package data and count
		res.Data = PaginatedResponse{Logs: logs, TotalCount: count}
		res.Message = "Success"
		respondJson(w, http.StatusOK, &res)

	case "device":
		logs, err := h.service.SearchDeviceLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to search device logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		count, err := h.service.CountDeviceLogs(r.Context(), filter)
		if err != nil {
			res.Message = "Failed to count device logs"
			slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
			respondJson(w, http.StatusInternalServerError, &res)
			return
		}

		if logs == nil {
			logs = []model.DeviceDataLogReport{}
		}

		// ⚡ Package data and count
		res.Data = PaginatedResponse{Logs: logs, TotalCount: count}
		res.Message = "Success"
		respondJson(w, http.StatusOK, &res)

	default:
		res.Message = "Invalid tab parameter"
		respondJson(w, http.StatusBadRequest, &res)
	}
}

func (h *LogReportHandler) GetEntityTypes(w http.ResponseWriter, r *http.Request) {
	var res Response
	types, err := h.service.GetAuditLogEntityTypes(r.Context())
	if err != nil {
		res.Message = "Failed to fetch entity types"
		slog.ErrorContext(r.Context(), res.Message, slog.String("track", err.Error()))
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	if types == nil {
		types = []string{}
	}

	res.Data = types
	res.Message = "Success"
	respondJson(w, http.StatusOK, &res)
}

// --- SYSTEM LOGS (Audit) EXPORTER ---
func (h *LogReportHandler) exportSystemLogs(w http.ResponseWriter, r *http.Request, format string, filter model.LogFilter) {
	var res Response
	logs, err := h.service.GetSystemLogsForExport(r.Context(), filter)
	if err != nil {
		res.Message = "Failed to fetch system logs"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=system_logs.csv")
		writer := csv.NewWriter(w)
		// ⚡ Removed Old/New Data headers
		writer.Write([]string{"Timestamp", "Entity Type", "Entity ID", "Action", "User"})

		for _, log := range logs {
			writer.Write([]string{
				log.CreatedAt.Format(time.RFC3339),
				log.EntityType,
				log.EntityId,
				log.Action,
				log.Username,
			})
		}
		writer.Flush()

	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment;filename=system_logs.json")

		// ⚡ Removed Old/New Data fields
		var exportData []struct {
			Timestamp  string `json:"timestamp"`
			EntityType string `json:"entityType"`
			EntityId   string `json:"entityId"`
			Action     string `json:"action"`
			User       string `json:"user"`
		}

		for _, log := range logs {
			exportData = append(exportData, struct {
				Timestamp  string `json:"timestamp"`
				EntityType string `json:"entityType"`
				EntityId   string `json:"entityId"`
				Action     string `json:"action"`
				User       string `json:"user"`
			}{
				Timestamp:  log.CreatedAt.Format(time.RFC3339),
				EntityType: log.EntityType,
				EntityId:   log.EntityId,
				Action:     log.Action,
				User:       log.Username,
			})
		}
		json.NewEncoder(w).Encode(exportData)

	case "excel":
		f := excelize.NewFile()
		// ⚡ Removed Old/New Data columns
		f.SetCellValue("Sheet1", "A1", "Timestamp")
		f.SetCellValue("Sheet1", "B1", "Entity Type")
		f.SetCellValue("Sheet1", "C1", "Entity ID")
		f.SetCellValue("Sheet1", "D1", "Action")
		f.SetCellValue("Sheet1", "E1", "User")

		for i, log := range logs {
			rowNum := i + 2
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), log.CreatedAt.Format("2006-01-02 15:04:05"))
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), log.EntityType)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), log.EntityId)
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowNum), log.Action)
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", rowNum), log.Username)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment;filename=system_logs.xlsx")
		f.Write(w)
	}
}

// --- DEVICE LOGS (Telemetry) EXPORTER ---
func (h *LogReportHandler) exportDeviceLogs(w http.ResponseWriter, r *http.Request, format string, filter model.LogFilter) {
	var res Response
	logs, err := h.service.GetDeviceLogsForExport(r.Context(), filter)
	if err != nil {
		res.Message = "Failed to fetch device logs"
		slog.ErrorContext(r.Context(), res.Message,
			slog.String("track", err.Error()),
		)

		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	switch format {
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=device_logs.csv")
		writer := csv.NewWriter(w)
		// ⚡ Added Device Name
		writer.Write([]string{"Timestamp", "Device ID", "Device Name", "Source", "Value"})

		for _, log := range logs {
			writer.Write([]string{
				log.ReceivedAt.Format(time.RFC3339),
				fmt.Sprintf("%d", log.DeviceId),
				log.DeviceName, // ⚡ Added
				log.Source,
				fmt.Sprintf("%d", log.ValueData),
			})
		}
		writer.Flush()

	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment;filename=device_logs.json")

		var exportData []struct {
			Timestamp  string `json:"timestamp"`
			DeviceId   int    `json:"deviceId"`
			DeviceName string `json:"deviceName"` // ⚡ Added
			Source     string `json:"source"`
			Value      int    `json:"value"`
		}

		for _, log := range logs {
			exportData = append(exportData, struct {
				Timestamp  string `json:"timestamp"`
				DeviceId   int    `json:"deviceId"`
				DeviceName string `json:"deviceName"`
				Source     string `json:"source"`
				Value      int    `json:"value"`
			}{
				Timestamp:  log.ReceivedAt.Format(time.RFC3339),
				DeviceId:   log.DeviceId,
				DeviceName: log.DeviceName, // ⚡ Added
				Source:     log.Source,
				Value:      log.ValueData,
			})
		}
		json.NewEncoder(w).Encode(exportData)

	case "excel":
		f := excelize.NewFile()
		f.SetCellValue("Sheet1", "A1", "Timestamp")
		f.SetCellValue("Sheet1", "B1", "Device ID")
		f.SetCellValue("Sheet1", "C1", "Device Name") // ⚡ Added
		f.SetCellValue("Sheet1", "D1", "Source")
		f.SetCellValue("Sheet1", "E1", "Value")

		for i, log := range logs {
			rowNum := i + 2
			f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowNum), log.ReceivedAt.Format("2006-01-02 15:04:05"))
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowNum), log.DeviceId)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowNum), log.DeviceName) // ⚡ Added
			f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowNum), log.Source)
			f.SetCellValue("Sheet1", fmt.Sprintf("E%d", rowNum), log.ValueData)
		}

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment;filename=device_logs.xlsx")
		f.Write(w)
	}
}
