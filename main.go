// main.go - Enhanced version with more features
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	accounthandler "aura-optimizer/api/handlers/account"
	stripehandler "aura-optimizer/api/handlers/payment"
	"aura-optimizer/internal/db"
	"aura-optimizer/internal/mailer"
	paysvc "aura-optimizer/internal/payment"
	"aura-optimizer/internal/repo"
)

// frontendDist is the built React app (Vite output). Embedded at compile time
// so a single Go binary deploys both backend and frontend. Run `npm run build`
// inside frontend/ before `go build`.
//
//go:embed all:frontend/dist
var frontendDist embed.FS

// Persistence + email layer. Initialised in main() before route registration.
var (
	licenseRepo  repo.LicenseRepository
	eventStore   repo.EventStore
	appMailer    mailer.Mailer
	userRepo     repo.UserRepository
	referralRepo repo.ReferralRepository
	leadRepo     repo.LeadRepository
)

// publicBaseURL is the canonical origin used to build referral share links.
// Empty falls back to the request's own scheme+host.
var publicBaseURL = getEnv("PUBLIC_BASE_URL", "")

// PayPal configuration - set these as environment variables.
// No defaults: leaving either credential empty disables the PayPal routes.
var (
	paypalClientID = getEnv("PAYPAL_CLIENT_ID", "")
	paypalSecret   = getEnv("PAYPAL_SECRET", "")
	paypalBaseURL  = normalizePayPalBaseURL(getEnv("PAYPAL_BASE_URL", "https://api-m.paypal.com")) // Use https://api-m.sandbox.paypal.com for sandbox
)

// Stripe publishable key (front-end). Secret key is read inside internal/payment.
// Empty value means the checkout page hides the Card option.
var stripePublishableKey = getEnv("STRIPE_PUBLISHABLE_KEY", "")

func getEnv(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

// loadDotEnv loads KEY=VALUE pairs from a .env file into the process environment
// for local development (production sets real env vars / Fly secrets). Blank
// lines and # comments are skipped; everything after the first '=' is the value
// with surrounding quotes trimmed. Existing env vars are NOT overridden, so a
// shell export always wins. A missing file is a silent no-op.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		val := strings.Trim(strings.TrimSpace(line[eq+1:]), `"'`)
		_ = os.Setenv(key, val)
	}
}

func normalizePayPalBaseURL(baseURL string) string {
	clean := strings.TrimSpace(baseURL)
	clean = strings.TrimRight(clean, "/")
	if clean == "" {
		return "https://api-m.paypal.com"
	}
	return clean
}

// setReferralCookieIfPresent reads ?ref=CODE from the URL (if present) and
// sets a first-party "ref" cookie. This is the server-side equivalent of
// the client-side useReferralCapture hook. It makes attribution work even
// for direct links, non-React pages, or when the JS hasn't run yet.
// The cookie is read later by CreatePaymentIntent (and passed to Stripe
// metadata) so the webhook can credit the correct referrer.
func setReferralCookieIfPresent(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		return
	}
	// Basic sanitization mirroring the client-side regex + the server's
	// sanitizeCode expectations. We don't reject here — CreditReferral
	// will still validate that the code actually belongs to someone.
	if len(ref) < 3 || len(ref) > 32 {
		return
	}
	// Only allow the same character set the app uses for codes.
	for _, ch := range ref {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return
		}
	}
	lowered := strings.ToLower(ref)

	http.SetCookie(w, &http.Cookie{
		Name:     "ref",
		Value:    lowered,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // 30 days, same as the JS hook
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
	})
}

// Enhanced Models
type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	Name       string    `json:"name"`
	Avatar     string    `json:"avatar"`
	IsPro      bool      `json:"is_pro"`
	Downloads  int       `json:"downloads"`
	JoinedAt   time.Time `json:"joined_at"`
	LastActive time.Time `json:"last_active"`
}

