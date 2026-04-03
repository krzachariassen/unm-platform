package handler

import (
	"errors"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// gapsResponse is the JSON shape for GET /api/views/gaps.
type gapsResponse struct {
	ModelID                string   `json:"model_id"`
	UnmappedNeeds          []string `json:"unmapped_needs"`
	UnrealizedCapabilities []string `json:"unrealized_capabilities"`
	UnownedServices        []string `json:"unowned_services"`
	UnneededCapabilities   []string `json:"unneeded_capabilities"`
	OrphanServices         []string `json:"orphan_services"`
}

// handleGaps handles GET /api/views/gaps?model_id=<id>
func (h *Handler) handleGaps(w http.ResponseWriter, r *http.Request) {
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

	report := h.gap.Analyze(stored.Model)

	unmappedNeeds := make([]string, 0, len(report.UnmappedNeeds))
	for _, n := range report.UnmappedNeeds {
		unmappedNeeds = append(unmappedNeeds, n.Name)
	}

	unrealizedCaps := make([]string, 0, len(report.UnrealizedCapabilities))
	for _, c := range report.UnrealizedCapabilities {
		unrealizedCaps = append(unrealizedCaps, c.Name)
	}

	unownedSvcs := make([]string, 0, len(report.UnownedServices))
	for _, s := range report.UnownedServices {
		unownedSvcs = append(unownedSvcs, s.Name)
	}

	unneededCaps := make([]string, 0, len(report.UnneededCapabilities))
	for _, c := range report.UnneededCapabilities {
		unneededCaps = append(unneededCaps, c.Name)
	}

	orphanSvcs := make([]string, 0, len(report.OrphanServices))
	for _, s := range report.OrphanServices {
		orphanSvcs = append(orphanSvcs, s.Name)
	}

	writeJSON(w, http.StatusOK, gapsResponse{
		ModelID:                modelID,
		UnmappedNeeds:          unmappedNeeds,
		UnrealizedCapabilities: unrealizedCaps,
		UnownedServices:        unownedSvcs,
		UnneededCapabilities:   unneededCaps,
		OrphanServices:         orphanSvcs,
	})
}
