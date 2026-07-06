package main

import (
	"context"
	"log"

	"github.com/FTATM/iot-monitor-interface/internal/app"
)

func main() {
	// Use a background context for the initialization phase
	ctx := context.Background()

	application, err := app.Initialize(ctx)
	if err != nil {
		log.Fatal("Failed to initialize app: ", err)
	}

	defer application.Close()

	if err := application.Run(); err != nil {
		log.Fatal("Server crashed: ", err)
	}
}
