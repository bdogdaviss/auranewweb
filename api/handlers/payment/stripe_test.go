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

	// Verify license was persisted.
	lic, err := licRepo.FindActive(context.Background(), "buyer@example.com", "lifetime")
	if err != nil || lic == nil {
		t.Fatalf("expected license to be inserted; lic=%+v err=%v", lic, err)
	}
	if lic.StripePaymentIntentID != "pi_full_flow" {
		t.Errorf("payment intent id not recorded: %q", lic.StripePaymentIntentID)
	}

	// Verify email was sent once.
	if got := mr.count(); got != 1 {
		t.Errorf("mailer calls after first delivery = %d, want 1", got)
	}

	// Resend: same event ID. EventStore should dedup before any license/email work.
	if code := deliver(); code != http.StatusOK {
		t.Fatalf("second delivery: status %d, want 200", code)
	}
	if got := mr.count(); got != 1 {
		t.Errorf("mailer calls after duplicate delivery = %d, want 1 (event-store dedup failed)", got)
	}
}

// Ensure the Mailer interface is implemented by our test fake at compile time.
var _ mailer.Mailer = (*recordingMailer)(nil)
