package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/serializer"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// registerModelRoutes registers model-related API routes.
func (h *Handler) registerModelRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/models/parse", h.handleParse)
	mux.HandleFunc("POST /api/models/validate", h.handleValidate)
	mux.HandleFunc("GET /api/models", h.handleListModels)
	mux.HandleFunc("GET /api/models/{id}/export", h.handleExport)
	mux.HandleFunc("GET /api/models/{id}/history", h.handleListVersions)
	mux.HandleFunc("GET /api/models/{id}/versions/{v}", h.handleGetVersion)
	mux.HandleFunc("GET /api/models/{id}/diff", h.handleDiffVersions)
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

// modelListItem is a single entry in the GET /api/models response.
type modelListItem struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	VersionCount int       `json:"version_count"`
}

// modelListResponse is the JSON response for GET /api/models.
type modelListResponse struct {
	Models []modelListItem `json:"models"`
	Total  int             `json:"total"`
}

// versionMetaItem is a single entry in the GET /api/models/{id}/history response.
type versionMetaItem struct {
	ID            string    `json:"id"`
	Version       int       `json:"version"`
	CommitMessage string    `json:"commit_message"`
	CommittedAt   time.Time `json:"committed_at"`
}

// versionHistoryResponse is the JSON response for GET /api/models/{id}/history.
type versionHistoryResponse struct {
	ModelID  string            `json:"model_id"`
	Versions []versionMetaItem `json:"versions"`
}

// diffResponse is the JSON response for GET /api/models/{id}/diff.
type diffResponse struct {
	ModelID     string            `json:"model_id"`
	FromVersion int               `json:"from_version"`
	ToVersion   int               `json:"to_version"`
	Added       service.DiffEntities `json:"added"`
	Removed     service.DiffEntities `json:"removed"`
	Changed     service.DiffEntities `json:"changed"`
}

// sniffIsDSL peeks at the first 64 bytes of the body to determine if the content
// looks like DSL format. DSL files begin with the keyword "system" followed by
// a space or double-quote (e.g. `system "Name" {`). Returns the peeked bytes so
// the caller can reconstitute the full reader with io.MultiReader.
func sniffIsDSL(body io.Reader) (bool, io.Reader) {
	peek := make([]byte, 64)
	n, _ := body.Read(peek)
	peek = peek[:n]
	full := io.MultiReader(bytes.NewReader(peek), body)
	trimmed := strings.TrimSpace(string(peek))
	isDSL := strings.HasPrefix(trimmed, "system ") || strings.HasPrefix(trimmed, `system"`)
	return isDSL, full
}

