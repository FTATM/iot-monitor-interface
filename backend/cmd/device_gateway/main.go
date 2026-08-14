package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/DeRuina/timberjack"
	"github.com/FTATM/iot-monitor-interface/internal/app"
	"github.com/FTATM/iot-monitor-interface/internal/middleware"
)

func main() {
	// Initialize the logger first so we can catch startup errors
	initLogger()

	slog.Info("Starting device gateway initialization...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.InitializeDeviceGateway(ctx)
	if err != nil {
		slog.Error("Failed to initialize device gateway", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer application.Close()

	go func() {
		if err := application.Run(); err != nil {
			slog.Error("Device gateway crashed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("\nShutdown signal received...")
}

func initLogger() {
	logDir := os.Getenv("LOG_DEVICE_GATEWAY_DIR")
	if logDir == "" {
		// Fallback to a local "log" folder if the env var is missing
		logDir = "log/gateway"
	}

	var err error
	logMaxSize, err := strconv.Atoi(os.Getenv("LOG_DEVICE_GATEWAY_MAX_SIZE"))
	if err != nil {
		panic("Failed to pass env MaxSize: " + err.Error())
	}
	logMaxBackup, err := strconv.Atoi(os.Getenv("LOG_DEVICE_GATEWAY_MAX_BACKUP"))
	if err != nil {
		panic("Failed to pass env MaxBackups: " + err.Error())
	}
	logMaxAge, err := strconv.Atoi(os.Getenv("LOG_DEVICE_GATEWAY_MAX_AGE"))
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
		Filename:         filepath.Join(logDir, "device_gateway_app.log"),
		MaxSize:          logMaxSize,
		MaxBackups:       logMaxBackup,
		MaxAge:           logMaxAge,
		LocalTime:        true,
		Compression:      "zstd",
		RotationInterval: 24 * time.Hour,
		RotateAt:         []string{"00:00"},
	}

	logLevelStr := os.Getenv("LOG_DEVICE_GATEWAY_LEVEL")
	var logLevel slog.Level
	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo // Default fallback for production
	}

	// Configure slog options
	opts := &slog.HandlerOptions{
		Level:     logLevel,
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
