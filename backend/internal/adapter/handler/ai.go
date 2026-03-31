package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/krzachariassen/unm-platform/internal/domain/entity"
	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// registerAIRoutes registers the AI advisor endpoints.
func (h *Handler) registerAIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/models/{id}/ask", h.handleAsk)
	mux.HandleFunc("POST /api/models/{id}/extract-actions", h.handleExtractActions)
}

// validAICategories maps user-supplied category names to template paths.
var validAICategories = map[string]string{
	"structural-load":    "advisor/structural-load",
	"service-placement":  "advisor/service-placement",
	"team-boundary":      "advisor/team-boundary",
	"fragmentation":      "advisor/fragmentation",
	"bottleneck":         "advisor/bottleneck",
	"coupling":           "advisor/coupling",
	"interaction-mode":   "advisor/interaction-mode",
	"value-stream":       "advisor/value-stream",
	"need-delivery-risk": "advisor/need-delivery-risk",
	"general":            "advisor/general",
	"recommendations":    "advisor/recommendations",
	"whatif-scenario":    "advisor/whatif-scenario",
	"extract-actions":   "advisor/extract-actions",
	"model-summary":      "query/model-summary",
	"health-summary":     "query/health-summary",
	"natural-language":   "query/natural-language",
}

// askRequest is the JSON body for POST /api/models/{id}/ask.
type askRequest struct {
	Question string `json:"question"`
	Category string `json:"category"` // optional; defaults to "general"
	Tier     string `json:"tier"`     // optional; forces a complexity tier ("simple", "medium", "complex")
}

// askResponse is the JSON response from POST /api/models/{id}/ask.
type askResponse struct {
	ModelID      string          `json:"model_id"`
	Category     string          `json:"category"`
	Question     string          `json:"question"`
	Answer       string          `json:"answer"`
	FinishReason string          `json:"finish_reason,omitempty"`
	Configured   bool            `json:"ai_configured"`
	Routing      *routingInfo    `json:"routing,omitempty"`
}

// routingInfo describes how the AI request was routed based on question complexity.
type routingInfo struct {
	Tier      string `json:"tier"`
	Model     string `json:"model"`
	Reasoning string `json:"reasoning"`
	Timeout   string `json:"timeout"`
}

// reasoningEffortForCategory returns the reasoning effort for a given template name.
// Lookup order: full path with "/" → "_" (e.g. "advisor_recommendations") → prefix → "default".
func (h *Handler) reasoningEffortForCategory(templateName string) string {
	flat := strings.ReplaceAll(templateName, "/", "_")
	if effort, ok := h.cfg.AI.Reasoning[flat]; ok {
		return effort
	}
	key := templateName
	if idx := strings.Index(templateName, "/"); idx >= 0 {
		key = templateName[:idx]
	}
	if effort, ok := h.cfg.AI.Reasoning[key]; ok {
		return effort
	}
	if effort, ok := h.cfg.AI.Reasoning["default"]; ok {
		return effort
	}
	return ""
}

