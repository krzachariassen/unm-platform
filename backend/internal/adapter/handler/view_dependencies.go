package handler

import (
	"errors"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/infrastructure/analyzer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// dependencyCycleJSON is the JSON shape for a single dependency cycle.
type dependencyCycleJSON struct {
	Path []string `json:"path"`
}

// dependenciesResponse is the JSON shape for GET /api/views/dependencies.
type dependenciesResponse struct {
	ModelID             string                `json:"model_id"`
	ServiceCycles       []dependencyCycleJSON `json:"service_cycles"`
	CapabilityCycles    []dependencyCycleJSON `json:"capability_cycles"`
	MaxServiceDepth     int                   `json:"max_service_depth"`
	MaxCapabilityDepth  int                   `json:"max_capability_depth"`
	CriticalServicePath []string              `json:"critical_service_path"`
}

// handleDependencies handles GET /api/views/dependencies?model_id=<id>
func (h *Handler) handleDependencies(w http.ResponseWriter, r *http.Request) {
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

	report := h.dependency.Analyze(stored.Model)

	writeJSON(w, http.StatusOK, dependenciesResponse{
		ModelID:             modelID,
		ServiceCycles:       mapCycles(report.ServiceCycles),
		CapabilityCycles:    mapCycles(report.CapabilityCycles),
		MaxServiceDepth:     report.MaxServiceDepth,
		MaxCapabilityDepth:  report.MaxCapabilityDepth,
		CriticalServicePath: nullSafeStringSlice(report.CriticalServicePath),
	})
}

// mapCycles converts analyzer.DependencyCycle slices to JSON-ready structs.
func mapCycles(cycles []analyzer.DependencyCycle) []dependencyCycleJSON {
	out := make([]dependencyCycleJSON, 0, len(cycles))
	for _, c := range cycles {
		out = append(out, dependencyCycleJSON{Path: c.Path})
	}
	return out
}

// nullSafeStringSlice returns an empty slice (never nil) for JSON null-safety.
func nullSafeStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
