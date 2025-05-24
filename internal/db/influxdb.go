package db

import (
	"context"
	"fmt"
	"log/slog"
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
	client := influxdb2.NewClient(cfg.InfluxURL, cfg.InfluxToken)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Health(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	writeAPI := client.WriteAPI(cfg.InfluxOrg, cfg.InfluxBucket)
	queryAPI := client.QueryAPI(cfg.InfluxOrg)

	return &InfluxDBClient{
		client:   client,
		writeAPI: writeAPI,
		queryAPI: queryAPI,
		bucket:   cfg.InfluxBucket,
		org:      cfg.InfluxOrg,
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

	c.writeAPI.WritePoint(p)
	c.writeAPI.Flush()

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
	`, c.bucket)

	result, err := c.queryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer result.Close()

	if !result.Next() {
		return nil, nil // No activities found
	}

	record := result.Record()
	activity := &strava.ActivityData{}

	// Parse the record and populate activity
	// This is a simplified version - in real implementation,
	// you'd need to handle all fields properly
	if val := record.ValueByKey("activity_id"); val != nil {
		if id, ok := val.(string); ok {
			// Parse ID from string
			_ = id
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
	defer result.Close()

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
