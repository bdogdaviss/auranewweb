package account

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"aura-optimizer/internal/db"
	"aura-optimizer/internal/repo"
)

// A user with zero referrals must serialize `"referrals": []`, never `null`.
// The frontend does `data.referrals.length` (AccountPage.tsx); a null there
// throws during render and — with no error boundary — unmounts the whole app,
// leaving the affiliate page blank/black on login. Regression test for that.
func TestSummary_ZeroReferralsSerializesAsEmptyArray(t *testing.T) {
	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "s.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	deps := SummaryDeps{
		GetUserEmail: func(*http.Request) string { return "newbie@example.com" },
		UserRepo:     repo.NewUserRepository(sqlDB),
		ReferralRepo: repo.NewReferralRepository(sqlDB),
	}
	h := NewSummaryHandler(deps)

	req := httptest.NewRequest(http.MethodGet, "/api/account/summary", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode body: %v; body: %s", err, rec.Body.String())
	}
	got, ok := raw["referrals"]
	if !ok {
		t.Fatalf("response missing `referrals` field; body: %s", rec.Body.String())
	}
	if string(got) != "[]" {
		t.Errorf("referrals = %s, want `[]` (null crashes the affiliate page)", got)
	}
}
