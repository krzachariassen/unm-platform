package handler

import (
	"sync"

	"github.com/krzachariassen/unm-platform/internal/adapter/repository"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// insightEntry is a cached insight result with its computation status.
type insightEntry struct {
	status   string // "computing", "ready", "failed"
	response InsightsResponse
}

// Handler holds all dependencies for HTTP request handling.
type Handler struct {
	cfg                entity.Config
	parseAndValidate   *usecase.ParseAndValidate
	fragmentation      *analyzer.FragmentationAnalyzer
	cognitiveLoad      *analyzer.CognitiveLoadAnalyzer
	dependency         *analyzer.DependencyAnalyzer
	gap                *analyzer.GapAnalyzer
	bottleneck         *analyzer.BottleneckAnalyzer
	coupling           *analyzer.CouplingAnalyzer
	complexity         *analyzer.ComplexityAnalyzer
	interactions       *analyzer.InteractionDiversityAnalyzer
	unlinked           *analyzer.UnlinkedCapabilityAnalyzer
	signalSuggestions  *analyzer.SignalSuggestionGenerator
	valueChain         *analyzer.ValueChainAnalyzer
	valueStream        *analyzer.ValueStreamAnalyzer
	changesetStore     *repository.ChangesetStore
	impactAnalyzer     *analyzer.ImpactAnalyzer
	aiClient           *ai.OpenAIClient // nil when API key not configured
	store              *repository.ModelStore
	insightCache       sync.Map // key: "modelId:domain" → insightEntry
}

// New constructs a Handler with all dependencies wired.
func New(
	cfg entity.Config,
	pv *usecase.ParseAndValidate,
	frag *analyzer.FragmentationAnalyzer,
	cl *analyzer.CognitiveLoadAnalyzer,
	dep *analyzer.DependencyAnalyzer,
	g *analyzer.GapAnalyzer,
	bn *analyzer.BottleneckAnalyzer,
	cp *analyzer.CouplingAnalyzer,
	cx *analyzer.ComplexityAnalyzer,
	intr *analyzer.InteractionDiversityAnalyzer,
	unl *analyzer.UnlinkedCapabilityAnalyzer,
	sg *analyzer.SignalSuggestionGenerator,
	vc *analyzer.ValueChainAnalyzer,
	vs *analyzer.ValueStreamAnalyzer,
	csStore *repository.ChangesetStore,
	ia *analyzer.ImpactAnalyzer,
	aiClient *ai.OpenAIClient,
	store *repository.ModelStore,
) *Handler {
	return &Handler{
		cfg:               cfg,
		parseAndValidate:  pv,
		fragmentation:     frag,
		cognitiveLoad:     cl,
		dependency:        dep,
		gap:               g,
		bottleneck:        bn,
		coupling:          cp,
		complexity:        cx,
		interactions:      intr,
		unlinked:          unl,
		signalSuggestions: sg,
		valueChain:        vc,
		valueStream:       vs,
		changesetStore:    csStore,
		impactAnalyzer:    ia,
		aiClient:          aiClient,
		store:             store,
	}
}
