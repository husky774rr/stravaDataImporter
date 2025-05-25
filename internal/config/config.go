package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port     string
	LogLevel string

	// Strava OAuth configuration
	StravaClientID     string
	StravaClientSecret string
	StravaRedirectURL  string

	// InfluxDB configuration
	InfluxDBURL    string
	InfluxDBToken  string
	InfluxDBOrg    string
	InfluxDBBucket string

	// Twitter configuration
	TwitterAPIKey            string
	TwitterAPISecret         string
	TwitterAccessToken       string
	TwitterAccessTokenSecret string

	// Scheduling intervals
	TokenRefreshInterval time.Duration
	DataImportInterval   time.Duration

	// Cron schedules
	TokenRefreshCron   string
	DataImportCron     string
	WeeklySummaryCron  string
	MonthlySummaryCron string
	YearlySummaryCron  string

	// FTP CSV file path
	FTPFilePath string
}

func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                     getEnv("PORT", "9090"),
		LogLevel:                 getEnv("LOG_LEVEL", "info"),
		StravaClientID:           getEnv("STRAVA_CLIENT_ID", ""),
		StravaClientSecret:       getEnv("STRAVA_CLIENT_SECRET", ""),
		StravaRedirectURL:        getEnv("STRAVA_REDIRECT_URL", "http://localhost:9090/auth/callback"),
		InfluxDBURL:              getEnv("INFLUXDB_URL", "http://localhost:8086"),
		InfluxDBToken:            getEnv("INFLUXDB_TOKEN", ""),
		InfluxDBOrg:              getEnv("INFLUXDB_ORG", "my-org"),
		InfluxDBBucket:           getEnv("INFLUXDB_BUCKET", "strava-data"),
		TwitterAPIKey:            getEnv("TWITTER_API_KEY", ""),
		TwitterAPISecret:         getEnv("TWITTER_API_SECRET", ""),
		TwitterAccessToken:       getEnv("TWITTER_ACCESS_TOKEN", ""),
		TwitterAccessTokenSecret: getEnv("TWITTER_ACCESS_TOKEN_SECRET", ""),
		FTPFilePath:              getEnv("FTP_FILE_PATH", "./conf/ftp.csv"),

		// Cron schedules with defaults
		TokenRefreshCron:   getEnv("TOKEN_REFRESH_CRON", "0 0 2 * * *"),   // 2 AM daily
		DataImportCron:     getEnv("DATA_IMPORT_CRON", "0 0 * * * *"),     // Every hour
		WeeklySummaryCron:  getEnv("WEEKLY_SUMMARY_CRON", "0 0 3 * * 1"),  // 3 AM every Monday
		MonthlySummaryCron: getEnv("MONTHLY_SUMMARY_CRON", "0 0 4 1 * *"), // 4 AM on the 1st of each month
		YearlySummaryCron:  getEnv("YEARLY_SUMMARY_CRON", "0 0 5 1 1 *"),  // 5 AM on January 1st
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

// ParseLogLevel converts a log level string to slog.Level
func (c *Config) ParseLogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // デフォルトはInfo
	}
}
