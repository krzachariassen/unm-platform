package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/domain/service"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// registerChangesetRoutes registers changeset-related API routes.
// Disambiguation routes resolve Go 1.22 mux conflict between
// POST /api/models/{id}/* and POST /api/models/analyze/{type}.
func (h *Handler) registerChangesetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/models/analyze/changesets", h.handleAnalyze)
	mux.HandleFunc("POST /api/models/analyze/ask", h.handleAnalyze)
	mux.HandleFunc("POST /api/models/{id}/changesets", h.handleCreateChangeset)
	mux.HandleFunc("GET /api/models/{id}/changesets/{csId}", h.handleGetChangeset)
	mux.HandleFunc("GET /api/models/{id}/changesets/{csId}/projected", h.handleProjectedModel)
	mux.HandleFunc("GET /api/models/{id}/changesets/{csId}/impact", h.handleImpact)
	mux.HandleFunc("POST /api/models/{id}/changesets/{csId}/apply", h.handleApply)
	mux.HandleFunc("POST /api/models/{id}/changesets/{csId}/commit", h.handleCommitChangeset)
	mux.HandleFunc("POST /api/models/{id}/changesets/{csId}/explain", h.handleExplainChangeset)
}

// changesetCreateRequest is the JSON body for POST /api/models/{id}/changesets.
type changesetCreateRequest struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Actions     []changeActionDTO `json:"actions"`
}

// changesetCreateResponse is the JSON response for POST /api/models/{id}/changesets.
type changesetCreateResponse struct {
	ID          string `json:"id"`
	ModelID     string `json:"model_id"`
	Description string `json:"description"`
	ActionCount int    `json:"action_count"`
}

// changesetGetResponse is the JSON response for GET /api/models/{id}/changesets/{csId}.
type changesetGetResponse struct {
	ID          string            `json:"id"`
	ModelID     string            `json:"model_id"`
	Description string            `json:"description"`
	Actions     []changeActionDTO `json:"actions"`
	CreatedAt   string            `json:"created_at"`
}

// impactDelta is a single dimension comparison in the impact response.
type impactDelta struct {
	Dimension string `json:"dimension"`
	Before    string `json:"before"`
	After     string `json:"after"`
	Change    string `json:"change"`
	Detail    string `json:"detail,omitempty"`
}

// impactResponse is the JSON response for GET /api/models/{id}/changesets/{csId}/impact.
type impactResponse struct {
	ChangesetID string        `json:"changeset_id"`
	Deltas      []impactDelta `json:"deltas"`
}

