package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"stravaDataImporter/internal/auth"
	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/db"
	"stravaDataImporter/internal/ftp"
	"stravaDataImporter/internal/strava"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	config       *config.Config
	stravaClient *strava.Client
	tokenStore   *auth.TokenStore
	stateStore   *auth.StateStore
	ftpManager   *ftp.FTPManager
	influxClient *db.InfluxDBClient
}

func NewHandler(cfg *config.Config, influxClient *db.InfluxDBClient) *Handler {
	var tokenStore *auth.TokenStore
	var ftpManager *ftp.FTPManager

	if influxClient != nil {
		tokenStore = auth.NewTokenStore(influxClient)
		ftpManager = ftp.NewFTPManager(cfg.FTPFilePath)
	}

	return &Handler{
		config:       cfg,
		stravaClient: strava.NewClient(cfg),
		tokenStore:   tokenStore,
		stateStore:   auth.NewStateStore(),
		ftpManager:   ftpManager,
		influxClient: influxClient,
	}
}

func (h *Handler) Home(c *gin.Context) {
	if h.tokenStore != nil && h.tokenStore.HasValidToken() {
		c.Redirect(http.StatusFound, "/portal")
		return
	}
	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Strava Data Importer - Login",
	})
}

func (h *Handler) Portal(c *gin.Context) {
	if h.tokenStore == nil || !h.tokenStore.HasValidToken() {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get latest activity
	if h.influxClient == nil {
		c.HTML(http.StatusOK, "portal.html", gin.H{
			"title":   "Strava Data Importer - Portal",
			"loading": true,
			"message": "Please wait while we fetch your activities...",
		})
		return
	}

	latestActivity, err := h.influxClient.GetLatestActivity()
	if err != nil {
		slog.Error("Failed to get latest activity", "error", err)
		c.HTML(http.StatusOK, "portal.html", gin.H{
			"title":   "Strava Data Importer - Portal",
			"loading": true,
			"message": "Please wait while we fetch your activities...",
		})
		return
	}

	if latestActivity == nil {
		c.HTML(http.StatusOK, "portal.html", gin.H{
			"title":   "Strava Data Importer - Portal",
			"loading": true,
			"message": "Please wait while we fetch your activities...",
		})
		return
	}

	c.HTML(http.StatusOK, "portal.html", gin.H{
		"title":    "Strava Data Importer - Portal",
		"loading":  false,
		"activity": latestActivity,
	})
}

func (h *Handler) AuthLogin(c *gin.Context) {
	state, err := h.stateStore.GenerateAndStore()
	if err != nil {
		slog.Error("Failed to generate state", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	authURL := h.stravaClient.GetAuthURL(state)
	c.Redirect(http.StatusFound, authURL)
}

func (h *Handler) AuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		slog.Error("OAuth error", "error", errorParam)
		c.Redirect(http.StatusFound, "/login?error=access_denied")
		return
	}

	if code == "" {
		slog.Error("No authorization code received")
		c.Redirect(http.StatusFound, "/login?error=no_code")
		return
	}

	if !h.stateStore.ValidateAndRemove(state) {
		slog.Error("Invalid state parameter", "state", state)
		c.Redirect(http.StatusFound, "/login?error=invalid_state")
		return
	}

	token, err := h.stravaClient.ExchangeCodeForToken(code)
	if err != nil {
		slog.Error("Failed to exchange code for token", "error", err)
		c.Redirect(http.StatusFound, "/login?error=token_exchange_failed")
		return
	}

	if err := h.tokenStore.SaveToken(token); err != nil {
		slog.Error("Failed to save token", "error", err)
		c.Redirect(http.StatusFound, "/login?error=token_save_failed")
		return
	}

	// Load FTP data
	if err := h.ftpManager.LoadFTPData(); err != nil {
		slog.Warn("Failed to load FTP data", "error", err)
	}

	slog.Info("Authentication successful")
	c.Redirect(http.StatusFound, "/portal")
}

func (h *Handler) AuthLogout(c *gin.Context) {
	if err := h.tokenStore.ClearToken(); err != nil {
		slog.Error("Failed to clear token", "error", err)
	}

	c.Redirect(http.StatusFound, "/login")
}

func (h *Handler) RefreshToken(c *gin.Context) {
	token, err := h.tokenStore.LoadToken()
	if err != nil || token == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
		return
	}

	newToken, err := h.stravaClient.RefreshToken(token.RefreshToken)
	if err != nil {
		slog.Error("Failed to refresh token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	if err := h.tokenStore.SaveToken(newToken); err != nil {
		slog.Error("Failed to save refreshed token", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}

func (h *Handler) GetActivities(c *gin.Context) {
	token, err := h.tokenStore.LoadToken()
	if err != nil || token == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token found"})
		return
	}

	perPageStr := c.DefaultQuery("per_page", "30")
	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 || perPage > 200 {
		perPage = 30
	}

	activities, err := h.stravaClient.GetActivities(token.AccessToken, token.ExpiresAt.AddDate(0, 0, -30), perPage)
	if err != nil {
		slog.Error("Failed to get activities", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"count":      len(activities),
	})
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": c.GetTime("request_time"),
	})
}
