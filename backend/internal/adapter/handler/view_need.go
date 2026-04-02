package handler

import (
	"sort"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
)

// ── 1. Enriched Need View ─────────────────────────────────────────────────────

type enrichedNeedResponse struct {
	ViewType      string           `json:"view_type"`
	TotalNeeds    int              `json:"total_needs"`
	UnmappedCount int              `json:"unmapped_count"`
	Groups        []needActorGroup `json:"groups"`
}

type needActorGroup struct {
	Actor needActorRef `json:"actor"`
	Needs []needEntry  `json:"needs"`
}

type needActorRef struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type needEntry struct {
	Need         needRef  `json:"need"`
	Capabilities []capRef `json:"capabilities"`
}

type needRef struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Data  map[string]any `json:"data"`
}

func buildEnrichedNeedView(m *entity.UNMModel, cfg ...entity.AnalysisConfig) enrichedNeedResponse {
	acfg := defaultAnalysisCfg(cfg)
	// Run value chain analysis to get per-need delivery risk data.
	vcReport := analyzer.NewValueChainAnalyzer(acfg.ValueChain).Analyze(m)
	riskByNeed := make(map[string]analyzer.NeedDeliveryRisk, len(vcReport.NeedRisks))
	for _, nr := range vcReport.NeedRisks {
		riskByNeed[nr.NeedName] = nr
	}

	// Group needs by actor — a need with multiple actors appears under each actor group
	actorNeeds := make(map[string][]*entity.Need)
	for _, n := range m.Needs {
		for _, actorName := range n.ActorNames {
			actorNeeds[actorName] = append(actorNeeds[actorName], n)
		}
	}

	// Sort actor names for determinism
	actorNames := make([]string, 0, len(actorNeeds))
	for name := range actorNeeds {
		actorNames = append(actorNames, name)
	}
	sort.Strings(actorNames)

	unmappedCount := 0
	groups := make([]needActorGroup, 0, len(actorNames))
	for _, actorName := range actorNames {
		needs := actorNeeds[actorName]
		sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })

		entries := make([]needEntry, 0, len(needs))
		for _, n := range needs {
			if !n.IsMapped() {
				unmappedCount++
			}

			caps := make([]capRef, 0)
			for _, rel := range n.SupportedBy {
				capName := rel.TargetID.String()
				if c, ok := m.Capabilities[capName]; ok {
					caps = append(caps, capRef{
						ID:    "cap-" + c.Name,
						Label: c.Name,
						Data:  map[string]any{"visibility": c.Visibility},
					})
				}
			}

			needData := map[string]any{
				"is_mapped": n.IsMapped(),
				"outcome":   n.Outcome,
			}
			// Team span data from value chain traversal.
			if nr, ok := riskByNeed[n.Name]; ok {
				needData["team_span"] = nr.TeamSpan
				needData["teams"] = nr.Teams
				needData["at_risk"] = nr.AtRisk
				needData["unbacked"] = nr.Unbacked
			}
			aps := detectNeedAntiPatterns(n)
			if len(aps) > 0 {
				needData["anti_patterns"] = aps
			}

			entries = append(entries, needEntry{
				Need: needRef{
					ID:    "need-" + n.Name,
					Label: n.Name,
					Data:  needData,
				},
				Capabilities: caps,
			})
		}
		groups = append(groups, needActorGroup{
			Actor: needActorRef{
				ID:    "actor-" + actorName,
				Label: actorName,
			},
			Needs: entries,
		})
	}

	return enrichedNeedResponse{
		ViewType:      "need",
		TotalNeeds:    len(m.Needs),
		UnmappedCount: unmappedCount,
		Groups:        groups,
	}
}
