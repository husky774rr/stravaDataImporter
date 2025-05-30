package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/strava"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxDBClient struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	queryAPI api.QueryAPI
	bucket   string
	org      string
}

func NewInfluxDBClient(cfg *config.Config) (*InfluxDBClient, error) {
	slog.Info("Creating InfluxDB client", "url", cfg.InfluxDBURL, "org", cfg.InfluxDBOrg, "bucket", cfg.InfluxDBBucket, "token_length", len(cfg.InfluxDBToken))

	client := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	slog.Info("InfluxDB health check passed", "status", health.Status, "message", health.Message)

	writeAPI := client.WriteAPI(cfg.InfluxDBOrg, cfg.InfluxDBBucket)
	queryAPI := client.QueryAPI(cfg.InfluxDBOrg)

	return &InfluxDBClient{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		bucket:   cfg.InfluxDBBucket,
		org:      cfg.InfluxDBOrg,
	}, nil
}

func (c *InfluxDBClient) Close() {
	c.client.Close()
}

func (c *InfluxDBClient) WriteActivity(activity *strava.ActivityData) error {
	p := influxdb2.NewPointWithMeasurement("activities").
		AddTag("activity_id", fmt.Sprintf("%d", activity.ID)).
		AddTag("activity_type", activity.Type).
		AddTag("activity_name", activity.Name).
		AddField("distance", activity.Distance).
		AddField("moving_time", activity.MovingTime).
		AddField("elapsed_time", activity.ElapsedTime).
		AddField("total_elevation_gain", activity.TotalElevationGain).
		AddField("average_speed", activity.AverageSpeed).
		AddField("max_speed", activity.MaxSpeed).
		AddField("calories", activity.Calories).
		AddField("average_heartrate", activity.AverageHeartrate).
		AddField("max_heartrate", activity.MaxHeartrate).
		AddField("average_watts", activity.AverageWatts).
		AddField("max_watts", activity.MaxWatts).
		AddField("weighted_average_watts", activity.WeightedAverageWatts).
		AddField("kilojoules", activity.Kilojoules).
		AddField("ftp", activity.FTP).
		AddField("tss", activity.TSS).
		AddField("np", activity.NP).
		SetTime(activity.StartDate)

	// Check for write errors
	errorsCh := c.writeAPI.Errors()
	go func() {
		for err := range errorsCh {
			slog.Error("InfluxDB write error", "error", err)
		}
	}()

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

	activityJsonStr, err := json.Marshal(activity)
	if err == nil {
		slog.Debug("Activity JSON marshaled", "activity_json", string(activityJsonStr))
	}

	slog.Info("Activity written to InfluxDB", "activity_id", activity.ID, "name", activity.Name)
	return nil
}

func (c *InfluxDBClient) WriteWeeklySummary(summary *strava.WeeklySummary) error {
	p := influxdb2.NewPointWithMeasurement("weekly_summary").
		AddTag("week_start", summary.WeekStart.Format("2006-01-02")).
		AddField("total_tss", summary.TotalTSS).
		AddField("total_moving_time", summary.TotalMovingTime).
		AddField("total_distance", summary.TotalDistance).
		AddField("total_elevation_gain", summary.TotalElevationGain).
		SetTime(summary.WeekStart)

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

	slog.Info("Weekly summary written to InfluxDB", "week_start", summary.WeekStart)
	return nil
}

func (c *InfluxDBClient) WriteMonthlySummary(summary *strava.MonthlySummary) error {
	p := influxdb2.NewPointWithMeasurement("monthly_summary").
		AddTag("month_start", summary.MonthStart.Format("2006-01-02")).
		AddField("total_tss", summary.TotalTSS).
		AddField("total_moving_time", summary.TotalMovingTime).
		AddField("total_distance", summary.TotalDistance).
		SetTime(summary.MonthStart)

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

	slog.Info("Monthly summary written to InfluxDB", "month_start", summary.MonthStart)
	return nil
}

