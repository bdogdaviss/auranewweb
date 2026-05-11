package payment

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stripe/stripe-go/v78/webhook"

	"aura-optimizer/internal/db"
	"aura-optimizer/internal/mailer"
	paysvc "aura-optimizer/internal/payment"
	"aura-optimizer/internal/repo"
)

func newTestHandler() *StripeHandler {
	return NewStripeHandler(Deps{
		LookupProduct: func(id string) (Product, bool) {
			if id == "pro" {
				return Product{ID: "pro", Name: "Pro Tier", Price: 19.99}, true
			}
			return Product{}, false
		},
	})
}

func TestCreatePaymentIntent_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/stripe/create-payment-intent", nil)
	h.CreatePaymentIntent(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestCreatePaymentIntent_BadJSON(t *testing.T) {
	h := newTestHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/stripe/create-payment-intent", strings.NewReader("not json"))
	h.CreatePaymentIntent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestCreatePaymentIntent_UnknownProduct(t *testing.T) {
	h := newTestHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/stripe/create-payment-intent", strings.NewReader(`{"product_id":"nope"}`))
	h.CreatePaymentIntent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestWebhook_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/stripe/webhook", nil)
	h.Webhook(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

// --- in-memory fakes for unit tests ---

type fakeEventStore struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newFakeEventStore() *fakeEventStore { return &fakeEventStore{seen: map[string]bool{}} }

func (f *fakeEventStore) IsProcessed(_ context.Context, eventID string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.seen[eventID], nil
}

func (f *fakeEventStore) MarkProcessed(_ context.Context, eventID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seen[eventID] = true
	return nil
}

type recordingMailer struct {
	mu    sync.Mutex
	calls []mailerCall
}

type mailerCall struct{ To, Key, ProductID string }

func (m *recordingMailer) SendLicense(_ context.Context, to, key, productID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, mailerCall{to, key, productID})
	return nil
}

func (m *recordingMailer) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func newStripeServiceForTest(t *testing.T, secret string) *paysvc.StripeService {
	t.Helper()
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_x")
	t.Setenv("STRIPE_WEBHOOK_SECRET", secret)
	svc, err := paysvc.NewStripeService()
	if err != nil {
		t.Fatalf("NewStripeService: %v", err)
	}
	return svc
}

func TestWebhook_IdempotentOnDuplicateEventID_SyntheticEvent(t *testing.T) {
	const whSecret = "whsec_test_idem"
	svc := newStripeServiceForTest(t, whSecret)

	var markProCalls int32
	h := NewStripeHandler(Deps{
		Service:     svc,
		MarkUserPro: func(string) { atomic.AddInt32(&markProCalls, 1) },
		EventStore:  newFakeEventStore(),
	})

	// Synthetic event from Stripe CLI: has user_id but no user_email/product_id.
	// Should fire MarkUserPro on first delivery, then be deduped on retry.
	payload := []byte(`{"id":"evt_dup_test","api_version":"2024-04-10","type":"payment_intent.succeeded","data":{"object":{"metadata":{"user_id":"u1"}}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payload,
		Secret:  whSecret,
	})

	deliver := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", bytes.NewReader(signed.Payload))
		req.Header.Set("Stripe-Signature", signed.Header)
		rec := httptest.NewRecorder()
		h.Webhook(rec, req)
		return rec.Code
	}

	if code := deliver(); code != http.StatusOK {
		t.Fatalf("first delivery: status %d, want 200", code)
	}
	if code := deliver(); code != http.StatusOK {
		t.Fatalf("second delivery: status %d, want 200", code)
	}

	if got := atomic.LoadInt32(&markProCalls); got != 1 {
		t.Errorf("MarkUserPro calls = %d, want 1 (duplicate event should be skipped)", got)
	}
}

func TestWebhook_FullFlow_InsertsLicenseAndEmails(t *testing.T) {
	const whSecret = "whsec_full_flow"
	svc := newStripeServiceForTest(t, whSecret)

	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	licRepo := repo.NewLicenseRepository(sqlDB)
	eventStore := repo.NewEventStore(sqlDB)
	mr := &recordingMailer{}

	// Seed the pool with one key so the assignment can succeed.
	if added, _, err := licRepo.AddAvailable(context.Background(), "lifetime",
		[]string{"AURA-TEST1-TEST1-TEST1"}); err != nil || added != 1 {
		t.Fatalf("seed pool: added=%d err=%v", added, err)
	}

	h := NewStripeHandler(Deps{
		Service:     svc,
		EventStore:  eventStore,
		LicenseRepo: licRepo,
		Mailer:      mr,
	})

	payload := []byte(`{"id":"evt_full_flow","api_version":"2024-04-10","type":"payment_intent.succeeded","data":{"object":{"id":"pi_full_flow","metadata":{"user_id":"u1","user_email":"buyer@example.com","product_id":"lifetime"}}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payload,
		Secret:  whSecret,
	})

	deliver := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", bytes.NewReader(signed.Payload))
		req.Header.Set("Stripe-Signature", signed.Header)
		rec := httptest.NewRecorder()
		h.Webhook(rec, req)
		return rec.Code
	}

	if code := deliver(); code != http.StatusOK {
		t.Fatalf("first delivery: status %d, want 200", code)
	}

	// Verify the pool-issued key landed in license_keys with the right fields.
	lic, err := licRepo.FindActive(context.Background(), "buyer@example.com", "lifetime")
	if err != nil || lic == nil {
		t.Fatalf("expected license to be issued; lic=%+v err=%v", lic, err)
	}
	if lic.Key != "AURA-TEST1-TEST1-TEST1" {
		t.Errorf("issued key = %q, want AURA-TEST1-TEST1-TEST1 (pool FIFO)", lic.Key)
	}
	if lic.StripePaymentIntentID != "pi_full_flow" {
		t.Errorf("payment intent id not recorded: %q", lic.StripePaymentIntentID)
	}

	// Pool count must drop to 0 (the sold key was deleted from available_keys).
	if n, _ := licRepo.CountAvailable(context.Background(), "lifetime"); n != 0 {
		t.Errorf("pool count after sale = %d, want 0", n)
	}

	if got := mr.count(); got != 1 {
		t.Errorf("mailer calls after first delivery = %d, want 1", got)
	}

	// Resend the same event: event-store dedup short-circuits, no new license,
	// no second email, pool still at 0.
	if code := deliver(); code != http.StatusOK {
		t.Fatalf("second delivery: status %d, want 200", code)
	}
	if got := mr.count(); got != 1 {
		t.Errorf("mailer calls after duplicate delivery = %d, want 1 (event-store dedup failed)", got)
	}
}

func TestWebhook_PoolEmpty_Returns5xx(t *testing.T) {
	const whSecret = "whsec_pool_empty"
	svc := newStripeServiceForTest(t, whSecret)

	sqlDB, err := db.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	licRepo := repo.NewLicenseRepository(sqlDB)
	eventStore := repo.NewEventStore(sqlDB)
	mr := &recordingMailer{}

	// Note: NO seedPool — pool is empty for "lifetime".

	h := NewStripeHandler(Deps{
		Service:     svc,
		EventStore:  eventStore,
		LicenseRepo: licRepo,
		Mailer:      mr,
	})

	payload := []byte(`{"id":"evt_empty_pool","api_version":"2024-04-10","type":"payment_intent.succeeded","data":{"object":{"id":"pi_empty","metadata":{"user_email":"buyer@example.com","product_id":"lifetime"}}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payload,
		Secret:  whSecret,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", bytes.NewReader(signed.Payload))
	req.Header.Set("Stripe-Signature", signed.Header)
	rec := httptest.NewRecorder()
	h.Webhook(rec, req)

	// Empty pool MUST result in a 5xx so Stripe retries (giving an admin
	// time to refill). Returning 200 would tell Stripe the payment is
	// handled and lose the retry.
	if rec.Code < 500 || rec.Code >= 600 {
		t.Errorf("empty-pool status = %d, want 5xx so Stripe retries", rec.Code)
	}
	if got := mr.count(); got != 0 {
		t.Errorf("mailer should not be called when pool is empty, got %d calls", got)
	}
}

// Ensure the Mailer interface is implemented by our test fake at compile time.
var _ mailer.Mailer = (*recordingMailer)(nil)
