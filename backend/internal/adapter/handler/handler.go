package handler

import (
	"log"
	"sync"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// insightEntry is a cached insight result with its computation status.
type insightEntry struct {
	status   string // "computing", "ready", "failed"
	response InsightsResponse
}

// HandlerDeps groups all dependencies for the Handler. Using a struct instead
// of positional parameters means tests only construct what they need (zero
// values are safe defaults) and adding a new dependency is a one-line change.
type HandlerDeps struct {
	Config              entity.Config
	ParseAndValidate    *usecase.ParseAndValidate
	ParseAndValidateDSL *usecase.ParseAndValidate // nil → built from parser.NewDSLParser() in New()
	Fragmentation       *analyzer.FragmentationAnalyzer
	CognitiveLoad       *analyzer.CognitiveLoadAnalyzer
	Dependency          *analyzer.DependencyAnalyzer
	Gap                 *analyzer.GapAnalyzer
	Bottleneck          *analyzer.BottleneckAnalyzer
	Coupling            *analyzer.CouplingAnalyzer
	Complexity          *analyzer.ComplexityAnalyzer
	Interactions        *analyzer.InteractionDiversityAnalyzer
	Unlinked            *analyzer.UnlinkedCapabilityAnalyzer
	SignalSuggestions   *analyzer.SignalSuggestionGenerator
	ValueChain          *analyzer.ValueChainAnalyzer
	ValueStream         *analyzer.ValueStreamAnalyzer
	ChangesetStore      usecase.ChangesetRepository
	ImpactAnalyzer      *analyzer.ImpactAnalyzer
	AIClient            *ai.OpenAIClient // nil when API key not configured
	Store               usecase.ModelRepository
}

// modelHandler groups the dependencies needed for model CRUD operations.
type modelHandler struct {
	store               usecase.ModelRepository
	parseAndValidate    *usecase.ParseAndValidate
	parseAndValidateDSL *usecase.ParseAndValidate
}

// changesetHandler groups the dependencies needed for changeset operations.
type changesetHandler struct {
	store          usecase.ModelRepository
	changesetStore usecase.ChangesetRepository
	impactAnalyzer *analyzer.ImpactAnalyzer
	insightCache   *sync.Map
	cfg            entity.Config
}

// viewHandler groups the dependencies needed for view projection.
type viewHandler struct {
	store usecase.ModelRepository
	cfg   entity.Config
}

// aiHandler groups the dependencies needed for AI advisor and insights.
type aiHandler struct {
	store          usecase.ModelRepository
	aiClient       *ai.OpenAIClient
	changesetStore usecase.ChangesetRepository
	promptRenderer *ai.PromptRenderer
	insightCache   *sync.Map
	cfg            entity.Config
}

// Handler is a thin router that holds focused sub-handlers and delegates to them.
// All HTTP handler methods live in their domain-specific files (model.go, changeset.go, etc.)
// and access dependencies through the appropriate sub-handler field.
type Handler struct {
	cfg entity.Config

	// Sub-handlers — each holds only the deps it needs.
	model *modelHandler
	cs    *changesetHandler
	view  *viewHandler
	aiH   *aiHandler

	// Analyzers used by signals, analysis, and AI context building.
	fragmentation     *analyzer.FragmentationAnalyzer
	cognitiveLoad     *analyzer.CognitiveLoadAnalyzer
	dependency        *analyzer.DependencyAnalyzer
	gap               *analyzer.GapAnalyzer
	bottleneck        *analyzer.BottleneckAnalyzer
	coupling          *analyzer.CouplingAnalyzer
	complexity        *analyzer.ComplexityAnalyzer
	interactions      *analyzer.InteractionDiversityAnalyzer
	unlinked          *analyzer.UnlinkedCapabilityAnalyzer
	signalSuggestions *analyzer.SignalSuggestionGenerator
	valueChain        *analyzer.ValueChainAnalyzer
	valueStream       *analyzer.ValueStreamAnalyzer

	// Convenience accessors delegated to sub-handlers (kept for backward-compat
	// across handler methods that reference h.store, h.aiClient, etc. directly).
	store          usecase.ModelRepository
	changesetStore usecase.ChangesetRepository
	impactAnalyzer *analyzer.ImpactAnalyzer
	aiClient       *ai.OpenAIClient
	insightCache   sync.Map

	// Singletons built once at startup.
	promptRenderer *ai.PromptRenderer
	runner         *usecase.AnalysisRunner

	// parser/validate (kept for direct access in model.go methods).
	parseAndValidate    *usecase.ParseAndValidate
	parseAndValidateDSL *usecase.ParseAndValidate
}

// New constructs a Handler from a HandlerDeps struct.
func New(deps HandlerDeps) *Handler {
	lib, err := ai.NewPromptLibrary()
	if err != nil {
		log.Fatalf("handler: failed to load prompt library: %v", err)
	}
	promptRenderer := ai.NewPromptRenderer(lib)

	h := &Handler{
		cfg: deps.Config,

		// Analyzers (used by signals, analysis, AI context).
		fragmentation:     deps.Fragmentation,
		cognitiveLoad:     deps.CognitiveLoad,
		dependency:        deps.Dependency,
		gap:               deps.Gap,
		bottleneck:        deps.Bottleneck,
		coupling:          deps.Coupling,
		complexity:        deps.Complexity,
		interactions:      deps.Interactions,
		unlinked:          deps.Unlinked,
		signalSuggestions: deps.SignalSuggestions,
		valueChain:        deps.ValueChain,
		valueStream:       deps.ValueStream,

		// Convenience flat accessors (shared across many handlers).
		store:          deps.Store,
		changesetStore: deps.ChangesetStore,
		impactAnalyzer: deps.ImpactAnalyzer,
		aiClient:       deps.AIClient,
		promptRenderer: promptRenderer,

		parseAndValidate:    deps.ParseAndValidate,
		parseAndValidateDSL: deps.ParseAndValidateDSL,
	}

	// Build sub-handlers by composing from HandlerDeps.
	h.model = &modelHandler{
		store:               deps.Store,
		parseAndValidate:    deps.ParseAndValidate,
		parseAndValidateDSL: deps.ParseAndValidateDSL,
	}

	h.cs = &changesetHandler{
		store:          deps.Store,
		changesetStore: deps.ChangesetStore,
		impactAnalyzer: deps.ImpactAnalyzer,
		insightCache:   &h.insightCache,
		cfg:            deps.Config,
	}

	h.view = &viewHandler{
		store: deps.Store,
		cfg:   deps.Config,
	}

	h.aiH = &aiHandler{
		store:          deps.Store,
		aiClient:       deps.AIClient,
		changesetStore: deps.ChangesetStore,
		promptRenderer: promptRenderer,
		insightCache:   &h.insightCache,
		cfg:            deps.Config,
	}

	if h.parseAndValidateDSL == nil {
		h.parseAndValidateDSL = usecase.NewParseAndValidate(parser.NewDSLParser(), service.NewValidationEngine())
		h.model.parseAndValidateDSL = h.parseAndValidateDSL
	}

	if deps.Fragmentation != nil {
		h.runner = usecase.NewAnalysisRunner(
			deps.Fragmentation,
			deps.CognitiveLoad,
			deps.Dependency,
			deps.Gap,
			deps.Bottleneck,
			deps.Coupling,
			deps.Complexity,
			deps.Interactions,
			deps.Unlinked,
			deps.SignalSuggestions,
		)
	}

	return h
}
