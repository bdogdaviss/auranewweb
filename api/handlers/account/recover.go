package account

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"aura-optimizer/internal/httpx"
	"aura-optimizer/internal/mailer"
	"aura-optimizer/internal/repo"
)

// ResendByEmailDeps backs the public, account-less "resend my license" flow for
// guest buyers who never created a login. Unlike ResendLicenseHandler the email
// comes from the request body, not a session, so the handler is deliberately
// tight-lipped (never reveals whether an email has a license) and rate-limited
// per email (so it can't be used to bomb a customer's inbox).
type ResendByEmailDeps struct {
	LicenseRepo repo.LicenseRepository
	Mailer      mailer.Mailer
	// Cooldown is the minimum gap between sends for the same email. <=0 uses 60s.
	Cooldown time.Duration
}

type ResendByEmailHandler struct {
	deps     ResendByEmailDeps
	mu       sync.Mutex
	lastSent map[string]time.Time
}

func NewResendByEmailHandler(deps ResendByEmailDeps) *ResendByEmailHandler {
	if deps.Cooldown <= 0 {
		deps.Cooldown = 60 * time.Second
	}
	return &ResendByEmailHandler{deps: deps, lastSent: make(map[string]time.Time)}
}

type resendByEmailRequest struct {
	Email string `json:"email"`
}

// genericOK is returned in every non-error case. It never confirms whether a
// license actually exists for the address, so the endpoint can't be used to
// enumerate which emails are customers.
const genericOK = "If a license exists for that email, we've just sent it. Check your inbox (and spam)."

func (h *ResendByEmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.deps.LicenseRepo == nil || h.deps.Mailer == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "license recovery is not configured")
		return
	}

	var req resendByEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}
	email := normalizeEmail(req.Email)
	if email == "" {
		httpx.WriteError(w, http.StatusBadRequest, "a valid email is required")
		return
	}

	// Per-email cooldown. A too-soon repeat is silently treated as success so an
	// attacker can neither bomb a customer's inbox nor time the response to learn
	// whether the email is a customer.
	if !h.allow(email) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": genericOK})
		return
	}

	lic, err := h.deps.LicenseRepo.FindLatestByEmail(r.Context(), email)
	if err != nil {
		log.Printf("resend-by-email: lookup %s: %v", email, err)
		// Respond generically regardless — don't leak failure detail.
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": genericOK})
		return
	}
	if lic != nil {
		if err := h.deps.Mailer.SendLicense(r.Context(), email, lic.Key, lic.ProductID); err != nil {
			log.Printf("resend-by-email: send %s: %v", email, err)
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": genericOK})
}

// allow reports whether email may be served now (outside the cooldown window),
// recording the time when it returns true. It opportunistically prunes stale
// entries so the map can't grow without bound from attacker-supplied addresses.
func (h *ResendByEmailHandler) allow(email string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	if last, ok := h.lastSent[email]; ok && now.Sub(last) < h.deps.Cooldown {
		return false
	}
	if len(h.lastSent) > 10000 {
		for k, t := range h.lastSent {
			if now.Sub(t) >= h.deps.Cooldown {
				delete(h.lastSent, k)
			}
		}
	}
	h.lastSent[email] = now
	return true
}

// normalizeEmail trims, validates, and lowercases an address, returning "" if
// it isn't a valid email.
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
