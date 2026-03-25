package handler

import (
	"fmt"
	"net/http"
)

// registerViewRoutes registers GET /api/models/{id}/views/{viewType}.
func (h *Handler) registerViewRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models/{id}/views/{viewType}", h.handleView)
}

// handleView returns pre-computed view data for a given viewType.
// Supported viewTypes: need, capability, realization, ownership, team-topology, cognitive-load.
func (h *Handler) handleView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	viewType := r.PathValue("viewType")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	m := stored.Model

	switch viewType {
	case "need":
		writeJSON(w, http.StatusOK, buildEnrichedNeedView(m, h.cfg.Analysis))
	case "capability":
		writeJSON(w, http.StatusOK, buildEnrichedCapabilityView(m, h.cfg.Analysis))
	case "realization":
		writeJSON(w, http.StatusOK, buildEnrichedRealizationView(m))
	case "ownership":
		writeJSON(w, http.StatusOK, buildEnrichedOwnershipView(m, h.cfg.Analysis))
	case "team-topology":
		writeJSON(w, http.StatusOK, buildEnrichedTeamTopologyView(m, h.cfg.Analysis))
	case "cognitive-load":
		writeJSON(w, http.StatusOK, buildEnrichedCognitiveLoadView(m, h.cfg.Analysis))
	case "unm-map":
		writeJSON(w, http.StatusOK, buildUNMMapView(m))
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown view type: %q", viewType))
		return
	}
}
