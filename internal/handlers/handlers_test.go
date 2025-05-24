package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stravaDataImporter/internal/config"

	"github.com/gin-gonic/gin"
)

func TestNewHandler(t *testing.T) {
	cfg := &config.Config{
		StravaClientID:     "test_client_id",
		StravaClientSecret: "test_client_secret",
		StravaRedirectURL:  "http://localhost:9090/auth/callback",
		FTPFilePath:        "./test_ftp.csv",
	}

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
