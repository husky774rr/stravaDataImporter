package twitter

import (
	"testing"
	"time"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/strava"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		TwitterBearerToken: "test_bearer_token",
		TwitterConsumerKey: "test_consumer_key",
	}

	client := NewClient(cfg)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.config != cfg {
		t.Error("Client config not set correctly")
	}
}

func TestFormatActivityTweet(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	activity := &strava.ActivityData{
		ID:                 123456,
		Name:               "Test Ride",
		Type:               "Ride",
		Distance:           50000, // 50km in meters
		MovingTime:         3600,  // 1 hour
		TotalElevationGain: 500,
		StartDate:          time.Date(2024, 10, 27, 10, 0, 0, 0, time.UTC),
		Calories:           1100,
		TSS:                100,
		NP:                 200,
	}

	tweet := client.formatActivityTweet(activity)

	if tweet == "" {
		t.Error("formatActivityTweet() returned empty string")
	}

	// Check if tweet contains expected elements
	expectedElements := []string{
		"TSS: 100",
		"NP: 200",
		"サイクリング",
		"1,100kcal",
		"1時間0分",
		"50.0km",
		"500m",
	}

	for _, element := range expectedElements {
		if !contains(tweet, element) {
			t.Errorf("Tweet does not contain expected element: %s", element)
		}
	}
}

func TestTranslateActivityType(t *testing.T) {
	cfg := &config.Config{}
	client := NewClient(cfg)

	tests := []struct {
		input    string
		expected string
	}{
		{"Ride", "サイクリング"},
		{"Run", "ランニング"},
		{"Swim", "水泳"},
		{"Walk", "ウォーキング"},
		{"VirtualRide", "バーチャルライド"},
		{"UnknownType", "UnknownType"}, // Should return original for unknown types
	}

	for _, test := range tests {
		result := client.translateActivityType(test.input)
		if result != test.expected {
			t.Errorf("translateActivityType(%s) = %s, want %s", test.input, result, test.expected)
		}
	}
}

func TestPostActivity(t *testing.T) {
	cfg := &config.Config{
		TwitterBearerToken: "test_bearer_token",
	}
	client := NewClient(cfg)

	activity := &strava.ActivityData{
		ID:                 123456,
		Name:               "Test Ride",
		Type:               "Ride",
		Distance:           50000,
		MovingTime:         3600,
		TotalElevationGain: 500,
		StartDate:          time.Now(),
		Calories:           1100,
		TSS:                100,
		NP:                 200,
	}

	// This should not fail even though we're not actually posting to Twitter
	err := client.PostActivity(activity)
	if err != nil {
		t.Errorf("PostActivity() error = %v", err)
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
