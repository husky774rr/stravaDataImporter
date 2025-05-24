package scheduler

import (
	"log/slog"
	"time"

	"stravaDataImporter/internal/auth"
	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/db"
	"stravaDataImporter/internal/ftp"
	"stravaDataImporter/internal/strava"
	"stravaDataImporter/internal/twitter"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	config        *config.Config
	cron          *cron.Cron
	stravaClient  *strava.Client
	tokenStore    *auth.TokenStore
	ftpManager    *ftp.FTPManager
	influxClient  *db.InfluxDBClient
	twitterClient *twitter.Client
}

func New(cfg *config.Config, influxClient *db.InfluxDBClient) *Scheduler {
	return &Scheduler{
		config:        cfg,
		cron:          cron.New(cron.WithSeconds()),
		stravaClient:  strava.NewClient(cfg),
		tokenStore:    auth.NewTokenStore("./data/token.json"),
		ftpManager:    ftp.NewFTPManager(cfg.FTPFilePath),
		influxClient:  influxClient,
		twitterClient: twitter.NewClient(cfg),
	}
}

func (s *Scheduler) Start() {
	slog.Info("Starting scheduler")

	// Load FTP data on startup
	if err := s.ftpManager.LoadFTPData(); err != nil {
		slog.Warn("Failed to load FTP data", "error", err)
	}

	// Schedule token refresh using config
	_, err := s.cron.AddFunc(s.config.TokenRefreshCron, s.refreshTokenJob)
	if err != nil {
		slog.Error("Failed to schedule token refresh job", "error", err, "cron", s.config.TokenRefreshCron)
	} else {
		slog.Info("Scheduled token refresh job", "cron", s.config.TokenRefreshCron)
	}

	// Schedule data import using config
	_, err = s.cron.AddFunc(s.config.DataImportCron, s.importDataJob)
	if err != nil {
		slog.Error("Failed to schedule data import job", "error", err, "cron", s.config.DataImportCron)
	} else {
		slog.Info("Scheduled data import job", "cron", s.config.DataImportCron)
	}

	// Schedule weekly summary calculation using config
	_, err = s.cron.AddFunc(s.config.WeeklySummaryCron, s.calculateWeeklySummaryJob)
	if err != nil {
		slog.Error("Failed to schedule weekly summary job", "error", err, "cron", s.config.WeeklySummaryCron)
	} else {
		slog.Info("Scheduled weekly summary job", "cron", s.config.WeeklySummaryCron)
	}

	// Schedule monthly summary calculation using config
	_, err = s.cron.AddFunc(s.config.MonthlySummaryCron, s.calculateMonthlySummaryJob)
	if err != nil {
		slog.Error("Failed to schedule monthly summary job", "error", err, "cron", s.config.MonthlySummaryCron)
	} else {
		slog.Info("Scheduled monthly summary job", "cron", s.config.MonthlySummaryCron)
	}

	// Schedule yearly summary calculation using config
	_, err = s.cron.AddFunc(s.config.YearlySummaryCron, s.calculateYearlySummaryJob)
	if err != nil {
		slog.Error("Failed to schedule yearly summary job", "error", err, "cron", s.config.YearlySummaryCron)
	} else {
		slog.Info("Scheduled yearly summary job", "cron", s.config.YearlySummaryCron)
	}

	s.cron.Start()
	slog.Info("Scheduler started successfully")
}

func (s *Scheduler) Stop() {
	slog.Info("Stopping scheduler")
	s.cron.Stop()
}

func (s *Scheduler) refreshTokenJob() {
	slog.Info("Starting token refresh job")

	token, err := s.tokenStore.LoadToken()
	if err != nil || token == nil {
		slog.Warn("No token found for refresh")
		return
	}

	newToken, err := s.stravaClient.RefreshToken(token.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh token", "error", err)
		return
	}

	if err := s.tokenStore.SaveToken(newToken); err != nil {
		slog.Error("Failed to save refreshed token", "error", err)
		return
	}

	slog.Info("Token refreshed successfully")
}

