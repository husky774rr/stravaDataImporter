package strava

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"stravaDataImporter/internal/config"

	"golang.org/x/oauth2"
)

const (
	stravaBaseURL  = "https://www.strava.com/api/v3"
	stravaAuthURL  = "https://www.strava.com/oauth/authorize"
	stravaTokenURL = "https://www.strava.com/oauth/token"
)

type Client struct {
	config      *config.Config
	httpClient  *http.Client
	oauthConfig *oauth2.Config
}

func NewClient(cfg *config.Config) *Client {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.StravaClientID,
		ClientSecret: cfg.StravaClientSecret,
		RedirectURL:  cfg.StravaRedirectURL,
		Scopes:       []string{"read,activity:read_all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  stravaAuthURL,
			TokenURL: stravaTokenURL,
		},
	}

	return &Client{
		config:      cfg,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		oauthConfig: oauthConfig,
	}
}

func (c *Client) GetAuthURL(state string) string {
	return c.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *Client) ExchangeCodeForToken(code string) (*TokenData, error) {
	data := url.Values{}
	data.Set("client_id", c.config.StravaClientID)
	data.Set("client_secret", c.config.StravaClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	resp, err := c.httpClient.PostForm(stravaTokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &TokenData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Unix(tokenResp.ExpiresAt, 0),
		TokenType:    tokenResp.TokenType,
		AthleteID:    tokenResp.Athlete.ID,
	}, nil
}

func (c *Client) RefreshToken(refreshToken string) (*TokenData, error) {
	data := url.Values{}
	data.Set("client_id", c.config.StravaClientID)
	data.Set("client_secret", c.config.StravaClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := c.httpClient.PostForm(stravaTokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &TokenData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    time.Unix(tokenResp.ExpiresAt, 0),
		TokenType:    tokenResp.TokenType,
		AthleteID:    tokenResp.Athlete.ID,
	}, nil
}

func (c *Client) GetActivities(accessToken string, after time.Time, perPage int) ([]StravaActivity, error) {
	req, err := http.NewRequest("GET", stravaBaseURL+"/athlete/activities", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("after", strconv.FormatInt(after.Unix(), 10))
	q.Add("per_page", strconv.Itoa(perPage))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activities: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("activities request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var activities []StravaActivity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities response: %w", err)
	}

	slog.Info("Fetched activities from Strava", "count", len(activities))
	return activities, nil
}

func (c *Client) GetActivity(accessToken string, activityID int64) (*StravaActivity, error) {
	url := fmt.Sprintf("%s/activities/%d", stravaBaseURL, activityID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activity: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("activity request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var activity StravaActivity
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return nil, fmt.Errorf("failed to decode activity response: %w", err)
	}

	return &activity, nil
}

func (c *Client) GetAthlete(accessToken string) (*AthleteInfo, error) {
	req, err := http.NewRequest("GET", stravaBaseURL+"/athlete", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch athlete: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("athlete request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var athlete AthleteInfo
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		return nil, fmt.Errorf("failed to decode athlete response: %w", err)
	}

	return &athlete, nil
}

// ConvertToActivityData converts StravaActivity to ActivityData
func ConvertToActivityData(stravaActivity StravaActivity, ftp float64) (*ActivityData, error) {
	startDate, err := time.Parse(time.RFC3339, stravaActivity.StartDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start date: %w", err)
	}

	activity := &ActivityData{
		ID:                   stravaActivity.ID,
		Name:                 stravaActivity.Name,
		Type:                 stravaActivity.Type,
		Distance:             stravaActivity.Distance,
		MovingTime:           stravaActivity.MovingTime,
		ElapsedTime:          stravaActivity.ElapsedTime,
		TotalElevationGain:   stravaActivity.TotalElevationGain,
		StartDate:            startDate,
		AverageSpeed:         stravaActivity.AverageSpeed,
		MaxSpeed:             stravaActivity.MaxSpeed,
		Calories:             stravaActivity.Calories,
		AverageHeartrate:     stravaActivity.AverageHeartrate,
		MaxHeartrate:         stravaActivity.MaxHeartrate,
		AverageWatts:         stravaActivity.AverageWatts,
		MaxWatts:             stravaActivity.MaxWatts,
		WeightedAverageWatts: stravaActivity.WeightedAverageWatts,
		Kilojoules:           stravaActivity.Kilojoules,
		FTP:                  ftp,
	}

	// Calculate TSS and NP
	if ftp > 0 && stravaActivity.WeightedAverageWatts > 0 {
		activity.NP = stravaActivity.WeightedAverageWatts
		intensityFactor := activity.NP / ftp
		activity.TSS = (float64(stravaActivity.MovingTime) * activity.NP * intensityFactor) / (ftp * 3600) * 100
	}

	return activity, nil
}
