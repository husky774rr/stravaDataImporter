package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("PORT", "9999")
	_ = os.Setenv("STRAVA_CLIENT_ID", "test_client_id")
	_ = os.Setenv("TOKEN_REFRESH_HOURS", "48")
	_ = os.Setenv("DATA_IMPORT_HOURS", "2")

	defer func() {
		_ = os.Unsetenv("PORT")
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

	if cfg.Port != "8080" {
		t.Errorf("Default Port = %v, want %v", cfg.Port, "8080")
	}

	if cfg.TokenRefreshInterval != 24*time.Hour {
		t.Errorf("Default TokenRefreshInterval = %v, want %v", cfg.TokenRefreshInterval, 24*time.Hour)
	}

	if cfg.DataImportInterval != 1*time.Hour {
		t.Errorf("Default DataImportInterval = %v, want %v", cfg.DataImportInterval, 1*time.Hour)
	}
}
