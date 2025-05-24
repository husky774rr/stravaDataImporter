package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/db"
	"stravaDataImporter/internal/strava"

	"github.com/gin-gonic/gin"
)

// InfluxDBClientの完全なモック実装
type mockInfluxDBClient struct{}

func (m *mockInfluxDBClient) Close()                                                   {}
func (m *mockInfluxDBClient) WriteActivity(activity *strava.ActivityData) error        { return nil }
func (m *mockInfluxDBClient) WriteWeeklySummary(summary *strava.WeeklySummary) error   { return nil }
func (m *mockInfluxDBClient) WriteMonthlySummary(summary *strava.MonthlySummary) error { return nil }
func (m *mockInfluxDBClient) WriteYearlySummary(summary *strava.YearlySummary) error   { return nil }
func (m *mockInfluxDBClient) GetLatestActivity() (*strava.ActivityData, error)         { return nil, nil }
func (m *mockInfluxDBClient) GetWeeklyTrend() ([]strava.WeeklySummary, error)          { return nil, nil }
func (m *mockInfluxDBClient) SaveToken(token *strava.TokenData) error                  { return nil }
func (m *mockInfluxDBClient) LoadToken() (*strava.TokenData, error)                    { return nil, nil }
func (m *mockInfluxDBClient) ClearToken() error                                        { return nil }

// mockInfluxDBClientを*db.InfluxDBClientに変換するヘルパー
func createMockInfluxDBClient() *db.InfluxDBClient {
	// 実際にはこれは型安全ではないので、テスト用にnilクライアントを受け入れるようにハンドラーを修正する方が良い
	return nil
}

func TestNewHandler(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
		StravaRedirectURL:  "http://localhost:9090/auth/callback",
		FTPFilePath:        "./test_ftp.csv",
	}

	// nilクライアントでテスト - これは実際のテストではハンドラーの作成をスキップする
	handler := NewHandler(cfg, nil)
	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}

	if handler.config != cfg {
		t.Error("Handler config not set correctly")
	}
}

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	handler := NewHandler(cfg, nil)

	router := gin.New()
	router.GET("/health", handler.Health)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Health handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	expectedContentType := "application/json; charset=utf-8"
	if rr.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Health handler returned wrong content type: got %v want %v", rr.Header().Get("Content-Type"), expectedContentType)
	}
}

func TestHomeRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	handler := NewHandler(cfg, nil)

	router := gin.New()
	router.GET("/", handler.Home)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Should redirect to login when no valid token
	if rr.Code != http.StatusFound {
		t.Errorf("Home handler returned wrong status code: got %v want %v", rr.Code, http.StatusFound)
	}

	location := rr.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Home handler redirected to wrong location: got %v want %v", location, "/login")
	}
}
