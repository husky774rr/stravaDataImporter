package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"stravaDataImporter/internal/strava"
)

type TokenStore struct {
	mu       sync.RWMutex
	filePath string
	token    *strava.TokenData
}

func NewTokenStore(filePath string) *TokenStore {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("Failed to create token directory", "error", err)
	}

	return &TokenStore{
		filePath: filePath,
	}
}

func (ts *TokenStore) SaveToken(token *strava.TokenData) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(ts.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	ts.token = token
	slog.Info("Token saved successfully")
	return nil
}

func (ts *TokenStore) LoadToken() (*strava.TokenData, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	if ts.token != nil {
		return ts.token, nil
	}

	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token exists
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token strava.TokenData
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	ts.token = &token
	return &token, nil
}

func (ts *TokenStore) HasValidToken() bool {
	token, err := ts.LoadToken()
	if err != nil || token == nil {
		return false
	}

	// Check if token is expired (with 5 minute buffer)
	return time.Now().Add(5 * time.Minute).Before(token.ExpiresAt)
}

func (ts *TokenStore) ClearToken() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if err := os.Remove(ts.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file: %w", err)
	}

	ts.token = nil
	slog.Info("Token cleared successfully")
	return nil
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type StateStore struct {
	mu     sync.RWMutex
	states map[string]time.Time
}

func NewStateStore() *StateStore {
	store := &StateStore{
		states: make(map[string]time.Time),
	}

	// Clean up expired states every 10 minutes
	go store.cleanupExpiredStates()

	return store
}

func (ss *StateStore) GenerateAndStore() (string, error) {
	state, err := GenerateState()
	if err != nil {
		return "", err
	}

	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.states[state] = time.Now().Add(10 * time.Minute) // Valid for 10 minutes
	return state, nil
}

func (ss *StateStore) ValidateAndRemove(state string) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	expiry, exists := ss.states[state]
	if !exists {
		return false
	}

	delete(ss.states, state)
	return time.Now().Before(expiry)
}

func (ss *StateStore) cleanupExpiredStates() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ss.mu.Lock()
		now := time.Now()
		for state, expiry := range ss.states {
			if now.After(expiry) {
				delete(ss.states, state)
			}
		}
		ss.mu.Unlock()
	}
}
