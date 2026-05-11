package account

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"aura-optimizer/internal/db"
	"aura-optimizer/internal/repo"
)

type recordingMailer struct {
	mu    sync.Mutex
	calls []string
}

func (m *recordingMailer) SendLicense(_ context.Context, to, _, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, to)
	return nil
}

func newDeps(t *testing.T, email string) (ResendLicenseDeps, repo.LicenseRepository, *recordingMailer) {
	t.Helper()
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "t.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	licRepo := repo.NewLicenseRepository(sqlDB)
	mr := &recordingMailer{}
	deps := ResendLicenseDeps{
		GetUserEmail: func(*http.Request) string { return email },
		LicenseRepo:  licRepo,
		Mailer:       mr,
	}
	return deps, licRepo, mr
}

func post(t *testing.T, h *ResendLicenseHandler) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/account/resend-license", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestResendLicense_MethodNotAllowed(t *testing.T) {
	deps, _, _ := newDeps(t, "u@example.com")
	h := NewResendLicenseHandler(deps)
	req := httptest.NewRequest(http.MethodGet, "/api/account/resend-license", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestResendLicense_Unauthenticated(t *testing.T) {
	deps, _, _ := newDeps(t, "") // GetUserEmail returns ""
	h := NewResendLicenseHandler(deps)
	rec := post(t, h)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestResendLicense_NoLicense(t *testing.T) {
	deps, _, mr := newDeps(t, "u@example.com")
	h := NewResendLicenseHandler(deps)
	rec := post(t, h)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	if len(mr.calls) != 0 {
		t.Errorf("mailer should not be called when no license exists, got %d calls", len(mr.calls))
	}
}

func TestResendLicense_SendsLicense(t *testing.T) {
	deps, licRepo, mr := newDeps(t, "u@example.com")
	// Seed: put a key in the pool, assign it to the user — same path the real
	// webhook handler takes when a purchase succeeds.
	if added, _, err := licRepo.AddAvailable(context.Background(), "lifetime",
		[]string{"AURA-AAAAA-BBBBB-CCCCC"}); err != nil || added != 1 {
		t.Fatalf("seed pool: added=%d err=%v", added, err)
	}
	if _, _, err := licRepo.AssignFromPool(context.Background(), "u@example.com", "lifetime", "pi_x"); err != nil {
		t.Fatalf("assign: %v", err)
	}
	h := NewResendLicenseHandler(deps)
	rec := post(t, h)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}
	if len(mr.calls) != 1 || mr.calls[0] != "u@example.com" {
		t.Errorf("mailer calls = %v, want [u@example.com]", mr.calls)
	}

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["status"] != "sent" || resp["to"] != "u@example.com" || resp["product"] != "lifetime" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestResendLicense_IgnoresAttackerSuppliedUserID(t *testing.T) {
	// Even if the request body claims a different user, GetUserEmail (which
	// reads the session cookie server-side) is the only source of identity.
	deps, licRepo, mr := newDeps(t, "victim@example.com")
	// Seed a license for the victim — attacker has no way to read it via this endpoint.
	if added, _, err := licRepo.AddAvailable(context.Background(), "lifetime",
		[]string{"AURA-VVVVV-IIIII-CCCCC"}); err != nil || added != 1 {
		t.Fatalf("seed pool: added=%d err=%v", added, err)
	}
	if _, _, err := licRepo.AssignFromPool(context.Background(), "victim@example.com", "lifetime", "pi_v"); err != nil {
		t.Fatalf("seed assignment: %v", err)
	}
	h := NewResendLicenseHandler(deps)

	// Attacker POSTs a body with a different identifier; handler must ignore it.
	req := httptest.NewRequest(http.MethodPost, "/api/account/resend-license", strings.NewReader(`{"user_id":"attacker@evil.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(mr.calls) != 1 || mr.calls[0] != "victim@example.com" {
		t.Errorf("expected mailer to use session-identified user, got calls=%v", mr.calls)
	}
}
