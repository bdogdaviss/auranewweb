package repo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// User is the durable identity row backing the affiliate program. It persists
// the referral_code and store-credit balance that the legacy in-memory user map
// (main.go) loses on every restart.
type User struct {
	ID               string
	Email            string
	Name             string
	ReferralCode     string
	ReferredByCode   string
	StoreCreditCents int64
	PasswordHash     string // bcrypt; empty for Discord-only accounts
	CreatedAt        time.Time
}

// ErrReferralCodeTaken is returned by UpdateReferralCode when the requested code
// already belongs to another user.
var ErrReferralCodeTaken = errors.New("referral code is already taken")

// ErrReferralCodeInvalid is returned when a requested code fails validation
// (not 3-32 chars of [a-z0-9_-] after sanitizing).
var ErrReferralCodeInvalid = errors.New("referral code is invalid")

type UserRepository interface {
	// EnsureUser upserts a durable user row keyed by email with a random code.
	EnsureUser(ctx context.Context, id, email, name string) (*User, error)

	// EnsureUserWithCode is like EnsureUser but, on first insert, prefers the
	// given code (e.g. a Discord username) — falling back to a random code if
	// it's blank, invalid, or already taken. A returning user (matched by email)
	// keeps their existing code untouched.
	EnsureUserWithCode(ctx context.Context, id, email, name, preferredCode string) (*User, error)

	// UpdateReferralCode changes a user's referral code (the editable "Social
	// Name"). Returns ErrReferralCodeTaken if another user already holds it.
	UpdateReferralCode(ctx context.Context, userID, code string) error

	// FindByEmail returns the user with this email, or nil.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByID returns the user with this id, or nil.
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByReferralCode returns the user owning this referral code, or nil.
	FindByReferralCode(ctx context.Context, code string) (*User, error)

	// GetBalance returns the user's store-credit balance in cents.
	GetBalance(ctx context.Context, userID string) (int64, error)

	// SetPassword stores a bcrypt hash for the user (used by email registration).
	SetPassword(ctx context.Context, userID, hash string) error

	// SetReferredBy records which referral code brought this user to the site
	// (the "referred_by_code" column). It only sets the value if the user
	// does not already have one. The code is sanitized; invalid/empty codes
	// are silently ignored.
	SetReferredBy(ctx context.Context, userID, referredByCode string) error
}

type sqliteUserRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &sqliteUserRepo{db: db}
}

const userColumns = `id, email, COALESCE(name, ''), referral_code,
	COALESCE(referred_by_code, ''), store_credit_cents, COALESCE(password_hash, ''), created_at`

func (r *sqliteUserRepo) EnsureUser(ctx context.Context, id, email, name string) (*User, error) {
	return r.EnsureUserWithCode(ctx, id, email, name, "")
}

