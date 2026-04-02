package handler

import (
	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// ── 6. Enriched Realization View ──────────────────────────────────────────────

type enrichedRealizationResponse struct {
	ViewType      string       `json:"view_type"`
	NoCapCount    int          `json:"no_cap_count"`
	MultiCapCount int          `json:"multi_cap_count"`
	ServiceRows   []serviceRow `json:"service_rows"`
}

func buildEnrichedRealizationView(m *entity.UNMModel) enrichedRealizationResponse {
	svcCapCount := buildSvcCapCount(m)
	rows := buildServiceRows(m, svcCapCount)

	noCapCount := 0
	multiCapCount := 0
	for _, svcName := range sortedServiceNames(m) {
		cc := svcCapCount[svcName]
		if cc == 0 {
			noCapCount++
		}
		if cc > 1 {
			multiCapCount++
		}
	}

	return enrichedRealizationResponse{
		ViewType:      "realization",
		NoCapCount:    noCapCount,
		MultiCapCount: multiCapCount,
		ServiceRows:   rows,
	}
}
