package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/valueobject"
)
// DepthChainThreshold is the default maximum service dependency chain depth before a coupling
// signal is suggested. Kept for backward compatibility; prefer SignalsConfig.DepthChainThreshold.
const DepthChainThreshold = 4

// SuggestedSignal is a candidate signal entry derived from analysis findings.
// It contains all information needed to write a signal block in the YAML model.
type SuggestedSignal struct {
	Category     string
	OnEntityName string
	Description  string
	Evidence     string
	Severity     valueobject.Severity
	Source       string                 // which analyzer generated this suggestion
	Explanation  string                 // human-readable why it was flagged and what threshold was breached
	SourceTag    valueobject.SourceType // trust layer: always SourceAnalyzerFinding for auto-generated signals
}

// SignalSuggestionsReport holds all auto-generated signal suggestions.
type SignalSuggestionsReport struct {
	Suggestions []SuggestedSignal
}

// SignalSuggestionGenerator synthesizes findings from multiple analyzers into candidate signals.
// It only suggests signals that do not already exist in the model (by category + entity name).
type SignalSuggestionGenerator struct {
	cfg entity.SignalsConfig
}

// NewSignalSuggestionGenerator constructs a SignalSuggestionGenerator.
func NewSignalSuggestionGenerator(cfg entity.SignalsConfig) *SignalSuggestionGenerator {
	return &SignalSuggestionGenerator{cfg: cfg}
}

