package ai_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"gopkg.in/yaml.v3"
)

// aiQuestion mirrors the YAML structure in testdata/ai_questions.yaml.
type aiQuestion struct {
	ID             int      `yaml:"id"`
	Category       string   `yaml:"category"`
	Template       string   `yaml:"template"`
	Question       string   `yaml:"question"`
	MustMention    []string `yaml:"must_mention"`
	MustNotMention []string `yaml:"must_not_mention"`
	Notes          string   `yaml:"notes"`
}

// teamSummary is prompt template data for teams.
type teamSummary struct {
	Name            string
	TeamType        string
	Size            int
	CognitiveLoad   string
	ServiceCount    int
	CapabilityCount int
	Services        []string
	Capabilities    []string
}

type serviceSummary struct {
	Name            string
	OwnerTeam       string
	DependencyCount int
	DependsOn       []string
	Capabilities    []string
}

type capSummary struct {
	Name              string
	Visibility        string
	OwnerTeams        []string
	RealizingServices []string
}

type needSummary struct {
	Name                    string
	Actor                   string
	SupportedByCapabilities []string
	TeamSpan                int
	AtRisk                  bool
}

type interactionSummary struct {
	From        string
	To          string
	Mode        string
	Via         string
	Description string
}

type signalSummary struct {
	Category         string
	Severity         string
	Description      string
	AffectedEntities string
	Evidence         string
}

