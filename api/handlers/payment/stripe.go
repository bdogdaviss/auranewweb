package payment

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"

	"github.com/stripe/stripe-go/v78"

	"aura-optimizer/internal/mailer"
	paysvc "aura-optimizer/internal/payment"
	"aura-optimizer/internal/repo"
)

type Product struct {
	ID    string
	Name  string
	Price float64
}

type Deps struct {
	Service       *paysvc.StripeService
	LookupProduct func(productID string) (Product, bool)
	GetUserID     func(r *http.Request) (id, email string)
	MarkUserPro   func(userID string)

	// Durable layer (nil-safe; handler degrades gracefully if absent).
	LicenseRepo repo.LicenseRepository
	EventStore  repo.EventStore
	Mailer      mailer.Mailer
}

type StripeHandler struct {
	deps Deps
}

func NewStripeHandler(deps Deps) *StripeHandler {
	return &StripeHandler{deps: deps}
}

type createIntentRequest struct {
	ProductID string `json:"product_id"`
}

func (h *StripeHandler) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	product, ok := h.deps.LookupProduct(req.ProductID)
	if !ok || product.Price <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid product")
		return
	}

	amountCents := int64(math.Round(product.Price * 100))

	var userID, email string
	if h.deps.GetUserID != nil {
		userID, email = h.deps.GetUserID(r)
	}

	result, err := h.deps.Service.CreatePaymentIntent(paysvc.CreateIntentInput{
		AmountCents: amountCents,
		Currency:    "usd",
		ProductID:   product.ID,
		ProductName: product.Name,
		UserID:      userID,
		UserEmail:   email,
	})
	if err != nil {
		log.Printf("stripe create intent: %v", err)
		writeJSONError(w, http.StatusBadGateway, "failed to create payment intent")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Webhook is the canonical idempotent flow for payment_intent.succeeded:
//  1. Verify Stripe signature.
//  2. Check the durable processed_events table — if seen, ACK fast.
//  3. Insert the license row (UNIQUE(user_email, product_id) makes this safe).
//  4. ALWAYS attempt the email; processed_events (not wasNew) gates duplicate sends.
//  5. On email success, record the event in processed_events.
//  6. On any persistent failure, return 5xx so Stripe retries. Retries are safe:
//     the license row is already there, the email re-attempts, processed_events
//     is only written on full success.
func (h *StripeHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "read body failed")
		return
	}

	event, err := h.deps.Service.ParseWebhook(payload, r.Header.Get("Stripe-Signature"))
	if err != nil {
		log.Printf("stripe webhook verify: %v", err)
		writeJSONError(w, http.StatusBadRequest, "signature verification failed")
		return
	}

	ctx := r.Context()

	if h.deps.EventStore != nil {
		seen, err := h.deps.EventStore.IsProcessed(ctx, event.ID)
		if err != nil {
			log.Printf("stripe webhook: event-store check: %v", err)
			// Don't 500 — InsertIfNew + the per-email-attempt logic remain
			// idempotent without this fast-path.
		} else if seen {
			log.Printf("stripe webhook: duplicate event %s, skipping", event.ID)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			log.Printf("stripe webhook unmarshal: %v", err)
			writeJSONError(w, http.StatusBadRequest, "bad payload")
			return
		}

		userID := pi.Metadata["user_id"]
		userEmail := pi.Metadata["user_email"]
		productID := pi.Metadata["product_id"]

		// Flip the in-memory IsPro flag (UI hint) regardless of license work.
		if userID != "" && h.deps.MarkUserPro != nil {
			h.deps.MarkUserPro(userID)
		}

		// Synthetic test events (stripe trigger / CLI) won't have our metadata.
		// Nothing to credit; mark processed so retries don't keep tripping in.
		if userEmail == "" || productID == "" {
			log.Printf("stripe webhook: %s event=%s pi=%s missing metadata, skipping license issue", event.Type, event.ID, pi.ID)
			if h.deps.EventStore != nil {
				_ = h.deps.EventStore.MarkProcessed(ctx, event.ID)
			}
			break
		}

		if h.deps.LicenseRepo != nil {
			license, wasNew, err := h.deps.LicenseRepo.AssignFromPool(ctx, userEmail, productID, pi.ID)
			if errors.Is(err, repo.ErrPoolEmpty) {
				// Out of inventory. Return 5xx so Stripe retries — by the
				// time it retries, an admin should have run the import-keys
				// CLI to refill the pool and the retry will succeed. The
				// customer paid and is waiting; this needs a monitoring
				// alert in any non-trivial deployment.
				log.Printf("stripe webhook: POOL EMPTY for product=%s (event=%s pi=%s buyer=%s) — Stripe will retry; REFILL THE POOL", productID, event.ID, pi.ID, userEmail)
				writeJSONError(w, http.StatusServiceUnavailable, "license inventory empty")
				return
			}
			if err != nil {
				log.Printf("stripe webhook: assign from pool (event=%s): %v", event.ID, err)
				writeJSONError(w, http.StatusInternalServerError, "license assignment failed")
				return
			}
			if wasNew {
				log.Printf("stripe webhook: new license %s for %s/%s (event=%s)", license.Key, userEmail, productID, event.ID)
			} else {
				log.Printf("stripe webhook: reusing existing license %s for %s/%s on retry (event=%s)", license.Key, userEmail, productID, event.ID)
			}

			// Always attempt the email; processed_events at the top is what
			// dedups duplicate sends. wasNew is logged for telemetry only.
			if h.deps.Mailer != nil {
				if err := h.deps.Mailer.SendLicense(ctx, userEmail, license.Key, productID); err != nil {
					log.Printf("stripe webhook: send email (event=%s): %v", event.ID, err)
					writeJSONError(w, http.StatusInternalServerError, "email send failed")
					return
				}
			}
		}

		if h.deps.EventStore != nil {
			if err := h.deps.EventStore.MarkProcessed(ctx, event.ID); err != nil {
				// Best-effort. On rare local-SQLite write failures this could
				// result in a single duplicate email if Stripe retries.
				log.Printf("stripe webhook: mark processed (event=%s): %v", event.ID, err)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
