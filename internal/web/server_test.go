package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stravaDataImporter/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{
		Port: "8080",
	}

	server := NewServer(cfg, nil)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.config != cfg {
		t.Error("Server config not set correctly")
	}

	if server.router == nil {
		t.Error("Router not initialized")
	}

	if server.httpServer == nil {
		t.Error("HTTP server not initialized")
	}
}

func TestHealthRoute(t *testing.T) {
	cfg := &config.Config{
		Port: "8080",
	}

	server := NewServer(cfg, nil)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Health route returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	expectedContentType := "application/json; charset=utf-8"
	if rr.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("Health route returned wrong content type: got %v want %v", rr.Header().Get("Content-Type"), expectedContentType)
	}
}

func TestCORSMiddleware(t *testing.T) {
	cfg := &config.Config{
		Port: "8080",
	}

	server := NewServer(cfg, nil)

	req, err := http.NewRequest("OPTIONS", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("OPTIONS request returned wrong status code: got %v want %v", rr.Code, http.StatusNoContent)
	}

	corsHeader := rr.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("CORS header not set correctly: got %v want %v", corsHeader, "*")
	}
}

func TestHomeRedirect(t *testing.T) {
	cfg := &config.Config{
		Port: "8080",
	}

	server := NewServer(cfg, nil)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("Home route returned wrong status code: got %v want %v", rr.Code, http.StatusFound)
	}

	location := rr.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Home route redirected to wrong location: got %v want %v", location, "/login")
	}
}

func TestProtectedRouteWithoutAuth(t *testing.T) {
	cfg := &config.Config{
		Port: "8080",
	}

	server := NewServer(cfg, nil)

	req, err := http.NewRequest("GET", "/portal", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Errorf("Protected route returned wrong status code: got %v want %v", rr.Code, http.StatusFound)
	}

	location := rr.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Protected route redirected to wrong location: got %v want %v", location, "/login")
	}
}
