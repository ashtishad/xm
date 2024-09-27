package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/ashtishad/xm/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// NewConnection creates a postgres db handle using pgx, returns *sql.DB
func NewConnection(ctx context.Context, l *slog.Logger, cfg common.DBConfig) (*sql.DB, error) {
	connConfig, err := pgx.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("invalid connection string: %w", err)
	}

	db := stdlib.OpenDB(*connConfig)

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.PingContext(ctx); err != nil {
		l.Error("failed to ping postgres connection", "err", err)
		return nil, err
	}

	l.Info("successfully connected to postgres", "dsn", connConfig.ConnString())

	return db, nil
}
