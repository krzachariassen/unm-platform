package handler

import (
	"errors"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// teamModeOverloadJSON is the JSON shape for a single over-reliant team entry.
type teamModeOverloadJSON struct {
	TeamName string `json:"team_name"`
	Mode     string `json:"mode"`
	Count    int    `json:"count"`
}

// interactionsResponse is the JSON shape for GET /api/views/interactions.
type interactionsResponse struct {
	ModelID          string                 `json:"model_id"`
	ModeDistribution map[string]int         `json:"mode_distribution"`
	IsolatedTeams    []string               `json:"isolated_teams"`
	OverReliantTeams []teamModeOverloadJSON  `json:"over_reliant_teams"`
	AllModesSame     bool                   `json:"all_modes_same"`
}

// handleInteractions handles GET /api/views/interactions?model_id=<id>
func (h *Handler) handleInteractions(w http.ResponseWriter, r *http.Request) {
	modelID := r.URL.Query().Get("model_id")
	if modelID == "" {
		writeError(w, http.StatusBadRequest, "missing required query parameter: model_id")
		return
	}

	stored, err := h.store.Get(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	report := h.interactions.Analyze(stored.Model)

	writeJSON(w, http.StatusOK, interactionsResponse{
		ModelID:          modelID,
		ModeDistribution: mapModeDistribution(report),
		IsolatedTeams:    nullSafeStringSlice(report.IsolatedTeams),
		OverReliantTeams: mapOverReliantTeams(report.OverReliantTeams),
		AllModesSame:     report.AllModesSame,
	})
}

// mapModeDistribution converts the analyzer's mode distribution (keyed by valueobject.InteractionMode)
// to a string-keyed map suitable for JSON serialization.
func mapModeDistribution(report analyzer.InteractionDiversityReport) map[string]int {
	out := make(map[string]int, len(report.ModeDistribution))
	for mode, count := range report.ModeDistribution {
		out[mode.String()] = count
	}
	return out
}

// mapOverReliantTeams converts analyzer.TeamModeOverload slices to JSON-ready structs.
func mapOverReliantTeams(teams []analyzer.TeamModeOverload) []teamModeOverloadJSON {
	out := make([]teamModeOverloadJSON, 0, len(teams))
	for _, t := range teams {
		out = append(out, teamModeOverloadJSON{
			TeamName: t.TeamName,
			Mode:     t.Mode.String(),
			Count:    t.Count,
		})
	}
	return out
}
