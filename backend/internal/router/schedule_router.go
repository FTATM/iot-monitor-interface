package router

import "net/http"

func SetupSchedule(handlers RouterHandlers) *http.ServeMux {
	mux := http.NewServeMux()

	scheduleEngine := http.NewServeMux()
	// 1. Change the base prefix to match the client
	mux.Handle("/scheduleengine/", http.StripPrefix("/scheduleengine", scheduleEngine))

	// 2. Adjust the remaining path to place {id} before /sync
	scheduleEngine.HandleFunc("POST /{id}/sync", handlers.ScheduleEngine.SyncSchedule)
	scheduleEngine.HandleFunc("DELETE /{id}/sync", handlers.ScheduleEngine.UnsyncSchedule)

	return mux
}
