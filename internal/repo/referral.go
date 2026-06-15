package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ReferralEvent is one earned referral, for the dashboard table.
type ReferralEvent struct {
	BuyerEmail  string
	ProductID   string
	CreditCents int64
	CreatedAt   time.Time
}

// CreditReferralInput carries everything CreditReferral needs to award store
// credit for one qualifying purchase. StripeEventID is the idempotency key.
type CreditReferralInput struct {
	StripeEventID   string
	PaymentIntentID string
	ReferralCode    string
	BuyerEmail      string
	ProductID       string
	CreditCents     int64
}

type ReferralRepository interface {
	// CreditReferral awards CreditCents of store credit to the user owning
	// ReferralCode, exactly once per StripeEventID. It is a single transaction:
	// insert the referral_events row (UNIQUE(stripe_event_id) makes a webhook
	// retry a clean no-op), resolve the referrer, refuse self-referral, bump the
	// balance, and append a credit_ledger row.
	//
	// Returns credited=false (err=nil) for the benign skips — duplicate event,
	// unknown/blank code, or self-referral — so the caller can log and move on
	// without treating them as failures. Returns a non-nil err only on real DB
	// faults; the webhook treats those as non-fatal (the buyer already has their
	// license) and the UNIQUE constraint keeps a later replay safe.
	CreditReferral(ctx context.Context, in CreditReferralInput) (credited bool, err error)

	// CountByReferrer returns how many qualifying referrals a user has earned.
	CountByReferrer(ctx context.Context, userID string) (int, error)

	// TotalEarnedCents returns the lifetime sum of referral credit a user earned.
	TotalEarnedCents(ctx context.Context, userID string) (int64, error)

	// ListByReferrer returns a user's most recent referral earnings (newest first).
	ListByReferrer(ctx context.Context, userID string, limit int) ([]ReferralEvent, error)

	// RedeemCredit debits AmountCents of store credit from the user identified by
	// Email, exactly once per StripeEventID. It is the spending counterpart of
	// CreditReferral: insert the redemptions row (UNIQUE(stripe_event_id) gates
	// retries), decrement the balance (clamped at zero so it can never go
	// negative even in a redemption race), and append a credit_ledger row.
	// Returns redeemed=false (err=nil) for benign skips — duplicate event,
	// unknown email, or non-positive amount.
	RedeemCredit(ctx context.Context, in RedeemInput) (redeemed bool, err error)
}

// RedeemInput carries what RedeemCredit needs to debit one purchase's applied
// store credit. StripeEventID is the idempotency key; Email identifies the
// credit holder (the buyer spending their own credit).
type RedeemInput struct {
	StripeEventID   string
	PaymentIntentID string
	Email           string
	ProductID       string
	AmountCents     int64
}

type sqliteReferralRepo struct {
	db *sql.DB
}

func NewReferralRepository(db *sql.DB) ReferralRepository {
	return &sqliteReferralRepo{db: db}
}