type Product struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Price        float64  `json:"price"`
	ComparePrice float64  `json:"compare_price,omitempty"`
	Features     []string `json:"features"`
	IsPopular    bool     `json:"is_popular"`
	IsNew        bool     `json:"is_new"`
	Badge        string   `json:"badge,omitempty"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
}

// Global stores
var (
	users    = make(map[string]*User)
	sessions = make(map[string]string)   // token -> userID
	carts    = make(map[string][]string) // userID -> productIDs
	mu       sync.RWMutex

	templates *template.Template

	Products []Product
)

func init() {
	initData()
}

func initData() {
	Products = []Product{
		{
			ID:          "free",
			Name:        "Starter",
			Description: "Perfect for casual gamers",
			Price:       0,
			Features:    []string{"Basic optimization", "RAM cleanup", "Startup manager", "Community support"},
			IsPopular:   false,
			Icon:        "zap",
			Color:       "gray",
		},
		{
			ID:           "lifetime",
			Name:         "Lifetime Pro",
			Description:  "One-time payment, forever access",
			Price:        15,
			ComparePrice: 99.99,
			Features:     []string{"Everything in Starter", "Advanced AI tuning", "FPS boost guarantee", "Priority 24/7 support", "Lifetime updates", "Multi-PC license"},
			IsPopular:    true,
			Badge:        "Best Value",
			Icon:         "crown",
			Color:        "aura",
		},
		{
			ID:           "team",
			Name:         "Team License",
			Description:  "For esports teams & cafes",
			Price:        149.99,
			ComparePrice: 299.99,
			Features:     []string{"5 PC licenses", "Team dashboard", "White-label option", "API access", "Dedicated manager"},
			IsNew:        true,
			Badge:        "New",
			Icon:         "users",
			Color:        "purple",
		},
	}

}

func main() {
	// Load .env for local dev. The env-derived package globals below were
	// initialised at package-init (before main), so re-read them now that the
	// file is loaded — otherwise Stripe/PayPal/base-URL would stay unset in dev.
	loadDotEnv(".env")
	stripePublishableKey = getEnv("STRIPE_PUBLISHABLE_KEY", "")
	publicBaseURL = getEnv("PUBLIC_BASE_URL", "")
	paypalClientID = getEnv("PAYPAL_CLIENT_ID", "")
	paypalSecret = getEnv("PAYPAL_SECRET", "")
	paypalBaseURL = normalizePayPalBaseURL(getEnv("PAYPAL_BASE_URL", "https://api-m.paypal.com"))
	discordClientID = getEnv("DISCORD_CLIENT_ID", "")
	discordClientSecret = getEnv("DISCORD_CLIENT_SECRET", "")
	discordRedirectURI = getEnv("DISCORD_REDIRECT_URI", "")

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b int) float64 { return float64(a) / float64(b) },
		"percentage": func(before, after int) int {
			if before == 0 {
				return 0
			}
			return ((after - before) * 100) / before
		},
		"lower": func(s string) string { return strings.ToLower(s) },
	}

	var err error
	templates, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Failed to parse templates:", err)
	}

	// Durable storage for licenses + processed webhook events. Required (the
	// payment provider routes still register independently; if the DB fails to
	// open, license issuance + the resend endpoint will be unavailable but the
	// rest of the site keeps working).
	sqlDB, err := db.Open(getEnv("SQLITE_PATH", "./data/auranewweb.sqlite"))
	if err != nil {
		log.Fatalf("open sqlite: %v", err)
	}
	defer sqlDB.Close()
	licenseRepo = repo.NewLicenseRepository(sqlDB)
	eventStore = repo.NewEventStore(sqlDB)
	userRepo = repo.NewUserRepository(sqlDB)
	referralRepo = repo.NewReferralRepository(sqlDB)
	leadRepo = repo.NewLeadRepository(sqlDB)
	log.Printf("✅ SQLite ready at %s", getEnv("SQLITE_PATH", "./data/auranewweb.sqlite"))

	// Email provider. Resend in prod; no-op (logs only) when not configured so
	// local dev doesn't require a Resend account.
	if resendKey, resendFrom := getEnv("RESEND_API_KEY", ""), getEnv("RESEND_FROM", ""); resendKey != "" && resendFrom != "" {
		rm, mailerErr := mailer.NewResendMailer(
			resendKey,
			resendFrom,
			getEnv("SUPPORT_EMAIL", ""),       // shown in "Need help?" line; defaults to From address
			getEnv("AURA_DOWNLOAD_URL", ""),   // the "Download Aura Optimizer" button target
		)
		if mailerErr != nil {
			log.Fatalf("resend mailer init: %v", mailerErr)
		}
		appMailer = rm
		log.Printf("✅ Email enabled via Resend (from %s)", resendFrom)
	} else {
		appMailer = mailer.NoopMailer{}
		log.Printf("⚠️  RESEND_API_KEY / RESEND_FROM not set – emails disabled (noop mailer)")
	}

	// Static files - serves CSS, JS, and images
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes — functional pages stay as Go templates (they depend on session
	// state, Stripe Elements, and the license-gated download flow).
	http.HandleFunc("/login", loginPageHandler)
	http.HandleFunc("/register", registerPageHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/download", downloadHandler)
	// /account is served by the React SPA (client-side AccountPage); data comes
	// from /api/account/summary. There is no server-rendered /account route.
	http.HandleFunc("/recover", recoverPageHandler)
	http.HandleFunc("/success", successPageHandler)
	// Marketing pages (/, /products, /pricing, /about, /features) are served by
	// the React SPA via the catch-all "/" handler registered after the API
	// routes below. Go's mux gives longest-prefix priority, so the specific
	// /login, /api/*, /static/* handlers still win.

	// API
	http.HandleFunc("/api/auth/login", apiLogin)
	http.HandleFunc("/api/auth/register", apiRegister)
	http.HandleFunc("/api/account/referral-code", apiUpdateReferralCode)
	http.HandleFunc("/api/leads", apiLeads)

	// Discord OAuth login (registers only when fully configured).
	if discordEnabled() {
		http.HandleFunc("/auth/discord/login", discordLoginHandler)
		http.HandleFunc("/auth/discord/callback", discordCallbackHandler)
		log.Printf("✅ Discord login enabled")
	} else {
		log.Printf("⚠️  Discord login disabled (set DISCORD_CLIENT_ID/SECRET/REDIRECT_URI)")
	}
	http.HandleFunc("/api/cart/add", apiAddToCart)
	http.HandleFunc("/api/cart/remove", apiRemoveFromCart)
	http.HandleFunc("/api/cart", apiGetCart)
	http.HandleFunc("/api/checkout", apiCheckout)

	// Checkout page is always available; payment-provider APIs register conditionally below.
	http.HandleFunc("/checkout", checkoutPageHandler)

	// PayPal API (registered only if PAYPAL_CLIENT_ID and PAYPAL_SECRET are set)
	if paypalClientID != "" && paypalSecret != "" {
		http.HandleFunc("/api/paypal/create-order", apiPayPalCreateOrder)
		http.HandleFunc("/api/paypal/capture-order", apiPayPalCaptureOrder)
		log.Printf("✅ PayPal routes enabled")
	} else {
		log.Printf("⚠️  PayPal credentials not set – PayPal routes disabled")
	}

	// Stripe Checkout (registered only if STRIPE_SECRET_KEY is set)
	if svc, err := paysvc.NewStripeService(); err != nil {
		log.Printf("stripe disabled: %v", err)
	} else {
		h := stripehandler.NewStripeHandler(stripehandler.Deps{
			Service:       svc,
			LookupProduct: lookupProductForStripe,
			GetUserID: func(r *http.Request) (string, string) {
				u := getUser(r)
				if u == nil {
					return "", ""
				}
				return u.ID, u.Email
			},
			MarkUserPro: func(userID string) {
				mu.Lock()
				if u, ok := users[userID]; ok {
					u.IsPro = true
				}
				mu.Unlock()
			},
			LicenseRepo:  licenseRepo,
			EventStore:   eventStore,
			Mailer:       appMailer,
			ReferralRepo: referralRepo,
			UserRepo:     userRepo,
		})
		http.HandleFunc("/api/stripe/create-payment-intent", h.CreatePaymentIntent)
		http.HandleFunc("/api/stripe/webhook", h.Webhook)
		log.Printf("✅ Stripe routes enabled")
	}

	// Resend-license endpoint: lets an authenticated user re-email themselves
	// their existing license key. Always registered; the handler degrades
	// gracefully if licenseRepo / appMailer aren't available.
	resendHandler := accounthandler.NewResendLicenseHandler(accounthandler.ResendLicenseDeps{
		GetUserEmail: func(r *http.Request) string {
			u := getUser(r)
			if u == nil {
				return ""
			}
			return u.Email
		},
		LicenseRepo: licenseRepo,
		Mailer:      appMailer,
	})
	http.Handle("/api/account/resend-license", resendHandler)

	// Affiliate dashboard data: the authenticated user's referral link + balance.
	summaryHandler := accounthandler.NewSummaryHandler(accounthandler.SummaryDeps{
		GetUserEmail: func(r *http.Request) string {
			u := getUser(r)
			if u == nil {
				return ""
			}
			return u.Email
		},
		UserRepo:      userRepo,
		ReferralRepo:  referralRepo,
		PublicBaseURL: publicBaseURL,
	})
	http.Handle("/api/account/summary", summaryHandler)

	// Public, account-less license recovery for guest buyers. Tight-lipped +
	// per-email rate-limited (see the handler).
	recoverHandler := accounthandler.NewResendByEmailHandler(accounthandler.ResendByEmailDeps{
		LicenseRepo: licenseRepo,
		Mailer:      appMailer,
	})
	http.Handle("/api/account/resend-by-email", recoverHandler)

	// SPA catch-all. Must be registered last (it's the "/" handler).
	// Specific /login, /checkout, /api/*, /static/* still take priority.
	http.HandleFunc("/", serveSPA)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getUser(r *http.Request) *User {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}
	mu.RLock()
	userID, ok := sessions[cookie.Value]
	mu.RUnlock()
	if !ok {
		return nil
	}
	mu.RLock()
	user := users[userID]
	mu.RUnlock()
	return user
}

// The home / pricing / about / features handlers used to render templates
// here; those routes are now owned by the React SPA in frontend/dist and
// served by serveSPA (see bottom of file). Their template files (index.html,
// pricing.html, about.html, features.html) remain on disk but are no longer
// referenced — left in place so a future revert is a one-line route
// registration, not a template rewrite.

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	setReferralCookieIfPresent(w, r)

	if r.Method == "POST" {
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}
	_ = templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"DiscordEnabled": discordEnabled(),
	})
}

func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	setReferralCookieIfPresent(w, r)

	_ = templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
		"DiscordEnabled": discordEnabled(),
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("session")
	if cookie != nil {
		mu.Lock()
		delete(sessions, cookie.Value)
		mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user != nil {
		mu.Lock()
		users[user.ID].Downloads++
		users[user.ID].LastActive = time.Now()
		mu.Unlock()
	}

	// Strict gate: only emit the download URL to clients with a logged-in user
	// and a non-revoked license in SQLite. IsPro (in-memory) is left as a
	// faster UI hint for other surfaces but is not authoritative here.
	var (
		hasLicense  bool
		downloadURL string
		licenseKey  string
	)
	if user != nil && licenseRepo != nil {
		lic, err := licenseRepo.FindLatestByEmail(r.Context(), user.Email)
		if err != nil {
			log.Printf("download license lookup for %s: %v", user.Email, err)
		} else if lic != nil {
			hasLicense = true
			downloadURL = os.Getenv("AURA_DOWNLOAD_URL")
			licenseKey = lic.Key
		}
	}

	_ = templates.ExecuteTemplate(w, "download.html", map[string]interface{}{
		"User":        user,
		"DownloadURL": downloadURL, // empty unless the user holds a valid license
		"HasLicense":  hasLicense,
		"LicenseKey":  licenseKey,
	})
}

// recoverPageHandler renders the public "resend my license" page for guest
// buyers who never made an account. No auth required.
func recoverPageHandler(w http.ResponseWriter, r *http.Request) {
	_ = templates.ExecuteTemplate(w, "recover.html", map[string]interface{}{
		"User": getUser(r),
	})
}

// successPageHandler renders the post-purchase confirmation. The order ref and
// product come from the checkout redirect query; nothing sensitive is shown, so
// no auth is required.
func successPageHandler(w http.ResponseWriter, r *http.Request) {
	productName := "Aura Optimizer"
	if id := r.URL.Query().Get("product"); id != "" {
		for _, p := range Products {
			if p.ID == id {
				productName = p.Name
				break
			}
		}
	}
	_ = templates.ExecuteTemplate(w, "success.html", map[string]interface{}{
		"User":        getUser(r),
		"Ref":         r.URL.Query().Get("ref"),
		"ProductName": productName,
	})
}

// buildReferralLink prefers PUBLIC_BASE_URL and falls back to the request's own
// scheme + host.
func buildReferralLink(r *http.Request, code string) string {
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	return base + "/?ref=" + code
}

// API Handlers
func apiLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	if userRepo == nil {
		http.Error(w, `{"error":"login unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// Verify against the durable account. No more auto-create on login: an
	// unknown email is rejected (register first), a wrong password is rejected,
	// and accounts with no password (Discord-only) must use Discord.
	dbUser, err := userRepo.FindByEmail(r.Context(), email)
	if err != nil {
		log.Printf("apiLogin: lookup %s: %v", email, err)
		http.Error(w, `{"error":"login failed"}`, http.StatusInternalServerError)
		return
	}
	if dbUser == nil {
		http.Error(w, `{"error":"No account found with that email. Please register."}`, http.StatusUnauthorized)
		return
	}
	if dbUser.PasswordHash == "" {
		http.Error(w, `{"error":"This account has no password set — sign in with Discord."}`, http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(req.Password)) != nil {
		http.Error(w, `{"error":"Invalid email or password."}`, http.StatusUnauthorized)
		return
	}

	// Verified — hydrate the in-memory user (which backs session resolution and
	// the UI hint fields) from the durable row.
	mu.Lock()
	var user *User
	for _, u := range users {
		if strings.EqualFold(u.Email, email) {
			user = u
			break
		}
	}
	if user == nil {
		user = &User{
			ID:       dbUser.ID,
			Email:    dbUser.Email,
			Name:     dbUser.Name,
			Avatar:   "https://ui-avatars.com/api/?name=" + email + "&background=10b981&color=fff",
			JoinedAt: dbUser.CreatedAt,
		}
		users[user.ID] = user
	}
	mu.Unlock()

	startSession(w, r, user.ID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user": user})
}

// apiLeads stores an email capture from a marketing form (the affiliate-signup
// banner and the payout waitlist). Public; source is allowlisted.
func apiLeads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	if leadRepo == nil {
		http.Error(w, `{"error":"unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	var req struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		Source string `json:"source"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		http.Error(w, `{"error":"a valid email is required"}`, http.StatusBadRequest)
		return
	}
	switch req.Source {
	case "affiliate-signup", "payout-waitlist":
		// allowed
	default:
		http.Error(w, `{"error":"unknown source"}`, http.StatusBadRequest)
		return
	}
	if err := leadRepo.AddLead(r.Context(), email, req.Name, req.Source); err != nil {
		log.Printf("apiLeads (%s): %v", req.Source, err)
		http.Error(w, `{"error":"could not save"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// startSession issues a session token for userID and sets the session cookie.
// Shared by the email login and the Discord OAuth callback.
func startSession(w http.ResponseWriter, r *http.Request, userID string) {
	token := uuid.New().String()
	mu.Lock()
	sessions[token] = userID
	mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
		MaxAge:   86400 * 7,
	})
}

// isHTTPS reports whether the request arrived over TLS (directly or behind a
// proxy that set X-Forwarded-Proto). Used to flag cookies Secure only when it
// won't break plain-HTTP local dev.
func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// apiUpdateReferralCode lets a logged-in user change their referral/discount
// code (the editable "Social Name" on the dashboard).
func apiUpdateReferralCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	user := getUser(r)
	if user == nil {
		http.Error(w, `{"error":"login required"}`, http.StatusUnauthorized)
		return
	}
	if userRepo == nil {
		http.Error(w, `{"error":"unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	var req struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Resolve the durable user row (id is keyed by email, stable across restarts).
	u, err := userRepo.FindByEmail(r.Context(), user.Email)
	if err != nil || u == nil {
		http.Error(w, `{"error":"account not found"}`, http.StatusInternalServerError)
		return
	}
	switch err := userRepo.UpdateReferralCode(r.Context(), u.ID, req.Code); {
	case err == nil:
		// fall through to success
	case errors.Is(err, repo.ErrReferralCodeTaken):
		http.Error(w, `{"error":"that code is already taken"}`, http.StatusConflict)
		return
	case errors.Is(err, repo.ErrReferralCodeInvalid):
		http.Error(w, `{"error":"code must be 3-32 characters: letters, numbers, _ or -"}`, http.StatusBadRequest)
		return
	default:
		// Don't echo raw DB errors to the client (leaks schema).
		log.Printf("apiUpdateReferralCode: update %s: %v", user.Email, err)
		http.Error(w, `{"error":"could not update code"}`, http.StatusInternalServerError)
		return
	}
	updated, ferr := userRepo.FindByEmail(r.Context(), user.Email)
	if ferr != nil || updated == nil {
		log.Printf("apiUpdateReferralCode: reload %s: %v", user.Email, ferr)
		http.Error(w, `{"error":"saved, but could not reload your code"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code": updated.ReferralCode,
		"link": buildReferralLink(r, updated.ReferralCode),
	})
}

func apiRegister(w http.ResponseWriter, r *http.Request) {
	setReferralCookieIfPresent(w, r)

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Durable duplicate check — this survives restarts (the in-memory users map
	// does not, which is why repeated registers used to silently succeed).
	if userRepo != nil {
		if existing, err := userRepo.FindByEmail(r.Context(), email); err == nil && existing != nil {
			http.Error(w, `{"error":"An account with this email already exists. Please sign in."}`, http.StatusConflict)
			return
		}
	}

	mu.Lock()
	for _, u := range users {
		if strings.EqualFold(u.Email, email) {
			mu.Unlock()
			http.Error(w, `{"error":"An account with this email already exists. Please sign in."}`, http.StatusConflict)
			return
		}
	}
	name := strings.Split(email, "@")[0]
	if n := strings.TrimSpace(req.Name); n != "" {
		name = n
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	user := &User{
		ID:       uuid.New().String(),
		Email:    email,
		Password: string(hashed),
		Name:     name,
		Avatar:   "https://ui-avatars.com/api/?name=" + email + "&background=10b981&color=fff",
		JoinedAt: time.Now(),
	}
	users[user.ID] = user
	mu.Unlock()

	if userRepo != nil {
		pu, err := userRepo.EnsureUser(r.Context(), user.ID, user.Email, user.Name)
		if err != nil {
			log.Printf("apiRegister: persist user %s: %v", user.Email, err)
		} else if pu != nil {
			if err := userRepo.SetPassword(r.Context(), pu.ID, user.Password); err != nil {
				log.Printf("apiRegister: set password %s: %v", user.Email, err)
			}

			// Record the referrer (from ?ref= cookie set by server-side capture or client hook).
			// Only the first referrer is kept.
			if refCookie, _ := r.Cookie("ref"); refCookie != nil && refCookie.Value != "" {
				if err := userRepo.SetReferredBy(r.Context(), pu.ID, refCookie.Value); err != nil {
					log.Printf("apiRegister: set referred_by_code for %s: %v", user.Email, err)
				}
			}
		}
	}

	startSession(w, r, user.ID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user": user})
}

func apiAddToCart(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Error(w, `{"error": "login required"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		ProductID string `json:"product_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	mu.Lock()
	carts[user.ID] = append(carts[user.ID], req.ProductID)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func buildCartProducts(items []string) ([]Product, float64) {
	var total float64
	var cartProducts []Product
	for _, id := range items {
		for _, p := range Products {
			if p.ID == id {
				cartProducts = append(cartProducts, p)
				total += p.Price
				break
			}
		}
	}
	return cartProducts, total
}

func apiRemoveFromCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := getUser(r)
	if user == nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		http.Error(w, `{"error": "product_id required"}`, http.StatusBadRequest)
		return
	}

	mu.Lock()
	items := carts[user.ID]
	updated := make([]string, 0, len(items))
	removed := false
	for _, id := range items {
		if !removed && id == productID {
			removed = true
			continue
		}
		updated = append(updated, id)
	}
	carts[user.ID] = updated
	mu.Unlock()

	cartProducts, total := buildCartProducts(updated)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": removed,
		"items":   cartProducts,
		"total":   total,
		"count":   len(updated),
	})
}

func apiGetCart(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	mu.RLock()
	items := carts[user.ID]
	mu.RUnlock()

	cartProducts, total := buildCartProducts(items)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items": cartProducts,
		"total": total,
		"count": len(items),
	})
}

func apiCheckout(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	if user == nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Simulate processing
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	for _, id := range carts[user.ID] {
		if id == "lifetime" || id == "team" {
			users[user.ID].IsPro = true
		}
	}
	carts[user.ID] = []string{}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Payment successful!",
		"redirect": "/download",
	})
}

// --- PayPal Checkout Integration ---

func checkoutPageHandler(w http.ResponseWriter, r *http.Request) {
	setReferralCookieIfPresent(w, r)

	user := getUser(r)
	productID := r.URL.Query().Get("product")
	if productID == "" {
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	var product *Product
	for _, p := range Products {
		if p.ID == productID {
			product = &p
			break
		}
	}
	if product == nil || product.Price == 0 {
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	savings := ""
	if product.ComparePrice > 0 {
		savings = fmt.Sprintf("%.2f", product.ComparePrice-product.Price)
	}

	// Store-credit context for logged-in users: their balance and whether they
	// already own this product (in which case credit can't be spent on it).
	var creditCents int64
	alreadyOwns := false
	if user != nil {
		if userRepo != nil {
			if u, _ := userRepo.FindByEmail(r.Context(), user.Email); u != nil {
				creditCents = u.StoreCreditCents
			}
		}
		if licenseRepo != nil {
			if lic, _ := licenseRepo.FindActive(r.Context(), user.Email, product.ID); lic != nil {
				alreadyOwns = true
			}
		}
	}

	_ = templates.ExecuteTemplate(w, "checkout.html", map[string]interface{}{
		"User":                 user,
		"Product":              product,
		"Savings":              savings,
		"PayPalClientID":       paypalClientID,
		"StripePublishableKey": stripePublishableKey,
		"StoreCreditCents":     creditCents,
		"StoreCreditDisplay":   fmt.Sprintf("$%d.%02d", creditCents/100, creditCents%100),
		"AlreadyOwns":          alreadyOwns,
	})
}

// getPayPalAccessToken fetches an OAuth2 token from PayPal
func getPayPalAccessToken() (string, error) {
	req, err := http.NewRequest("POST", paypalBaseURL+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(paypalClientID, paypalSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("failed to get PayPal access token")
	}
	return result.AccessToken, nil
}

func apiPayPalCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProductID string `json:"product_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	var product *Product
	for _, p := range Products {
		if p.ID == req.ProductID {
			product = &p
			break
		}
	}
	if product == nil || product.Price == 0 {
		http.Error(w, `{"error":"invalid product"}`, http.StatusBadRequest)
		return
	}

	accessToken, err := getPayPalAccessToken()
	if err != nil {
		log.Printf("PayPal auth error: %v", err)
		http.Error(w, `{"error":"payment service unavailable"}`, http.StatusInternalServerError)
		return
	}

	orderPayload := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": product.ID,
				"description":  product.Name + " - Aura Optimizer",
				"amount": map[string]interface{}{
					"currency_code": "USD",
					"value":         fmt.Sprintf("%.2f", product.Price),
					"breakdown": map[string]interface{}{
						"item_total": map[string]string{
							"currency_code": "USD",
							"value":         fmt.Sprintf("%.2f", product.Price),
						},
					},
				},
				"items": []map[string]interface{}{
					{
						"name":        product.Name,
						"description": product.Description,
						"unit_amount": map[string]string{
							"currency_code": "USD",
							"value":         fmt.Sprintf("%.2f", product.Price),
						},
						"quantity": "1",
						"category": "DIGITAL_GOODS",
					},
				},
			},
		},
		"payment_source": map[string]interface{}{
			"paypal": map[string]interface{}{
				"experience_context": map[string]interface{}{
					"payment_method_preference": "IMMEDIATE_PAYMENT_REQUIRED",
					"landing_page":              "LOGIN",
					"user_action":               "PAY_NOW",
					"return_url":                "https://example.com/return",
					"cancel_url":                "https://example.com/cancel",
				},
			},
		},
	}

	body, _ := json.Marshal(orderPayload)
	ppReq, err := http.NewRequest("POST", paypalBaseURL+"/v2/checkout/orders", bytes.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":"failed to create order"}`, http.StatusInternalServerError)
		return
	}
	ppReq.Header.Set("Content-Type", "application/json")
	ppReq.Header.Set("Authorization", "Bearer "+accessToken)
	ppReq.Header.Set("PayPal-Request-Id", uuid.New().String())

	resp, err := http.DefaultClient.Do(ppReq)
	if err != nil {
		http.Error(w, `{"error":"payment service error"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		log.Printf("PayPal create order error: %s", string(respBody))
		http.Error(w, `{"error":"failed to create payment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBody)
}

func apiPayPalCaptureOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID string `json:"order_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.OrderID == "" {
		http.Error(w, `{"error":"order_id required"}`, http.StatusBadRequest)
		return
	}

	accessToken, err := getPayPalAccessToken()
	if err != nil {
		log.Printf("PayPal auth error: %v", err)
		http.Error(w, `{"error":"payment service unavailable"}`, http.StatusInternalServerError)
		return
	}

	ppReq, err := http.NewRequest("POST", paypalBaseURL+"/v2/checkout/orders/"+req.OrderID+"/capture", nil)
	if err != nil {
		http.Error(w, `{"error":"capture failed"}`, http.StatusInternalServerError)
		return
	}
	ppReq.Header.Set("Content-Type", "application/json")
	ppReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(ppReq)
	if err != nil {
		http.Error(w, `{"error":"capture request failed"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var captureResult map[string]interface{}
	_ = json.Unmarshal(respBody, &captureResult)

	status, _ := captureResult["status"].(string)
	if status == "COMPLETED" {
		// Mark user as Pro if logged in
		user := getUser(r)
		if user != nil {
			mu.Lock()
			users[user.ID].IsPro = true
			mu.Unlock()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"order_id": req.OrderID,
			"status":   status,
		})
		return
	}

	log.Printf("PayPal capture status: %s, body: %s", status, string(respBody))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   "Payment could not be completed. Status: " + status,
	})
}

func lookupProductForStripe(productID string) (stripehandler.Product, bool) {
	for _, p := range Products {
		if p.ID == productID {
			return stripehandler.Product{ID: p.ID, Name: p.Name, Price: p.Price}, true
		}
	}
	return stripehandler.Product{}, false
}

// serveSPA serves the React app embedded under frontend/dist. For any path
// that exists as a static file in dist (e.g. /assets/index-xxx.js), it streams
// the file. For anything else that isn't an /api/ route, it falls back to
// index.html so React Router can handle client-side routing. /api/* requests
// that didn't match an earlier specific handler get a real 404.
func serveSPA(w http.ResponseWriter, r *http.Request) {
	setReferralCookieIfPresent(w, r)

	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	distFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		http.Error(w, "frontend not built", http.StatusInternalServerError)
		return
	}

	// Try the requested path as a static asset first.
	requested := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if requested != "" && requested != "." {
		if f, err := distFS.Open(requested); err == nil {
			info, _ := f.Stat()
			_ = f.Close()
			if info != nil && !info.IsDir() {
				http.FileServer(http.FS(distFS)).ServeHTTP(w, r)
				return
			}
		}
	}

	// SPA fallback: send index.html so React Router renders the route.
	idx, err := distFS.Open("index.html")
	if err != nil {
		http.Error(w, "index.html missing from frontend/dist (did you run `npm run build`?)", http.StatusInternalServerError)
		return
	}
	defer idx.Close()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = io.Copy(w, idx)
}
