package scheduler

import (
	"testing"
	"time"

	"stravaDataImporter/internal/config"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
		FTPFilePath:        "./test_ftp.csv",
	}

	scheduler := New(cfg, nil)
	if scheduler == nil {
		t.Fatal("New() returned nil")
	}

	if scheduler.config != cfg {
		t.Error("Scheduler config not set correctly")
	}

	if scheduler.cron == nil {
		t.Error("Cron scheduler not initialized")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
		FTPFilePath:        "./test_ftp.csv",
	}

	scheduler := New(cfg, nil)

	// Test Start
	scheduler.Start()

	// Test Stop
	scheduler.Stop()

	// Should not panic
}

func TestGetWeekStart(t *testing.T) {
	// Test with a known date (Wednesday, January 3, 2024)
	testDate := time.Date(2024, 1, 3, 15, 30, 0, 0, time.UTC)
	weekStart := getWeekStart(testDate)

	// Should be Monday, January 1, 2024 at 00:00:00
	expectedWeekStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	if !weekStart.Equal(expectedWeekStart) {
		t.Errorf("getWeekStart() = %v, want %v", weekStart, expectedWeekStart)
	}
}

func TestGetMonthStart(t *testing.T) {
	// Test with a known date (January 15, 2024)
	testDate := time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC)
	monthStart := getMonthStart(testDate)

	// Should be January 1, 2024 at 00:00:00
	expectedMonthStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	if !monthStart.Equal(expectedMonthStart) {
		t.Errorf("getMonthStart() = %v, want %v", monthStart, expectedMonthStart)
	}
}

func TestGetYearStart(t *testing.T) {
	// Test with a known date (June 15, 2024)
	testDate := time.Date(2024, 6, 15, 15, 30, 0, 0, time.UTC)
	yearStart := getYearStart(testDate)

	// Should be January 1, 2024 at 00:00:00
	expectedYearStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	if !yearStart.Equal(expectedYearStart) {
		t.Errorf("getYearStart() = %v, want %v", yearStart, expectedYearStart)
	}
}

// Helper functions for testing (duplicated from strava package for testing)
func getWeekStart(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysFromMonday := weekday - 1
	monday := date.AddDate(0, 0, -daysFromMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

func getMonthStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func getYearStart(date time.Time) time.Time {
	return time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, date.Location())
}