// handleAsk answers a question about a stored model using AI.
func (h *Handler) handleAsk(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	var req askRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Category == "" {
		req.Category = "general"
	}

	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	templateName, ok := validAICategories[req.Category]
	if !ok {
		cats := make([]string, 0, len(validAICategories))
		for k := range validAICategories {
			cats = append(cats, k)
		}
		sort.Strings(cats)
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown category: %q — valid categories: %s", req.Category, strings.Join(cats, ", ")))
		return
	}

	if h.aiClient == nil || !h.aiClient.IsConfigured() {
		writeJSON(w, http.StatusOK, askResponse{
			ModelID:    id,
			Category:   req.Category,
			Question:   req.Question,
			Answer:     "AI advisor is not configured. Set the UNM_OPENAI_API_KEY environment variable to enable AI features.",
			Configured: false,
		})
		return
	}

	m := stored.Model
	data := buildAIPromptData(m, req.Question, h)

	rendered, err := h.promptRenderer.Render(templateName, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("prompt rendering failed: %v", err))
		return
	}

	// Smart routing: classify question complexity and pick appropriate config.
	// If the client sends an explicit tier, use that instead of auto-classification.
	var tier ComplexityTier
	switch req.Tier {
	case "simple":
		tier = TierSimple
	case "medium":
		tier = TierMedium
	case "complex":
		tier = TierComplex
	default:
		tier = ClassifyComplexity(req.Question)
	}
	configKey := templateName
	if req.Category == "general" {
		configKey = TierConfigKey(tier)
	}

	timeout := h.cfg.AI.TimeoutForCategory(configKey)
	model := h.cfg.AI.ModelForCategory(configKey)
	reasoning := h.reasoningEffortForCategory(configKey)
	log.Printf("[AI-ASK] category=%s template=%s tier=%s model=%s reasoning=%s timeout=%s prompt_len=%d",
		req.Category, templateName, tier, model, reasoning, timeout, len(rendered))

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	chatResp, err := h.aiClient.Complete(ctx, rendered, req.Question,
		ai.WithModel(model),
		ai.WithReasoning(reasoning))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("AI request failed: %v", err))
		return
	}

	var routing *routingInfo
	if req.Category == "general" {
		routing = &routingInfo{
			Tier:      string(tier),
			Model:     model,
			Reasoning: reasoning,
			Timeout:   timeout.String(),
		}
	}

	writeJSON(w, http.StatusOK, askResponse{
		ModelID:      id,
		Category:     req.Category,
		Question:     req.Question,
		Answer:       chatResp.Content,
		FinishReason: chatResp.FinishReason,
		Configured:   true,
		Routing:      routing,
	})
}

// extractActionsRequest is the JSON body for POST /api/models/{id}/extract-actions.
type extractActionsRequest struct {
	AdvisorResponse string `json:"advisor_response"`
}

// extractedAction extends ChangeAction with a human-readable reason.
type extractedAction struct {
	entity.ChangeAction
	Reason string `json:"reason"`
}

// extractActionsResponse is the JSON response from POST /api/models/{id}/extract-actions.
type extractActionsResponse struct {
	Actions      []extractedAction `json:"actions"`
	Summary      string            `json:"summary"`
	AIConfigured bool              `json:"ai_configured"`
}

// aiExtractedJSON is the raw JSON returned by the AI for action extraction.
type aiExtractedJSON struct {
	Actions []json.RawMessage `json:"actions"`
	Summary string            `json:"summary"`
}

// handleExtractActions extracts structured ChangeActions from an AI advisor response.
func (h *Handler) handleExtractActions(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	var req extractActionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.AdvisorResponse = strings.TrimSpace(req.AdvisorResponse)
	if req.AdvisorResponse == "" {
		writeError(w, http.StatusBadRequest, "advisor_response is required")
		return
	}

	if h.aiClient == nil || !h.aiClient.IsConfigured() {
		writeJSON(w, http.StatusOK, extractActionsResponse{
			Actions:      nil,
			Summary:      "AI advisor is not configured.",
			AIConfigured: false,
		})
		return
	}

	m := stored.Model
	data := buildAIPromptData(m, "", h)
	data["AdvisorResponse"] = req.AdvisorResponse

	templateName := "advisor/extract-actions"
	rendered, err := h.promptRenderer.Render(templateName, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("prompt rendering failed: %v", err))
		return
	}

	timeout := h.cfg.AI.TimeoutForCategory(templateName)
	model := h.cfg.AI.ModelForCategory(templateName)
	reasoning := h.reasoningEffortForCategory(templateName)
	log.Printf("[AI-EXTRACT] template=%s model=%s reasoning=%s timeout=%s prompt_len=%d",
		templateName, model, reasoning, timeout, len(rendered))

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	rawJSON, err := h.aiClient.CompleteJSON(ctx, rendered,
		"Extract all concrete structural actions from the advisor recommendation as JSON.",
		ai.WithModel(model),
		ai.WithReasoning(reasoning))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("AI request failed: %v", err))
		return
	}

	var parsed aiExtractedJSON
	if err := json.Unmarshal([]byte(rawJSON), &parsed); err != nil {
		log.Printf("[AI-EXTRACT] failed to parse AI JSON response: %v — raw: %s", err, rawJSON)
		writeError(w, http.StatusInternalServerError, "AI returned invalid JSON for action extraction")
		return
	}

	var validActions []extractedAction
	for _, raw := range parsed.Actions {
		var act struct {
			entity.ChangeAction
			Reason string `json:"reason"`
		}
		if err := json.Unmarshal(raw, &act); err != nil {
			log.Printf("[AI-EXTRACT] skipping unparseable action: %v", err)
			continue
		}
		if err := act.ChangeAction.Validate(); err != nil {
			log.Printf("[AI-EXTRACT] skipping invalid action (type=%s): %v", act.Type, err)
			continue
		}
		validActions = append(validActions, extractedAction{
			ChangeAction: act.ChangeAction,
			Reason:       act.Reason,
		})
	}

	log.Printf("[AI-EXTRACT] extracted %d valid actions from %d raw (skipped %d)",
		len(validActions), len(parsed.Actions), len(parsed.Actions)-len(validActions))

	writeJSON(w, http.StatusOK, extractActionsResponse{
		Actions:      validActions,
		Summary:      parsed.Summary,
		AIConfigured: true,
	})
}

