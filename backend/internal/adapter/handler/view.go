package handler

import (
	"fmt"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
)

// registerViewRoutes registers GET /api/models/{id}/views/{viewType}.
func (h *Handler) registerViewRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models/{id}/views/{viewType}", h.handleView)
}

// viewBuilder is a function that builds a view response for a given model and analysis config.
// Adding a new view type requires only one entry in viewRegistry below.
type viewBuilder func(m *entity.UNMModel, analysis entity.AnalysisConfig) any

// viewRegistry maps view-type keys to their builder functions.
// Builders that do not need analysis config simply ignore the second argument.
var viewRegistry = map[string]viewBuilder{
	"need":           func(m *entity.UNMModel, a entity.AnalysisConfig) any { return buildEnrichedNeedView(m, a) },
	"capability":     func(m *entity.UNMModel, a entity.AnalysisConfig) any { return buildEnrichedCapabilityView(m, a) },
	"realization":    func(m *entity.UNMModel, _ entity.AnalysisConfig) any { return buildEnrichedRealizationView(m) },
	"ownership":      func(m *entity.UNMModel, a entity.AnalysisConfig) any { return buildEnrichedOwnershipView(m, a) },
	"team-topology":  func(m *entity.UNMModel, a entity.AnalysisConfig) any { return buildEnrichedTeamTopologyView(m, a) },
	"cognitive-load": func(m *entity.UNMModel, a entity.AnalysisConfig) any { return buildEnrichedCognitiveLoadView(m, a) },
	"unm-map":        func(m *entity.UNMModel, _ entity.AnalysisConfig) any { return buildUNMMapView(m) },
}

// handleView returns pre-computed view data for a given viewType.
// Supported viewTypes are defined in viewRegistry above.
func (h *Handler) handleView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	viewType := r.PathValue("viewType")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	build, ok := viewRegistry[viewType]
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown view type: %q", viewType))
		return
	}

	writeJSON(w, http.StatusOK, build(stored.Model, h.cfg.Analysis))
}
