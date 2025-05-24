package strava

import (
	"time"
)

// ActivityData represents a Strava activity with calculated metrics
type ActivityData struct {
	ID                   int64     `json:"id"`
	Name                 string    `json:"name"`
	Type                 string    `json:"type"`
	Distance             float64   `json:"distance"`
	MovingTime           int       `json:"moving_time"`
	ElapsedTime          int       `json:"elapsed_time"`
	TotalElevationGain   float64   `json:"total_elevation_gain"`
	StartDate            time.Time `json:"start_date"`
	AverageSpeed         float64   `json:"average_speed"`
	MaxSpeed             float64   `json:"max_speed"`
	Calories             float64   `json:"calories"`
	AverageHeartrate     float64   `json:"average_heartrate"`
	MaxHeartrate         float64   `json:"max_heartrate"`
	AverageWatts         float64   `json:"average_watts"`
	MaxWatts             float64   `json:"max_watts"`
	WeightedAverageWatts float64   `json:"weighted_average_watts"`
	Kilojoules           float64   `json:"kilojoules"`

	// Calculated fields
	FTP float64 `json:"ftp"`
	TSS float64 `json:"tss"`
	NP  float64 `json:"np"`
}

// WeeklySummary represents weekly aggregated data
type WeeklySummary struct {
	WeekStart          time.Time `json:"week_start"`
	TotalTSS           float64   `json:"total_tss"`
	TotalMovingTime    int       `json:"total_moving_time"`
	TotalDistance      float64   `json:"total_distance"`
	TotalElevationGain float64   `json:"total_elevation_gain"`
}

// MonthlySummary represents monthly aggregated data
type MonthlySummary struct {
	MonthStart      time.Time `json:"month_start"`
	TotalTSS        float64   `json:"total_tss"`
	TotalMovingTime int       `json:"total_moving_time"`
	TotalDistance   float64   `json:"total_distance"`
}

// YearlySummary represents yearly aggregated data
type YearlySummary struct {
	YearStart       time.Time `json:"year_start"`
	TotalTSS        float64   `json:"total_tss"`
	TotalMovingTime int       `json:"total_moving_time"`
	TotalDistance   float64   `json:"total_distance"`
}

// TokenData represents OAuth token information
type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// StravaActivity represents the raw activity data from Strava API
type StravaActivity struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	Type                 string  `json:"type"`
	Distance             float64 `json:"distance"`
	MovingTime           int     `json:"moving_time"`
	ElapsedTime          int     `json:"elapsed_time"`
	TotalElevationGain   float64 `json:"total_elevation_gain"`
	StartDate            string  `json:"start_date"`
	AverageSpeed         float64 `json:"average_speed"`
	MaxSpeed             float64 `json:"max_speed"`
	Calories             float64 `json:"calories"`
	AverageHeartrate     float64 `json:"average_heartrate"`
	MaxHeartrate         float64 `json:"max_heartrate"`
	AverageWatts         float64 `json:"average_watts"`
	MaxWatts             float64 `json:"max_watts"`
	WeightedAverageWatts float64 `json:"weighted_average_watts"`
	Kilojoules           float64 `json:"kilojoules"`
}

// AthleteInfo represents basic athlete information from Strava
type AthleteInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// TokenResponse represents OAuth token response from Strava
type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    int64       `json:"expires_at"`
	TokenType    string      `json:"token_type"`
	Athlete      AthleteInfo `json:"athlete"`
}
