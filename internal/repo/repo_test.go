package repo

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"aura-optimizer/internal/db"
)

func mustOpen(t *testing.T) *sql.DB {
	t.Helper()
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}

func seedPool(t *testing.T, r LicenseRepository, productID string, keys ...string) {
	t.Helper()
	added, skipped, err := r.AddAvailable(context.Background(), productID, keys)
	if err != nil {
		t.Fatalf("seed pool: %v", err)
	}
	if added != len(keys) || skipped != 0 {
		t.Fatalf("seed: added=%d skipped=%d (want added=%d skipped=0)", added, skipped, len(keys))
	}
}

// --- AssignFromPool ---

func TestAssignFromPool_HappyPath(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC", "AURA-DDDDD-EEEEE-FFFFF")

	if n, _ := r.CountAvailable(context.Background(), "lifetime"); n != 2 {
		t.Fatalf("pool should have 2, got %d", n)
	}

	lic, wasNew, err := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_abc")
	if err != nil {
		t.Fatalf("assign: %v", err)
	}
	if !wasNew {
		t.Fatal("expected wasNew=true on first assignment")
	}
	if lic.Key != "AURA-AAAAA-BBBBB-CCCCC" {
		t.Errorf("expected FIFO pool pick AURA-AAAAA..., got %q", lic.Key)
	}
	if lic.UserEmail != "u@example.com" || lic.ProductID != "lifetime" {
		t.Errorf("license fields wrong: %+v", lic)
	}

	if n, _ := r.CountAvailable(context.Background(), "lifetime"); n != 1 {
		t.Errorf("pool should be 1 after assignment, got %d", n)
	}
}

func TestAssignFromPool_IdempotentForSameBuyer(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC", "AURA-DDDDD-EEEEE-FFFFF")

	first, _, err := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_1")
	if err != nil {
		t.Fatalf("first: %v", err)
	}

	// Stripe webhook retry: same buyer/product. Must NOT burn a second pool key.
	second, wasNew, err := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_2")
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if wasNew {
		t.Error("expected wasNew=false on retry")
	}
	if second.Key != first.Key {
		t.Errorf("retry returned different key: first=%q second=%q", first.Key, second.Key)
	}

	if n, _ := r.CountAvailable(context.Background(), "lifetime"); n != 1 {
		t.Errorf("pool should still be 1 (only one assignment happened), got %d", n)
	}
}

func TestAssignFromPool_DifferentBuyersGetDifferentKeys(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC", "AURA-DDDDD-EEEEE-FFFFF")

	a, _, _ := r.AssignFromPool(context.Background(), "alice@example.com", "lifetime", "pi_a")
	b, _, _ := r.AssignFromPool(context.Background(), "bob@example.com", "lifetime", "pi_b")

	if a.Key == b.Key {
		t.Errorf("two buyers got the same key %q (pool double-issued!)", a.Key)
	}
	if n, _ := r.CountAvailable(context.Background(), "lifetime"); n != 0 {
		t.Errorf("pool should be empty after 2 assignments, got %d", n)
	}
}

func TestAssignFromPool_EmptyPoolReturnsErrPoolEmpty(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	// No seeding — pool is empty.

	_, _, err := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_x")
	if !errors.Is(err, ErrPoolEmpty) {
		t.Errorf("expected ErrPoolEmpty, got %v", err)
	}
}

func TestAssignFromPool_PerProductPools(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-LIFE1-LIFE1-LIFE1")
	seedPool(t, r, "team", "AURA-TEAM1-TEAM1-TEAM1")

	lifeLic, _, _ := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_1")
	if lifeLic.Key != "AURA-LIFE1-LIFE1-LIFE1" {
		t.Errorf("got %q from lifetime pool", lifeLic.Key)
	}
	teamLic, _, _ := r.AssignFromPool(context.Background(), "u@example.com", "team", "pi_2")
	if teamLic.Key != "AURA-TEAM1-TEAM1-TEAM1" {
		t.Errorf("got %q from team pool", teamLic.Key)
	}
}

