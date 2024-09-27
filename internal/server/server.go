package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ashtishad/xm/common"
	"github.com/ashtishad/xm/infra/postgres"
	"github.com/gin-gonic/gin"
)

type Server struct {
	router *gin.Engine
	db     *sql.DB
	Logger *slog.Logger
	server *http.Server
	Config *common.AppConfig
}

func NewServer(ctx context.Context) (*Server, error) {
	logger := common.NewSlogger()
	slog.SetDefault(logger)

	cfg, err := common.LoadConfig(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	db, err := postgres.NewConnection(ctx, logger, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection pool: %w", err)
	}

	if err := postgres.RunMigrations(ctx, logger, db); err != nil {
		return nil, err
	}

	gin.SetMode(cfg.Server.GinMode)
	router := gin.New()

	s := &Server{
		router: router,
		Logger: logger,
		db:     db,
		Config: cfg,
		server: &http.Server{
			Addr:         cfg.Server.Address,
			Handler:      router,
			IdleTimeout:  time.Minute,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
	}

	s.setupMiddleware()
	s.setupRoutes()

	s.Logger.Info(fmt.Sprintf("Swagger Specs available at %s/swagger/index.html", s.server.Addr))

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.db.Close(); err != nil {
		s.Logger.Error("failed to close database connection", "error", err)
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.Logger.Info("Server stopped gracefully")
	return nil
}

// HealthCheck godoc
// @Summary Check the health of the database connection.
// @Description Check the health of the database connection.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]any
// @Router /health [get]
func (s *Server) dbHealthHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := struct {
		Status    string             `json:"status"`
		Database  *postgres.DBHealth `json:"database"`
		Timestamp time.Time          `json:"timestamp"`
	}{
		Status:    "OK",
		Timestamp: time.Now(),
	}

	dbHealth, err := postgres.Health(ctx, s.db, s.Logger)
	if err != nil {
		health.Status = "Error"
		health.Database = &postgres.DBHealth{Status: "Error", ResponseTime: dbHealth.ResponseTime}
		s.Logger.ErrorContext(ctx, "database health check failed", "err", err)
	} else {
		health.Database = dbHealth
	}

	c.JSON(http.StatusOK, health)
}
