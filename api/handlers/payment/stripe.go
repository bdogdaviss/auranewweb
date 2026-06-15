package payment

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"

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
	LicenseRepo  repo.LicenseRepository
	EventStore   repo.EventStore
	Mailer       mailer.Mailer
	ReferralRepo repo.ReferralRepository
	UserRepo     repo.UserRepository
}

// ReferralCreditCents is the flat store credit awarded to a referrer for each
// qualifying purchase made through their link.
const ReferralCreditCents = 599

// referralQualifyingProduct is the only product whose sale earns referral
// credit (the "optimizer"). Other products do not pay out.
const referralQualifyingProduct = "lifetime"

// stripeMinChargeCents is Stripe's minimum charge ($0.50). When applying store
// credit we leave at least this much to charge, so a redemption can't drive the
// total below what Stripe will accept.
const stripeMinChargeCents = 50

type StripeHandler struct {
	deps Deps
}

func NewStripeHandler(deps Deps) *StripeHandler {
	return &StripeHandler{deps: deps}
}

type createIntentRequest struct {
	ProductID string `json:"product_id"`
	// GuestEmail is used when the buyer is not logged in (guest checkout). It's
	// where the license is delivered and the identity referral self-referral
	// checks compare against. Ignored when a session is present.
	GuestEmail string `json:"guest_email"`
	// GuestEmailConfirm is sent from the client for guest checkout and must
	// match GuestEmail (after normalization). This is enforced on the server
	// (in addition to client-side check) to reduce typos and obvious abuse.
	GuestEmailConfirm string `json:"guest_email_confirm"`
	// ApplyCredit asks to spend the logged-in user's store credit on this
	// purchase. The amount is computed server-side from the real balance — the
	// client never dictates the discount.
	ApplyCredit bool `json:"apply_credit"`
}

// normalizeEmail trims, validates, and lowercases an email address, returning
// "" if it isn't a valid address.
func normalizeEmail(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return ""
	}
	return strings.ToLower(addr.Address)
}

