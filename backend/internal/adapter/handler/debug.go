package handler

import (
	"net/http"
	"os"

	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// registerDebugRoutes registers development-only helper routes.
func (h *Handler) registerDebugRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/debug/load-example", h.handleLoadExample)
}

// handleLoadExample parses the bundled INCA extended example and stores it,
// returning the same response shape as POST /api/models/parse.
// The server must be started from the backend/ directory for the relative path to resolve.
func (h *Handler) handleLoadExample(w http.ResponseWriter, r *http.Request) {
	// Try a few candidate paths relative to cwd
	candidates := []string{
		"../examples/inca.unm.extended.yaml",
		"../../examples/inca.unm.extended.yaml",
		"examples/inca.unm.extended.yaml",
	}

	var (
		f   *os.File
		err error
	)
	for _, path := range candidates {
		f, err = os.Open(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not find example file: "+err.Error())
		return
	}
	defer f.Close()

	uc := usecase.NewParseAndValidate(parser.NewYAMLParser(), service.NewValidationEngine())
	model, result, err := uc.Execute(f)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if replaceID := r.Header.Get("X-Replace-Model"); replaceID != "" {
		h.store.Delete(replaceID)
	}

	id, err := h.store.Store(model)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store model: "+err.Error())
		return
	}

	// Eagerly compute all AI insights in the background so they are ready
	// by the time the user navigates to any view page.
	go h.precomputeInsights(id, h.store.Get(id))

	summary := model.Summary()
	writeJSON(w, http.StatusOK, parseResponse{
		ID:         id,
		SystemName: summary.SystemName,
		Summary: parseSummary{
			Actors:       summary.ActorCount,
			Needs:        summary.NeedCount,
			Capabilities: summary.CapabilityCount,
			Services:     summary.ServiceCount,
			Teams:        summary.TeamCount,
		},
		Validation: validatePayload{
			IsValid:  result.IsValid(),
			Errors:   toValidationErrors(result.Errors),
			Warnings: toValidationWarnings(result.Warnings),
		},
	})
}

func toValidationErrors(items []service.ValidationError) []validationItem {
	out := make([]validationItem, 0, len(items))
	for _, e := range items {
		out = append(out, validationItem{Code: string(e.Code), Message: e.Message, Entity: e.Entity})
	}
	return out
}

func toValidationWarnings(items []service.ValidationWarning) []validationItem {
	out := make([]validationItem, 0, len(items))
	for _, w := range items {
		out = append(out, validationItem{Code: string(w.Code), Message: w.Message, Entity: w.Entity})
	}
	return out
}
