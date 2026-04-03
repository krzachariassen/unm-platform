package handler

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// NewRouter registers all routes and wraps the mux in middleware.
// The config controls CORS origins, debug route registration, and
// optional static file serving (for Docker/production single-binary mode).
func NewRouter(h *Handler, cfgs ...entity.Config) http.Handler {
	var cfg entity.Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	} else {
		cfg = entity.DefaultConfig()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /api/config", h.handleGetConfig)

	// Auth routes (public — not behind auth middleware).
	if h.authH != nil {
		h.registerAuthRoutes(mux)
	}

	h.registerModelRoutes(mux)
	h.registerChangesetRoutes(mux)
	h.registerAnalysisRoutes(mux)
	h.registerQueryRoutes(mux)
	h.registerViewRoutes(mux)
	h.registerAnalyzerViewRoutes(mux)
	h.registerSignalsRoutes(mux)
	h.registerAIRoutes(mux)
	h.registerInsightsRoutes(mux)

	// Org + workspace management routes (registered when org store is configured).
	if h.orgH != nil {
		h.registerOrgRoutes(mux)
	}

	if cfg.Features.DebugRoutes {
		h.registerDebugRoutes(mux)
	}

	if cfg.Server.StaticDir != "" {
		mux.Handle("/", spaFileServer(cfg.Server.StaticDir))
	}

	cors := makeCORSMiddleware(cfg.Server.CORSOrigins)

	// Build session store for middleware (may be nil when auth not configured).
	var sessionStore usecase.SessionRepository
	if h.authH != nil {
		sessionStore = h.authH.sessionStore
	} else {
		// Fallback no-op store — never reached in normal flow.
		sessionStore = &noopSessionStore{}
	}

	authMw := makeAuthMiddleware(cfg.Auth, sessionStore)
	devModeMw := makeDevModeMiddleware(cfg.Auth)

	return chain(mux, recoveryMiddleware, loggingMiddleware, cors, authMw, devModeMw)
}

// spaFileServer serves static files from dir. If the requested file does not
// exist it falls back to index.html, which lets the React SPA handle routing.
func spaFileServer(dir string) http.Handler {
	fsys := os.DirFS(dir)
	fileServer := http.FileServerFS(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(fsys, filepath.Clean(path)); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

// noopSessionStore is a do-nothing SessionRepository used when auth is not configured.
// It always returns ErrNotFound, ensuring no sessions are found (safe default).
type noopSessionStore struct{}

func (n *noopSessionStore) Create(_, _, _, _ string, _ time.Duration) (*usecase.UserSession, error) {
	return nil, usecase.ErrNotFound
}
func (n *noopSessionStore) Get(_ string) (*usecase.UserSession, error) { return nil, usecase.ErrNotFound }
func (n *noopSessionStore) Delete(_ string) error                      { return nil }

