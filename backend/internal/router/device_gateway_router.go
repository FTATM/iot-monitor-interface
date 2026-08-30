package router

import (
	"net/http"
	"os"

	"github.com/FTATM/iot-monitor-interface/internal/middleware"
)

func SetupDeviceGateway(handlers RouterHandlers) *http.ServeMux {
	mux := http.NewServeMux()
	gatewayMux := http.NewServeMux()
	internalSecret := os.Getenv("INTERNAL_API_SECRET")

	mux.Handle("/devicegateway/", http.StripPrefix("/devicegateway", gatewayMux))

	// PUBLIC ROUTES (Exposed to Physical IoT Devices)
	gatewayMux.HandleFunc("POST /telemetry", handlers.DeviceGateway.HTTPTelemetry)
	gatewayMux.HandleFunc("GET /command", handlers.DeviceGateway.HTTPCommandPolling)

	// Initialize the middleware with the secret
	secureMiddleware := middleware.InternalOnly(internalSecret)

	DeviceStatus := http.HandlerFunc(handlers.DeviceGateway.DeviceStatus)
	Command := http.HandlerFunc(handlers.DeviceGateway.Command)

	// 3. Wrap your handlers
	gatewayMux.Handle("GET /internal/devicestatus", secureMiddleware(DeviceStatus))
	gatewayMux.Handle("POST /internal/command", secureMiddleware(Command))

	return mux
}
