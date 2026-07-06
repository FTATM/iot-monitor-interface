package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/FTATM/iot-monitor-interface/internal/handler"
	"github.com/FTATM/iot-monitor-interface/internal/repo"
	"github.com/FTATM/iot-monitor-interface/internal/router"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	DB     *pgxpool.Pool // Updated to use pgxpool
	Server *http.Server
}

func Initialize(ctx context.Context) (*App, error) {

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("CRITICAL ERROR: DATABASE_URL environment variable is not set!")
	}

	// Create a context for the connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the connection pool
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected via pgxpool!")

	// 2. Dependency Injection
	widgetRepo := repo.NewWidgetRepository(db)

	handlers := router.AllHandlers{
		Widget: handler.NewWidgetHandler(widgetRepo),
	}

	// 3. Initialize Router (using the struct bundle from earlier)
	mux := router.Setup(handlers)
	wrappedMux := router.LoggingMiddleware(mux)

	// 4. Configure HTTP Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: wrappedMux,
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
