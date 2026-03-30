package handler

import (
	"log"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/parser"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/serializer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// registerModelRoutes registers POST /api/models/parse and POST /api/models/validate.
func (h *Handler) registerModelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/models/parse", h.handleParse)
	mux.HandleFunc("POST /api/models/validate", h.handleValidate)
	mux.HandleFunc("GET /api/models/{id}/export", h.handleExport)
}

// parseResponse is the JSON shape returned by POST /api/models/parse.
type parseResponse struct {
	ID                string          `json:"id"`
	SystemName        string          `json:"system_name"`
	SystemDescription string          `json:"system_description"`
	Summary           parseSummary    `json:"summary"`
	Validation        validatePayload `json:"validation"`
	Warnings          []string        `json:"warnings,omitempty"`
}

// parseSummary holds entity counts for the parse response.
type parseSummary struct {
	Actors       int `json:"actors"`
	Needs        int `json:"needs"`
	Capabilities int `json:"capabilities"`
	Services     int `json:"services"`
	Teams        int `json:"teams"`
}

// validatePayload is the JSON shape returned by POST /api/models/validate (and embedded in parse).
type validatePayload struct {
	IsValid  bool             `json:"is_valid"`
	Errors   []validationItem `json:"errors"`
	Warnings []validationItem `json:"warnings"`
}

// validationItem is a single error or warning entry.
type validationItem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Entity  string `json:"entity"`
}

// handleParse parses a submitted UNM model, stores it, and returns JSON.
// Pass ?format=dsl to parse DSL (.unm) format; default is YAML.
// Send X-Replace-Model header with a previous model ID to delete the old model
// and its changesets before storing the new one (prevents memory leaks).
func (h *Handler) handleParse(w http.ResponseWriter, r *http.Request) {
	pv := h.parseAndValidate
	if r.URL.Query().Get("format") == "dsl" {
		pv = usecase.NewParseAndValidate(parser.NewDSLParser(), service.NewValidationEngine())
	}
	model, result, err := pv.Execute(r.Body)
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

	summary := model.Summary()

	writeJSON(w, http.StatusOK, parseResponse{
		ID:                id,
		SystemName:        summary.SystemName,
		SystemDescription: summary.SystemDescription,
		Summary: parseSummary{
			Actors:       summary.ActorCount,
			Needs:        summary.NeedCount,
			Capabilities: summary.CapabilityCount,
			Services:     summary.ServiceCount,
			Teams:        summary.TeamCount,
		},
		Validation: buildValidatePayload(result),
		Warnings:   model.Warnings,
	})
}

// handleValidate parses and validates a submitted UNM model, returning errors/warnings.
// Pass ?format=dsl to parse DSL (.unm) format; default is YAML.
// The model is NOT stored.
func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	pv := h.parseAndValidate
	if r.URL.Query().Get("format") == "dsl" {
		pv = usecase.NewParseAndValidate(parser.NewDSLParser(), service.NewValidationEngine())
	}
	_, result, err := pv.Execute(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, buildValidatePayload(result))
}

// buildValidatePayload converts a service.ValidationResult into the HTTP response shape.
func buildValidatePayload(result service.ValidationResult) validatePayload {
	errors := make([]validationItem, 0, len(result.Errors))
	for _, e := range result.Errors {
		errors = append(errors, validationItem{
			Code:    string(e.Code),
			Message: e.Message,
			Entity:  e.Entity,
		})
	}

	warnings := make([]validationItem, 0, len(result.Warnings))
	for _, w := range result.Warnings {
		warnings = append(warnings, validationItem{
			Code:    string(w.Code),
			Message: w.Message,
			Entity:  w.Entity,
		})
	}

	return validatePayload{
		IsValid:  result.IsValid(),
		Errors:   errors,
		Warnings: warnings,
	}
}

// handleExport serializes the stored model back to YAML for download.
func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")

	stored := h.store.Get(modelID)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	data, err := serializer.MarshalYAML(stored.Model)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to export model: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=model.unm.yaml")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("export write: %v", err)
	}
}
