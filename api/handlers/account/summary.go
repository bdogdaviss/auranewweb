package account

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"aura-optimizer/internal/httpx"
	"aura-optimizer/internal/repo"
)

// SummaryDeps are injected from main.go. GetUserEmail returns the authenticated
// user's email or "" — we never accept a user identifier from the request
// (IDOR). PublicBaseURL is the canonical site origin used to build share links;
// when empty it is derived from the incoming request.
type SummaryDeps struct {
	GetUserEmail  func(r *http.Request) string
	UserRepo      repo.UserRepository
	ReferralRepo  repo.ReferralRepository
	PublicBaseURL string
}

type SummaryHandler struct {
	deps SummaryDeps
}

func NewSummaryHandler(deps SummaryDeps) *SummaryHandler {
	return &SummaryHandler{deps: deps}
}

func (h *SummaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.deps.GetUserEmail == nil || h.deps.UserRepo == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "account summary is not configured")
		return
	}

	email := h.deps.GetUserEmail(r)
	if email == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "login required")
		return
	}

	user, err := h.deps.UserRepo.EnsureUser(r.Context(), "", email, "")
	if err != nil || user == nil {
		log.Printf("account summary: ensure user %s: %v", email, err)
		httpx.WriteError(w, http.StatusInternalServerError, "lookup failed")
		return
	}

	referralCount := 0
	totalEarned := int64(0)
	var referrals []map[string]any
	if h.deps.ReferralRepo != nil {
		if n, err := h.deps.ReferralRepo.CountByReferrer(r.Context(), user.ID); err == nil {
			referralCount = n
		}
		if t, err := h.deps.ReferralRepo.TotalEarnedCents(r.Context(), user.ID); err == nil {
			totalEarned = t
		}
		if evts, err := h.deps.ReferralRepo.ListByReferrer(r.Context(), user.ID, 50); err == nil {
			for _, e := range evts {
				referrals = append(referrals, map[string]any{
					"date":       e.CreatedAt.Format("Jan 2, 2006"),
					"product":    e.ProductID,
					"commission": formatCents(e.CreditCents),
					"status":     "Credited",
				})
			}
		}
	}

	// Resolve the person who referred this user (if any) for nicer display in the dashboard.
	referredByName := ""
	if user.ReferredByCode != "" {
		if refUser, err := h.deps.UserRepo.FindByReferralCode(r.Context(), user.ReferredByCode); err == nil && refUser != nil {
			referredByName = refUser.Name
		}
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"referral_code":        user.ReferralCode,
		"referral_link":        h.referralLink(r, user.ReferralCode),
		"store_credit_cents":   user.StoreCreditCents,
		"store_credit_display": formatCents(user.StoreCreditCents),
		"total_earned_cents":   totalEarned,
		"total_earned_display": formatCents(totalEarned),
		"referral_count":       referralCount,
		"referrals":            referrals,
		"commission_display":   "$5.99",
		"referred_by_code":     user.ReferredByCode,
		"referred_by_name":     referredByName,
		"user": map[string]any{
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

// referralLink builds the shareable URL for a code, preferring the configured
// PublicBaseURL and falling back to the request's own scheme + host.
func (h *SummaryHandler) referralLink(r *http.Request, code string) string {
	base := strings.TrimRight(strings.TrimSpace(h.deps.PublicBaseURL), "/")
	if base == "" {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	return base + "/?ref=" + code
}

func formatCents(cents int64) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}
