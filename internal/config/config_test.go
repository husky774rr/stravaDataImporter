package config

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("PORT", "9999")
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("STRAVA_CLIENT_ID", "test_client_id")
	_ = os.Setenv("TOKEN_REFRESH_HOURS", "48")
	_ = os.Setenv("DATA_IMPORT_HOURS", "2")

	defer func() {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("STRAVA_CLIENT_ID")
		_ = os.Unsetenv("TOKEN_REFRESH_HOURS")
		_ = os.Unsetenv("DATA_IMPORT_HOURS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "9999" {
		t.Errorf("Port = %v, want %v", cfg.Port, "9999")
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, "debug")
	}

	if cfg.StravaClientID != "test_client_id" {
		t.Errorf("StravaClientID = %v, want %v", cfg.StravaClientID, "test_client_id")
	}

	if cfg.TokenRefreshInterval != 48*time.Hour {
		t.Errorf("TokenRefreshInterval = %v, want %v", cfg.TokenRefreshInterval, 48*time.Hour)
	}

	if cfg.DataImportInterval != 2*time.Hour {
		t.Errorf("DataImportInterval = %v, want %v", cfg.DataImportInterval, 2*time.Hour)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Default Port = %v, want %v", cfg.Port, "9090")
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Default LogLevel = %v, want %v", cfg.LogLevel, "info")
	}

	if cfg.TokenRefreshInterval != 24*time.Hour {
		t.Errorf("Default TokenRefreshInterval = %v, want %v", cfg.TokenRefreshInterval, 24*time.Hour)
	}

	if cfg.DataImportInterval != 1*time.Hour {
		t.Errorf("Default DataImportInterval = %v, want %v", cfg.DataImportInterval, 1*time.Hour)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{"debug level", "debug", slog.LevelDebug},
		{"info level", "info", slog.LevelInfo},
		{"warn level", "warn", slog.LevelWarn},
		{"warning level", "warning", slog.LevelWarn},
		{"error level", "error", slog.LevelError},
		{"uppercase debug", "DEBUG", slog.LevelDebug},
		{"uppercase info", "INFO", slog.LevelInfo},
		{"invalid level defaults to info", "invalid", slog.LevelInfo},
		{"empty level defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LogLevel: tt.logLevel}
			result := cfg.ParseLogLevel()
			if result != tt.expected {
				t.Errorf("ParseLogLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}
