package handler

import (
	"fmt"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// registerAnalysisRoutes registers analysis endpoints:
//   - POST /api/analyze/{type}             – submit YAML, get analysis (no stored model)
//   - GET  /api/models/{id}/analyze/{type} – analyze a stored model by ID
func (h *Handler) registerAnalysisRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/analyze/{type}", h.handleAnalyze)
	mux.HandleFunc("GET /api/models/{id}/analyze/{type}", h.handleAnalyzeStored)
}

// handleAnalyze runs the requested analysis type on a submitted YAML model.
func (h *Handler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	analyzeType := r.PathValue("type")
	if !usecase.ValidAnalysisType(analyzeType) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown analysis type: %q", analyzeType))
		return
	}

	model, _, err := h.parseAndValidate.Execute(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.runner.RunAnalysis(analyzeType, model)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// handleAnalyzeStored runs analysis on a model already stored by ID.
func (h *Handler) handleAnalyzeStored(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	analyzeType := r.PathValue("type")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	if !usecase.ValidAnalysisType(analyzeType) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown analysis type: %q", analyzeType))
		return
	}

	result, err := h.runner.RunAnalysis(analyzeType, stored.Model)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
