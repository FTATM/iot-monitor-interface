package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/FTATM/iot-monitor-interface/internal/client"
	"github.com/FTATM/iot-monitor-interface/internal/handler"
	"github.com/FTATM/iot-monitor-interface/internal/middleware"
	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/FTATM/iot-monitor-interface/internal/repo"
	"github.com/FTATM/iot-monitor-interface/internal/router"
	"github.com/FTATM/iot-monitor-interface/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServerApi struct {
	DB                *pgxpool.Pool
	Server            *http.Server
	DeviceStartPublic func(ctx context.Context)
	cancelBroadcaster context.CancelFunc
}

func (a *ServerApi) Run() error {
	slog.Info(fmt.Sprintf("Server starting on %s", a.Server.Addr))
	ctx, cancel := context.WithCancel(context.Background())

	// 2. Store the cancel function in the struct so Close() can use it
	a.cancelBroadcaster = cancel
	go a.DeviceStartPublic(ctx)

	return a.Server.ListenAndServe()
}

func (a *ServerApi) Close() {
	slog.Info("Executing graceful shutdown...")

	if a.cancelBroadcaster != nil {
		a.cancelBroadcaster()
	}

	if a.DB != nil {
		a.DB.Close()
	}
	slog.Info(fmt.Sprintf("Server successfully stopped on %s", a.Server.Addr))
}

func InitializeApi(ctx context.Context) (App, error) {

	var dbConfig config.DB = &config.PostgresConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}

	// JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtKey := []byte(jwtSecret)

	//schedul engine
	scheduleEngineURL := os.Getenv("SCHEDULE_ENGINE_URL")
	if scheduleEngineURL == "" {
		slog.Error("SCHEDULE_ENGINE_URL is not set in environment")
		os.Exit(1)
	}

	//schedul engine
	deviceGatewayURL := os.Getenv("DEVICE_GATEWAY_URL")
	if deviceGatewayURL == "" {
		slog.Error("DEVICE_GATEWAY_URL is not set in environment")
		os.Exit(1)
	}

	// Create a context for the connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the connection pool
	db, err := pgxpool.New(ctx, dbConfig.ConnectString())
	if err != nil {
		slog.Error("Unable to initialize database pool", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := db.Ping(ctx); err != nil {
		slog.Error("Database is unreachable", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Database connected!")

	// Dependency Injection
	//? DB
	txManager := repo.NewTxManager(db)
	widgetRepo := repo.NewWidgetRepository(db)
	widgetTypeRepo := repo.NewWidgetTypeRepository(db)
	canvasRepo := repo.NewCanvasRepository(db)
	userRepo := repo.NewUserRepository(db)
	deviceRepo := repo.NewDeviceRepository(db)
	auditLogRepo := repo.NewAuditLogRepository(db)
	roleRepo := repo.NewRoleRepository(db)
	scheduleRepo := repo.NewScheduleRepository(db)

	//? client
	// scheduleClient := client.NewScheduleClient(scheduleEngineURL)
	deviceGatewayClient := client.NewDeviceGatewayClient(deviceGatewayURL)

	//? service
	widgetService := service.NewWidgetService(txManager, widgetRepo, widgetTypeRepo, canvasRepo)
	canvasService := service.NewCanvasService(txManager, widgetRepo, canvasRepo, auditLogRepo)
	widgetTypeService := service.NewWidgetTypeService(txManager, widgetTypeRepo)
	userService := service.NewUserService(txManager, userRepo, jwtKey, roleRepo, auditLogRepo)
	deviceService := service.NewDeviceService(txManager, deviceRepo, auditLogRepo)
	roleService := service.NewRoleService(txManager, roleRepo)
	scheduleService := service.NewScheduleService(txManager, scheduleRepo, auditLogRepo)

	handlers := router.RouterHandlers{
		Widget:     handler.NewWidgetHandler(widgetService, roleService),
		Canvas:     handler.NewCanvasHandler(canvasService, roleService),
		WidgetType: handler.NewWidgetTypeHandler(widgetTypeService),
		User:       handler.NewUserHandler(userService, roleService),
		Device:     handler.NewDeviceHandler(deviceService, roleService, deviceGatewayClient),
		Role:       handler.NewRoleHandler(roleService),
		Schedule:   handler.NewScheduleHandler(scheduleService, roleService),
	}

	// Initialize Router (using the struct bundle from earlier)
	mux := router.SetupApi(handlers)

	// Chain Middleware
	// AuthMiddleware is global check method for open public route
	protectedMux := middleware.AuthApi(jwtKey, mux)
	wrappedMux := middleware.LoggingApi(protectedMux)

	// Configure HTTP Server

	serverPort := os.Getenv("SERVER_API_PORT")
	if serverPort == "" {
		serverPort = "8080"
		slog.Info(fmt.Sprintf("port fail default at %s", serverPort))
	}
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: wrappedMux,
	}

	count, err := userRepo.UserCount(ctx)
	if err != nil {
		slog.Error("Error at count user", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if count == 0 {
		adminUser := os.Getenv("INITIAL_ADMIN_USERNAME")
		adminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
		createUserReq := model.CreateUser{
			FirstName: "system",
			LastName:  "admin",
			Username:  adminUser,
			Password:  adminPassword,
			Active:    true,
			RoleId:    1,
		}
		user, err := userService.CreateUser(ctx, &createUserReq, 0)
		if err != nil {
			slog.Error("Error at init admin user", slog.String("error", err.Error()))
			os.Exit(1)
		}

		slog.Info(fmt.Sprintf("init admin user: %s  %s", user.FirstName, user.LastName))
	}

	return &ServerApi{
		DB:                db,
		Server:            server,
		DeviceStartPublic: deviceService.StartPublic,
	}, nil
}
