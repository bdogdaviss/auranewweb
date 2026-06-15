package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aura-optimizer/internal/db"
	"aura-optimizer/internal/repo"
)

func newRecoverDeps(t *testing.T) (ResendByEmailDeps, repo.LicenseRepository, *recordingMailer) {
	t.Helper()
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "r.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	licRepo := repo.NewLicenseRepository(sqlDB)
	mr := &recordingMailer{}
	return ResendByEmailDeps{LicenseRepo: licRepo, Mailer: mr, Cooldown: time.Minute}, licRepo, mr
}

func seedAndAssign(t *testing.T, r repo.LicenseRepository, email string) {
	t.Helper()
	if added, _, err := r.AddAvailable(context.Background(), "lifetime", []string{"AURA-RRRRR-EEEEE-CCCCC"}); err != nil || added != 1 {
		t.Fatalf("seed pool: added=%d err=%v", added, err)
	}
	if _, _, err := r.AssignFromPool(context.Background(), email, "lifetime", "pi_r"); err != nil {
		t.Fatalf("assign: %v", err)
	}
}

func postEmail(t *testing.T, h *ResendByEmailHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/account/resend-by-email", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestResendByEmail_MethodNotAllowed(t *testing.T) {
	deps, _, _ := newRecoverDeps(t)
	h := NewResendByEmailHandler(deps)
	req := httptest.NewRequest(http.MethodGet, "/api/account/resend-by-email", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestResendByEmail_InvalidEmailRejected(t *testing.T) {
	deps, _, mr := newRecoverDeps(t)
	h := NewResendByEmailHandler(deps)
	rec := postEmail(t, h, `{"email":"not-an-email"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if len(mr.calls) != 0 {
		t.Errorf("no mail expected for invalid email, got %v", mr.calls)
	}
}

func TestResendByEmail_NoLicenseIsGenericOK(t *testing.T) {
	// Unknown email must look identical to a known one (no enumeration) and send
	// nothing.
	deps, _, mr := newRecoverDeps(t)
	h := NewResendByEmailHandler(deps)
	rec := postEmail(t, h, `{"email":"ghost@example.com"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}
	if len(mr.calls) != 0 {
		t.Errorf("no mail expected when no license, got %v", mr.calls)
	}
}

func TestResendByEmail_SendsWhenLicenseExists(t *testing.T) {
	deps, licRepo, mr := newRecoverDeps(t)
	seedAndAssign(t, licRepo, "guest@example.com")
	h := NewResendByEmailHandler(deps)
	rec := postEmail(t, h, `{"email":"guest@example.com"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}
	if len(mr.calls) != 1 || mr.calls[0] != "guest@example.com" {
		t.Errorf("mailer calls = %v, want [guest@example.com]", mr.calls)
	}
}

func TestResendByEmail_CooldownBlocksRepeat(t *testing.T) {
	deps, licRepo, mr := newRecoverDeps(t)
	seedAndAssign(t, licRepo, "guest@example.com")
	h := NewResendByEmailHandler(deps)

	_ = postEmail(t, h, `{"email":"guest@example.com"}`)
	// Same address (different case) within the cooldown window must not re-send.
	_ = postEmail(t, h, `{"email":"GUEST@Example.com"}`)
	if len(mr.calls) != 1 {
		t.Errorf("cooldown should block the 2nd send, got %d calls", len(mr.calls))
	}
}