func (h *Handler) handleCreateChangeset(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")

	// Verify model exists.
	if _, err := h.store.Get(modelID); errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	var req changesetCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	cs, err := entity.NewChangeset(req.ID, req.Description)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	for _, actionDTO := range req.Actions {
		if err := cs.AddAction(fromChangeActionDTO(actionDTO)); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	id, err := h.changesetStore.Store(modelID, cs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store changeset: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, changesetCreateResponse{
		ID:          id,
		ModelID:     modelID,
		Description: cs.Description,
		ActionCount: len(cs.Actions),
	})
}

func (h *Handler) handleGetChangeset(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	csID := r.PathValue("csId")

	if _, err := h.store.Get(modelID); errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	sc, err := h.changesetStore.Get(csID)
	if errors.Is(err, usecase.ErrNotFound) || (err == nil && sc.ModelID != modelID) {
		writeError(w, http.StatusNotFound, "changeset not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	actionDTOs := make([]changeActionDTO, len(sc.Changeset.Actions))
	for i, a := range sc.Changeset.Actions {
		actionDTOs[i] = toChangeActionDTO(a)
	}

	writeJSON(w, http.StatusOK, changesetGetResponse{
		ID:          sc.ID,
		ModelID:     sc.ModelID,
		Description: sc.Changeset.Description,
		Actions:     actionDTOs,
		CreatedAt:   sc.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *Handler) handleProjectedModel(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	csID := r.PathValue("csId")

	stored, err := h.store.Get(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	sc, err := h.changesetStore.Get(csID)
	if errors.Is(err, usecase.ErrNotFound) || (err == nil && sc.ModelID != modelID) {
		writeError(w, http.StatusNotFound, "changeset not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	applier := service.NewChangesetApplier()
	projected, err := applier.Apply(stored.Model, sc.Changeset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to apply changeset: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, buildEnrichedNeedView(projected, h.cfg.Analysis))
}

func (h *Handler) handleImpact(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	csID := r.PathValue("csId")

	stored, err := h.store.Get(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	sc, err := h.changesetStore.Get(csID)
	if errors.Is(err, usecase.ErrNotFound) || (err == nil && sc.ModelID != modelID) {
		writeError(w, http.StatusNotFound, "changeset not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	report, err := h.impactAnalyzer.Analyze(stored.Model, sc.Changeset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "impact analysis failed: "+err.Error())
		return
	}

	deltas := make([]impactDelta, 0, len(report.Deltas))
	for _, d := range report.Deltas {
		deltas = append(deltas, impactDelta{
			Dimension: d.Dimension,
			Before:    d.Before,
			After:     d.After,
			Change:    string(d.Change),
			Detail:    d.Detail,
		})
	}

	writeJSON(w, http.StatusOK, impactResponse{
		ChangesetID: sc.Changeset.ID,
		Deltas:      deltas,
	})
}

func (h *Handler) handleApply(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	csID := r.PathValue("csId")

	stored, err := h.store.Get(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	sc, err := h.changesetStore.Get(csID)
	if errors.Is(err, usecase.ErrNotFound) || (err == nil && sc.ModelID != modelID) {
		writeError(w, http.StatusNotFound, "changeset not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	applier := service.NewChangesetApplier()
	projected, err := applier.Apply(stored.Model, sc.Changeset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to apply changeset: "+err.Error())
		return
	}

	summary := projected.Summary()
	writeJSON(w, http.StatusOK, map[string]any{
		"system_name": summary.SystemName,
		"summary": map[string]int{
			"actors":       summary.ActorCount,
			"needs":        summary.NeedCount,
			"capabilities": summary.CapabilityCount,
			"services":     summary.ServiceCount,
			"teams":        summary.TeamCount,
		},
	})
}

// commitResponse is returned by the commit endpoint.
type commitResponse struct {
	ModelID    string                `json:"model_id"`
	SystemName string                `json:"system_name"`
	Summary    map[string]int        `json:"summary"`
	Validation commitValidationResult `json:"validation"`
}

type commitValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func (h *Handler) handleCommitChangeset(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("id")
	csID := r.PathValue("csId")

	stored, err := h.store.Get(modelID)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	sc, err := h.changesetStore.Get(csID)
	if errors.Is(err, usecase.ErrNotFound) || (err == nil && sc.ModelID != modelID) {
		writeError(w, http.StatusNotFound, "changeset not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	applier := service.NewChangesetApplier()
	projected, err := applier.Apply(stored.Model, sc.Changeset)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to apply changeset: "+err.Error())
		return
	}

	validator := service.NewValidationEngine()
	valResult := validator.Validate(projected)

	if !valResult.IsValid() {
		errs := make([]string, 0, len(valResult.Errors))
		for _, e := range valResult.Errors {
			errs = append(errs, fmt.Sprintf("[%s] %s: %s", e.Code, e.Entity, e.Message))
		}
		writeJSON(w, http.StatusConflict, commitResponse{
			ModelID: modelID,
			Validation: commitValidationResult{
				Valid:  false,
				Errors: errs,
			},
		})
		return
	}

	if err := h.store.ReplaceWithMessage(modelID, projected, sc.Changeset.Description); errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model disappeared during commit")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	// Clear insight cache for this model
	h.insightCache.Range(func(key, _ any) bool {
		if k, ok := key.(string); ok && len(k) > len(modelID) && k[:len(modelID)] == modelID {
			h.insightCache.Delete(key)
		}
		return true
	})

	summary := projected.Summary()
	warnings := make([]string, 0, len(valResult.Warnings))
	for _, w := range valResult.Warnings {
		warnings = append(warnings, fmt.Sprintf("[%s] %s: %s", w.Code, w.Entity, w.Message))
	}

	writeJSON(w, http.StatusOK, commitResponse{
		ModelID:    modelID,
		SystemName: summary.SystemName,
		Summary: map[string]int{
			"actors":       summary.ActorCount,
			"needs":        summary.NeedCount,
			"capabilities": summary.CapabilityCount,
			"services":     summary.ServiceCount,
			"teams":        summary.TeamCount,
		},
		Validation: commitValidationResult{
			Valid:    true,
			Warnings: warnings,
		},
	})
}
