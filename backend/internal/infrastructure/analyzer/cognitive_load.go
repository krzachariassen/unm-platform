package analyzer

import (
	"math"
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)

// LoadLevel represents a traffic-light assessment of a single dimension.
type LoadLevel string

const (
	LoadLow    LoadLevel = "low"
	LoadMedium LoadLevel = "medium"
	LoadHigh   LoadLevel = "high"
)

// InteractionWeight returns the cognitive overhead of an interaction mode,
// derived from Team Topologies: collaboration requires tight coupling and
// continuous coordination (high), facilitating is temporary but needs attention
// (medium), x-as-a-service has well-defined boundaries (low).
// When called with default weights (Collaboration=3, Facilitating=2, XAsAService=1).
func InteractionWeight(mode valueobject.InteractionMode) int {
	return InteractionWeightWithConfig(mode, entity.DefaultConfig().Analysis.InteractionWeights)
}

// InteractionWeightWithConfig returns the cognitive overhead for mode using the given weights.
func InteractionWeightWithConfig(mode valueobject.InteractionMode, w entity.InteractionWeightConfig) int {
	switch mode {
	case valueobject.Collaboration:
		return w.Collaboration
	case valueobject.Facilitating:
		return w.Facilitating
	case valueobject.XAsAService:
		return w.XAsAService
	default:
		return w.XAsAService
	}
}

// LoadDimension is a single axis of the structural load assessment.
// Value holds the assessed metric: raw count for most dimensions,
// per-person ratio for ServiceLoad.
type LoadDimension struct {
	Value float64   `json:"value"`
	Level LoadLevel `json:"level"`
}

// TeamLoad holds the multi-dimensional structural load assessment for a team.
// Each dimension is assessed independently as low/medium/high. The composite
// OverallLevel is the worst (highest) of the four dimensions.
type TeamLoad struct {
	Team *entity.Team

	// Raw counts
	CapabilityCount  int
	ServiceCount     int
	DependencyCount  int
	InteractionCount int // raw count of interactions (unweighted)

	// Interaction score weighted by mode (collab=3, facilitating=2, xaas=1)
	InteractionScore int

	// Team metadata
	TeamSize       int
	SizeIsExplicit bool

	// Four-dimension assessment
	DomainSpread    LoadDimension // capability count
	ServiceLoad     LoadDimension // services per person
	InteractionLoad LoadDimension // weighted interaction score
	DependencyLoad  LoadDimension // outbound dependency count

	// Composite
	OverallLevel LoadLevel
}

// CognitiveLoadReport holds the result of a cognitive load analysis pass.
type CognitiveLoadReport struct {
	TeamLoads []TeamLoad
}

// CognitiveLoadAnalyzer computes structural load metrics per team using a
// multi-dimensional model inspired by Team Topologies. It does NOT compute
// a single "percentage" — cognitive load is inherently multi-faceted.
type CognitiveLoadAnalyzer struct {
	cfg     entity.CognitiveLoadConfig
	weights entity.InteractionWeightConfig
}

func NewCognitiveLoadAnalyzer(cfg entity.CognitiveLoadConfig, weights entity.InteractionWeightConfig) *CognitiveLoadAnalyzer {
	return &CognitiveLoadAnalyzer{cfg: cfg, weights: weights}
}

