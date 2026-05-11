package repo

import (
	"context"
	"database/sql"
	"errors"
)

// EventStore tracks Stripe webhook events that have been fully processed
// (signature verified, license inserted, email delivered). Used to short-circuit
// duplicate deliveries that Stripe sends on retry.
type EventStore interface {
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string) error
}

type sqliteEventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) EventStore {
	return &sqliteEventStore{db: db}
}

func (s *sqliteEventStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	if eventID == "" {
		return false, nil
	}
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM processed_events WHERE event_id = ?`, eventID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *sqliteEventStore) MarkProcessed(ctx context.Context, eventID string) error {
	if eventID == "" {
		return errors.New("eventID is required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO processed_events (event_id) VALUES (?)
		ON CONFLICT(event_id) DO NOTHING
	`, eventID)
	return err
}
