package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// DBHealth represents the health status of the database
// Time-based fields (waitDuration and responseTime) are in milliseconds
type DBHealth struct {
	Status            string `json:"status"`
	OpenConnections   int    `json:"openConnections"`
	InUseConnections  int    `json:"inUseConnections"`
	IdleConnections   int    `json:"idleConnections"`
	WaitCount         int64  `json:"waitCount"`
	WaitDuration      int64  `json:"waitDuration"`
	MaxIdleClosed     int64  `json:"maxIdleClosed"`
	MaxLifetimeClosed int64  `json:"maxLifetimeClosed"`
	ResponseTime      int64  `json:"responseTime"`
}

// Health checks the database health and returns a DBHealth struct
// Checks if the database is reachable, response time and important aspects of the connection.
func Health(ctx context.Context, db *sql.DB, l *slog.Logger) (*DBHealth, error) {
	start := time.Now()

	health := &DBHealth{}

	if err := db.PingContext(ctx); err != nil {
		health.Status = "Unreachable"
		l.ErrorContext(ctx, "Database ping failed", "error", err)
		return health, fmt.Errorf("database unreachable: %w", err)
	}

	health.ResponseTime = time.Since(start).Milliseconds()

	stats := db.Stats()
	health.OpenConnections = stats.OpenConnections
	health.InUseConnections = stats.InUse
	health.IdleConnections = stats.Idle
	health.WaitCount = stats.WaitCount
	health.WaitDuration = stats.WaitDuration.Milliseconds()
	health.MaxIdleClosed = stats.MaxIdleClosed
	health.MaxLifetimeClosed = stats.MaxLifetimeClosed

	if stats.OpenConnections >= (db.Stats().MaxOpenConnections * 90 / 100) {
		health.Status = "Near Capacity"
		l.WarnContext(ctx, "Database connections near capacity",
			"openConnections", stats.OpenConnections,
			"maxOpenConnections", db.Stats().MaxOpenConnections)
	} else {
		health.Status = "Healthy"
	}

	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		health.Status = "Query Error"
		l.ErrorContext(ctx, "Database query test failed", "error", err)
		return health, fmt.Errorf("query test failed: %w", err)
	}

	return health, nil
}
