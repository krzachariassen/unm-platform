package handler

import (
	"net/http"
)

// registerQueryRoutes registers GET /api/models/{id}/... query endpoints.
func (h *Handler) registerQueryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/models/{id}/capabilities", h.handleQueryCapabilities)
	mux.HandleFunc("GET /api/models/{id}/teams", h.handleQueryTeams)
	mux.HandleFunc("GET /api/models/{id}/needs", h.handleQueryNeeds)
	mux.HandleFunc("GET /api/models/{id}/services", h.handleQueryServices)
	mux.HandleFunc("GET /api/models/{id}/actors", h.handleQueryActors)
}

// capabilityResponse is the JSON shape for a single capability.
type capabilityResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	IsLeaf      bool   `json:"is_leaf"`
}

// handleQueryCapabilities returns all capabilities in the stored model.
func (h *Handler) handleQueryCapabilities(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	caps := make([]capabilityResponse, 0, len(stored.Model.Capabilities))
	for _, c := range stored.Model.Capabilities {
		caps = append(caps, capabilityResponse{
			Name:        c.Name,
			Description: c.Description,
			Visibility:  c.Visibility,
			IsLeaf:      c.IsLeaf(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"capabilities": caps,
	})
}

// teamResponse is the JSON shape for a single team.
type teamResponse struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	CapabilityCount int    `json:"capability_count"`
	IsOverloaded    bool   `json:"is_overloaded"`
}

// handleQueryTeams returns all teams in the stored model.
func (h *Handler) handleQueryTeams(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	teams := make([]teamResponse, 0, len(stored.Model.Teams))
	for _, t := range stored.Model.Teams {
		teams = append(teams, teamResponse{
			Name:            t.Name,
			Type:            t.TeamType.String(),
			CapabilityCount: t.CapabilityCount(),
			IsOverloaded:    t.IsOverloaded(h.cfg.Analysis.OverloadedCapabilityThreshold),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"teams": teams,
	})
}

// needResponse is the JSON shape for a single need.
type needResponse struct {
	Name      string `json:"name"`
	ActorName string `json:"actor_name"`
	IsMapped  bool   `json:"is_mapped"`
}

// handleQueryNeeds returns all needs in the stored model.
func (h *Handler) handleQueryNeeds(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	needs := make([]needResponse, 0, len(stored.Model.Needs))
	for _, n := range stored.Model.Needs {
		needs = append(needs, needResponse{
			Name:      n.Name,
			ActorName: n.ActorName,
			IsMapped:  n.IsMapped(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"needs": needs,
	})
}

// serviceResponse is the JSON shape for a single service.
type serviceResponse struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	OwnerTeamName string `json:"owner_team_name"`
}

// handleQueryServices returns all services in the stored model.
func (h *Handler) handleQueryServices(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	services := make([]serviceResponse, 0, len(stored.Model.Services))
	for _, s := range stored.Model.Services {
		services = append(services, serviceResponse{
			Name:          s.Name,
			Description:   s.Description,
			OwnerTeamName: s.OwnerTeamName,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"services": services,
	})
}

// actorResponse is the JSON shape for a single actor.
type actorResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// handleQueryActors returns all actors in the stored model.
func (h *Handler) handleQueryActors(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	actors := make([]actorResponse, 0, len(stored.Model.Actors))
	for _, a := range stored.Model.Actors {
		actors = append(actors, actorResponse{
			Name:        a.Name,
			Description: a.Description,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"actors": actors,
	})
}
