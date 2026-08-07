package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/DeRuina/timberjack"

	"github.com/FTATM/iot-monitor-interface/internal/app"
	"github.com/FTATM/iot-monitor-interface/internal/middleware"
)

func main() {
	// Initialize the logger first so we can catch startup errors
	initLogger()

	slog.Info("Starting application initialization...")

	ctx := context.Background()
	application, err := app.InitializeApi(ctx)
	if err != nil {
		slog.Error("Failed to initialize app", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer application.Close()

	if err := application.Run(); err != nil {
		slog.Error("Server api crashed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func initLogger() {
	// Read the path from the environment variable (e.g., LOG_DIR=/var/logs/myapi)
	logDir := os.Getenv("LOG_API_DIR")
	if logDir == "" {
		// Fallback to a local "log" folder if the env var is missing
		logDir = "log"
	}
	var err error
	logMaxSize, err := strconv.Atoi(os.Getenv("LOG_API_MAX_SIZE"))
	if err != nil {
		panic("Failed to pass env MaxSize: " + err.Error())
	}
	logMaxBackup, err := strconv.Atoi(os.Getenv("LOG_API_MAX_BACKUP"))
	if err != nil {
		panic("Failed to pass env MaxBackups: " + err.Error())
	}
	logMaxAge, err := strconv.Atoi(os.Getenv("LOG_API_MAX_AGE"))
	if err != nil {
		panic("Failed to pass env MaxAge: " + err.Error())
	}

	// Safely create the directory if it doesn't exist yet
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the log folder, panic before the app starts
		panic("Failed to create log directory: " + err.Error())
	}

	// Set up the lumberjack file writer
	logRotator := &timberjack.Logger{
		Filename:         filepath.Join(logDir, "api_app.log"), // Uses the dynamic logDir
		MaxSize:          logMaxSize,                           // Max size in Megabytes before rotating
		MaxBackups:       logMaxBackup,                         // Max number of old files to keep
		MaxAge:           logMaxAge,                            // Max number of days to keep old files
		LocalTime:        true,                                 // false will use UTC
		Compression:      "zstd",                               // Explicitly set compression type
		RotationInterval: 24 * time.Hour,                       // Rotate every 24 hours
		RotateAt:         []string{"00:00"},                    // Force rotation exactly at midnight
	}

	// Configure slog options
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Clean up the "source" file path so it's not overly long
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				source.File = filepath.Base(source.File)
				source.Function = filepath.Base(source.Function)
			}
			return a
		},
	}

	// MULTI-WRITER FIX: Write to both os.Stdout (Console) AND logRotator (File)
	multiWriter := io.MultiWriter(os.Stdout, logRotator)

	// Create the standard JSON Handler and set it directly
	jsonHandler := slog.NewJSONHandler(multiWriter, opts)

	// FIX: Use the ContextHandler from your middleware package
	ctxHandler := middleware.ContextHandler{Handler: jsonHandler}

	slog.SetDefault(slog.New(ctxHandler))
}