// testDir returns the directory of this test file.
func testDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func TestAIAdvisor_30Questions(t *testing.T) {
	if os.Getenv("UNM_AI_TESTS") != "true" {
		t.Skip("UNM_AI_TESTS not enabled — skipping AI advisor test suite (set UNM_AI_TESTS=true to run)")
	}
	key := os.Getenv("UNM_OPENAI_API_KEY")
	if key == "" {
		t.Skip("UNM_OPENAI_API_KEY not set — cannot run AI tests without API key")
	}

	// Load INCA model
	incaPath := filepath.Join(testDir(), "..", "..", "..", "..", "examples", "inca.unm.yaml")
	m, err := parser.ParseFile(incaPath)
	require.NoError(t, err, "failed to parse INCA model")
	require.NotNil(t, m)

	// Load questions from YAML
	questionsPath := filepath.Join(testDir(), "..", "..", "..", "testdata", "ai_questions.yaml")
	questionsData, err := os.ReadFile(questionsPath)
	require.NoError(t, err, "failed to read ai_questions.yaml")

	var questions []aiQuestion
	require.NoError(t, yaml.Unmarshal(questionsData, &questions))
	require.Len(t, questions, 30, "expected 30 questions")

	// Build prompt data from the model (mirrors handler.buildAIPromptData)
	clAnalyzer := analyzer.NewCognitiveLoadAnalyzer(entity.DefaultConfig().Analysis.CognitiveLoad, entity.DefaultConfig().Analysis.InteractionWeights)
	clReport := clAnalyzer.Analyze(m)
	loadByTeam := make(map[string]string)
	for _, tl := range clReport.TeamLoads {
		loadByTeam[tl.Team.Name] = string(tl.OverallLevel)
	}

	vcAnalyzer := analyzer.NewValueChainAnalyzer(entity.DefaultConfig().Analysis.ValueChain)
	vcReport := vcAnalyzer.Analyze(m)
	vcByNeed := make(map[string]analyzer.NeedDeliveryRisk)
	for _, ndr := range vcReport.NeedRisks {
		vcByNeed[ndr.NeedName] = ndr
	}

	teams := make([]teamSummary, 0, len(m.Teams))
	for _, t := range m.Teams {
		var svcNames []string
		for _, svc := range m.Services {
			if svc.OwnerTeamName == t.Name {
				svcNames = append(svcNames, svc.Name)
			}
		}
		capNames := make([]string, 0, len(t.Owns))
		for _, rel := range t.Owns {
			capNames = append(capNames, rel.TargetID.String())
		}
		teams = append(teams, teamSummary{
			Name:            t.Name,
			TeamType:        string(t.TeamType),
			Size:            t.EffectiveSize(),
			CognitiveLoad:   loadByTeam[t.Name],
			ServiceCount:    len(svcNames),
			CapabilityCount: len(capNames),
			Services:        svcNames,
			Capabilities:    capNames,
		})
	}
	sort.Slice(teams, func(i, j int) bool { return teams[i].Name < teams[j].Name })

	svcToCaps := make(map[string][]string)
	for _, cap := range m.Capabilities {
		for _, rel := range cap.RealizedBy {
			svcToCaps[rel.TargetID.String()] = append(svcToCaps[rel.TargetID.String()], cap.Name)
		}
	}
	services := make([]serviceSummary, 0, len(m.Services))
	for _, svc := range m.Services {
		depTargets := make([]string, 0, len(svc.DependsOn))
		for _, rel := range svc.DependsOn {
			depTargets = append(depTargets, rel.TargetID.String())
		}
		services = append(services, serviceSummary{
			Name:            svc.Name,
			OwnerTeam:       svc.OwnerTeamName,
			DependencyCount: len(svc.DependsOn),
			DependsOn:       depTargets,
			Capabilities:    svcToCaps[svc.Name],
		})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })

	caps := make([]capSummary, 0, len(m.Capabilities))
	for _, cap := range m.Capabilities {
		// Find owner teams via team.Owns relationships
		ownerTeams := make([]string, 0)
		for _, team := range m.Teams {
			for _, rel := range team.Owns {
				if rel.TargetID.String() == cap.Name {
					ownerTeams = append(ownerTeams, team.Name)
				}
			}
		}
		realizingServices := make([]string, 0, len(cap.RealizedBy))
		for _, rel := range cap.RealizedBy {
			realizingServices = append(realizingServices, rel.TargetID.String())
		}
		caps = append(caps, capSummary{
			Name:              cap.Name,
			Visibility:        cap.Visibility,
			OwnerTeams:        ownerTeams,
			RealizingServices: realizingServices,
		})
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })

	needs := make([]needSummary, 0, len(m.Needs))
	for _, n := range m.Needs {
		suppBy := make([]string, 0, len(n.SupportedBy))
		for _, rel := range n.SupportedBy {
			suppBy = append(suppBy, rel.TargetID.String())
		}
		ndr := vcByNeed[n.Name]
		needs = append(needs, needSummary{
			Name:                    n.Name,
			Actor:                   n.ActorName,
			SupportedByCapabilities: suppBy,
			TeamSpan:                ndr.TeamSpan,
			AtRisk:                  ndr.AtRisk,
		})
	}
	sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })

	interactions := make([]interactionSummary, 0, len(m.Interactions))
	for _, ix := range m.Interactions {
		interactions = append(interactions, interactionSummary{
			From:        ix.FromTeamName,
			To:          ix.ToTeamName,
			Mode:        string(ix.Mode),
			Via:         ix.Via,
			Description: ix.Description,
		})
	}

	signals := make([]signalSummary, 0, len(m.Signals))
	for _, s := range m.Signals {
		affectedStr := strings.Join(s.AffectedEntities, ", ")
		signals = append(signals, signalSummary{
			Category:         string(s.Category),
			Severity:         string(s.Severity),
			Description:      s.Description,
			AffectedEntities: affectedStr,
			Evidence:         s.Evidence,
		})
	}

	type gapSummary struct {
		UnmappedNeeds          []string
		UnrealizedCapabilities []string
		UnownedServices        []string
		UnneededCapabilities   []string
	}
	type vcCounts struct {
		CrossTeamNeeds int
		AtRiskNeeds    int
		UnbackedNeeds  int
	}
	baseData := map[string]any{
		"SystemName":            m.System.Name,
		"ModelDescription":      m.System.Description,
		"Teams":                 teams,
		"Services":              services,
		"Capabilities":          caps,
		"Needs":                 needs,
		"Interactions":          interactions,
		"Signals":               signals,
		"FragmentedCapabilities": []any{},
		"CognitiveLoadDetails":  []any{},
		"Bottlenecks":           []any{},
		"Couplings":             []any{},
		"Gaps":                  gapSummary{},
		"Complexities":          []any{},
		"ServiceCycles":         [][]string{},
		"CapabilityCycles":      [][]string{},
		"CriticalServicePath":   []string{},
		"MaxServiceDepth":       0,
		"MaxCapabilityDepth":    0,
		"ValueStreamCoherence":  []any{},
		"LowCoherenceCount":     0,
		"ModeDistribution":      map[string]int{},
		"OverReliantTeams":      []any{},
		"IsolatedTeams":         []string{},
		"AllModesSame":          false,
		"UnlinkedCapabilities":  []any{},
		"ExternalDependencies":  []any{},
		"ValueChainCounts":      vcCounts{},
		"ValueChains":           []any{},
	}

	// Set up AI client and prompt renderer
	client, err := ai.NewOpenAIClient()
	require.NoError(t, err)
	require.True(t, client.IsConfigured())

	lib, err := ai.NewPromptLibrary()
	require.NoError(t, err)
	renderer := ai.NewPromptRenderer(lib)

	passed := 0
	failed := 0

	for _, q := range questions {
		q := q // capture range variable
		t.Run(fmt.Sprintf("Q%d_%s", q.ID, q.Category), func(t *testing.T) {
			// Build per-question data map
			data := make(map[string]any, len(baseData)+2)
			for k, v := range baseData {
				data[k] = v
			}
			data["UserQuestion"] = q.Question
			data["Question"] = q.Question

			// Render system prompt
			rendered, err := renderer.Render(q.Template, data)
			require.NoError(t, err, "failed to render template %s", q.Template)

			// Call OpenAI
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := client.Complete(ctx, rendered, q.Question)
			require.NoError(t, err, "OpenAI call failed for Q%d", q.ID)
			require.NotEmpty(t, resp.Content, "empty response for Q%d", q.ID)

			answer := strings.ToLower(resp.Content)

			// Check must_mention
			for _, phrase := range q.MustMention {
				assert.Contains(t, answer, strings.ToLower(phrase),
					"Q%d: answer should mention %q", q.ID, phrase)
			}

			// Check must_not_mention
			for _, phrase := range q.MustNotMention {
				assert.NotContains(t, answer, strings.ToLower(phrase),
					"Q%d: answer should NOT mention %q", q.ID, phrase)
			}

			if !t.Failed() {
				passed++
			} else {
				failed++
			}

			// Small delay to avoid rate limiting
			time.Sleep(100 * time.Millisecond)
		})
	}

	t.Logf("Results: %d passed, %d failed, %d total", passed, failed, len(questions))
}
