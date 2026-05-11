// main.go - Enhanced version with more features
package main

import (
	"bytes"
	"embed"
	"encoding/json"
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
	licenseRepo repo.LicenseRepository
	eventStore  repo.EventStore
	appMailer   mailer.Mailer
)

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

func normalizePayPalBaseURL(baseURL string) string {
	clean := strings.TrimSpace(baseURL)
	clean = strings.TrimRight(clean, "/")
	if clean == "" {
		return "https://api-m.paypal.com"
	}
	return clean
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

type Testimonial struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	Avatar    string `json:"avatar"`
	Content   string `json:"content"`
	Rating    int    `json:"rating"`
	Game      string `json:"game"`
	FPSBefore int    `json:"fps_before"`
	FPSAfter  int    `json:"fps_after"`
}

type Stat struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Change string `json:"change,omitempty"`
}

// Global stores
var (
	users    = make(map[string]*User)
	sessions = make(map[string]string)   // token -> userID
	carts    = make(map[string][]string) // userID -> productIDs
	mu       sync.RWMutex

	templates *template.Template

	Products     []Product
	Testimonials []Testimonial
	Stats        []Stat
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

	Testimonials = []Testimonial{
		{
			ID:        "1",
			Name:      "Alex",
			Handle:    "@alexfn",
			Avatar:    "A",
			Content:   "Went from 120 FPS to 240 FPS in Valorant. Game changer!",
			Rating:    5,
			Game:      "Valorant",
			FPSBefore: 120,
			FPSAfter:  240,
		},
		{
			ID:        "2",
			Name:      "Sarah",
			Handle:    "@sarahgaming",
			Avatar:    "S",
			Content:   "Finally no more stutters in Fortnite. Smooth 60fps on my old laptop!",
			Rating:    5,
			Game:      "Fortnite",
			FPSBefore: 45,
			FPSAfter:  75,
		},
		{
			ID:        "3",
			Name:      "Mike",
			Handle:    "@mikepro",
			Avatar:    "M",
			Content:   "The AI tuning actually works. My setup has never been this responsive.",
			Rating:    5,
			Game:      "Apex Legends",
			FPSBefore: 80,
			FPSAfter:  144,
		},
	}

	Stats = []Stat{
		{Label: "Active Users", Value: "50K+", Change: "+12% this month"},
		{Label: "Avg FPS Boost", Value: "+47%", Change: "Verified by users"},
		{Label: "Countries", Value: "120+", Change: "Global reach"},
		{Label: "Support Rating", Value: "4.9/5", Change: "24/7 support"},
	}
}

func main() {
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
	// Marketing pages (/, /products, /pricing, /about, /features) are served by
	// the React SPA via the catch-all "/" handler registered after the API
	// routes below. Go's mux gives longest-prefix priority, so the specific
	// /login, /api/*, /static/* handlers still win.

	// API
	http.HandleFunc("/api/auth/login", apiLogin)
	http.HandleFunc("/api/auth/register", apiRegister)
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
			LicenseRepo: licenseRepo,
			EventStore:  eventStore,
			Mailer:      appMailer,
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	data := map[string]interface{}{
		"User":         user,
		"Products":     Products,
		"Testimonials": Testimonials,
		"Stats":        Stats,
		"ShowCart":     true,
	}
	templates.ExecuteTemplate(w, "index.html", data)
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		http.Redirect(w, r, "/api/auth/login", http.StatusSeeOther)
		return
	}
	templates.ExecuteTemplate(w, "login.html", nil)
}

func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
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

func pricingHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	templates.ExecuteTemplate(w, "pricing.html", map[string]interface{}{
		"User":     user,
		"Products": Products,
	})
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

	templates.ExecuteTemplate(w, "download.html", map[string]interface{}{
		"User":        user,
		"DownloadURL": downloadURL, // empty unless the user holds a valid license
		"HasLicense":  hasLicense,
		"LicenseKey":  licenseKey,
	})
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	templates.ExecuteTemplate(w, "about.html", map[string]interface{}{"User": user})
}

func featuresHandler(w http.ResponseWriter, r *http.Request) {
	user := getUser(r)
	templates.ExecuteTemplate(w, "features.html", map[string]interface{}{"User": user})
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
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Demo: auto-create or verify
	mu.Lock()
	var user *User
	for _, u := range users {
		if u.Email == req.Email {
			user = u
			break
		}
	}

	if user == nil {
		// Auto-register for demo
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
		user = &User{
			ID:       uuid.New().String(),
			Email:    req.Email,
			Password: string(hashed),
			Name:     strings.Split(req.Email, "@")[0],
			Avatar:   "https://ui-avatars.com/api/?name=" + req.Email + "&background=10b981&color=fff",
			JoinedAt: time.Now(),
		}
		users[user.ID] = user
	}
	mu.Unlock()

	token := uuid.New().String()
	mu.Lock()
	sessions[token] = user.ID
	mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400 * 7,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "user": user})
}

func apiRegister(w http.ResponseWriter, r *http.Request) {
	apiLogin(w, r) // Same logic for demo
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
	json.NewDecoder(r.Body).Decode(&req)

	mu.Lock()
	carts[user.ID] = append(carts[user.ID], req.ProductID)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
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
	json.NewEncoder(w).Encode(map[string]interface{}{
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
	json.NewEncoder(w).Encode(map[string]interface{}{
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Payment successful!",
		"redirect": "/download",
	})
}

// --- PayPal Checkout Integration ---

func checkoutPageHandler(w http.ResponseWriter, r *http.Request) {
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

	templates.ExecuteTemplate(w, "checkout.html", map[string]interface{}{
		"User":                 user,
		"Product":              product,
		"Savings":              savings,
		"PayPalClientID":       paypalClientID,
		"StripePublishableKey": stripePublishableKey,
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
	json.NewDecoder(r.Body).Decode(&req)

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
	w.Write(respBody)
}

func apiPayPalCaptureOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderID string `json:"order_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

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
	json.Unmarshal(respBody, &captureResult)

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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"order_id": req.OrderID,
			"status":   status,
		})
		return
	}

	log.Printf("PayPal capture status: %s, body: %s", status, string(respBody))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
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