func (c *InfluxDBClient) WriteYearlySummary(summary *strava.YearlySummary) error {
	p := influxdb2.NewPointWithMeasurement("yearly_summary").
		AddTag("year_start", summary.YearStart.Format("2006-01-02")).
		AddField("total_tss", summary.TotalTSS).
		AddField("total_moving_time", summary.TotalMovingTime).
		AddField("total_distance", summary.TotalDistance).
		SetTime(summary.YearStart)

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

	slog.Info("Yearly summary written to InfluxDB", "year_start", summary.YearStart)
	return nil
}

func (c *InfluxDBClient) GetLatestActivity() (*strava.ActivityData, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -30d)
		|> filter(fn: (r) => r._measurement == "activities")
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: 1)
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
	`, c.bucket)

	result, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = result.Close() }()

	if !result.Next() {
		return nil, nil // No activities found
	}

	record := result.Record()
	activity := &strava.ActivityData{}

	// Parse activity ID from tag
	if val := record.ValueByKey("activity_id"); val != nil {
		if idStr, ok := val.(string); ok {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				activity.ID = id
			}
		}
	}

	// Parse activity name from tag
	if val := record.ValueByKey("activity_name"); val != nil {
		if name, ok := val.(string); ok {
			activity.Name = name
		}
	}

	// Parse activity type from tag
	if val := record.ValueByKey("activity_type"); val != nil {
		if actType, ok := val.(string); ok {
			activity.Type = actType
		}
	}

	// Parse numeric fields
	activity.StartDate = record.Time()

	if val := record.ValueByKey("distance"); val != nil {
		if distance, ok := val.(float64); ok {
			activity.Distance = distance
		}
	}

	if val := record.ValueByKey("moving_time"); val != nil {
		if movingTime, ok := val.(float64); ok {
			activity.MovingTime = int(movingTime)
		}
	}

	if val := record.ValueByKey("elapsed_time"); val != nil {
		if elapsedTime, ok := val.(float64); ok {
			activity.ElapsedTime = int(elapsedTime)
		}
	}

	if val := record.ValueByKey("total_elevation_gain"); val != nil {
		if elevation, ok := val.(float64); ok {
			activity.TotalElevationGain = elevation
		}
	}

	if val := record.ValueByKey("calories"); val != nil {
		if calories, ok := val.(float64); ok {
			activity.Calories = calories
		}
	}

	if val := record.ValueByKey("average_watts"); val != nil {
		if watts, ok := val.(float64); ok {
			activity.AverageWatts = watts
		}
	}

	if val := record.ValueByKey("max_watts"); val != nil {
		if watts, ok := val.(float64); ok {
			activity.MaxWatts = watts
		}
	}

	if val := record.ValueByKey("tss"); val != nil {
		if tss, ok := val.(float64); ok {
			activity.TSS = tss
		}
	}

	if val := record.ValueByKey("np"); val != nil {
		if np, ok := val.(float64); ok {
			activity.NP = np
		}
	}

	if val := record.ValueByKey("ftp"); val != nil {
		if ftp, ok := val.(float64); ok {
			activity.FTP = ftp
		}
	}

	return activity, nil
}

func (c *InfluxDBClient) GetWeeklyTrend() ([]strava.WeeklySummary, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -8w)
		|> filter(fn: (r) => r._measurement == "weekly_summary")
		|> sort(columns: ["_time"], desc: false)
	`, c.bucket)

	result, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer func() { _ = result.Close() }()

	var summaries []strava.WeeklySummary
	for result.Next() {
		record := result.Record()
		summary := strava.WeeklySummary{
			WeekStart: record.Time(),
		}

		if val := record.ValueByKey("total_tss"); val != nil {
			if tss, ok := val.(float64); ok {
				summary.TotalTSS = tss
			}
		}

		if val := record.ValueByKey("total_moving_time"); val != nil {
			if movingTime, ok := val.(int); ok {
				summary.TotalMovingTime = movingTime
			}
		}

		if val := record.ValueByKey("total_distance"); val != nil {
			if distance, ok := val.(float64); ok {
				summary.TotalDistance = distance
			}
		}

		if val := record.ValueByKey("total_elevation_gain"); val != nil {
			if elevation, ok := val.(float64); ok {
				summary.TotalElevationGain = elevation
			}
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Token management methods
func (c *InfluxDBClient) SaveToken(token *strava.TokenData) error {
	p := influxdb2.NewPointWithMeasurement("tokens").
		AddTag("token_type", "strava_access").
		AddField("access_token", token.AccessToken).
		AddField("refresh_token", token.RefreshToken).
		AddField("expires_at", token.ExpiresAt.Unix()).
		AddField("athlete_id", token.AthleteID).
		SetTime(time.Now())

	// Check for write errors
	errorsCh := c.writeAPI.Errors()
	go func() {
		for err := range errorsCh {
			slog.Error("InfluxDB token write error", "error", err)
		}
	}()

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

	slog.Info("Token saved to InfluxDB", "athlete_id", token.AthleteID)
	return nil
}

func (c *InfluxDBClient) LoadToken() (*strava.TokenData, error) {
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -7d)
		|> filter(fn: (r) => r._measurement == "tokens")
		|> filter(fn: (r) => r.token_type == "strava_access")
		|> sort(columns: ["_time"], desc: true)
		|> limit(n: 1)
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
	`, c.bucket)

	slog.Debug("Executing token query", "bucket", c.bucket)
	result, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		slog.Error("Token query failed", "error", err)
		return nil, fmt.Errorf("token query failed: %w", err)
	}
	defer func() { _ = result.Close() }()

	if !result.Next() {
		slog.Debug("No token found in InfluxDB")
		return nil, nil // No token found
	}

	record := result.Record()
	token := &strava.TokenData{}

	slog.Debug("Found token record", "time", record.Time(), "measurement", record.Measurement())

	// Parse access token
	if val := record.ValueByKey("access_token"); val != nil {
		if accessToken, ok := val.(string); ok {
			token.AccessToken = accessToken
			slog.Debug("Parsed access token", "length", len(accessToken))
		}
	}

	// Parse refresh token
	if val := record.ValueByKey("refresh_token"); val != nil {
		if refreshToken, ok := val.(string); ok {
			token.RefreshToken = refreshToken
			slog.Debug("Parsed refresh token", "length", len(refreshToken))
		}
	}

	// Parse expires_at
	if val := record.ValueByKey("expires_at"); val != nil {
		if expiresAt, ok := val.(float64); ok {
			token.ExpiresAt = time.Unix(int64(expiresAt), 0)
			slog.Debug("Parsed expires_at", "expires_at", token.ExpiresAt)
		}
	}

	// Parse athlete_id
	if val := record.ValueByKey("athlete_id"); val != nil {
		if athleteID, ok := val.(float64); ok {
			token.AthleteID = int64(athleteID)
			slog.Debug("Parsed athlete_id", "athlete_id", token.AthleteID)
		}
	}

	slog.Info("Token loaded from InfluxDB", "athlete_id", token.AthleteID)
	return token, nil
}

func (c *InfluxDBClient) ClearToken() error {
	// InfluxDBでは過去のデータを直接削除するのは複雑なので、
	// 代わりに無効化フラグを立てるアプローチを取ります
	p := influxdb2.NewPointWithMeasurement("tokens").
		AddTag("token_type", "strava_access").
		AddField("access_token", "").
		AddField("refresh_token", "").
		AddField("expires_at", time.Now().Add(-24*time.Hour).Unix()). // 過去の時間に設定
		AddField("athlete_id", int64(0)).
		AddField("invalidated", true).
		SetTime(time.Now())

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

	slog.Info("Token invalidated in InfluxDB")
	return nil
}
