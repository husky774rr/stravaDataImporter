package twitter

import (
	"fmt"
	"log/slog"
	"time"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/strava"
)

type Client struct {
	config *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

func (c *Client) PostActivity(activity *strava.ActivityData) error {
	slog.Info("Posting activity to Twitter", "activity_id", activity.ID)

	// Format the tweet text
	tweetText := c.formatActivityTweet(activity)

	// In a real implementation, you would use the Twitter API
	// For now, we'll just log the tweet
	slog.Info("Tweet content", "text", tweetText)

	// TODO: Implement actual Twitter API posting
	// This would require using the Twitter API v2 client
	// and handling image generation for the weekly trend

	return nil
}

func (c *Client) formatActivityTweet(activity *strava.ActivityData) string {
	// Format the date in Japanese style
	dateStr := activity.StartDate.Format("2006年01月02日(Mon)")
	weekdays := map[string]string{
		"Mon": "月", "Tue": "火", "Wed": "水", "Thu": "木",
		"Fri": "金", "Sat": "土", "Sun": "日",
	}

	for en, jp := range weekdays {
		dateStr = fmt.Sprintf(activity.StartDate.Format("2006年01月02日(%s)"), jp)
		if activity.StartDate.Format("Mon") == en {
			break
		}
	}

	// Convert activity type to Japanese
	activityTypeJP := c.translateActivityType(activity.Type)

	// Format duration
	duration := time.Duration(activity.MovingTime) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	var durationStr string
	if hours > 0 {
		durationStr = fmt.Sprintf("%d時間%d分", hours, minutes)
	} else {
		durationStr = fmt.Sprintf("%d分", minutes)
	}

	// Format distance
	distanceKm := activity.Distance / 1000
	distanceStr := fmt.Sprintf("%.1fkm", distanceKm)

	// Format elevation
	elevationStr := fmt.Sprintf("%.0fm", activity.TotalElevationGain)

	// Format calories with comma separator
	caloriesStr := fmt.Sprintf("%skcal", formatNumber(int(activity.Calories)))

	tweet := fmt.Sprintf(`%s
TSS: %.0f
NP: %.0f
アクティビティタイプ: %s
消費カロリー: %s
運動時間: %s
走行距離: %s
獲得標高: %s`,
		dateStr,
		activity.TSS,
		activity.NP,
		activityTypeJP,
		caloriesStr,
		durationStr,
		distanceStr,
		elevationStr,
	)

	return tweet
}

func (c *Client) translateActivityType(activityType string) string {
	translations := map[string]string{
		"Ride":            "サイクリング",
		"Run":             "ランニング",
		"Swim":            "水泳",
		"Walk":            "ウォーキング",
		"Hike":            "ハイキング",
		"AlpineSki":       "アルペンスキー",
		"BackcountrySki":  "バックカントリースキー",
		"Canoeing":        "カヌー",
		"Crossfit":        "クロスフィット",
		"EBikeRide":       "電動自転車",
		"Elliptical":      "エリプティカル",
		"Golf":            "ゴルフ",
		"Handcycle":       "ハンドサイクル",
		"IceSkate":        "アイススケート",
		"InlineSkate":     "インラインスケート",
		"Kayaking":        "カヤック",
		"Kitesurf":        "カイトサーフィン",
		"NordicSki":       "ノルディックスキー",
		"RockClimbing":    "ロッククライミング",
		"RollerSki":       "ローラースキー",
		"Rowing":          "ローイング",
		"Sailing":         "セーリング",
		"Skateboard":      "スケートボード",
		"Snowboard":       "スノーボード",
		"Snowshoe":        "スノーシュー",
		"Soccer":          "サッカー",
		"StairStepper":    "ステアステッパー",
		"StandUpPaddling": "SUP",
		"Surfing":         "サーフィン",
		"Tennis":          "テニス",
		"TrailRun":        "トレイルラン",
		"Velomobile":      "ベロモービル",
		"VirtualRide":     "バーチャルライド",
		"VirtualRun":      "バーチャルラン",
		"WeightTraining":  "ウェイトトレーニング",
		"Wheelchair":      "車椅子",
		"Windsurf":        "ウィンドサーフィン",
		"Workout":         "ワークアウト",
		"Yoga":            "ヨガ",
	}

	if translated, exists := translations[activityType]; exists {
		return translated
	}
	return activityType // Return original if no translation found
}

func (c *Client) generateWeeklyTrendImage(summaries []strava.WeeklySummary) ([]byte, error) {
	// TODO: Implement chart generation
	// This would use a charting library like go-echarts to generate
	// a trend chart showing the weekly TSS, distance, time, etc.

	slog.Info("Generating weekly trend image", "summaries_count", len(summaries))

	// For now, return nil as this is a complex implementation
	// that would require additional dependencies
	return nil, fmt.Errorf("image generation not implemented")
}

// formatNumber adds comma separators to numbers for better readability
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	// Add commas every 3 digits from right
	result := ""
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}
	return result
}