// Analyze computes a CognitiveLoadReport for every team in m.
//
// Four dimensions (each assessed as low/medium/high):
//
//	Domain Spread:     capability count            → low(1-3), medium(4-5), high(6+)
//	Service Load:      services ÷ team_size        → low(≤2), medium(2-3), high(>3)
//	Interaction Load:  Σ interaction_weight(mode)   → low(≤3), medium(4-6), high(7+)
//	Dependency Load:   outbound dependency count    → low(≤4), medium(5-8), high(9+)
//
// OverallLevel = worst of the four dimensions.
// Teams are sorted by overall severity (high first), then by composite score.
func (a *CognitiveLoadAnalyzer) Analyze(m *entity.UNMModel) CognitiveLoadReport {
	if len(m.Teams) == 0 {
		return CognitiveLoadReport{}
	}

	// Precompute weighted interaction scores per team
	interactionScores := make(map[string]int)
	interactionCounts := make(map[string]int)
	for _, ix := range m.Interactions {
		w := InteractionWeightWithConfig(ix.Mode, a.weights)
		interactionScores[ix.FromTeamName] += w
		interactionScores[ix.ToTeamName] += w
		interactionCounts[ix.FromTeamName]++
		interactionCounts[ix.ToTeamName]++
	}

	var loads []TeamLoad
	for _, team := range m.Teams {
		capCount := len(team.Owns)
		svcCount := 0
		depCount := 0
		for _, svc := range m.Services {
			if svc.OwnerTeamName == team.Name {
				svcCount++
				depCount += len(svc.DependsOn)
			}
		}
		ixScore := interactionScores[team.Name]
		ixCount := interactionCounts[team.Name]
		teamSize := team.EffectiveSize()
		sizeExplicit := team.SizeExplicit

		// Assess each dimension using config thresholds
		domainSpread := assessDomainSpreadCfg(capCount, a.cfg.DomainSpreadThresholds)
		serviceLoad := assessServiceLoadCfg(svcCount, teamSize, a.cfg.ServiceLoadThresholds)
		interactionLoad := assessInteractionLoadCfg(ixScore, a.cfg.InteractionLoadThresholds)
		dependencyLoad := assessDependencyLoadCfg(depCount, a.cfg.DependencyLoadThresholds)

		overall := worstLevel(domainSpread.Level, serviceLoad.Level, interactionLoad.Level, dependencyLoad.Level)

		loads = append(loads, TeamLoad{
			Team:             team,
			CapabilityCount:  capCount,
			ServiceCount:     svcCount,
			DependencyCount:  depCount,
			InteractionCount: ixCount,
			InteractionScore: ixScore,
			TeamSize:         teamSize,
			SizeIsExplicit:   sizeExplicit,
			DomainSpread:     domainSpread,
			ServiceLoad:      serviceLoad,
			InteractionLoad:  interactionLoad,
			DependencyLoad:   dependencyLoad,
			OverallLevel:     overall,
		})
	}

	sort.Slice(loads, func(i, j int) bool {
		li := levelRank(loads[i].OverallLevel)
		lj := levelRank(loads[j].OverallLevel)
		if li != lj {
			return li > lj
		}
		return compositeScore(loads[i]) > compositeScore(loads[j])
	})

	return CognitiveLoadReport{TeamLoads: loads}
}

func assessDomainSpreadCfg(caps int, thresholds [2]int) LoadDimension {
	level := LoadLow
	if caps >= thresholds[1] {
		level = LoadHigh
	} else if caps >= thresholds[0] {
		level = LoadMedium
	}
	return LoadDimension{Value: float64(caps), Level: level}
}

func assessServiceLoadCfg(services, teamSize int, thresholds [2]float64) LoadDimension {
	if teamSize <= 0 {
		teamSize = 1
	}
	ratio := float64(services) / float64(teamSize)
	level := LoadLow
	if ratio > thresholds[1] {
		level = LoadHigh
	} else if ratio > thresholds[0] {
		level = LoadMedium
	}
	return LoadDimension{Value: math.Round(ratio*10) / 10, Level: level}
}

func assessInteractionLoadCfg(weightedScore int, thresholds [2]int) LoadDimension {
	level := LoadLow
	if weightedScore >= thresholds[1] {
		level = LoadHigh
	} else if weightedScore >= thresholds[0] {
		level = LoadMedium
	}
	return LoadDimension{Value: float64(weightedScore), Level: level}
}

func assessDependencyLoadCfg(deps int, thresholds [2]int) LoadDimension {
	level := LoadLow
	if deps >= thresholds[1] {
		level = LoadHigh
	} else if deps >= thresholds[0] {
		level = LoadMedium
	}
	return LoadDimension{Value: float64(deps), Level: level}
}

func worstLevel(levels ...LoadLevel) LoadLevel {
	worst := LoadLow
	for _, l := range levels {
		if levelRank(l) > levelRank(worst) {
			worst = l
		}
	}
	return worst
}

func levelRank(l LoadLevel) int {
	switch l {
	case LoadHigh:
		return 3
	case LoadMedium:
		return 2
	default:
		return 1
	}
}

// compositeScore produces a single number for sort tiebreaking.
// It sums dimension values — NOT a "cognitive load percentage."
func compositeScore(tl TeamLoad) float64 {
	return tl.DomainSpread.Value + tl.ServiceLoad.Value + tl.InteractionLoad.Value + tl.DependencyLoad.Value
}
