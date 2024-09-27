package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ashtishad/xm/common"
	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID       `json:"id"`
	EventType string          `json:"eventType"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type EventRepository interface {
	StoreEvent(ctx context.Context, eventType string, data json.RawMessage) error
}

type eventRepository struct {
	db *sql.DB
	l  *slog.Logger
}

func NewEventRepository(db *sql.DB, l *slog.Logger) EventRepository {
	return &eventRepository{
		db: db,
		l:  l,
	}
}

func (r *eventRepository) StoreEvent(ctx context.Context, eventType string, data json.RawMessage) error {
	query := `
        INSERT INTO events (event_type, data)
        VALUES ($1, $2)
    `
	if _, err := r.db.ExecContext(ctx, query, eventType, data); err != nil {
		r.l.Error("failed to store event", "err", err)
		return common.NewInternalServerError(common.ErrUnexpectedEvent, err)
	}

	return nil
}
