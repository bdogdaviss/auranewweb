package payment

import (
	"os"
	"testing"

	"github.com/stripe/stripe-go/v78/webhook"
)

func TestNewStripeService_MissingKey(t *testing.T) {
	t.Setenv("STRIPE_SECRET_KEY", "")
	if _, err := NewStripeService(); err == nil {
		t.Fatal("expected error when STRIPE_SECRET_KEY unset")
	}
}

func TestNewStripeService_OK(t *testing.T) {
	t.Setenv("STRIPE_SECRET_KEY", "sk_test_dummy")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_dummy")
	s, err := NewStripeService()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.secretKey != "sk_test_dummy" || s.webhookSecret != "whsec_dummy" {
		t.Fatalf("env not captured: %+v", s)
	}
}

func TestParseWebhook_NoSecret(t *testing.T) {
	s := &StripeService{secretKey: "sk_test_x"}
	if _, err := s.ParseWebhook([]byte("{}"), "t=1,v1=deadbeef"); err == nil {
		t.Fatal("expected error when webhook secret unset")
	}
}

func TestParseWebhook_BadSignature(t *testing.T) {
	s := &StripeService{secretKey: "sk_test_x", webhookSecret: "whsec_real"}
	if _, err := s.ParseWebhook([]byte(`{"id":"evt_1"}`), "t=1,v1=deadbeef"); err == nil {
		t.Fatal("expected signature verification error")
	}
}

func TestParseWebhook_VerifiesValidSignature(t *testing.T) {
	const secret = "whsec_testsecret"
	s := &StripeService{secretKey: "sk_test_x", webhookSecret: secret}

	payload := []byte(`{"id":"evt_1","api_version":"2024-04-10","type":"payment_intent.succeeded","data":{"object":{}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload: payload,
		Secret:  secret,
	})

	evt, err := s.ParseWebhook(signed.Payload, signed.Header)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(evt.Type) != "payment_intent.succeeded" {
		t.Errorf("got event type %q, want payment_intent.succeeded", evt.Type)
	}
}

func TestStripeService_EnvIsolation(t *testing.T) {
	// Sanity check: t.Setenv restores env after test.
	t.Setenv("STRIPE_SECRET_KEY", "sk_xx")
	if got := os.Getenv("STRIPE_SECRET_KEY"); got != "sk_xx" {
		t.Fatalf("got %q", got)
	}
}
