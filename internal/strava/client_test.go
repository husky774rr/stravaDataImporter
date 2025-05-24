package strava

import (
	"testing"

	"stravaDataImporter/internal/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
		StravaRedirectURL:  "http://localhost:8080/auth/callback",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config != cfg {
		t.Error("Client config not set correctly")
	}
}

func TestGetAuthURL(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
		StravaRedirectURL:  "http://localhost:8080/auth/callback",
	}

	client := NewClient(cfg)
	authURL := client.GetAuthURL("test_state")

	if authURL == "" {
		t.Error("GetAuthURL() returned empty string")
	}

	// Check if URL contains expected parameters
	if !contains(authURL, "client_id=test_client_id") {
		t.Error("Auth URL does not contain client_id")
	}

	if !contains(authURL, "state=test_state") {
		t.Error("Auth URL does not contain state")
	}
}

func TestConvertToActivityData(t *testing.T) {
	stravaActivity := StravaActivity{
		ID:                   123456,
		Name:                 "Test Ride",
		Type:                 "Ride",
		Distance:             50000, // 50km in meters
		MovingTime:           3600,  // 1 hour
		ElapsedTime:          3660,
		TotalElevationGain:   500,
		StartDate:            "2024-01-01T10:00:00Z",
		AverageSpeed:         13.89, // ~50km/h
		MaxSpeed:             25.0,
		Calories:             1000,
		AverageHeartrate:     150,
		MaxHeartrate:         180,
		AverageWatts:         200,
		MaxWatts:             400,
		WeightedAverageWatts: 220,
		Kilojoules:           720,
	}

	ftp := 250.0

	activity, err := ConvertToActivityData(stravaActivity, ftp)
	if err != nil {
		t.Fatalf("ConvertToActivityData() error = %v", err)
	}

	if activity.ID != 123456 {
		t.Errorf("Activity ID = %v, want %v", activity.ID, 123456)
	}

	if activity.FTP != ftp {
		t.Errorf("Activity FTP = %v, want %v", activity.FTP, ftp)
	}

	if activity.NP != 220 {
		t.Errorf("Activity NP = %v, want %v", activity.NP, 220)
	}

	// TSS should be calculated as (movingTime * NP * IF) / (FTP * 3600) * 100
	// IF = NP / FTP = 220 / 250 = 0.88
	// TSS = (3600 * 220 * 0.88) / (250 * 3600) * 100 = 77.44
	expectedTSS := (3600.0 * 220.0 * (220.0 / 250.0)) / (250.0 * 3600.0) * 100.0
	if activity.TSS != expectedTSS {
		t.Errorf("Activity TSS = %v, want %v", activity.TSS, expectedTSS)
	}
}

func TestConvertToActivityDataInvalidDate(t *testing.T) {
	stravaActivity := StravaActivity{
		ID:        123456,
		Name:      "Test Ride",
		StartDate: "invalid-date",
	}

	_, err := ConvertToActivityData(stravaActivity, 250.0)
	if err == nil {
		t.Error("ConvertToActivityData() expected error for invalid date, got nil")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
