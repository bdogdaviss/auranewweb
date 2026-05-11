package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type License struct {
	ID                    int64
	Key                   string
	UserEmail             string
	ProductID             string
	StripePaymentIntentID string
	CreatedAt             time.Time
	RevokedAt             sql.NullTime
}

type LicenseRepository interface {
	// InsertIfNew tries to insert a license for (userEmail, productID). If one
	// already exists (UNIQUE constraint), it returns the existing row with
	// wasNew=false. Webhook handlers rely on this being idempotent.
	InsertIfNew(ctx context.Context, userEmail, productID, stripePaymentIntentID string) (lic *License, wasNew bool, err error)

	// FindActive returns the active (non-revoked) license for (userEmail, productID),
	// or nil if none exists.
	FindActive(ctx context.Context, userEmail, productID string) (*License, error)

	// FindLatestByEmail returns the most recent active license for the email
	// across any product, or nil. Used by the resend-license endpoint.
	FindLatestByEmail(ctx context.Context, userEmail string) (*License, error)

	// RevokeByKey marks the license as revoked. No-op if already revoked.
	RevokeByKey(ctx context.Context, key string) error
}

type sqliteLicenseRepo struct {
	db *sql.DB
}

func NewLicenseRepository(db *sql.DB) LicenseRepository {
	return &sqliteLicenseRepo{db: db}
}

func (r *sqliteLicenseRepo) InsertIfNew(ctx context.Context, userEmail, productID, paymentIntentID string) (*License, bool, error) {
	if userEmail == "" || productID == "" {
		return nil, false, errors.New("userEmail and productID are required")
	}

	newKey := uuid.NewString()

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO license_keys (key, user_email, product_id, stripe_payment_intent_id)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_email, product_id) DO NOTHING
	`, newKey, userEmail, productID, nullableString(paymentIntentID))
	if err != nil {
		return nil, false, fmt.Errorf("insert license: %w", err)
	}
	affected, _ := res.RowsAffected()
	wasNew := affected > 0

	lic, err := r.findOne(ctx, `
		SELECT id, key, user_email, product_id, COALESCE(stripe_payment_intent_id, ''),
		       created_at, revoked_at
		FROM license_keys
		WHERE user_email = ? AND product_id = ?
	`, userEmail, productID)
	if err != nil {
		return nil, false, fmt.Errorf("re-read license: %w", err)
	}
	if lic == nil {
		return nil, false, errors.New("license vanished between insert and select")
	}
	return lic, wasNew, nil
}

func (r *sqliteLicenseRepo) FindActive(ctx context.Context, userEmail, productID string) (*License, error) {
	return r.findOne(ctx, `
		SELECT id, key, user_email, product_id, COALESCE(stripe_payment_intent_id, ''),
		       created_at, revoked_at
		FROM license_keys
		WHERE user_email = ? AND product_id = ? AND revoked_at IS NULL
		LIMIT 1
	`, userEmail, productID)
}

func (r *sqliteLicenseRepo) FindLatestByEmail(ctx context.Context, userEmail string) (*License, error) {
	return r.findOne(ctx, `
		SELECT id, key, user_email, product_id, COALESCE(stripe_payment_intent_id, ''),
		       created_at, revoked_at
		FROM license_keys
		WHERE user_email = ? AND revoked_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userEmail)
}

func (r *sqliteLicenseRepo) RevokeByKey(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("license key is required")
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE license_keys SET revoked_at = CURRENT_TIMESTAMP
		WHERE key = ? AND revoked_at IS NULL
	`, key)
	return err
}

func (r *sqliteLicenseRepo) findOne(ctx context.Context, query string, args ...interface{}) (*License, error) {
	var lic License
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&lic.ID, &lic.Key, &lic.UserEmail, &lic.ProductID, &lic.StripePaymentIntentID,
		&lic.CreatedAt, &lic.RevokedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lic, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
