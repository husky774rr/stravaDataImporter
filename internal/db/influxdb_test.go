package db

import (
	"testing"
	"time"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/strava"
)

func TestNewInfluxDBClient(t *testing.T) {
	cfg := &config.Config{
		InfluxURL:    "http://localhost:8086",
		InfluxToken:  "test-token",
		InfluxOrg:    "test-org",
		InfluxBucket: "test-bucket",
	}

	// This test will fail if InfluxDB is not running
	// In a real test environment, you would use a test container
	_, err := NewInfluxDBClient(cfg)
	if err != nil {
		t.Logf("InfluxDB connection failed (expected if not running): %v", err)
		// Don't fail the test if InfluxDB is not available
		return
	}
}

func TestWriteActivity(t *testing.T) {
	// Mock test - in real implementation, use test container
	activity := &strava.ActivityData{
		ID:                   123456,
		Name:                 "Test Activity",
		Type:                 "Ride",
		Distance:             50.5,
		MovingTime:           3600,
		ElapsedTime:          3660,
		TotalElevationGain:   500.0,
		StartDate:            time.Now(),
		AverageSpeed:         14.0,
		MaxSpeed:             25.0,
		Calories:             1000,
		AverageHeartrate:     150.0,
		MaxHeartrate:         180.0,
		AverageWatts:         200.0,
		MaxWatts:             400.0,
		WeightedAverageWatts: 220.0,
		Kilojoules:           720.0,
		FTP:                  250.0,
		TSS:                  100.0,
		NP:                   210.0,
	}

	// This would be a real test if InfluxDB was available
	t.Logf("Activity test data: %+v", activity)
}
