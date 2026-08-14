package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/FTATM/iot-monitor-interface/internal/handler"
	"github.com/FTATM/iot-monitor-interface/internal/listener"
	"github.com/FTATM/iot-monitor-interface/internal/middleware"
	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/FTATM/iot-monitor-interface/internal/repo"
	"github.com/FTATM/iot-monitor-interface/internal/router"
	"github.com/FTATM/iot-monitor-interface/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerDeviceGateway struct {
	DB             *pgxpool.Pool
	Server         *http.Server
	GatewayService model.DeviceGatewayService
	SessionService model.SessionManagerService
	ctx            context.Context
}

func (a *ServerDeviceGateway) Run() error {
	slog.Info("Device Gateway starting listeners...")

	gatewayPort := os.Getenv("DEVICE_GATEWAY_PORT")
	if gatewayPort == "" {
		gatewayPort = "8090"
	}

	// Start Protocol Listeners (Injecting SessionService so they can register connections)
	go listener.StartTCPServer(a.ctx, ":"+gatewayPort, a.GatewayService, a.SessionService)
	go listener.StartUDPServer(a.ctx, ":"+gatewayPort, a.GatewayService, a.SessionService)

	brokerURL := os.Getenv("MQTT_BROKER_URL")
	if brokerURL != "" {
		go listener.StartMQTTClient(a.ctx, brokerURL, a.GatewayService, a.SessionService)
	}

	// Start HTTP Server for Telemetry and Internal Commands
	slog.Info(fmt.Sprintf("Gateway HTTP Server starting on %s", a.Server.Addr))
	return a.Server.ListenAndServe()
}

func (a *ServerDeviceGateway) Close() {
	slog.Info("Executing graceful shutdown for Device Gateway...")

	// Create a new context specifically for the HTTP shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down HTTP Server
	if a.Server != nil {
		if err := a.Server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Gateway HTTP server shutdown error", slog.String("error", err.Error()))
		}
	}

	// Shut down the in-memory batcher engine (flushes remaining records to DB)
	if a.GatewayService != nil {
		a.GatewayService.Stop()
	}

	// Shut down PostgreSQL connection pool
	if a.DB != nil {
		a.DB.Close()
	}

	slog.Info(fmt.Sprintf("Device Gateway successfully stopped on %s", a.Server.Addr))
}

func InitializeDeviceGateway(ctx context.Context) (App, error) {
	// 1. Initialize Database Config EXACTLY like schedule engine
	var dbConfig config.DB = &config.PostgresConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Initialize the connection pool
	db, err := pgxpool.New(dbCtx, dbConfig.ConnectString())
	if err != nil {
		slog.Error("Unable to initialize database pool", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := db.Ping(dbCtx); err != nil {
		slog.Error("Database is unreachable", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Device Gateway Database connected!")

	gatewayRepo := repo.NewDeviceGatewayRepository(db)
	qBuffer, err := strconv.Atoi(os.Getenv("DEVICE_GATEWAY_QUEUE_BUFFER"))
	if err != nil {
		slog.Error("QUEUE BUFFER env is not correct", slog.String("error", err.Error()))
		os.Exit(1)
	}
	qTimeout, err := time.ParseDuration(os.Getenv("DEVICE_GATEWAY_QUEUE_TIMEOUT_SECOND") + "s")
	if err != nil {
		slog.Error("QUEUE TIMEOUT env is not correct", slog.String("error", err.Error()))
		os.Exit(1)
	}
	gatewayService := service.NewDeviceGatewayService(gatewayRepo, qBuffer, qTimeout)

	sessionService := service.NewSessionManagerService()

	if err := gatewayService.Start(ctx); err != nil {
		slog.Error("Failed to start device gateway service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handlers := router.RouterHandlers{
		DeviceGateway: handler.NewDeviceGatewayHandler(gatewayService, sessionService),
	}

	mux := router.SetupDeviceGateway(handlers)
	wrappedMux := middleware.LoggingApi(mux)

	// 5. HTTP Server Setup
	serverPort := os.Getenv("DEVICE_GATEWAY_HTTP_PORT")
	if serverPort == "" {
		serverPort = "8091"
		slog.Info(fmt.Sprintf("Gateway HTTP port failed, defaulting to %s", serverPort))
	}

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: wrappedMux,
	}

	return &ServerDeviceGateway{
		DB:             db,
		Server:         server,
		GatewayService: gatewayService,
		SessionService: sessionService,
		ctx:            ctx,
	}, nil
}
