package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrPoolEmpty is returned by AssignFromPool when no unsold keys remain for the
// requested product. The webhook handler maps this to a 500 so Stripe retries;
// once the pool is refilled, the retry will succeed and the buyer gets their key.
var ErrPoolEmpty = errors.New("license pool is empty for this product")

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
	// AssignFromPool atomically issues a license to (userEmail, productID).
	// If the buyer already has an active license for the product, returns it
	// with wasNew=false (idempotent against Stripe retries). Otherwise pops
	// one row from available_keys, inserts it into license_keys, and returns
	// the new License with wasNew=true. Returns ErrPoolEmpty if the pool has
	// no keys for productID.
	AssignFromPool(ctx context.Context, userEmail, productID, stripePaymentIntentID string) (lic *License, wasNew bool, err error)

	// FindActive returns the active license for (userEmail, productID), or nil.
	FindActive(ctx context.Context, userEmail, productID string) (*License, error)

	// FindLatestByEmail returns the most recent active license for the email
	// across any product, or nil. Used by the resend-license endpoint.
	FindLatestByEmail(ctx context.Context, userEmail string) (*License, error)

	// RevokeByKey marks the license as revoked. No-op if already revoked.
	RevokeByKey(ctx context.Context, key string) error

	// AddAvailable bulk-inserts pool keys for a product. Returns counts of
	// keys added vs skipped (duplicates of either pool or already-issued
	// keys; blank lines also count as skipped). Used by cmd/import-keys.
	AddAvailable(ctx context.Context, productID string, keys []string) (added, skipped int, err error)

	// CountAvailable returns the number of unsold keys for the product.
	CountAvailable(ctx context.Context, productID string) (int, error)
}

type sqliteLicenseRepo struct {
	db *sql.DB
}

func NewLicenseRepository(db *sql.DB) LicenseRepository {
	return &sqliteLicenseRepo{db: db}
}

func (r *sqliteLicenseRepo) AssignFromPool(ctx context.Context, userEmail, productID, paymentIntentID string) (*License, bool, error) {
	if userEmail == "" || productID == "" {
		return nil, false, errors.New("userEmail and productID are required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // safe on already-committed tx

	// Step 1: idempotency — does this buyer already have an active license
	// for this product? If so, return it unchanged. This is what makes the
	// webhook safe against Stripe retries; we don't burn another pool key.
	if existing, err := scanLicense(tx.QueryRowContext(ctx, queryByUserAndProductActive, userEmail, productID)); err != nil {
		return nil, false, fmt.Errorf("check existing: %w", err)
	} else if existing != nil {
		if err := tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit tx: %w", err)
		}
		return existing, false, nil
	}

	// Step 2: pop one row from the pool (FIFO by id).
	var poolID int64
	var poolKey string
	err = tx.QueryRowContext(ctx, `
		SELECT id, key FROM available_keys
		WHERE product_id = ?
		ORDER BY id ASC
		LIMIT 1
	`, productID).Scan(&poolID, &poolKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, ErrPoolEmpty
	}
	if err != nil {
		return nil, false, fmt.Errorf("pick pool key: %w", err)
	}

	// Step 3: delete from pool — this is the "key disappears from the database"
	// behavior the product owner asked for. Pool count decrements by exactly 1.
	if _, err := tx.ExecContext(ctx, `DELETE FROM available_keys WHERE id = ?`, poolID); err != nil {
		return nil, false, fmt.Errorf("delete from pool: %w", err)
	}

	// Step 4: insert receipt row into license_keys.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO license_keys (key, user_email, product_id, stripe_payment_intent_id)
		VALUES (?, ?, ?, ?)
	`, poolKey, userEmail, productID, nullableString(paymentIntentID))
	if err != nil {
		return nil, false, fmt.Errorf("insert license_keys: %w", err)
	}
	newID, _ := res.LastInsertId()

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit tx: %w", err)
	}

	return &License{
		ID:                    newID,
		Key:                   poolKey,
		UserEmail:             userEmail,
		ProductID:             productID,
		StripePaymentIntentID: paymentIntentID,
		CreatedAt:             time.Now(),
	}, true, nil
}

func (r *sqliteLicenseRepo) FindActive(ctx context.Context, userEmail, productID string) (*License, error) {
	return scanLicense(r.db.QueryRowContext(ctx, queryByUserAndProductActive, userEmail, productID))
}

func (r *sqliteLicenseRepo) FindLatestByEmail(ctx context.Context, userEmail string) (*License, error) {
	return scanLicense(r.db.QueryRowContext(ctx, `
		SELECT id, key, user_email, product_id, COALESCE(stripe_payment_intent_id, ''),
		       created_at, revoked_at
		FROM license_keys
		WHERE user_email = ? AND revoked_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userEmail))
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

func (r *sqliteLicenseRepo) AddAvailable(ctx context.Context, productID string, keys []string) (int, int, error) {
	if productID == "" {
		return 0, 0, errors.New("productID is required")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var added, skipped int
	for _, raw := range keys {
		key := strings.TrimSpace(raw)
		if key == "" {
			skipped++
			continue
		}

		// Refuse to re-insert a key that's already been issued (sold). The
		// duplicate check against available_keys is handled by the UNIQUE
		// constraint; the issued check is explicit because the two tables
		// can't share a constraint.
		var existingIssued int
		if err := tx.QueryRowContext(ctx,
			`SELECT 1 FROM license_keys WHERE key = ?`, key,
		).Scan(&existingIssued); err == nil {
			skipped++
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return added, skipped, fmt.Errorf("check issued: %w", err)
		}

		res, err := tx.ExecContext(ctx, `
			INSERT INTO available_keys (key, product_id) VALUES (?, ?)
			ON CONFLICT(key) DO NOTHING
		`, key, productID)
		if err != nil {
			return added, skipped, fmt.Errorf("insert pool key: %w", err)
		}
		aff, _ := res.RowsAffected()
		if aff > 0 {
			added++
		} else {
			skipped++ // already in pool
		}
	}

	if err := tx.Commit(); err != nil {
		return added, skipped, fmt.Errorf("commit tx: %w", err)
	}
	return added, skipped, nil
}

func (r *sqliteLicenseRepo) CountAvailable(ctx context.Context, productID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM available_keys WHERE product_id = ?`, productID,
	).Scan(&n)
	return n, err
}

// --- helpers ---

const queryByUserAndProductActive = `
	SELECT id, key, user_email, product_id, COALESCE(stripe_payment_intent_id, ''),
	       created_at, revoked_at
	FROM license_keys
	WHERE user_email = ? AND product_id = ? AND revoked_at IS NULL
	LIMIT 1
`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanLicense(row rowScanner) (*License, error) {
	var lic License
	err := row.Scan(
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
