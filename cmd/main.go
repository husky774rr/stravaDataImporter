package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/db"
	"stravaDataImporter/internal/scheduler"
	"stravaDataImporter/internal/web"
)

var log *slog.Logger

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration first
	cfg, err := config.Load()
	if err != nil {
		// Use temporary logger for config load error
		tempLog := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		tempLog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logger with configured log level
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.ParseLogLevel(),
	}))
	slog.SetDefault(log)

	log.Info("Starting stravaDataImporter", "logLevel", cfg.LogLevel)

	// Initialize InfluxDB client
	influxClient, err := db.NewInfluxDBClient(cfg)
	if err != nil {
		log.Error("Failed to initialize InfluxDB client", "error", err)
		os.Exit(1)
	}
	defer influxClient.Close()

	// Initialize scheduler
	scheduler := scheduler.New(cfg, influxClient)
	scheduler.Start()
	defer scheduler.Stop()

	// Initialize web server
	server, err := web.NewServer(cfg, influxClient)
	if err != nil {
		log.Error("Failed to create web server", "error", err)
		return
	}

	// Start web server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Error("Web server error", "error", err)
			cancel()
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("Received interrupt signal, shutting down...")
	case <-ctx.Done():
		log.Info("Context cancelled, shutting down...")
	}

	// Graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Error during server shutdown", "error", err)
	}

	log.Info("Application stopped")
}