func (r *sqliteUserRepo) EnsureUserWithCode(ctx context.Context, id, email, name, preferred string) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	// Fast path: already persisted. Return the stored row as-is (existing code
	// and balance untouched).
	if existing, err := r.FindByEmail(ctx, email); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	// Callers normally pass the uuid minted in main.go, but the summary endpoint
	// may reach here for a session whose user was never persisted; mint an id so
	// the primary key is never blank.
	if id == "" {
		genID, err := newReferralCode()
		if err != nil {
			return nil, fmt.Errorf("generate user id: %w", err)
		}
		id = "u_" + genID
	}

	// First insert: try the preferred code (e.g. Discord username) first, then
	// fall back to random codes, retrying on the rare UNIQUE collision.
	for attempt := 0; attempt < 6; attempt++ {
		var code string
		if attempt == 0 {
			code = sanitizeCode(preferred)
		}
		if code == "" {
			rc, err := newReferralCode()
			if err != nil {
				return nil, fmt.Errorf("generate referral code: %w", err)
			}
			code = rc
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO users (id, email, name, referral_code)
			VALUES (?, ?, ?, ?)
		`, id, email, nullableString(name), code)
		if err == nil {
			return r.FindByEmail(ctx, email)
		}
		// A racing login may have inserted the same email; treat as success.
		if existing, ferr := r.FindByEmail(ctx, email); ferr == nil && existing != nil {
			return existing, nil
		}
		// Only retry on a referral_code uniqueness collision; surface anything
		// else (locked DB, disk full, a different constraint) instead of looping.
		if !strings.Contains(err.Error(), "referral_code") {
			return nil, fmt.Errorf("insert user: %w", err)
		}
	}
	return nil, errors.New("could not allocate a unique referral code")
}

func (r *sqliteUserRepo) UpdateReferralCode(ctx context.Context, userID, code string) error {
	c := sanitizeCode(code)
	if c == "" {
		return ErrReferralCodeInvalid
	}
	var owner string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE referral_code = ? COLLATE NOCASE`, c).Scan(&owner)
	if err == nil && owner != userID {
		return ErrReferralCodeTaken
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("check code: %w", err)
	}
	// The UNIQUE constraint is the real guard against a concurrent claim; map a
	// collision there back to ErrReferralCodeTaken rather than leaking the error.
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET referral_code = ? WHERE id = ?`, c, userID); err != nil {
		if strings.Contains(err.Error(), "referral_code") {
			return ErrReferralCodeTaken
		}
		return fmt.Errorf("update code: %w", err)
	}
	return nil
}

// sanitizeCode lowercases and strips a desired referral code to URL-safe
// characters, returning "" if the result isn't a usable 3-32 char code.
func sanitizeCode(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			b.WriteRune(ch)
		}
	}
	out := b.String()
	if len(out) > 32 {
		out = out[:32]
	}
	if len(out) < 3 {
		return ""
	}
	return out
}

func (r *sqliteUserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE email = ? LIMIT 1`, email))
}

func (r *sqliteUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, nil
	}
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = ? LIMIT 1`, id))
}

func (r *sqliteUserRepo) FindByReferralCode(ctx context.Context, code string) (*User, error) {
	if code == "" {
		return nil, nil
	}
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT `+userColumns+` FROM users WHERE referral_code = ? LIMIT 1`, code))
}

func (r *sqliteUserRepo) GetBalance(ctx context.Context, userID string) (int64, error) {
	var cents int64
	err := r.db.QueryRowContext(ctx,
		`SELECT store_credit_cents FROM users WHERE id = ?`, userID).Scan(&cents)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return cents, err
}

func (r *sqliteUserRepo) SetPassword(ctx context.Context, userID, hash string) error {
	if userID == "" || hash == "" {
		return errors.New("userID and hash are required")
	}
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = ? WHERE id = ?`, hash, userID)
	return err
}

func (r *sqliteUserRepo) SetReferredBy(ctx context.Context, userID, referredByCode string) error {
	if userID == "" {
		return nil
	}
	code := sanitizeCode(referredByCode)
	if code == "" {
		return nil // nothing to set
	}

	// Only set if the user doesn't already have a referred_by_code.
	// This keeps the first referrer even if the user later visits other links.
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET referred_by_code = ? WHERE id = ? AND (referred_by_code IS NULL OR referred_by_code = '')`,
		code, userID)
	if err != nil {
		return fmt.Errorf("set referred_by_code: %w", err)
	}
	return nil
}

func scanUser(row rowScanner) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.ReferralCode,
		&u.ReferredByCode, &u.StoreCreditCents, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// referralCodeAlphabet excludes visually ambiguous characters (0/O, 1/I/L) so
// codes are safe to read aloud and type from a share link.
const referralCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// newReferralCode returns an 8-character URL-safe code drawn uniformly from the
// no-confusables alphabet using crypto/rand.
func newReferralCode() (string, error) {
	const n = 8
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = referralCodeAlphabet[int(b)%len(referralCodeAlphabet)]
	}
	return string(out), nil
}
