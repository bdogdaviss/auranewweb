package payment

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeService struct {
	secretKey     string
	webhookSecret string
}

func NewStripeService() (*StripeService, error) {
	key := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if key == "" {
		return nil, errors.New("STRIPE_SECRET_KEY is not set")
	}
	stripe.Key = key
	return &StripeService{
		secretKey:     key,
		webhookSecret: strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET")),
	}, nil
}

type CreateIntentInput struct {
	AmountCents int64
	Currency    string
	ProductID   string
	ProductName string
	UserID      string
	UserEmail   string
}

type CreateIntentResult struct {
	ID           string `json:"id"`
	ClientSecret string `json:"client_secret"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
}

func (s *StripeService) CreatePaymentIntent(in CreateIntentInput) (*CreateIntentResult, error) {
	if in.AmountCents <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if in.Currency == "" {
		in.Currency = "usd"
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(in.AmountCents),
		Currency: stripe.String(strings.ToLower(in.Currency)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
		Description: stripe.String(in.ProductName),
	}
	if in.UserEmail != "" {
		params.ReceiptEmail = stripe.String(in.UserEmail)
	}
	params.AddMetadata("product_id", in.ProductID)
	if in.UserID != "" {
		params.AddMetadata("user_id", in.UserID)
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, err
	}
	return &CreateIntentResult{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Status:       string(pi.Status),
	}, nil
}

func (s *StripeService) ParseWebhook(payload []byte, signatureHeader string) (stripe.Event, error) {
	if s.webhookSecret == "" {
		return stripe.Event{}, errors.New("STRIPE_WEBHOOK_SECRET is not set")
	}
	// IgnoreAPIVersionMismatch: the Stripe account's webhook/CLI may use a newer
	// API version than stripe-go v78 pins (e.g. 2026-04-22.dahlia vs 2024-04-10).
	// We only consume payment_intent.succeeded.metadata, which is schema-stable;
	// strict matching would force a re-pin on every Stripe API release.
	return webhook.ConstructEventWithOptions(payload, signatureHeader, s.webhookSecret,
		webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true})
}
