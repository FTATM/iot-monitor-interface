package router

import (
	"net/http"

	"github.com/FTATM/iot-monitor-interface/internal/handler"
)

// ApiHandlers bundles everything together so Setup() stays clean
type ApiHandlers struct {
	Widget         *handler.WidgetHandler
	Canvas         *handler.CanvasHandler
	WidgetType     *handler.WidgetTypeHandler
	User           *handler.UserHandler
	Device         *handler.DeviceHandler
	Schedule       *handler.ScheduleHandler
	ScheduleEngine *handler.ScheduleEngineHandler
	Role           *handler.RoleHandler
	// Add more as the app grows...
}

// SetupApi now takes the bundle and applies middleware
func SetupApi(handlers ApiHandlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.WriteHeader(http.StatusOK)
	})

	widgetMux := http.NewServeMux()
	mux.Handle("/widget/", http.StripPrefix("/widget", widgetMux))
	widgetMux.HandleFunc("GET /getbyid/{id}", handlers.Widget.GetById)
	widgetMux.HandleFunc("POST /upsert", handlers.Widget.Upsert)

	widgetTypeMux := http.NewServeMux()
	mux.Handle("/widgettype/", http.StripPrefix("/widgettype", widgetTypeMux))
	widgetTypeMux.HandleFunc("GET /getbyid", handlers.WidgetType.GetById)
	widgetTypeMux.HandleFunc("GET /getall", handlers.WidgetType.GetAll)

	canvasMux := http.NewServeMux()
	mux.Handle("/canvas/", http.StripPrefix("/canvas", canvasMux))
	canvasMux.HandleFunc("GET /getall", handlers.Canvas.GetAll)
	canvasMux.HandleFunc("GET /getdetailbyid/{id}", handlers.Canvas.GetDetailById)
	canvasMux.HandleFunc("GET /getalldetailbyuser", handlers.Canvas.GetAllDetailByUser)
	canvasMux.HandleFunc("GET /getallcanvasroledetail", handlers.Canvas.GetAllCanvasRoleDetail)
	canvasMux.HandleFunc("POST /upsertcanvasrole", handlers.Canvas.UpsertCanvasRole)

	userMux := http.NewServeMux()
	mux.Handle("/user/", http.StripPrefix("/user", userMux))
	userMux.HandleFunc("POST /create", handlers.User.Create)
	userMux.HandleFunc("PUT /update", handlers.User.Update)
	userMux.HandleFunc("DELETE /delete/{id}", handlers.User.Delete)
	userMux.HandleFunc("POST /login", handlers.User.Login)
	userMux.HandleFunc("POST /logout", handlers.User.Logout)
	userMux.HandleFunc("GET /getalldetail", handlers.User.GetAllDetail)
	userMux.HandleFunc("GET /permission", handlers.User.Permission)

	deviceMux := http.NewServeMux()
	mux.Handle("/device/", http.StripPrefix("/device", deviceMux))
	deviceMux.HandleFunc("GET /getalldetail", handlers.Device.GetAllDetail)
	deviceMux.HandleFunc("POST /create", handlers.Device.Create)
	deviceMux.HandleFunc("PUT /update", handlers.Device.Update)
	deviceMux.HandleFunc("DELETE /delete/{id}", handlers.Device.Delete)

	scheduleMux := http.NewServeMux()
	mux.Handle("/schedule/", http.StripPrefix("/schedule", scheduleMux))
	scheduleMux.HandleFunc("GET /getalldetail", handlers.Schedule.GetAllDetail)
	scheduleMux.HandleFunc("POST /create", handlers.Schedule.Create)
	scheduleMux.HandleFunc("PUT /update", handlers.Schedule.Update)

	roleMux := http.NewServeMux()
	mux.Handle("/role/", http.StripPrefix("/role", roleMux))
	roleMux.HandleFunc("GET /getall", handlers.Role.GetAll)
	roleMux.HandleFunc("GET /getmenuavailable", handlers.Role.GetMenuAvailable)
	roleMux.HandleFunc("GET /getdetailbyid/{id}", handlers.Role.GetDetailById)
	roleMux.HandleFunc("POST /upsert", handlers.Role.Upsert)

	return mux
}

func SetupSchedule(handlers ApiHandlers) *http.ServeMux {
	mux := http.NewServeMux()

	scheduleEngine := http.NewServeMux()
	// 1. Change the base prefix to match the client
	mux.Handle("/scheduleengine/", http.StripPrefix("/scheduleengine", scheduleEngine))

	// 2. Adjust the remaining path to place {id} before /sync
	scheduleEngine.HandleFunc("POST /{id}/sync", handlers.ScheduleEngine.SyncSchedule)
	scheduleEngine.HandleFunc("DELETE /{id}/sync", handlers.ScheduleEngine.UnsyncSchedule)

	return mux
}
