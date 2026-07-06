package router

import (
	"log"
	"net/http"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/handler"
)

// AllHandlers bundles everything together so Setup() stays clean
type AllHandlers struct {
	Widget *handler.WidgetHandler
	// User   *handler.UserHandler
	// Add more as the app grows...
}

// Setup now takes the bundle and applies middleware
func Setup(handlers AllHandlers) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /widget", handlers.Widget.GetWidgetById)
	// mux.HandleFunc("POST /widget", handlers.Widget.CreateWidget)

	return mux
}

// LoggingMiddleware is a standard way to wrap handlers in Go
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
