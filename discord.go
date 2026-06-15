package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Discord OAuth2 config. Routes register only when client id + secret + redirect
// are all set (graceful degradation, same as Stripe/PayPal).
var (
	discordClientID     = getEnv("DISCORD_CLIENT_ID", "")
	discordClientSecret = getEnv("DISCORD_CLIENT_SECRET", "")
	discordRedirectURI  = getEnv("DISCORD_REDIRECT_URI", "")
)

func discordEnabled() bool {
	return discordClientID != "" && discordClientSecret != "" && discordRedirectURI != ""
}

const (
	discordAuthorizeURL = "https://discord.com/api/oauth2/authorize"
	discordTokenURL     = "https://discord.com/api/oauth2/token"
	discordUserURL      = "https://discord.com/api/users/@me"
)

// discordLoginHandler kicks off the OAuth dance: set a CSRF state cookie and
// redirect to Discord's consent screen (identify + email scopes).
func discordLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	http.SetCookie(w, &http.Cookie{
		Name:     "discord_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
		MaxAge:   600,
	})
	q := url.Values{}
	q.Set("client_id", discordClientID)
	q.Set("redirect_uri", discordRedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "identify email")
	q.Set("state", state)
	http.Redirect(w, r, discordAuthorizeURL+"?"+q.Encode(), http.StatusSeeOther)
}

type discordUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
}

// discordCallbackHandler completes OAuth: verify state, exchange the code for a
// token, fetch the Discord profile, then create/find the local user (keyed by
// the Discord email) with their Discord username as the default referral code,
// and start a session.
func discordCallbackHandler(w http.ResponseWriter, r *http.Request) {
	setReferralCookieIfPresent(w, r)

	// CSRF: the returned state must match the cookie we set.
	stateCookie, err := r.Cookie("discord_state")
	if err != nil || stateCookie.Value == "" || r.URL.Query().Get("state") != stateCookie.Value {
		http.Redirect(w, r, "/login?error=discord_state", http.StatusSeeOther)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "discord_state", Value: "", Path: "/", MaxAge: -1})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, "/login?error=discord_code", http.StatusSeeOther)
		return
	}

	du, err := discordExchange(code)
	if err != nil {
		log.Printf("discord callback: %v", err)
		http.Redirect(w, r, "/login?error=discord_auth", http.StatusSeeOther)
		return
	}
	email := strings.ToLower(strings.TrimSpace(du.Email))
	if email == "" {
		// The account has no email (or didn't grant the scope) — we key users by
		// email, so we can't proceed.
		http.Redirect(w, r, "/login?error=discord_no_email", http.StatusSeeOther)
		return
	}

	name := du.GlobalName
	if name == "" {
		name = du.Username
	}

	// Find or create the in-memory user (keyed by email), then persist durably.
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
			ID:       uuid.New().String(),
			Email:    email,
			Name:     name,
			Avatar:   "https://ui-avatars.com/api/?name=" + url.QueryEscape(name) + "&background=10b981&color=fff",
			JoinedAt: time.Now(),
		}
		users[user.ID] = user
	}
	mu.Unlock()

	// Persist; the Discord username becomes the default referral code on first
	// creation (falls back to random if taken/invalid).
	if userRepo != nil {
		pu, err := userRepo.EnsureUserWithCode(r.Context(), user.ID, email, name, du.Username)
		if err != nil {
			log.Printf("discord callback: persist user %s: %v", email, err)
		} else if pu != nil {
			// Record the referrer (from ?ref= cookie). Server-side capture or client hook
			// will have set the cookie if the user arrived via a referral link.
			if refCookie, _ := r.Cookie("ref"); refCookie != nil && refCookie.Value != "" {
				if err := userRepo.SetReferredBy(r.Context(), pu.ID, refCookie.Value); err != nil {
					log.Printf("discord callback: set referred_by_code for %s: %v", email, err)
				}
			}
		}
	}

	startSession(w, r, user.ID)
	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

// discordExchange swaps an auth code for a token and returns the Discord profile.
func discordExchange(code string) (*discordUser, error) {
	form := url.Values{}
	form.Set("client_id", discordClientID)
	form.Set("client_secret", discordClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", discordRedirectURI)

	resp, err := http.PostForm(discordTokenURL, form)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint status %d", resp.StatusCode)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	if tok.AccessToken == "" {
		return nil, fmt.Errorf("empty access token")
	}

	req, err := http.NewRequest(http.MethodGet, discordUserURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	uresp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user request: %w", err)
	}
	defer uresp.Body.Close()
	if uresp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user endpoint status %d", uresp.StatusCode)
	}
	var du discordUser
	if err := json.NewDecoder(uresp.Body).Decode(&du); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &du, nil
}
