package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"stravaDataImporter/internal/config"
	"stravaDataImporter/internal/db"
	"stravaDataImporter/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *config.Config
	router     *gin.Engine
	httpServer *http.Server
	handler    *handlers.Handler
}

func NewServer(cfg *config.Config, influxClient *db.InfluxDBClient) *Server {
	// Set Gin mode
	if cfg.Port == "8080" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(requestTimeMiddleware())

	// Load HTML templates
	templatePath := "internal/web/templates/*"
	// Try to load templates, handle case where path doesn't exist (for tests)
	if templates, err := filepath.Glob(templatePath); err != nil || len(templates) == 0 {
		slog.Warn("Templates not found at primary path, trying alternative paths", "path", templatePath)
		// Try alternative paths for tests
		altPaths := []string{"templates/*", "web/templates/*", "../web/templates/*"}
		for _, altPath := range altPaths {
			if templates, err := filepath.Glob(altPath); err == nil && len(templates) > 0 {
				templatePath = altPath
				break
			}
		}
	}
	router.LoadHTMLGlob(templatePath)

	// Serve static files (if any)
	router.Static("/static", "./static")

	handler := handlers.NewHandler(cfg, influxClient)

	server := &Server{
		config:  cfg,
		router:  router,
		handler: handler,
	}

	server.setupRoutes()

	server.httpServer = &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        server.router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return server
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.handler.Health)

	// Public routes
	s.router.GET("/", s.handler.Home)
	s.router.GET("/login", s.handler.Login)

	// Auth routes
	auth := s.router.Group("/auth")
	{
		auth.GET("/login", s.handler.AuthLogin)
		auth.GET("/callback", s.handler.AuthCallback)
		auth.POST("/logout", s.handler.AuthLogout)
		auth.POST("/refresh", s.handler.RefreshToken)
	}

	// Protected routes
	protected := s.router.Group("/")
	protected.Use(authMiddleware(s.handler))
	{
		protected.GET("/portal", s.handler.Portal)
		protected.GET("/activities", s.handler.GetActivities)
	}

	// API routes
	api := s.router.Group("/api/v1")
	api.Use(authMiddleware(s.handler))
	{
		api.GET("/health", s.handler.Health)
		api.GET("/activities", s.handler.GetActivities)
		api.POST("/auth/refresh", s.handler.RefreshToken)
	}
}

func (s *Server) Start() error {
	slog.Info("Starting web server", "port", s.config.Port)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down web server")

	return s.httpServer.Shutdown(ctx)
}

// Middleware functions

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func requestTimeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("request_time", time.Now())
		c.Next()
	}
}

func authMiddleware(handler *handlers.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is an API route
		if c.FullPath() != "" && c.FullPath()[:4] == "/api" {
			// For API routes, return JSON error
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// For web routes, redirect to login
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
	}
}
