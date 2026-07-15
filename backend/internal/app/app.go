package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/FTATM/iot-monitor-interface/config"
	"github.com/FTATM/iot-monitor-interface/internal/handler"
	"github.com/FTATM/iot-monitor-interface/internal/model"
	"github.com/FTATM/iot-monitor-interface/internal/repo"
	"github.com/FTATM/iot-monitor-interface/internal/router"
	"github.com/FTATM/iot-monitor-interface/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB     *pgxpool.Pool
	Server *http.Server
}

func Initialize(ctx context.Context) (*App, error) {

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

	// Create a context for the connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the connection pool
	db, err := pgxpool.New(ctx, dbConfig.ConnectString())
	if err != nil {
		log.Fatalf("Unable to initialize database pool: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
	}

	log.Println("Database connected via pgxpool!")

	// Dependency Injection
	//? DB
	txManager := repo.NewTxManager(db)
	widgetRepo := repo.NewWidgetRepository(db)
	widgetTypeRepo := repo.NewWidgetTypeRepository(db)
	canvasRepo := repo.NewCanvasRepository(db)
	userCanvasGroupRepo := repo.NewUserCanvasGroupRepository(db)
	userRepo := repo.NewUserRepository(db)

	//? service
	widgetService := service.NewWidgetService(txManager, widgetRepo, widgetTypeRepo, canvasRepo)
	canvasService := service.NewCanvasService(txManager, widgetRepo, canvasRepo, userCanvasGroupRepo)
	widgetTypeService := service.NewWidgetTypeService(txManager, widgetTypeRepo)
	userService := service.NewUserService(txManager, userRepo, jwtKey)

	handlers := router.AllHandlers{
		Widget:     handler.NewWidgetHandler(widgetService),
		Canvas:     handler.NewCanvasHandler(canvasService),
		WidgetType: handler.NewWidgetTypeHandler(widgetTypeService),
		User:       handler.NewUserHandler(userService),
	}

	// Initialize Router (using the struct bundle from earlier)
	mux := router.Setup(handlers)

	// Chain Middleware
	// AuthMiddleware is global check method for open public route
	protectedMux := router.AuthMiddleware(jwtKey, mux)
	wrappedMux := router.LoggingMiddleware(protectedMux)

	// Configure HTTP Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: wrappedMux,
	}

	count, err := userRepo.UserCount(ctx)
	if err != nil {
		log.Fatalf("Error at count user %v", err)
	}

	if count == 0 {
		adminUser := os.Getenv("INITIAL_ADMIN_USERNAME")
		adminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
		createUserReq := model.CreateUserRequest{
			FirstName: "system",
			LastName:  "admin",
			Username:  adminUser,
			Password:  adminPassword,
			Active:    true,
		}
		log.Println(createUserReq)

		user, err := userService.Create(ctx, &createUserReq)
		if err != nil {
			log.Fatalf("Error at init admin user %v", err)
		}

		log.Println(user)
	}

	return &App{
		DB:     db,
		Server: server,
	}, nil
}

func (a *App) Run() error {
	log.Printf("Server starting on %s", a.Server.Addr)
	return a.Server.ListenAndServe()
}

func (a *App) Close() {
	if a.DB != nil {
		// Close the pgxpool to gracefully shut down DB connections
		a.DB.Close()
	}
}
