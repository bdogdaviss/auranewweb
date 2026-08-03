package repo

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// SessionRepository persists login sessions so a server restart (or a
// multi-machine deploy) doesn't invalidate everyone's cookies. main.go keeps
// its in-memory map as a fast path; this is the durable source of truth.
type SessionRepository interface {
	// Create stores a session token for a user with an expiry.
	Create(ctx context.Context, token, userID string, expiresAt time.Time) error

	// GetUserID resolves a token to a user id. Returns "" (no error) for
	// unknown or expired tokens.
	GetUserID(ctx context.Context, token string) (string, error)

	// Delete removes a session (logout). Deleting a missing token is a no-op.
	Delete(ctx context.Context, token string) error

	// DeleteExpired prunes expired rows. Safe to call opportunistically.
	DeleteExpired(ctx context.Context) error
}

type sqliteSessionRepo struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sqliteSessionRepo{db: db}
}

func (r *sqliteSessionRepo) Create(ctx context.Context, token, userID string, expiresAt time.Time) error {
	if token == "" || userID == "" {
		return errors.New("sessions: token and userID are required")
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt.UTC())
	return err
}

func (r *sqliteSessionRepo) GetUserID(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", nil
	}
	var userID string
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM sessions WHERE token = ? AND expires_at > ? LIMIT 1`,
		token, time.Now().UTC()).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *sqliteSessionRepo) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (r *sqliteSessionRepo) DeleteExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= ?`, time.Now().UTC())
	return err
}
