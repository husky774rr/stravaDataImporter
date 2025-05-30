package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"stravaDataImporter/internal/strava"
)

// InfluxDBTokenStore interface for dependency injection
type InfluxDBTokenStore interface {
	SaveToken(token *strava.TokenData) error
	LoadToken() (*strava.TokenData, error)
	ClearToken() error
}

type TokenStore struct {
	mu          sync.RWMutex
	influxStore InfluxDBTokenStore
	token       *strava.TokenData
}

func NewTokenStore(influxStore InfluxDBTokenStore) *TokenStore {
	return &TokenStore{
		influxStore: influxStore,
	}
}

func (ts *TokenStore) SaveToken(token *strava.TokenData) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if err := ts.influxStore.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token to InfluxDB: %w", err)
	}

	ts.token = token
	slog.Info("Token saved successfully to InfluxDB")
	return nil
}

func (ts *TokenStore) LoadToken() (*strava.TokenData, error) {
	// 最初に読み取りロックでメモリキャッシュをチェック
	ts.mu.RLock()
	if ts.token != nil {
		cachedToken := ts.token
		ts.mu.RUnlock()
		slog.Debug("Token loaded from memory cache", "athlete_id", cachedToken.AthleteID)
		return cachedToken, nil
	}
	ts.mu.RUnlock()

	// メモリキャッシュにない場合は書き込みロックでInfluxDBから読み込み
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// ダブルチェック（他のgoroutineが既に読み込んでいる可能性）
	if ts.token != nil {
		slog.Debug("Token loaded from memory cache (double check)", "athlete_id", ts.token.AthleteID)
		return ts.token, nil
	}

	slog.Debug("Loading token from InfluxDB")
	token, err := ts.influxStore.LoadToken()
	if err != nil {
		slog.Error("Failed to load token from InfluxDB", "error", err)
		return nil, fmt.Errorf("failed to load token from InfluxDB: %w", err)
	}

	if token != nil {
		ts.token = token
		slog.Debug("Token loaded from InfluxDB and cached", "athlete_id", token.AthleteID)
	} else {
		slog.Debug("No token found in InfluxDB")
	}

	return token, nil
}

func (ts *TokenStore) HasValidToken() bool {
	token, err := ts.LoadToken()
	if err != nil {
		slog.Debug("Failed to load token", "error", err)
		return false
	}
	if token == nil {
		slog.Debug("No token found")
		return false
	}

	// Check if token is expired (with 5 minute buffer)
	isValid := time.Now().Add(5 * time.Minute).Before(token.ExpiresAt)
	slog.Debug("Token validity check", "expires_at", token.ExpiresAt, "is_valid", isValid, "athlete_id", token.AthleteID)
	return isValid
}

func (ts *TokenStore) ClearToken() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if err := ts.influxStore.ClearToken(); err != nil {
		return fmt.Errorf("failed to clear token from InfluxDB: %w", err)
	}

	ts.token = nil
	slog.Info("Token cleared successfully from InfluxDB")
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