// handleExplainChangeset explains the impact of a changeset using AI.
// Route: POST /api/models/{id}/changesets/{csId}/explain
// Note: This is registered by registerChangesetRoutes after changesetStore is available.
func (h *Handler) handleExplainChangeset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	csID := r.PathValue("csId")

	stored := h.store.Get(id)
	if stored == nil {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}

	if h.changesetStore == nil {
		writeError(w, http.StatusServiceUnavailable, "changeset store not available")
		return
	}

	storedCS := h.changesetStore.Get(csID)
	if storedCS == nil {
		writeError(w, http.StatusNotFound, "changeset not found")
		return
	}

	if h.aiClient == nil || !h.aiClient.IsConfigured() {
		writeJSON(w, http.StatusOK, map[string]any{
			"changeset_id":  csID,
			"explanation":   "AI advisor is not configured. Set the UNM_OPENAI_API_KEY environment variable to enable AI features.",
			"ai_configured": false,
		})
		return
	}

	// Delegate impact analysis + data preparation to the ChangesetExplainer use case.
	explainer := usecase.NewChangesetExplainer(h.impactAnalyzer)
	data, err := explainer.PrepareExplainData(stored.Model, storedCS.Changeset, csID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	rendered, err := h.promptRenderer.Render("whatif/impact-assessment", data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("prompt rendering failed: %v", err))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.cfg.AI.TimeoutForCategory("whatif/impact-assessment"))
	defer cancel()

	chatResp, err := h.aiClient.Complete(ctx, rendered, "Explain the impact of this changeset in natural language.",
		ai.WithModel(h.cfg.AI.ModelForCategory("whatif/impact-assessment")),
		ai.WithReasoning(h.reasoningEffortForCategory("whatif/impact-assessment")))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("AI request failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"changeset_id":  csID,
		"explanation":   chatResp.Content,
		"finish_reason": chatResp.FinishReason,
		"ai_configured": true,
	})
}

// buildAIPromptData delegates to the AIContextBuilder use case.
// Kept as a package-level function to preserve its call site in insights.go.
func buildAIPromptData(m *entity.UNMModel, userQuestion string, h *Handler) map[string]any {
	builder := usecase.NewAIContextBuilder(
		h.cognitiveLoad,
		h.valueChain,
		h.fragmentation,
		h.dependency,
		h.gap,
		h.bottleneck,
		h.coupling,
		h.complexity,
		h.interactions,
		h.unlinked,
		h.valueStream,
	)
	data, err := builder.BuildPromptData(m, userQuestion)
	if err != nil {
		// Return minimal data on error — AI prompt will work with less context.
		return map[string]any{
			"SystemName":   m.System.Name,
			"UserQuestion": userQuestion,
			"Question":     userQuestion,
		}
	}
	return data
}
