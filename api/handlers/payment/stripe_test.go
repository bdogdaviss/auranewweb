package payment

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stripe/stripe-go/v78/webhook"

	paysvc "aura-optimizer/internal/payment"
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

func TestWebhook_IdempotentOnDuplicateEventID(t *testing.T) {
	const whSecret = "whsec_test_idem"
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_x")
	t.Setenv("STRIPE_WEBHOOK_SECRET", whSecret)

	svc, err := paysvc.NewStripeService()
	if err != nil {
		t.Fatalf("NewStripeService: %v", err)
	}

	var markProCalls int32
	h := NewStripeHandler(Deps{
		Service:     svc,
		MarkUserPro: func(string) { atomic.AddInt32(&markProCalls, 1) },
	})

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
