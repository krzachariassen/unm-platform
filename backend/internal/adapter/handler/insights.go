package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/krzachariassen/unm-platform/internal/infrastructure/ai"
	"github.com/krzachariassen/unm-platform/internal/usecase"
)

// signalFinding is a single labelled finding sent to the AI for the signals domain.
// The key is pre-computed so the AI never needs to derive it — it just fills in
// explanation and suggestion and echoes the key back.
type signalFinding struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	Details     string `json:"details,omitempty"`
}

// buildSignalsFindingsList converts a signalsResponse into a flat list of findings
// with pre-computed keys that exactly match what the frontend expects.
func buildSignalsFindingsList(s signalsResponse) []signalFinding {
	var findings []signalFinding
	for _, nr := range s.UserExperienceLayer.NeedsRequiring3PlusTeams {
		findings = append(findings, signalFinding{
			Key:         "need-cross-team:" + slugifyInsightKey(nr.NeedName),
			Type:        "need-cross-team",
			DisplayName: nr.NeedName,
			Details:     fmt.Sprintf("Served by %d teams: %s", nr.TeamSpan, strings.Join(nr.Teams, ", ")),
		})
	}
	for _, nr := range s.UserExperienceLayer.NeedsAtRisk {
		findings = append(findings, signalFinding{
			Key:         "need-unbacked:" + slugifyInsightKey(nr.NeedName),
			Type:        "need-unbacked",
			DisplayName: nr.NeedName,
			Details:     fmt.Sprintf("Served by %d teams: %s", nr.TeamSpan, strings.Join(nr.Teams, ", ")),
		})
	}
	for _, nr := range s.UserExperienceLayer.NeedsWithNoCapBacking {
		findings = append(findings, signalFinding{
			Key:         "need-unbacked:" + slugifyInsightKey(nr.NeedName),
			Type:        "need-no-cap-backing",
			DisplayName: nr.NeedName,
		})
	}
	for _, c := range s.ArchitectureLayer.CapabilitiesNotConnectedToAnyNeed {
		findings = append(findings, signalFinding{
			Key:         "cap-disconnected:" + slugifyInsightKey(c.CapabilityName),
			Type:        "cap-disconnected",
			DisplayName: c.CapabilityName,
			Details:     c.Visibility,
		})
	}
	for _, c := range s.ArchitectureLayer.CapabilitiesFragmentedAcrossTeams {
		findings = append(findings, signalFinding{
			Key:         "cap-fragmented:" + slugifyInsightKey(c.CapabilityName),
			Type:        "cap-fragmented",
			DisplayName: c.CapabilityName,
			Details:     fmt.Sprintf("%d teams: %s", c.TeamCount, strings.Join(c.Teams, ", ")),
		})
	}
	for _, svc := range s.OrganizationLayer.CriticalBottleneckServices {
		findings = append(findings, signalFinding{
			Key:         "bottleneck:" + slugifyInsightKey(svc.ServiceName),
			Type:        "bottleneck",
			DisplayName: svc.ServiceName,
			Details:     fmt.Sprintf("%d dependents", svc.FanIn),
		})
	}
	for _, t := range s.OrganizationLayer.TopTeamsByStructuralLoad {
		findings = append(findings, signalFinding{
			Key:         "team-load:" + slugifyInsightKey(t.TeamName),
			Type:        "team-load",
			DisplayName: t.TeamName,
			Details:     fmt.Sprintf("%s team, %d caps, %d services, load: %s", t.TeamType, t.CapabilityCount, t.ServiceCount, t.OverallLevel),
		})
	}
	for _, t := range s.OrganizationLayer.LowCoherenceTeams {
		findings = append(findings, signalFinding{
			Key:         "team-coherence:" + slugifyInsightKey(t.TeamName),
			Type:        "team-coherence",
			DisplayName: t.TeamName,
			Details:     fmt.Sprintf("coherence: %.0f%%", t.CoherenceScore*100),
		})
	}
	return findings
}


type InsightItem struct {
	Explanation string `json:"explanation"`
	Suggestion  string `json:"suggestion"`
}

// slugifyInsightKey converts a string to a URL-safe slug.
// "Catalog Entity Management" → "catalog-entity-management"
// "actor-Downstream Platform Team" → "actor-downstream-platform-team"
func slugifyInsightKey(s string) string {
	var b strings.Builder
	prevHyphen := true // start true to suppress leading hyphens
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevHyphen = false
		} else if !prevHyphen {
			b.WriteByte('-')
			prevHyphen = true
		}
	}
	return strings.TrimRight(b.String(), "-")
}

