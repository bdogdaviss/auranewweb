package account

import (
	"log"
	"net/http"

	"aura-optimizer/internal/httpx"
	"aura-optimizer/internal/mailer"
	"aura-optimizer/internal/repo"
)

// ResendLicenseDeps are injected from main.go. GetUserEmail must return the
// authenticated user's email or "" if the request is unauthenticated. We do
// NOT accept a user identifier in the request body — that would be an IDOR.
type ResendLicenseDeps struct {
	GetUserEmail func(r *http.Request) string
	LicenseRepo  repo.LicenseRepository
	Mailer       mailer.Mailer
}

type ResendLicenseHandler struct {
	deps ResendLicenseDeps
}

func NewResendLicenseHandler(deps ResendLicenseDeps) *ResendLicenseHandler {
	return &ResendLicenseHandler{deps: deps}
}

func (h *ResendLicenseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.deps.GetUserEmail == nil || h.deps.LicenseRepo == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "license resend is not configured")
		return
	}

	email := h.deps.GetUserEmail(r)
	if email == "" {
		httpx.WriteError(w, http.StatusUnauthorized, "login required")
		return
	}

	lic, err := h.deps.LicenseRepo.FindLatestByEmail(r.Context(), email)
	if err != nil {
		log.Printf("resend-license: find license for %s: %v", email, err)
		httpx.WriteError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	if lic == nil {
		httpx.WriteError(w, http.StatusNotFound, "no active license for this account")
		return
	}

	if h.deps.Mailer == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "email is not configured")
		return
	}

	if err := h.deps.Mailer.SendLicense(r.Context(), email, lic.Key, lic.ProductID); err != nil {
		log.Printf("resend-license: send to %s: %v", email, err)
		httpx.WriteError(w, http.StatusBadGateway, "email send failed")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "sent", "to": email, "product": lic.ProductID})
}
