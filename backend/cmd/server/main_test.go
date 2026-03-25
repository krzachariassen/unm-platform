package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/uber/unm-platform/internal/adapter/handler"
	"github.com/uber/unm-platform/internal/adapter/repository"
	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/domain/service"
	"github.com/uber/unm-platform/internal/infrastructure/analyzer"
	"github.com/uber/unm-platform/internal/infrastructure/parser"
	"github.com/uber/unm-platform/internal/usecase"
)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	cfg := entity.DefaultConfig()
	store := repository.NewModelStore()
	h := handler.New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		repository.NewChangesetStore(),
		analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		nil, // aiClient
		store,
	)
	return handler.NewRouter(h, cfg)
}

func TestHealthCheck(t *testing.T) {
	srv := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestCORSHeaders(t *testing.T) {
	srv := newTestRouter(t)
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin == "" {
		t.Error("expected Access-Control-Allow-Origin header to be set")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", w.Code)
	}
}

func TestUnknownRoute(t *testing.T) {
	srv := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown route, got %d", w.Code)
	}
}

func TestNewRouter_DebugRoutesDisabled_Returns404(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Features.DebugRoutes = false

	store := repository.NewModelStore()
	h := handler.New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		repository.NewChangesetStore(),
		analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		nil,
		store,
	)
	srv := handler.NewRouter(h, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/debug/load-example", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when debug routes disabled, got %d", w.Code)
	}
}

func TestNewRouter_DebugRoutesEnabled_Registers(t *testing.T) {
	cfg := entity.DefaultConfig()
	cfg.Features.DebugRoutes = true

	store := repository.NewModelStore()
	h := handler.New(
		cfg,
		usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine()),
		analyzer.NewFragmentationAnalyzer(),
		analyzer.NewCognitiveLoadAnalyzer(cfg.Analysis.CognitiveLoad, cfg.Analysis.InteractionWeights),
		analyzer.NewDependencyAnalyzer(),
		analyzer.NewGapAnalyzer(),
		analyzer.NewBottleneckAnalyzer(cfg.Analysis.Bottleneck),
		analyzer.NewCouplingAnalyzer(),
		analyzer.NewComplexityAnalyzer(),
		analyzer.NewInteractionDiversityAnalyzer(cfg.Analysis.Signals),
		analyzer.NewUnlinkedCapabilityAnalyzer(),
		analyzer.NewSignalSuggestionGenerator(cfg.Analysis.Signals),
		analyzer.NewValueChainAnalyzer(cfg.Analysis.ValueChain),
		analyzer.NewValueStreamAnalyzer(),
		repository.NewChangesetStore(),
		analyzer.NewImpactAnalyzer(entity.DefaultConfig().Analysis),
		nil,
		store,
	)
	srv := handler.NewRouter(h, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/debug/load-example", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	// Should NOT be 404 — the route is registered (it might fail to find the file, but that's 500, not 404)
	if w.Code == http.StatusNotFound {
		t.Error("expected debug route to be registered when DebugRoutes=true, got 404")
	}
}
