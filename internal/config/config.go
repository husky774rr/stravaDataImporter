package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port string

	// Strava OAuth configuration
	StravaClientID     string
	StravaClientSecret string
	StravaRedirectURL  string

	// InfluxDB configuration
	InfluxURL    string
	InfluxToken  string
	InfluxOrg    string
	InfluxBucket string

	// Twitter configuration
	TwitterBearerToken       string
	TwitterConsumerKey       string
	TwitterConsumerSecret    string
	TwitterAccessToken       string
	TwitterAccessTokenSecret string

	// Scheduling intervals
	TokenRefreshInterval time.Duration
	DataImportInterval   time.Duration

	// FTP CSV file path
	FTPFilePath string
}

func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                     getEnv("PORT", "8080"),
		StravaClientID:           getEnv("STRAVA_CLIENT_ID", ""),
		StravaClientSecret:       getEnv("STRAVA_CLIENT_SECRET", ""),
		StravaRedirectURL:        getEnv("STRAVA_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		InfluxURL:                getEnv("INFLUX_URL", "http://localhost:8086"),
		InfluxToken:              getEnv("INFLUX_TOKEN", ""),
		InfluxOrg:                getEnv("INFLUX_ORG", "my-org"),
		InfluxBucket:             getEnv("INFLUX_BUCKET", "strava-data"),
		TwitterBearerToken:       getEnv("TWITTER_BEARER_TOKEN", ""),
		TwitterConsumerKey:       getEnv("TWITTER_CONSUMER_KEY", ""),
		TwitterConsumerSecret:    getEnv("TWITTER_CONSUMER_SECRET", ""),
		TwitterAccessToken:       getEnv("TWITTER_ACCESS_TOKEN", ""),
		TwitterAccessTokenSecret: getEnv("TWITTER_ACCESS_TOKEN_SECRET", ""),
		FTPFilePath:              getEnv("FTP_FILE_PATH", "./conf/ftp.csv"),
	}

	// Parse intervals
	tokenRefreshHours, err := strconv.Atoi(getEnv("TOKEN_REFRESH_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid TOKEN_REFRESH_HOURS: %w", err)
	}
	cfg.TokenRefreshInterval = time.Duration(tokenRefreshHours) * time.Hour

	dataImportHours, err := strconv.Atoi(getEnv("DATA_IMPORT_HOURS", "1"))
	if err != nil {
		return nil, fmt.Errorf("invalid DATA_IMPORT_HOURS: %w", err)
	}
	cfg.DataImportInterval = time.Duration(dataImportHours) * time.Hour

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