// normalizeInsightKeys rewrites insight map keys so the suffix after the first ":"
// is always in slug format. This makes frontend lookups reliable regardless of
// whether the AI returned a human-readable name, entity ID, or slug.
// "interaction:" keys are handled specially — their suffix contains two team IDs
// separated by ":", and each segment is slugified independently.
// e.g. "cap:Administrative & Bulk Operations" → "cap:administrative-bulk-operations"
//      "interaction:INCA Core:INCA Publisher"  → "interaction:inca-core:inca-publisher"
func normalizeInsightKeys(raw map[string]InsightItem) map[string]InsightItem {
	out := make(map[string]InsightItem, len(raw))
	for k, v := range raw {
		idx := strings.IndexByte(k, ':')
		if idx < 0 {
			out[slugifyInsightKey(k)] = v
			continue
		}
		prefix := k[:idx]
		suffix := k[idx+1:]
		if prefix == "interaction" {
			// suffix = "teamA:teamB" — slugify each segment separately
			parts := strings.SplitN(suffix, ":", 2)
			for i, p := range parts {
				parts[i] = slugifyInsightKey(p)
			}
			out[prefix+":"+strings.Join(parts, ":")] = v
		} else {
			out[prefix+":"+slugifyInsightKey(suffix)] = v
		}
	}
	return out
}

// InsightsResponse is the JSON response from GET /api/models/{id}/insights/{domain}.
type InsightsResponse struct {
	Domain       string                 `json:"domain"`
	Status       string                 `json:"status,omitempty"` // "computing", "ready", "failed"
	Insights     map[string]InsightItem `json:"insights"`
	AIConfigured bool                   `json:"ai_configured"`
	Error        string                 `json:"error,omitempty"` // "ai_unavailable", "template_error", "ai_parse_error", "internal_error"
}

// validInsightDomains maps domain names to their template paths.
var validInsightDomains = map[string]string{
	"signals":        "insights/signals",
	"cognitive-load": "insights/cognitive-load",
	"needs":          "insights/needs",
	"capabilities":   "insights/capabilities",
	"ownership":      "insights/ownership",
	"topology":       "insights/topology",
	"dashboard":      "insights/dashboard",
}

// registerInsightsRoutes registers the insights endpoint.
func (h *Handler) registerInsightsRoutes(mux *http.ServeMux) {
	// Register literal /status before wildcard /{domain} (more specific wins in Go 1.22+)
	mux.HandleFunc("GET /api/models/{id}/insights/status", h.handleGetInsightsStatus)
	mux.HandleFunc("GET /api/models/{id}/insights/{domain}", h.handleGetInsights)
}

// handleGetInsightsStatus returns the computation status of all insight domains for a model.
// The frontend polls this endpoint during upload to know when AI pre-computation is done.
func (h *Handler) handleGetInsightsStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := h.store.Get(id); errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	// If AI not configured, all domains are immediately "ready" (no AI to wait for)
	if h.aiClient == nil || !h.aiClient.IsConfigured() {
		domains := make(map[string]string, len(validInsightDomains))
		for domain := range validInsightDomains {
			domains[domain] = "ready"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"domains":   domains,
			"all_ready": true,
		})
		return
	}

	domains := make(map[string]string, len(validInsightDomains))
	allReady := true
	for domain := range validInsightDomains {
		key := id + ":" + domain
		if raw, ok := h.insightCache.Load(key); ok {
			entry, ok := raw.(insightEntry)
			if !ok {
				writeError(w, http.StatusInternalServerError, "insight cache error")
				return
			}
			domains[domain] = entry.status
			if entry.status != "ready" && entry.status != "failed" {
				allReady = false
			}
		} else {
			// With lazy computation, not-yet-requested domains are considered
			// ready (no AI work needed until the user actually views them).
			domains[domain] = "ready"
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"domains":   domains,
		"all_ready": allReady,
	})
}

// precomputeInsights is a no-op kept for backward compatibility. Insights are now
// computed lazily on first request per domain, avoiding 7 concurrent OpenAI calls
// on every model upload.
func (h *Handler) precomputeInsights(modelID string, stored *usecase.StoredModel) {
	// Intentionally empty — insights are computed on demand in handleGetInsights.
}

// computeAndCacheInsight triggers background computation for a single domain if
// not already in progress, and returns immediately. The caller should serve a
// "computing" response while the goroutine runs.
func (h *Handler) computeAndCacheInsight(modelID string, stored *usecase.StoredModel, domain string) {
	key := modelID + ":" + domain
	h.insightCache.Store(key, insightEntry{status: "computing"})
	go func() {
		templateName := validInsightDomains[domain]
		ctx, cancel := context.WithTimeout(context.Background(), h.cfg.AI.TimeoutForCategory(templateName))
		defer cancel()
		result := h.computeInsightForDomain(ctx, stored, domain)
		status := "ready"
		if result.Error != "" {
			status = "failed"
			result.Status = "failed"
		}
		h.insightCache.Store(key, insightEntry{status: status, response: result})
	}()
}