// isDisposableEmail returns true for common throwaway/temporary email domains.
// Add more as you encounter abuse. This is a simple first line of defense.
func isDisposableEmail(email string) bool {
	disposables := []string{
		"mailinator.com", "tempmail.com", "10minutemail.com", "guerrillamail.com",
		"yopmail.com", "maildrop.cc", "throwawaymail.com", "mailtemp.info",
		"fakeinbox.com", "inbox.lv", "trashmail.com", "mailcatch.com",
		"maildrop.cc", "mintemail.com", "mailinator.net",
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := strings.ToLower(strings.TrimSpace(parts[1]))
	for _, d := range disposables {
		if domain == d || strings.HasSuffix(domain, "."+d) {
			return true
		}
	}
	return false
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
	loggedIn := email != ""

	// Guest checkout: with no logged-in session, fall back to the email entered
	// on the form. An email is required either way — the webhook can't issue a
	// license (or credit a referrer) without one.
	if email == "" {
		email = normalizeEmail(req.GuestEmail)
	}
	if email == "" {
		writeJSONError(w, http.StatusBadRequest, "a valid email is required")
		return
	}

	// Server-side enforcement of email confirmation for guest checkout.
	// This prevents obvious typos and makes it slightly harder to abuse.
	if !loggedIn {
		confirm := normalizeEmail(req.GuestEmailConfirm)
		if confirm == "" || confirm != email {
			writeJSONError(w, http.StatusBadRequest, "email addresses do not match")
			return
		}
	}

	// Block common disposable / temporary email domains.
	// This is not foolproof but catches a lot of throwaway addresses used for abuse.
	if isDisposableEmail(email) {
		writeJSONError(w, http.StatusBadRequest, "disposable email addresses are not allowed")
		return
	}

	// Attribution: the referral code is captured into a first-party "ref" cookie
	// when a visitor lands via someone's share link. Reading it server-side keeps
	// the amount/identity authoritative — the client never sends a code we trust.
	// Lowercase to match stored codes (always lowercased by sanitizeCode); a
	// share link like /?ref=BdogBTW would otherwise never match and the referrer
	// would silently go uncredited.
	var referralCode string
	if c, err := r.Cookie("ref"); err == nil {
		referralCode = strings.ToLower(strings.TrimSpace(c.Value))
	}

	// Store-credit redemption (logged-in users only). The discount is computed
	// here from the real balance and applied to the charged amount; the balance
	// itself is debited in the webhook once the payment settles.
	var appliedCents int64
	if req.ApplyCredit && loggedIn && h.deps.UserRepo != nil {
		appliedCents = h.computeAppliedCredit(r.Context(), email, product.ID, amountCents)
	}
	netCents := amountCents - appliedCents

	result, err := h.deps.Service.CreatePaymentIntent(paysvc.CreateIntentInput{
		AmountCents:        netCents,
		Currency:           "usd",
		ProductID:          product.ID,
		ProductName:        product.Name,
		UserID:             userID,
		UserEmail:          email,
		ReferralCode:       referralCode,
		AppliedCreditCents: appliedCents,
	})
	if err != nil {
		log.Printf("stripe create intent: %v", err)
		writeJSONError(w, http.StatusBadGateway, "failed to create payment intent")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// computeAppliedCredit decides how much store credit to apply to a purchase —
// authoritative on the server. It returns 0 if the user already owns the
// product (the one-license rule means they can't re-buy it, so credit would be
// wasted); otherwise the balance capped at the price, always leaving at least
// Stripe's minimum charge.
func (h *StripeHandler) computeAppliedCredit(ctx context.Context, email, productID string, amountCents int64) int64 {
	if h.deps.LicenseRepo != nil {
		if lic, _ := h.deps.LicenseRepo.FindActive(ctx, email, productID); lic != nil {
			return 0
		}
	}
	u, err := h.deps.UserRepo.FindByEmail(ctx, email)
	if err != nil || u == nil || u.StoreCreditCents <= 0 {
		return 0
	}
	applied := u.StoreCreditCents
	if applied > amountCents {
		applied = amountCents
	}
	if amountCents-applied < stripeMinChargeCents {
		applied = amountCents - stripeMinChargeCents
	}
	if applied < 0 {
		applied = 0
	}
	return applied
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

		var license *repo.License
		if h.deps.LicenseRepo != nil {
			var (
				wasNew bool
				err    error
			)
			license, wasNew, err = h.deps.LicenseRepo.AssignFromPool(ctx, userEmail, productID, pi.ID)
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
		}

		// Crediting + redemption run BEFORE the email so they always execute on
		// the first successful pass — even if the email later fails and forces a
		// Stripe retry. Both are idempotent (gated on event.ID), so a retry never
		// double-credits or double-debits. (Previously these ran AFTER the email;
		// a permanently-failing email skipped them, leaving the referrer
		// uncredited and the buyer's applied store credit never debited.)
		if h.deps.ReferralRepo != nil && productID == referralQualifyingProduct {
			if code := pi.Metadata["referral_code"]; code != "" {
				credited, err := h.deps.ReferralRepo.CreditReferral(ctx, repo.CreditReferralInput{
					StripeEventID:   event.ID,
					PaymentIntentID: pi.ID,
					ReferralCode:    code,
					BuyerEmail:      userEmail,
					ProductID:       productID,
					CreditCents:     ReferralCreditCents,
				})
				if err != nil {
					log.Printf("stripe webhook: referral credit failed (event=%s code=%s): %v", event.ID, code, err)
				} else if credited {
					log.Printf("stripe webhook: credited %d¢ to referrer of code=%s (event=%s)", ReferralCreditCents, code, event.ID)

					// Notify the referrer (best-effort; don't fail the webhook on email problems)
					if h.deps.Mailer != nil && h.deps.UserRepo != nil {
						if referrer, err := h.deps.UserRepo.FindByReferralCode(ctx, code); err == nil && referrer != nil {
							_ = h.deps.Mailer.SendReferralEarned(ctx, referrer.Email, "$5.99", userEmail, productID)
						}
					}
				}
			}
		}
		if h.deps.ReferralRepo != nil {
			if appliedStr := pi.Metadata["applied_credit_cents"]; appliedStr != "" {
				if applied, perr := strconv.ParseInt(appliedStr, 10, 64); perr == nil && applied > 0 {
					redeemed, err := h.deps.ReferralRepo.RedeemCredit(ctx, repo.RedeemInput{
						StripeEventID:   event.ID,
						PaymentIntentID: pi.ID,
						Email:           userEmail,
						ProductID:       productID,
						AmountCents:     applied,
					})
					if err != nil {
						log.Printf("stripe webhook: redeem credit failed (event=%s email=%s): %v", event.ID, userEmail, err)
					} else if redeemed {
						log.Printf("stripe webhook: redeemed %d¢ store credit from %s (event=%s)", applied, userEmail, event.ID)
					}
				}
			}
		}

		// Email last. A transient failure returns 5xx so Stripe retries delivery;
		// crediting/redemption above already happened idempotently. processed_events
		// dedups duplicate sends on retry.
		if license != nil && h.deps.Mailer != nil {
			if err := h.deps.Mailer.SendLicense(ctx, userEmail, license.Key, productID); err != nil {
				log.Printf("stripe webhook: send email (event=%s): %v", event.ID, err)
				writeJSONError(w, http.StatusInternalServerError, "email send failed")
				return
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
