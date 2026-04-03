package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

const sessionCookieName = "unm_session"
const stateCookieName = "oauth_state"
const sessionTTL = 24 * time.Hour

// contextKey is an unexported type for context keys in this package.
type contextKey string

const authUserKey contextKey = "authUser"

// AuthUserFromContext extracts the authenticated user from the context.
// Returns nil if no user is present (auth disabled or unauthenticated).
func AuthUserFromContext(ctx context.Context) *usecase.AuthUser {
	u, _ := ctx.Value(authUserKey).(*usecase.AuthUser)
	return u
}

// authHandler holds dependencies for authentication-related HTTP handlers.
type authHandler struct {
	cfg          entity.AuthConfig
	sessionStore usecase.SessionRepository
	oauthCfg     *oauth2.Config
	// userinfoURL and tokenURL can be overridden in tests.
	userinfoURL string
	tokenURL    string
}

func newAuthHandler(cfg entity.AuthConfig, sessionStore usecase.SessionRepository) *authHandler {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &authHandler{
		cfg:          cfg,
		sessionStore: sessionStore,
		oauthCfg:     oauthCfg,
		userinfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
	}
}

// generateState creates a random CSRF state token.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// handleGoogleLogin redirects the browser to Google's OAuth consent page.
func (h *Handler) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.authH == nil {
		http.Error(w, "auth not configured", http.StatusInternalServerError)
		return
	}
	state, err := generateState()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 min
	})
	url := h.authH.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

// handleGoogleCallback processes the OAuth2 callback from Google.
func (h *Handler) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.authH == nil {
		http.Error(w, "auth not configured", http.StatusInternalServerError)
		return
	}
	// Verify state.
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")

	// Exchange code for token, overriding the endpoint in tests.
	var cfg *oauth2.Config
	if h.authH.tokenURL != "" {
		cfg = &oauth2.Config{
			ClientID:     h.authH.oauthCfg.ClientID,
			ClientSecret: h.authH.oauthCfg.ClientSecret,
			RedirectURL:  h.authH.oauthCfg.RedirectURL,
			Scopes:       h.authH.oauthCfg.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  google.Endpoint.AuthURL,
				TokenURL: h.authH.tokenURL,
			},
		}
	} else {
		cfg = h.authH.oauthCfg
	}

	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("auth: token exchange error: %v", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	// Fetch user info.
	info, err := h.authH.fetchUserInfo(r.Context(), token.AccessToken)
	if err != nil {
		log.Printf("auth: userinfo error: %v", err)
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// Create session.
	sess, err := h.authH.sessionStore.Create(info.id, info.email, info.name, info.picture, sessionTTL)
	if err != nil {
		log.Printf("auth: session create error: %v", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	// Set session cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleLogout clears the session.
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && h.authH != nil {
		_ = h.authH.sessionStore.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusOK)
}

// handleMe returns the current authenticated user and their org memberships.
func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	user := AuthUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	type orgResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		Role string `json:"role"`
	}
	type meResp struct {
		ID        string    `json:"id"`
		Email     string    `json:"email"`
		Name      string    `json:"name"`
		AvatarURL string    `json:"avatar_url"`
		Orgs      []orgResp `json:"orgs"`
	}

	orgs := make([]orgResp, 0, len(user.Orgs))
	for _, o := range user.Orgs {
		orgs = append(orgs, orgResp{ID: o.OrgID, Name: o.OrgName, Slug: o.OrgSlug, Role: o.Role})
	}

	writeJSON(w, http.StatusOK, meResp{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Orgs:      orgs,
	})
}

// googleUserInfo holds raw fields from Google userinfo endpoint.
type googleUserInfo struct {
	id      string
	email   string
	name    string
	picture string
}

func (h *authHandler) fetchUserInfo(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", h.userinfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("parse userinfo: %w", err)
	}

	info := &googleUserInfo{}
	if v, ok := m["id"].(string); ok {
		info.id = v
	} else if v, ok := m["sub"].(string); ok {
		info.id = v
	}
	if v, ok := m["email"].(string); ok {
		info.email = v
	}
	if v, ok := m["name"].(string); ok {
		info.name = v
	}
	if v, ok := m["picture"].(string); ok {
		info.picture = v
	}
	if info.id == "" || info.email == "" {
		return nil, fmt.Errorf("userinfo missing required fields (id=%q email=%q)", info.id, info.email)
	}
	return info, nil
}

// makeAuthMiddleware returns middleware that enforces or bypasses authentication.
// When auth.enabled=true: requires a valid unm_session cookie; returns 401 otherwise.
// When auth.enabled=false: injects the AuthUser from the session if present, but never rejects.
// Always skips /health and /auth/* paths.
func makeAuthMiddleware(cfg entity.AuthConfig, sessions usecase.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip non-API paths.
			path := r.URL.Path
			if path == "/health" || strings.HasPrefix(path, "/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			// Try to load the session.
			var user *usecase.AuthUser
			if cookie, err := r.Cookie(sessionCookieName); err == nil {
				if sess, err := sessions.Get(cookie.Value); err == nil {
					user = &usecase.AuthUser{
						ID:        sess.UserID,
						Email:     sess.Email,
						Name:      sess.Name,
						AvatarURL: sess.AvatarURL,
					}
				}
			}

			if cfg.Enabled && user == nil {
				http.Error(w, `{"error":"unauthenticated"}`, http.StatusUnauthorized)
				return
			}

			if user != nil {
				ctx := context.WithValue(r.Context(), authUserKey, user)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// handleDevLogin creates a real session for the hardcoded dev user without
// going through Google OAuth. Only available when auth.dev_login=true.
// Works regardless of auth.enabled — lets developers test auth-enabled mode
// without Google credentials.
func (h *Handler) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	if h.authH == nil || !h.authH.cfg.DevLogin {
		http.Error(w, "dev login not available", http.StatusNotFound)
		return
	}
	const devUserID = "00000000-0000-0000-0000-000000000001"
	sess, err := h.authH.sessionStore.Create(devUserID, "local@dev", "Local Dev User", "", sessionTTL)
	if err != nil {
		log.Printf("auth: dev-login session create error: %v", err)
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sess.ID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	w.WriteHeader(http.StatusOK)
}

// makeDevModeMiddleware injects a hard-coded default user when auth.enabled=false.
// This allows all existing routes to work without any auth setup during development.
func makeDevModeMiddleware(cfg entity.AuthConfig) func(http.Handler) http.Handler {
	devUser := &usecase.AuthUser{
		ID:    "00000000-0000-0000-0000-000000000001",
		Email: "local@dev",
		Name:  "Local Dev User",
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Enabled {
				ctx := context.WithValue(r.Context(), authUserKey, devUser)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// registerAuthRoutes registers the OAuth and session routes.
func (h *Handler) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/google", h.handleGoogleLogin)
	mux.HandleFunc("GET /auth/callback", h.handleGoogleCallback)
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
	mux.HandleFunc("POST /auth/dev-login", h.handleDevLogin)
	mux.HandleFunc("GET /api/me", h.handleMe)
}
