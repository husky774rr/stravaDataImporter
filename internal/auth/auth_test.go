package auth

import (
	"path/filepath"
	"testing"
	"time"

	"stravaDataImporter/internal/strava"
)

func TestTokenStore(t *testing.T) {
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "token.json")

	store := NewTokenStore(tokenFile)

	// Test saving and loading token
	token := &strava.TokenData{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
	}

	err := store.SaveToken(token)
	if err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	loadedToken, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}

	if loadedToken.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loadedToken.AccessToken, token.AccessToken)
	}

	if loadedToken.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %v, want %v", loadedToken.RefreshToken, token.RefreshToken)
	}

	// Test HasValidToken
	if !store.HasValidToken() {
		t.Error("HasValidToken() = false, want true")
	}

	// Test with expired token
	expiredToken := &strava.TokenData{
		AccessToken:  "expired_token",
		RefreshToken: "expired_refresh",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		TokenType:    "Bearer",
	}

	err = store.SaveToken(expiredToken)
	if err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	if store.HasValidToken() {
		t.Error("HasValidToken() = true, want false for expired token")
	}

	// Test ClearToken
	err = store.ClearToken()
	if err != nil {
		t.Fatalf("ClearToken() error = %v", err)
	}

	if store.HasValidToken() {
		t.Error("HasValidToken() = true, want false after clearing token")
	}
}

func TestTokenStoreNoFile(t *testing.T) {
	tempDir := t.TempDir()
	tokenFile := filepath.Join(tempDir, "nonexistent.json")

	store := NewTokenStore(tokenFile)

	token, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error = %v", err)
	}

	if token != nil {
		t.Error("LoadToken() = non-nil, want nil for nonexistent file")
	}

	if store.HasValidToken() {
		t.Error("HasValidToken() = true, want false for nonexistent token")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error = %v", err)
	}

	if state1 == "" {
		t.Error("GenerateState() returned empty string")
	}

	state2, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() error = %v", err)
	}

	if state1 == state2 {
		t.Error("GenerateState() returned same state twice")
	}
}

func TestStateStore(t *testing.T) {
	store := NewStateStore()

	// Generate and store state
	state, err := store.GenerateAndStore()
	if err != nil {
		t.Fatalf("GenerateAndStore() error = %v", err)
	}

	if state == "" {
		t.Error("GenerateAndStore() returned empty string")
	}

	// Validate state
	if !store.ValidateAndRemove(state) {
		t.Error("ValidateAndRemove() = false, want true for valid state")
	}

	// Try to validate same state again (should fail)
	if store.ValidateAndRemove(state) {
		t.Error("ValidateAndRemove() = true, want false for already used state")
	}

	// Test invalid state
	if store.ValidateAndRemove("invalid_state") {
		t.Error("ValidateAndRemove() = true, want false for invalid state")
	}
}
