package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubSessionStore is an in-memory session store for testing.
type stubSessionStore struct {
	sessions map[string]*usecase.UserSession
}

func newStubSessionStore() *stubSessionStore {
	return &stubSessionStore{sessions: make(map[string]*usecase.UserSession)}
}

func (s *stubSessionStore) Create(userID, email, name, avatarURL string, ttl time.Duration) (*usecase.UserSession, error) {
	id := fmt.Sprintf("sess-%s", userID)
	sess := &usecase.UserSession{
		ID:        id,
		UserID:    userID,
		Email:     email,
		Name:      name,
		AvatarURL: avatarURL,
		ExpiresAt: time.Now().Add(ttl),
	}
	s.sessions[id] = sess
	return sess, nil
}

func (s *stubSessionStore) Get(id string) (*usecase.UserSession, error) {
	sess, ok := s.sessions[id]
	if !ok {
		return nil, usecase.ErrNotFound
	}
	if time.Now().After(sess.ExpiresAt) {
		return nil, usecase.ErrNotFound
	}
	return sess, nil
}

func (s *stubSessionStore) Delete(id string) error {
	delete(s.sessions, id)
	return nil
}

// stubUserinfoServer serves fake Google userinfo JSON.
func stubUserinfoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/token") {
			// Fake token exchange endpoint.
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"access_token":"fake-token","token_type":"Bearer","expires_in":3600}`)
			return
		}
		// Fake userinfo endpoint.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"id":"google-sub-123","email":"test@example.com","name":"Test User","picture":"https://example.com/pic.jpg"}`)
	}))
}

func newAuthTestHandler(t *testing.T, cfg entity.Config, sessionStore usecase.SessionRepository) *Handler {
	t.Helper()
	deps := HandlerDeps{
		Config:       cfg,
		SessionStore: sessionStore,
		Store:        repository.NewModelStore(),
	}
	return New(deps)
}

// TestHandleGoogleLogin_Redirect verifies the login endpoint redirects to Google with state.
func TestHandleGoogleLogin_Redirect(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true
	cfg.Auth.Google.ClientID = "test-client-id"
	cfg.Auth.Google.ClientSecret = "test-client-secret"
	cfg.Auth.Google.RedirectURL = "http://localhost:8080/auth/callback"

	h := newAuthTestHandler(t, cfg, newStubSessionStore())

	req := httptest.NewRequest("GET", "/auth/google", nil)
	rec := httptest.NewRecorder()
	h.handleGoogleLogin(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	loc := rec.Header().Get("Location")
	assert.Contains(t, loc, "accounts.google.com")
	assert.Contains(t, loc, "state=")

	// State cookie must be set.
	cookies := rec.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			stateCookie = c
		}
	}
	require.NotNil(t, stateCookie, "oauth_state cookie must be set")
	assert.True(t, stateCookie.HttpOnly)
}

// TestHandleGoogleCallback_Success verifies a valid callback creates a session.
func TestHandleGoogleCallback_Success(t *testing.T) {
	stub := stubUserinfoServer(t)
	defer stub.Close()

	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true
	cfg.Auth.Google.ClientID = "test-client-id"
	cfg.Auth.Google.ClientSecret = "test-client-secret"
	cfg.Auth.Google.RedirectURL = "http://localhost:8080/auth/callback"

	sessStore := newStubSessionStore()
	h := newAuthTestHandler(t, cfg, sessStore)
	// Override the userinfo URL to use the stub.
	h.authH.userinfoURL = stub.URL + "/userinfo"
	h.authH.tokenURL = stub.URL + "/token"

	state := "test-state-value"
	req := httptest.NewRequest("GET", "/auth/callback?code=fake-code&state="+state, nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
	rec := httptest.NewRecorder()
	h.handleGoogleCallback(rec, req)

	// Should redirect to frontend.
	assert.Equal(t, http.StatusFound, rec.Code)
	// Session cookie must be set.
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "unm_session" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie, "unm_session cookie must be set")
	assert.True(t, sessionCookie.HttpOnly)
}

