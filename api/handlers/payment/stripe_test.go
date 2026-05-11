package payment

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