func (s *Scheduler) importDataJob() {
	slog.Info("Starting data import job")

	token, err := s.tokenStore.LoadToken()
	if err != nil || token == nil {
		slog.Warn("No token found for data import")
		return
	}

	// Get activities from the last 2 days to ensure we don't miss any
	since := time.Now().AddDate(0, 0, -2)
	activities, err := s.stravaClient.GetActivities(token.AccessToken, since, 200)
	if err != nil {
		slog.Error("Failed to fetch activities", "error", err)
		return
	}

	slog.Info("Fetched activities", "count", len(activities))

	for _, activity := range activities {
		ftp := s.ftpManager.GetFTPForDate(time.Now()) // Use current FTP for simplicity

		activityData, err := strava.ConvertToActivityData(activity, ftp)
		if err != nil {
			slog.Error("Failed to convert activity data", "activity_id", activity.ID, "error", err)
			continue
		}

		if err := s.influxClient.WriteActivity(activityData); err != nil {
			slog.Error("Failed to write activity to InfluxDB", "activity_id", activity.ID, "error", err)
		}

		// Post to Twitter for new activities (within last hour)
		if activityData.StartDate.After(time.Now().Add(-1 * time.Hour)) {
			go s.postToTwitter(activityData)
		}
	}

	slog.Info("Data import job completed")
}

func (s *Scheduler) calculateWeeklySummaryJob() {
	slog.Info("Starting weekly summary calculation job")

	// Calculate for the current week and previous week
	now := time.Now()
	currentWeek := strava.GetWeekStart(now)
	previousWeek := currentWeek.AddDate(0, 0, -7)

	s.calculateWeeklySummary(currentWeek)
	s.calculateWeeklySummary(previousWeek)

	slog.Info("Weekly summary calculation job completed")
}

func (s *Scheduler) calculateMonthlySummaryJob() {
	slog.Info("Starting monthly summary calculation job")

	// Calculate for the current month and previous month
	now := time.Now()
	currentMonth := strava.GetMonthStart(now)
	previousMonth := currentMonth.AddDate(0, -1, 0)

	s.calculateMonthlySummary(currentMonth)
	s.calculateMonthlySummary(previousMonth)

	slog.Info("Monthly summary calculation job completed")
}

func (s *Scheduler) calculateYearlySummaryJob() {
	slog.Info("Starting yearly summary calculation job")

	// Calculate for the current year and previous year
	now := time.Now()
	currentYear := strava.GetYearStart(now)
	previousYear := currentYear.AddDate(-1, 0, 0)

	s.calculateYearlySummary(currentYear)
	s.calculateYearlySummary(previousYear)

	slog.Info("Yearly summary calculation job completed")
}

func (s *Scheduler) calculateWeeklySummary(weekStart time.Time) {
	// This is a simplified implementation
	// In a real scenario, you would query InfluxDB for activities in this week
	slog.Info("Calculating weekly summary", "week_start", weekStart)

	summary := strava.WeeklySummary{
		WeekStart:          weekStart,
		TotalTSS:           0,
		TotalMovingTime:    0,
		TotalDistance:      0,
		TotalElevationGain: 0,
	}

	if err := s.influxClient.WriteWeeklySummary(&summary); err != nil {
		slog.Error("Failed to write weekly summary", "error", err)
	}
}

func (s *Scheduler) calculateMonthlySummary(monthStart time.Time) {
	slog.Info("Calculating monthly summary", "month_start", monthStart)

	summary := strava.MonthlySummary{
		MonthStart:      monthStart,
		TotalTSS:        0,
		TotalMovingTime: 0,
		TotalDistance:   0,
	}

	if err := s.influxClient.WriteMonthlySummary(&summary); err != nil {
		slog.Error("Failed to write monthly summary", "error", err)
	}
}

func (s *Scheduler) calculateYearlySummary(yearStart time.Time) {
	slog.Info("Calculating yearly summary", "year_start", yearStart)

	summary := strava.YearlySummary{
		YearStart:       yearStart,
		TotalTSS:        0,
		TotalMovingTime: 0,
		TotalDistance:   0,
	}

	if err := s.influxClient.WriteYearlySummary(&summary); err != nil {
		slog.Error("Failed to write yearly summary", "error", err)
	}
}

func (s *Scheduler) postToTwitter(activity *strava.ActivityData) {
	slog.Info("Posting activity to Twitter", "activity_id", activity.ID)

	if err := s.twitterClient.PostActivity(activity); err != nil {
		slog.Error("Failed to post to Twitter", "activity_id", activity.ID, "error", err)
	}
}
