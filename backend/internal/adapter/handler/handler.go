package handler

import (
	"log"
	"sync"

	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
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
	Config            entity.Config
	ParseAndValidate  *usecase.ParseAndValidate
	ParseAndValidateDSL *usecase.ParseAndValidate // nil → built from parser.NewDSLParser() in New()
	Fragmentation     *analyzer.FragmentationAnalyzer
	CognitiveLoad     *analyzer.CognitiveLoadAnalyzer
	Dependency        *analyzer.DependencyAnalyzer
	Gap               *analyzer.GapAnalyzer
	Bottleneck        *analyzer.BottleneckAnalyzer
	Coupling          *analyzer.CouplingAnalyzer
	Complexity        *analyzer.ComplexityAnalyzer
	Interactions      *analyzer.InteractionDiversityAnalyzer
	Unlinked          *analyzer.UnlinkedCapabilityAnalyzer
	SignalSuggestions *analyzer.SignalSuggestionGenerator
	ValueChain        *analyzer.ValueChainAnalyzer
	ValueStream       *analyzer.ValueStreamAnalyzer
	ChangesetStore    *repository.ChangesetStore
	ImpactAnalyzer    *analyzer.ImpactAnalyzer
	AIClient          *ai.OpenAIClient // nil when API key not configured
	Store             *repository.ModelStore
}

// Handler holds all dependencies for HTTP request handling.
type Handler struct {
	cfg               entity.Config
	parseAndValidate  *usecase.ParseAndValidate
	parseAndValidateDSL *usecase.ParseAndValidate
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
	changesetStore    *repository.ChangesetStore
	impactAnalyzer    *analyzer.ImpactAnalyzer
	aiClient          *ai.OpenAIClient // nil when API key not configured
	store             *repository.ModelStore
	insightCache      sync.Map // key: "modelId:domain" → insightEntry

	// Singletons built once at startup.
	promptRenderer *ai.PromptRenderer
	runner         *usecase.AnalysisRunner
}

// New constructs a Handler from a HandlerDeps struct.
func New(deps HandlerDeps) *Handler {
	lib, err := ai.NewPromptLibrary()
	if err != nil {
		log.Fatalf("handler: failed to load prompt library: %v", err)
	}

	h := &Handler{
		cfg:                 deps.Config,
		parseAndValidate:    deps.ParseAndValidate,
		parseAndValidateDSL: deps.ParseAndValidateDSL,
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
		changesetStore:    deps.ChangesetStore,
		impactAnalyzer:    deps.ImpactAnalyzer,
		aiClient:          deps.AIClient,
		store:             deps.Store,
		promptRenderer:    ai.NewPromptRenderer(lib),
	}

	if h.parseAndValidateDSL == nil {
		h.parseAndValidateDSL = usecase.NewParseAndValidate(parser.NewDSLParser(), service.NewValidationEngine())
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