// --- AddAvailable ---

func TestAddAvailable_SkipsDuplicates(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC")

	added, skipped, err := r.AddAvailable(context.Background(), "lifetime", []string{
		"AURA-AAAAA-BBBBB-CCCCC", // dup of pool
		"AURA-DDDDD-EEEEE-FFFFF", // new
		"",                       // blank, skip
		"  AURA-AAAAA-BBBBB-CCCCC  ", // dup with whitespace
	})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if added != 1 {
		t.Errorf("added = %d, want 1", added)
	}
	if skipped != 3 {
		t.Errorf("skipped = %d, want 3 (dup + blank + whitespace-dup)", skipped)
	}
}

func TestAddAvailable_RefusesAlreadyIssuedKey(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC")
	_, _, _ = r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_1")

	// Trying to re-import the same key into the pool should be skipped.
	added, skipped, err := r.AddAvailable(context.Background(), "lifetime",
		[]string{"AURA-AAAAA-BBBBB-CCCCC"})
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if added != 0 || skipped != 1 {
		t.Errorf("added=%d skipped=%d, want added=0 skipped=1", added, skipped)
	}
}

// --- FindActive / FindLatestByEmail / RevokeByKey ---

func TestFindActive(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC")

	if lic, _ := r.FindActive(context.Background(), "nobody@x.com", "lifetime"); lic != nil {
		t.Errorf("expected nil for unknown buyer, got %+v", lic)
	}
	original, _, _ := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_1")
	found, err := r.FindActive(context.Background(), "u@example.com", "lifetime")
	if err != nil || found == nil || found.Key != original.Key {
		t.Errorf("expected key %q, got %+v err=%v", original.Key, found, err)
	}
}

func TestFindLatestByEmail(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC")
	seedPool(t, r, "team", "AURA-DDDDD-EEEEE-FFFFF")

	if lic, _ := r.FindLatestByEmail(context.Background(), "u@example.com"); lic != nil {
		t.Errorf("expected nil for unknown email, got %+v", lic)
	}
	_, _, _ = r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_a")
	_, _, _ = r.AssignFromPool(context.Background(), "u@example.com", "team", "pi_b")

	lic, err := r.FindLatestByEmail(context.Background(), "u@example.com")
	if err != nil || lic == nil {
		t.Fatalf("find latest: %v / %+v", err, lic)
	}
	if lic.UserEmail != "u@example.com" {
		t.Errorf("wrong email: %q", lic.UserEmail)
	}
}

func TestRevokeByKey(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))
	seedPool(t, r, "lifetime", "AURA-AAAAA-BBBBB-CCCCC")
	lic, _, _ := r.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_1")

	if err := r.RevokeByKey(context.Background(), lic.Key); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if active, _ := r.FindActive(context.Background(), "u@example.com", "lifetime"); active != nil {
		t.Errorf("revoked license should not appear active, got %+v", active)
	}
}

// --- event store (unchanged behaviour, kept here for full coverage) ---

func TestEventStore_RoundTrip(t *testing.T) {
	s := NewEventStore(mustOpen(t))
	ctx := context.Background()

	if seen, _ := s.IsProcessed(ctx, "evt_1"); seen {
		t.Fatal("expected unseen")
	}
	if err := s.MarkProcessed(ctx, "evt_1"); err != nil {
		t.Fatalf("mark: %v", err)
	}
	if seen, _ := s.IsProcessed(ctx, "evt_1"); !seen {
		t.Fatal("expected seen after mark")
	}
	if err := s.MarkProcessed(ctx, "evt_1"); err != nil {
		t.Fatalf("second mark (idempotent): %v", err)
	}
}

func TestEventStore_EmptyEventID(t *testing.T) {
	s := NewEventStore(mustOpen(t))
	if seen, _ := s.IsProcessed(context.Background(), ""); seen {
		t.Error("empty event_id should not be seen")
	}
	if err := s.MarkProcessed(context.Background(), ""); err == nil {
		t.Error("MarkProcessed should reject empty event_id")
	}
}