// Generate produces signal suggestions from multiple analysis reports.
// Existing signals in m are consulted to skip duplicates (same category + entity name).
// Suggestions are sorted: severity descending (critical first), then category, then entity name.
func (g *SignalSuggestionGenerator) Generate(
	bottleneckReport BottleneckReport,
	cognitiveLoadReport CognitiveLoadReport,
	fragmentationReport FragmentationReport,
	depReport DependencyReport,
	unlinkedReport UnlinkedCapabilityReport,
	m *entity.UNMModel,
) SignalSuggestionsReport {
	// Build set of (category|entityName) for existing signals to avoid duplicates.
	existing := make(map[string]bool, len(m.Signals))
	for _, sig := range m.Signals {
		existing[sig.Category+"|"+sig.OnEntityName] = true
	}
	alreadyExists := func(category, entityName string) bool {
		return existing[category+"|"+entityName]
	}

	var suggestions []SuggestedSignal

	// ── Rule 1: Bottleneck signals ──────────────────────────────────────────────
	for _, sb := range bottleneckReport.ServiceBottlenecks {
		if !sb.IsCritical && !sb.IsWarning {
			continue
		}
		if alreadyExists(entity.CategoryBottleneck, sb.Service.Name) {
			continue
		}
		var sev valueobject.Severity
		var desc, explanation string
		if sb.IsCritical {
			sev = valueobject.SeverityCritical
			desc = fmt.Sprintf("%s is a critical bottleneck: %d services depend on it", sb.Service.Name, sb.FanIn)
			explanation = fmt.Sprintf("Service '%s' has %d dependents (critical threshold exceeded). This creates a high-blast-radius deployment bottleneck where a single failure cascades to many downstream services.", sb.Service.Name, sb.FanIn)
		} else {
			sev = valueobject.SeverityHigh
			desc = fmt.Sprintf("%s is a bottleneck candidate: %d services depend on it", sb.Service.Name, sb.FanIn)
			explanation = fmt.Sprintf("Service '%s' has %d dependents (warning threshold exceeded). High fan-in indicates this service is becoming a coordination bottleneck and may slow team velocity.", sb.Service.Name, sb.FanIn)
		}
		suggestions = append(suggestions, SuggestedSignal{
			Category:     entity.CategoryBottleneck,
			OnEntityName: sb.Service.Name,
			Description:  desc,
			Evidence:     fmt.Sprintf("Fan-in: %d (fan-out: %d)", sb.FanIn, sb.FanOut),
			Severity:     sev,
			Source:       "bottleneck-analyzer",
			Explanation:  explanation,
			SourceTag:    valueobject.SourceAnalyzerFinding,
		})
	}

	// ── Rule 2: Cognitive-load signals ─────────────────────────────────────────
	// Flag teams where the overall structural load level is "high" (at least
	// one dimension exceeds the high threshold from Team Topologies guidance).
	for _, tl := range cognitiveLoadReport.TeamLoads {
		if tl.OverallLevel != LoadHigh {
			continue
		}
		if alreadyExists(entity.CategoryCognitiveLoad, tl.Team.Name) {
			continue
		}
		var highDims []string
		if tl.DomainSpread.Level == LoadHigh {
			highDims = append(highDims, fmt.Sprintf("domain-spread=%g", tl.DomainSpread.Value))
		}
		if tl.ServiceLoad.Level == LoadHigh {
			highDims = append(highDims, fmt.Sprintf("service-load=%d/%d", tl.ServiceCount, tl.TeamSize))
		}
		if tl.InteractionLoad.Level == LoadHigh {
			highDims = append(highDims, fmt.Sprintf("interaction-load=%d", tl.InteractionScore))
		}
		if tl.DependencyLoad.Level == LoadHigh {
			highDims = append(highDims, fmt.Sprintf("dependency-fanout=%d", tl.DependencyCount))
		}
		suggestions = append(suggestions, SuggestedSignal{
			Category:     entity.CategoryCognitiveLoad,
			OnEntityName: tl.Team.Name,
			Description:  fmt.Sprintf("%s has high structural load — consider splitting responsibilities", tl.Team.Name),
			Evidence:     fmt.Sprintf("High dimensions: %s", strings.Join(highDims, ", ")),
			Severity:     valueobject.SeverityHigh,
			Source:       "cognitive-load-analyzer",
			Explanation:  fmt.Sprintf("Team '%s' exceeds structural load thresholds across %d dimension(s): %s. According to Team Topologies principles, high cognitive load reduces team responsiveness and increases delivery risk.", tl.Team.Name, len(highDims), strings.Join(highDims, ", ")),
			SourceTag:    valueobject.SourceAnalyzerFinding,
		})
	}

	// ── Rule 3: Fragmentation signals ──────────────────────────────────────────
	for _, fc := range fragmentationReport.FragmentedCapabilities {
		if alreadyExists(entity.CategoryFragmentation, fc.Capability.Name) {
			continue
		}
		teamNames := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teamNames = append(teamNames, t.Name)
		}
		sort.Strings(teamNames)
		suggestions = append(suggestions, SuggestedSignal{
			Category:     entity.CategoryFragmentation,
			OnEntityName: fc.Capability.Name,
			Description:  fmt.Sprintf("%s is delivered by services from %d teams — fragmented ownership", fc.Capability.Name, len(fc.Teams)),
			Evidence:     fmt.Sprintf("Teams involved: %s", strings.Join(teamNames, ", ")),
			Severity:     valueobject.SeverityHigh,
			Source:       "fragmentation-analyzer",
			Explanation:  fmt.Sprintf("Capability '%s' is realized by services owned across %d teams (%s). Fragmented ownership increases coordination overhead, creates release coupling, and reduces reliability of the capability end-to-end.", fc.Capability.Name, len(fc.Teams), strings.Join(teamNames, ", ")),
			SourceTag:    valueobject.SourceAnalyzerFinding,
		})
	}

	// ── Rule 4: Coupling signal from deep dependency chains ────────────────────
	if depReport.MaxServiceDepth > g.cfg.DepthChainThreshold && len(depReport.CriticalServicePath) > 0 {
		deepest := depReport.CriticalServicePath[len(depReport.CriticalServicePath)-1]
		if !alreadyExists(entity.CategoryCoupling, deepest) {
			suggestions = append(suggestions, SuggestedSignal{
				Category:     entity.CategoryCoupling,
				OnEntityName: deepest,
				Description:  fmt.Sprintf("Service dependency chain depth of %d exceeds threshold of %d — high blast-radius risk", depReport.MaxServiceDepth, g.cfg.DepthChainThreshold),
				Evidence:     fmt.Sprintf("Critical path (%d hops): %s", depReport.MaxServiceDepth, strings.Join(depReport.CriticalServicePath, " → ")),
				Severity:     valueobject.SeverityMedium,
				Source:       "dependency-analyzer",
				Explanation:  fmt.Sprintf("The service dependency chain reaches a depth of %d (threshold: %d), with '%s' at the deepest point. Long dependency chains amplify the blast radius of failures and complicate independent deployment.", depReport.MaxServiceDepth, g.cfg.DepthChainThreshold, deepest),
				SourceTag:    valueobject.SourceAnalyzerFinding,
			})
		}
	}

	// ── Rule 5: Gap signals for unlinked domain/foundational capabilities ───────
	for _, uc := range unlinkedReport.UnlinkedLeafCapabilities {
		if uc.IsExpected {
			continue // infrastructure-visibility unlinked caps are normal
		}
		if alreadyExists(entity.CategoryGap, uc.Capability.Name) {
			continue
		}
		suggestions = append(suggestions, SuggestedSignal{
			Category:     entity.CategoryGap,
			OnEntityName: uc.Capability.Name,
			Description:  fmt.Sprintf("No user need drives the %q capability (%s) — add a need or reclassify to infrastructure", uc.Capability.Name, uc.Visibility),
			Evidence:     fmt.Sprintf("Visibility: %s; capability is a leaf with no need.supportedBy reference", uc.Visibility),
			Severity:     valueobject.SeverityHigh,
			Source:       "unlinked-capability-analyzer",
			Explanation:  fmt.Sprintf("Capability '%s' (visibility: %s) is a leaf capability with no user need pointing to it via 'supportedBy'. This suggests the capability may be orphaned, misclassified, or that a user need is missing from the model.", uc.Capability.Name, uc.Visibility),
			SourceTag:    valueobject.SourceAnalyzerFinding,
		})
	}

	// Sort: severity descending, then category, then entity name.
	severityLevel := func(s valueobject.Severity) int {
		switch s {
		case valueobject.SeverityCritical:
			return 4
		case valueobject.SeverityHigh:
			return 3
		case valueobject.SeverityMedium:
			return 2
		default:
			return 1
		}
	}
	sort.Slice(suggestions, func(i, j int) bool {
		si := severityLevel(suggestions[i].Severity)
		sj := severityLevel(suggestions[j].Severity)
		if si != sj {
			return si > sj // higher severity first
		}
		if suggestions[i].Category != suggestions[j].Category {
			return suggestions[i].Category < suggestions[j].Category
		}
		return suggestions[i].OnEntityName < suggestions[j].OnEntityName
	})

	return SignalSuggestionsReport{Suggestions: suggestions}
}
