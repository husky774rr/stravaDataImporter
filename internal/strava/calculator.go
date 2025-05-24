package strava

import (
	"fmt"
	"time"
)

// CalculateWeeklySummary calculates weekly summary from activities
func CalculateWeeklySummary(activities []ActivityData, weekStart time.Time) WeeklySummary {
	weekEnd := weekStart.AddDate(0, 0, 7)

	summary := WeeklySummary{
		WeekStart: weekStart,
	}

	for _, activity := range activities {
		if activity.StartDate.After(weekStart) && activity.StartDate.Before(weekEnd) {
			summary.TotalTSS += activity.TSS
			summary.TotalMovingTime += activity.MovingTime
			summary.TotalDistance += activity.Distance
			summary.TotalElevationGain += activity.TotalElevationGain
		}
	}

	return summary
}

// CalculateMonthlySummary calculates monthly summary from activities
func CalculateMonthlySummary(activities []ActivityData, monthStart time.Time) MonthlySummary {
	monthEnd := monthStart.AddDate(0, 1, 0)

	summary := MonthlySummary{
		MonthStart: monthStart,
	}

	for _, activity := range activities {
		if activity.StartDate.After(monthStart) && activity.StartDate.Before(monthEnd) {
			summary.TotalTSS += activity.TSS
			summary.TotalMovingTime += activity.MovingTime
			summary.TotalDistance += activity.Distance
		}
	}

	return summary
}

// CalculateYearlySummary calculates yearly summary from activities
func CalculateYearlySummary(activities []ActivityData, yearStart time.Time) YearlySummary {
	yearEnd := yearStart.AddDate(1, 0, 0)

	summary := YearlySummary{
		YearStart: yearStart,
	}

	for _, activity := range activities {
		if activity.StartDate.After(yearStart) && activity.StartDate.Before(yearEnd) {
			summary.TotalTSS += activity.TSS
			summary.TotalMovingTime += activity.MovingTime
			summary.TotalDistance += activity.Distance
		}
	}

	return summary
}

// GetWeekStart returns the Monday 00:00:00 of the week containing the given date
func GetWeekStart(date time.Time) time.Time {
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	daysFromMonday := weekday - 1
	monday := date.AddDate(0, 0, -daysFromMonday)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// GetMonthStart returns the first day 00:00:00 of the month containing the given date
func GetMonthStart(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

// GetYearStart returns January 1st 00:00:00 of the year containing the given date
func GetYearStart(date time.Time) time.Time {
	return time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, date.Location())
}

// CalculateTSS calculates Training Stress Score
func CalculateTSS(normalizedPower, ftp float64, durationSeconds int) float64 {
	if ftp <= 0 {
		return 0
	}

	intensityFactor := normalizedPower / ftp
	return (float64(durationSeconds) * normalizedPower * intensityFactor) / (ftp * 3600) * 100
}

// CalculateIntensityFactor calculates Intensity Factor
func CalculateIntensityFactor(normalizedPower, ftp float64) float64 {
	if ftp <= 0 {
		return 0
	}
	return normalizedPower / ftp
}

// FormatDuration formats duration in seconds to human readable format
func FormatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d時間%d分", hours, minutes)
	}
	return fmt.Sprintf("%d分", minutes)
}

// FormatDistance formats distance in meters to kilometers
func FormatDistance(meters float64) string {
	km := meters / 1000
	return fmt.Sprintf("%.1fkm", km)
}

// FormatElevation formats elevation gain
func FormatElevation(meters float64) string {
	return fmt.Sprintf("%.0fm", meters)
}
