package repo

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"aura-optimizer/internal/db"
)

func mustOpen(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.sqlite")
	sqlDB, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return sqlDB
}

// --- license repo ---

func TestLicenseRepo_InsertIfNew_FirstInsertCreatesRow(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))

	lic, wasNew, err := r.InsertIfNew(context.Background(), "u@example.com", "lifetime", "pi_abc")
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if !wasNew {
		t.Fatal("wasNew=false on first insert")
	}
	if lic.Key == "" || lic.UserEmail != "u@example.com" || lic.ProductID != "lifetime" {
		t.Fatalf("unexpected license: %+v", lic)
	}
	if lic.StripePaymentIntentID != "pi_abc" {
		t.Errorf("payment intent id not stored: %q", lic.StripePaymentIntentID)
	}
}

func TestLicenseRepo_InsertIfNew_DuplicateReturnsExisting(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))

	first, _, err := r.InsertIfNew(context.Background(), "u@example.com", "lifetime", "pi_1")
	if err != nil {
		t.Fatalf("first insert: %v", err)
	}
	second, wasNew, err := r.InsertIfNew(context.Background(), "u@example.com", "lifetime", "pi_2")
	if err != nil {
		t.Fatalf("second insert: %v", err)
	}
	if wasNew {
		t.Fatal("wasNew=true on duplicate")
	}
	if second.Key != first.Key {
		t.Errorf("duplicate insert returned different key: %q vs %q", first.Key, second.Key)
	}
	if second.StripePaymentIntentID != "pi_1" {
		t.Errorf("payment intent id changed on duplicate insert: %q", second.StripePaymentIntentID)
	}
}

func TestLicenseRepo_InsertIfNew_RejectsEmptyArgs(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))

	if _, _, err := r.InsertIfNew(context.Background(), "", "lifetime", "pi"); err == nil {
		t.Error("expected error for empty userEmail")
	}
	if _, _, err := r.InsertIfNew(context.Background(), "u@x", "", "pi"); err == nil {
		t.Error("expected error for empty productID")
	}
}

func TestLicenseRepo_FindActive(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))

	if lic, err := r.FindActive(context.Background(), "nobody@x.com", "lifetime"); err != nil || lic != nil {
		t.Errorf("expected nil license for unknown email, got %v err=%v", lic, err)
	}

	original, _, _ := r.InsertIfNew(context.Background(), "u@example.com", "lifetime", "pi_a")
	found, err := r.FindActive(context.Background(), "u@example.com", "lifetime")
	if err != nil {
		t.Fatalf("find active: %v", err)
	}
	if found == nil || found.Key != original.Key {
		t.Errorf("expected to find license %q, got %+v", original.Key, found)
	}
}

func TestLicenseRepo_FindLatestByEmail(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))

	if lic, _ := r.FindLatestByEmail(context.Background(), "u@example.com"); lic != nil {
		t.Errorf("expected nil for unknown email, got %+v", lic)
	}

	_, _, _ = r.InsertIfNew(context.Background(), "u@example.com", "lifetime", "pi_a")
	_, _, _ = r.InsertIfNew(context.Background(), "u@example.com", "team", "pi_b")

	lic, err := r.FindLatestByEmail(context.Background(), "u@example.com")
	if err != nil {
		t.Fatalf("find latest: %v", err)
	}
	if lic == nil {
		t.Fatal("expected a license")
	}
	if lic.UserEmail != "u@example.com" {
		t.Errorf("wrong email: %q", lic.UserEmail)
	}
}

func TestLicenseRepo_RevokeByKey(t *testing.T) {
	r := NewLicenseRepository(mustOpen(t))

	lic, _, _ := r.InsertIfNew(context.Background(), "u@example.com", "lifetime", "pi_a")
	if err := r.RevokeByKey(context.Background(), lic.Key); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	active, _ := r.FindActive(context.Background(), "u@example.com", "lifetime")
	if active != nil {
		t.Errorf("revoked license should not appear in FindActive, got %+v", active)
	}
}

// --- event store ---

func TestEventStore_RoundTrip(t *testing.T) {
	s := NewEventStore(mustOpen(t))
	ctx := context.Background()

	seen, err := s.IsProcessed(ctx, "evt_1")
	if err != nil || seen {
		t.Fatalf("expected unseen, got seen=%v err=%v", seen, err)
	}

	if err := s.MarkProcessed(ctx, "evt_1"); err != nil {
		t.Fatalf("mark: %v", err)
	}

	seen, err = s.IsProcessed(ctx, "evt_1")
	if err != nil || !seen {
		t.Fatalf("expected seen, got seen=%v err=%v", seen, err)
	}

	// Double-mark is idempotent.
	if err := s.MarkProcessed(ctx, "evt_1"); err != nil {
		t.Fatalf("second mark: %v", err)
	}
}

func TestEventStore_EmptyEventID(t *testing.T) {
	s := NewEventStore(mustOpen(t))

	if seen, _ := s.IsProcessed(context.Background(), ""); seen {
		t.Error("empty event_id should not be marked seen")
	}
	if err := s.MarkProcessed(context.Background(), ""); err == nil {
		t.Error("MarkProcessed should reject empty event_id")
	}
}