// TestHandleGoogleCallback_StateMismatch returns 400 when state doesn't match.
func TestHandleGoogleCallback_StateMismatch(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true
	cfg.Auth.Google.ClientID = "test-client-id"

	h := newAuthTestHandler(t, cfg, newStubSessionStore())

	req := httptest.NewRequest("GET", "/auth/callback?code=fake-code&state=bad-state", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "good-state"})
	rec := httptest.NewRecorder()
	h.handleGoogleCallback(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandleLogout_ClearsSession verifies logout deletes the session and clears the cookie.
func TestHandleLogout_ClearsSession(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true

	sessStore := newStubSessionStore()
	// Pre-seed a session.
	sess, err := sessStore.Create("user-1", "u@example.com", "U", "", 24*time.Hour)
	require.NoError(t, err)

	h := newAuthTestHandler(t, cfg, sessStore)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "unm_session", Value: sess.ID})
	rec := httptest.NewRecorder()
	h.handleLogout(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Session must be gone.
	_, err = sessStore.Get(sess.ID)
	assert.ErrorIs(t, err, usecase.ErrNotFound)

	// Cookie must be cleared.
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "unm_session" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, "", sessionCookie.Value)
	assert.True(t, sessionCookie.MaxAge < 0 || sessionCookie.Expires.Before(time.Now()))
}

// TestHandleMe_AuthEnabled returns user info when session is valid.
func TestHandleMe_AuthEnabled(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true

	sessStore := newStubSessionStore()
	sess, err := sessStore.Create("user-1", "me@example.com", "Me", "https://pic.jpg", 24*time.Hour)
	require.NoError(t, err)

	h := newAuthTestHandler(t, cfg, sessStore)

	// Inject auth user via context (as middleware would).
	req := httptest.NewRequest("GET", "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "unm_session", Value: sess.ID})

	// Apply auth middleware manually to populate context.
	authMw := makeAuthMiddleware(cfg.Auth, sessStore)
	var called bool
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		h.handleMe(w, r)
	})
	rec := httptest.NewRecorder()
	authMw(inner).ServeHTTP(rec, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "me@example.com", body.Email)
	assert.Equal(t, "Me", body.Name)
}

// TestAuthMiddleware_401_WhenEnabled returns 401 for requests with no valid session.
func TestAuthMiddleware_401_WhenEnabled(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true

	sessStore := newStubSessionStore()
	mw := makeAuthMiddleware(cfg.Auth, sessStore)

	req := httptest.NewRequest("GET", "/api/models", nil)
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestAuthMiddleware_PassThrough_WhenDisabled does not reject requests when auth.enabled=false.
func TestAuthMiddleware_PassThrough_WhenDisabled(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = false

	sessStore := newStubSessionStore()
	mw := makeAuthMiddleware(cfg.Auth, sessStore)

	req := httptest.NewRequest("GET", "/api/models", nil)
	rec := httptest.NewRecorder()
	var reached bool
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	assert.True(t, reached)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHandleDevLogin_CreatesSession verifies dev-login creates a session when dev_login=true.
func TestHandleDevLogin_CreatesSession(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = true  // enabled — yet dev login still works
	cfg.Auth.DevLogin = true

	sessStore := newStubSessionStore()
	h := newAuthTestHandler(t, cfg, sessStore)

	req := httptest.NewRequest("POST", "/auth/dev-login", nil)
	rec := httptest.NewRecorder()
	h.handleDevLogin(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Session cookie must be set.
	var sessionCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "unm_session" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie, "unm_session cookie must be set")
	assert.True(t, sessionCookie.HttpOnly)

	// Session must exist in the store.
	sess, err := sessStore.Get(sessionCookie.Value)
	require.NoError(t, err)
	assert.Equal(t, "local@dev", sess.Email)
}

// TestHandleDevLogin_NotAvailable verifies dev-login returns 404 when disabled.
func TestHandleDevLogin_NotAvailable(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.DevLogin = false

	h := newAuthTestHandler(t, cfg, newStubSessionStore())

	req := httptest.NewRequest("POST", "/auth/dev-login", nil)
	rec := httptest.NewRecorder()
	h.handleDevLogin(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDevModeMiddleware_InjectsDefaultUser(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Auth.Enabled = false

	mw := makeDevModeMiddleware(cfg.Auth)

	req := httptest.NewRequest("GET", "/api/models", nil)
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := AuthUserFromContext(r.Context())
		require.NotNil(t, u)
		assert.Equal(t, "local@dev", u.Email)
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