// handleParse parses a submitted UNM model, stores it, and returns JSON.
// Pass ?format=dsl to parse DSL (.unm) format. If no format param is given,
// the content is sniffed: bodies starting with `system ` or `system"` are
// treated as DSL automatically; everything else is parsed as YAML.
// Send X-Replace-Model header with a previous model ID to delete the old model
// and its changesets before storing the new one (prevents memory leaks).
func (h *Handler) handleParse(w http.ResponseWriter, r *http.Request) {
	var body io.Reader = r.Body
	pv := h.parseAndValidate
	switch r.URL.Query().Get("format") {
	case "dsl":
		pv = h.parseAndValidateDSL
	case "yaml", "yml":
		// explicit YAML — keep default
	default:
		// auto-detect
		var isDSL bool
		isDSL, body = sniffIsDSL(body)
		if isDSL {
			pv = h.parseAndValidateDSL
		}
	}

	model, result, err := pv.Execute(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if replaceID := r.Header.Get("X-Replace-Model"); replaceID != "" {
		_ = h.store.Delete(replaceID)
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
// The model is NOT stored. Format auto-detection follows the same rules as handleParse.
func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	var body io.Reader = r.Body
	pv := h.parseAndValidate
	switch r.URL.Query().Get("format") {
	case "dsl":
		pv = h.parseAndValidateDSL
	case "yaml", "yml":
		// explicit YAML — keep default
	default:
		// auto-detect
		var isDSL bool
		isDSL, body = sniffIsDSL(body)
		if isDSL {
			pv = h.parseAndValidateDSL
		}
	}

	_, result, err := pv.Execute(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, buildValidatePayload(result))
}

// handleListModels returns a list of all stored models with version counts.
func (h *Handler) handleListModels(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list models: "+err.Error())
		return
	}

	models := make([]modelListItem, 0, len(items))
	for _, item := range items {
		name := ""
		if item.Model != nil {
			name = item.Model.System.Name
		}
		models = append(models, modelListItem{
			ID:           item.ID,
			Name:         name,
			CreatedAt:    item.CreatedAt,
			VersionCount: item.VersionCount,
		})
	}

	writeJSON(w, http.StatusOK, modelListResponse{
		Models: models,
		Total:  len(models),
	})
}

// handleListVersions returns the version history for a model.
func (h *Handler) handleListVersions(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")

	versions, err := h.store.ListVersions(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list versions: "+err.Error())
		return
	}

	items := make([]versionMetaItem, 0, len(versions))
	for _, v := range versions {
		items = append(items, versionMetaItem{
			ID:            v.ID,
			Version:       v.Version,
			CommitMessage: v.CommitMessage,
			CommittedAt:   v.CommittedAt,
		})
	}

	writeJSON(w, http.StatusOK, versionHistoryResponse{
		ModelID:  modelID,
		Versions: items,
	})
}

// handleGetVersion retrieves a model at a specific version number.
func (h *Handler) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	vStr := r.PathValue("v")

	v, err := strconv.Atoi(vStr)
	if err != nil || v < 1 {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid version %q: must be a positive integer", vStr))
		return
	}

	model, err := h.store.GetVersion(modelID, v)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("version %d not found", v))
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get version: "+err.Error())
		return
	}

	summary := model.Summary()
	valResult := service.NewValidationEngine().Validate(model)

	writeJSON(w, http.StatusOK, parseResponse{
		ID:                modelID,
		SystemName:        summary.SystemName,
		SystemDescription: summary.SystemDescription,
		Summary: parseSummary{
			Actors:       summary.ActorCount,
			Needs:        summary.NeedCount,
			Capabilities: summary.CapabilityCount,
			Services:     summary.ServiceCount,
			Teams:        summary.TeamCount,
		},
		Validation: buildValidatePayload(valResult),
	})
}

// handleDiffVersions computes a structural diff between two versions of a model.
// Query params: from (version number) and to (version number).
func (h *Handler) handleDiffVersions(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	fromV, err := strconv.Atoi(fromStr)
	if err != nil || fromV < 1 {
		writeError(w, http.StatusBadRequest, "from must be a positive integer version number")
		return
	}
	toV, err := strconv.Atoi(toStr)
	if err != nil || toV < 1 {
		writeError(w, http.StatusBadRequest, "to must be a positive integer version number")
		return
	}

	diff, err := h.store.DiffVersions(modelID, fromV, toV)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model or version not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute diff: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, diffResponse{
		ModelID:     modelID,
		FromVersion: diff.FromVersion,
		ToVersion:   diff.ToVersion,
		Added:       diff.Added,
		Removed:     diff.Removed,
		Changed:     diff.Changed,
	})
}

// buildValidatePayload converts a service.ValidationResult into the HTTP response shape.
func buildValidatePayload(result service.ValidationResult) validatePayload {
	errs := make([]validationItem, 0, len(result.Errors))
	for _, e := range result.Errors {
		errs = append(errs, validationItem{
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
		Errors:   errs,
		Warnings: warnings,
	}
}

// handleExport serializes the stored model for download.
func (h *Handler) handleExport(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")

	stored, err := h.store.Get(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	if r.URL.Query().Get("format") == "dsl" {
		data, err := serializer.MarshalDSL(stored.Model)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to export model: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=model.unm")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			log.Printf("export write: %v", err)
		}
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
