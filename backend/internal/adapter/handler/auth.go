package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

const sessionCookieName = "unm_session"

// contextKey is an unexported type for context keys in this package.
type contextKey string

const authUserKey contextKey = "authUser"

// AuthUserFromContext extracts the authenticated user from the context.
// Returns nil if no user is present (auth disabled or unauthenticated).
func AuthUserFromContext(ctx context.Context) *usecase.AuthUser {
	u, _ := ctx.Value(authUserKey).(*usecase.AuthUser)
	return u
}

// authHandler is a placeholder until the full OAuth2 implementation from
// feat/phase-15a-auth is merged. Only the types and stub are needed here
// to allow router.go and handler.go to compile cleanly.
type authHandler struct {
	cfg          entity.AuthConfig
	sessionStore usecase.SessionRepository
}

func newAuthHandler(cfg entity.AuthConfig, sessionStore usecase.SessionRepository) *authHandler {
	return &authHandler{cfg: cfg, sessionStore: sessionStore}
}

// registerAuthRoutes registers stub auth endpoints.
// Full OAuth2 implementation lives on feat/phase-15a-auth.
func (h *Handler) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/logout", h.handleLogout)
	mux.HandleFunc("POST /auth/dev-login", h.handleDevLogin)
	mux.HandleFunc("GET /api/me", h.handleMe)
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

// handleMe returns the current authenticated user.
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

// handleDevLogin creates a real session for the hardcoded dev user.
func (h *Handler) handleDevLogin(w http.ResponseWriter, r *http.Request) {
	if h.authH == nil || !h.authH.cfg.DevLogin {
		http.Error(w, "dev login not available", http.StatusNotFound)
		return
	}
	const devUserID = "00000000-0000-0000-0000-000000000001"
	const sessionTTL = 24 * time.Hour
	sess, err := h.authH.sessionStore.Create(devUserID, "local@dev", "Local Dev User", "", sessionTTL)
	if err != nil {
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

// makeAuthMiddleware returns middleware that enforces or bypasses authentication.
// When auth.enabled=true: requires a valid unm_session cookie; returns 401 otherwise.
// When auth.enabled=false: injects the AuthUser from the session if present, but never rejects.
// Always skips /health and /auth/* paths.
func makeAuthMiddleware(cfg entity.AuthConfig, sessions usecase.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/health" || strings.HasPrefix(path, "/auth/") {
				next.ServeHTTP(w, r)
				return
			}

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

// makeDevModeMiddleware injects a hardcoded dev user when auth.enabled=false.
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
