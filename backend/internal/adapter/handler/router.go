package handler

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
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

	h.registerModelRoutes(mux)
	h.registerChangesetRoutes(mux)
	h.registerAnalysisRoutes(mux)
	h.registerQueryRoutes(mux)
	h.registerViewRoutes(mux)
	h.registerAnalyzerViewRoutes(mux)
	h.registerSignalsRoutes(mux)
	h.registerAIRoutes(mux)
	h.registerInsightsRoutes(mux)

	if cfg.Features.DebugRoutes {
		h.registerDebugRoutes(mux)
	}

	if cfg.Server.StaticDir != "" {
		mux.Handle("/", spaFileServer(cfg.Server.StaticDir))
	}

	cors := makeCORSMiddleware(cfg.Server.CORSOrigins)
	return chain(mux, recoveryMiddleware, loggingMiddleware, cors)
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
