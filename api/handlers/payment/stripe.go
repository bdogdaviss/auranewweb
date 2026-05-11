package payment

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"sync"

	"github.com/stripe/stripe-go/v78"

	paysvc "aura-optimizer/internal/payment"
)

type Product struct {
	ID    string
	Name  string
	Price float64
}

type Deps struct {
	Service      *paysvc.StripeService
	LookupProduct func(productID string) (Product, bool)
	GetUserID    func(r *http.Request) (id, email string)
	MarkUserPro  func(userID string)
}

type StripeHandler struct {
	deps Deps
	// processedEvents tracks event.IDs already handled, so Stripe's retry-on-5xx
	// behaviour doesn't double-fire MarkUserPro. In-memory: resets on restart,
	// which is acceptable while users/IsPro also live in memory. Swap for a
	// durable store before introducing a real DB or running multiple replicas.
	processedEvents sync.Map
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

	if _, alreadySeen := h.processedEvents.LoadOrStore(event.ID, struct{}{}); alreadySeen {
		log.Printf("stripe webhook: duplicate event %s, skipping", event.ID)
		w.WriteHeader(http.StatusOK)
		return
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
		if userID != "" && h.deps.MarkUserPro != nil {
			h.deps.MarkUserPro(userID)
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