func (r *sqliteReferralRepo) CreditReferral(ctx context.Context, in CreditReferralInput) (bool, error) {
	if in.StripeEventID == "" {
		return false, errors.New("stripe event id is required")
	}
	code := strings.TrimSpace(in.ReferralCode)
	if code == "" || in.CreditCents <= 0 {
		return false, nil // nothing to credit
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // safe on already-committed tx

	// Resolve the referrer first so we can record their id on the ledger row.
	var referrerID, referrerEmail string
	err = tx.QueryRowContext(ctx,
		`SELECT id, email FROM users WHERE referral_code = ? COLLATE NOCASE LIMIT 1`, code,
	).Scan(&referrerID, &referrerEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // stale or invalid code — skip
	}
	if err != nil {
		return false, fmt.Errorf("lookup referrer: %w", err)
	}

	// Self-referral guard: a buyer cannot earn credit on their own purchase.
	if strings.EqualFold(strings.TrimSpace(in.BuyerEmail), strings.TrimSpace(referrerEmail)) {
		return false, nil
	}

	// Idempotency gate: one referral_events row per Stripe event. A retry hits
	// the UNIQUE(stripe_event_id) conflict, we roll back, and no balance moves.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO referral_events
			(stripe_event_id, payment_intent_id, referrer_user_id, referrer_code,
			 buyer_email, product_id, credit_cents)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(stripe_event_id) DO NOTHING
	`, in.StripeEventID, nullableString(in.PaymentIntentID), referrerID, code,
		in.BuyerEmail, in.ProductID, in.CreditCents)
	if err != nil {
		return false, fmt.Errorf("insert referral_event: %w", err)
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return false, nil // already credited for this event
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET store_credit_cents = store_credit_cents + ? WHERE id = ?
	`, in.CreditCents, referrerID); err != nil {
		return false, fmt.Errorf("bump balance: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO credit_ledger (user_id, delta_cents, reason, ref_event_id)
		VALUES (?, ?, 'referral_earn', ?)
	`, referrerID, in.CreditCents, in.StripeEventID); err != nil {
		return false, fmt.Errorf("insert credit_ledger: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}
	return true, nil
}

func (r *sqliteReferralRepo) CountByReferrer(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM referral_events WHERE referrer_user_id = ?`, userID,
	).Scan(&n)
	return n, err
}

func (r *sqliteReferralRepo) TotalEarnedCents(ctx context.Context, userID string) (int64, error) {
	var n sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(credit_cents), 0) FROM referral_events WHERE referrer_user_id = ?`, userID,
	).Scan(&n)
	return n.Int64, err
}

func (r *sqliteReferralRepo) ListByReferrer(ctx context.Context, userID string, limit int) ([]ReferralEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT buyer_email, product_id, credit_cents, created_at
		FROM referral_events
		WHERE referrer_user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ReferralEvent
	for rows.Next() {
		var e ReferralEvent
		if err := rows.Scan(&e.BuyerEmail, &e.ProductID, &e.CreditCents, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *sqliteReferralRepo) RedeemCredit(ctx context.Context, in RedeemInput) (bool, error) {
	if in.StripeEventID == "" {
		return false, errors.New("stripe event id is required")
	}
	email := strings.TrimSpace(in.Email)
	if email == "" || in.AmountCents <= 0 {
		return false, nil // nothing to redeem
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // safe on already-committed tx

	var userID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil // no such user — nothing to debit
	}
	if err != nil {
		return false, fmt.Errorf("lookup user: %w", err)
	}

	// Idempotency gate: one redemption row per Stripe event.
	res, err := tx.ExecContext(ctx, `
		INSERT INTO redemptions
			(stripe_event_id, payment_intent_id, user_id, product_id, amount_cents)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(stripe_event_id) DO NOTHING
	`, in.StripeEventID, nullableString(in.PaymentIntentID), userID, in.ProductID, in.AmountCents)
	if err != nil {
		return false, fmt.Errorf("insert redemption: %w", err)
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return false, nil // already redeemed for this event
	}

	// Clamp at zero: in the (rare, single-user) case where a concurrent redeem
	// already drew the balance down, we debit what's left rather than going
	// negative. The applied amount was computed from the balance at PI-create,
	// so the common case debits exactly AmountCents.
	if _, err := tx.ExecContext(ctx, `
		UPDATE users SET store_credit_cents = MAX(0, store_credit_cents - ?) WHERE id = ?
	`, in.AmountCents, userID); err != nil {
		return false, fmt.Errorf("debit balance: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO credit_ledger (user_id, delta_cents, reason, ref_event_id)
		VALUES (?, ?, 'redeem', ?)
	`, userID, -in.AmountCents, in.StripeEventID); err != nil {
		return false, fmt.Errorf("insert credit_ledger: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}
	return true, nil
}
