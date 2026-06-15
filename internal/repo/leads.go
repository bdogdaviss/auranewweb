package repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// LeadRepository stores marketing email captures (affiliate signup, payout
// waitlist). It's intentionally minimal — just durable storage so the "you're
// on the list" message the user sees is actually true.
type LeadRepository interface {
	// AddLead records an email for a source. Idempotent per (email, source).
	AddLead(ctx context.Context, email, name, source string) error
}

type sqliteLeadRepo struct {
	db *sql.DB
}

func NewLeadRepository(db *sql.DB) LeadRepository {
	return &sqliteLeadRepo{db: db}
}

func (r *sqliteLeadRepo) AddLead(ctx context.Context, email, name, source string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	source = strings.TrimSpace(source)
	if email == "" || source == "" {
		return errors.New("email and source are required")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO leads (email, name, source) VALUES (?, ?, ?)
		ON CONFLICT(email, source) DO NOTHING
	`, email, nullableString(name), source)
	return err
}