// computeInsightForDomain runs the full AI insight pipeline for one domain and
// returns an InsightsResponse. It never writes to http.ResponseWriter; callers
// decide how to deliver the result.
func (h *Handler) computeInsightForDomain(ctx context.Context, stored *usecase.StoredModel, domain string) InsightsResponse {
	empty := InsightsResponse{Domain: domain, Insights: map[string]InsightItem{}, AIConfigured: true}

	templateName := validInsightDomains[domain]
	m := stored.Model

	var contextData any
	var precomputedKeys []string
	switch domain {
	case "signals":
		findings := buildSignalsFindingsList(h.buildSignalsData(m))
		for _, f := range findings {
			precomputedKeys = append(precomputedKeys, f.Key)
		}
		contextData = findings
	case "cognitive-load":
		contextData = buildEnrichedCognitiveLoadView(m, h.cfg.Analysis)
	case "needs":
		contextData = buildEnrichedNeedView(m, h.cfg.Analysis)
	case "capabilities":
		contextData = buildEnrichedCapabilityView(m, h.cfg.Analysis)
	case "ownership":
		contextData = buildEnrichedOwnershipView(m, h.cfg.Analysis)
	case "topology":
		contextData = buildEnrichedTeamTopologyView(m, h.cfg.Analysis)
	case "dashboard":
		contextData = buildAIPromptData(m, "", h)
	}

	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		empty.Error = "internal_error"
		return empty
	}

	systemPrompt, err := h.promptRenderer.Render(templateName, string(contextJSON))
	if err != nil {
		empty.Error = "template_error"
		return empty
	}

	result, err := h.aiClient.CompleteJSON(ctx, systemPrompt, string(contextJSON),
		ai.WithModel(h.cfg.AI.ModelForCategory(templateName)),
		ai.WithReasoning(h.reasoningEffortForCategory(templateName)))
	if err != nil {
		empty.Error = "ai_unavailable"
		return empty
	}

	var insights map[string]InsightItem
	if precomputedKeys != nil {
		var raw map[string]InsightItem
		if err := json.Unmarshal([]byte(result), &raw); err != nil {
			empty.Error = "ai_parse_error"
			return empty
		}
		normalized := normalizeInsightKeys(raw)
		canonical := make(map[string]struct{}, len(precomputedKeys))
		for _, k := range precomputedKeys {
			canonical[k] = struct{}{}
		}
		insights = make(map[string]InsightItem, len(normalized))
		for k, v := range normalized {
			if _, ok := canonical[k]; ok {
				insights[k] = v
				continue
			}
			matched := false
			for _, pk := range precomputedKeys {
				if strings.HasPrefix(pk, k) {
					insights[pk] = v
					matched = true
					break
				}
			}
			if !matched {
				insights[k] = v
			}
		}
	} else {
		if err := json.Unmarshal([]byte(result), &insights); err != nil {
			empty.Error = "ai_parse_error"
			return empty
		}
		insights = normalizeInsightKeys(insights)
	}

	return InsightsResponse{
		Domain:       domain,
		Status:       "ready",
		Insights:     insights,
		AIConfigured: true,
	}
}

// handleGetInsights returns AI-generated insights for a specific domain of a stored model.
// It serves from the pre-computed cache when available, returning status "computing" while
// the background goroutine is still running, or falls back to on-demand computation.
func (h *Handler) handleGetInsights(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	domain := r.PathValue("domain")

	// Validate domain
	_, ok := validInsightDomains[domain]
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown insight domain: %q", domain))
		return
	}

	// Load model
	stored, err := h.store.Get(id)
	if errors.Is(err, usecase.ErrNotFound) {
		writeError(w, http.StatusNotFound, "model not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "store error: "+err.Error())
		return
	}

	// If AI not configured, return empty insights
	if h.aiClient == nil || !h.aiClient.IsConfigured() {
		writeJSON(w, http.StatusOK, InsightsResponse{
			Domain:       domain,
			Insights:     map[string]InsightItem{},
			AIConfigured: false,
		})
		return
	}

	key := id + ":" + domain
	if raw, ok := h.insightCache.Load(key); ok {
		entry, ok := raw.(insightEntry)
		if !ok {
			writeError(w, http.StatusInternalServerError, "insight cache error")
			return
		}
		if entry.status == "computing" {
			writeJSON(w, http.StatusOK, InsightsResponse{
				Domain:       domain,
				Status:       "computing",
				Insights:     map[string]InsightItem{},
				AIConfigured: true,
			})
			return
		}
		writeJSON(w, http.StatusOK, entry.response)
		return
	}

	// No cached entry — trigger background computation and return "computing"
	h.computeAndCacheInsight(id, stored, domain)
	writeJSON(w, http.StatusOK, InsightsResponse{
		Domain:       domain,
		Status:       "computing",
		Insights:     map[string]InsightItem{},
		AIConfigured: true,
	})
}
