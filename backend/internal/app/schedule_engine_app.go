package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/FTATM/iot-monitor-interface/internal/handler"
	"github.com/FTATM/iot-monitor-interface/internal/middleware"
	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/FTATM/iot-monitor-interface/internal/repo"
	"github.com/FTATM/iot-monitor-interface/internal/router"
	"github.com/FTATM/iot-monitor-interface/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerScheduler struct {
	DB              *pgxpool.Pool
	Server          *http.Server
	ScheduleService model.ScheduleEngineService
}

func (a *ServerScheduler) Run() error {
	slog.Info(fmt.Sprintf("Server starting on %s", a.Server.Addr))
	return a.Server.ListenAndServe()
}

func (a *ServerScheduler) Close() {
	slog.Info("Executing graceful shutdown...")

	// Create a new context specifically for the HTTP shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down HTTP Server
	if a.Server != nil {
		if err := a.Server.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown error", slog.String("error", err.Error()))
		}
	}

	// Shut down the in-memory GoCron engine
	if a.ScheduleService != nil {
		a.ScheduleService.Stop() // Assumes you have a Stop() method on your service
	}

	// Shut down PostgreSQL connection pool
	if a.DB != nil {
		a.DB.Close()
	}

	slog.Info(fmt.Sprintf("Server successfully stopped on %s", a.Server.Addr))
}

func InitializeScheduleEngine(ctx context.Context) (App, error) {
	var dbConfig config.DB = &config.PostgresConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second) // Use the passed-in ctx as the parent
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

	slog.Info("Database connected!")

	// Dependency Injection
	scheduleRepo := repo.NewScheduleEngineRepository(db)

	scheduleService, err := service.NewSchedulerEngineService(scheduleRepo)
	if err != nil {
		slog.Error("Failed to initialize service", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Start the engine
	if err := scheduleService.Start(ctx); err != nil {
		slog.Error("Failed to start schedule engine", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handlers := router.RouterHandlers{
		ScheduleEngine: handler.NewScheduleEngineHandler(scheduleService),
	}

	mux := router.SetupApi(handlers)
	wrappedMux := middleware.LoggingApi(mux)

	serverPort := os.Getenv("SERVER_SCHEDULE_ENGINE_PORT")
	if serverPort == "" {
		serverPort = "8081"
		slog.Info(fmt.Sprintf("port fail default at %s", serverPort))
	}

	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: wrappedMux,
	}

	return &ServerScheduler{
		DB:              db,
		Server:          server,
		ScheduleService: scheduleService,
	}, nil
}
