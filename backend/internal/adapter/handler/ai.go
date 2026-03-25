package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/uber/unm-platform/internal/domain/entity"
	"github.com/uber/unm-platform/internal/infrastructure/ai"
	"github.com/uber/unm-platform/internal/infrastructure/analyzer"
)

// registerAIRoutes registers the AI advisor endpoints.
func (h *Handler) registerAIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/models/{id}/ask", h.handleAsk)
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
	"whatif-scenario":     "advisor/whatif-scenario",
	"model-summary":      "query/model-summary",
	"health-summary":     "query/health-summary",
	"natural-language":   "query/natural-language",
}

// askRequest is the JSON body for POST /api/models/{id}/ask.
type askRequest struct {
	Question string `json:"question"`
	Category string `json:"category"` // optional; defaults to "general"
}

// askResponse is the JSON response from POST /api/models/{id}/ask.
type askResponse struct {
	ModelID      string `json:"model_id"`
	Category     string `json:"category"`
	Question     string `json:"question"`
	Answer       string `json:"answer"`
	FinishReason string `json:"finish_reason,omitempty"`
	Configured   bool   `json:"ai_configured"`
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
		writeError(w, http.StatusBadRequest, "question must not be empty")
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

	lib, err := ai.NewPromptLibrary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load prompt library")
		return
	}
	renderer := ai.NewPromptRenderer(lib)

	rendered, err := renderer.Render(templateName, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("prompt rendering failed: %v", err))
		return
	}

	timeout := h.cfg.AI.TimeoutForCategory(templateName)
	model := h.cfg.AI.ModelForCategory(templateName)
	reasoning := h.reasoningEffortForCategory(templateName)
	log.Printf("[AI-ASK] category=%s template=%s model=%s reasoning=%s timeout=%s prompt_len=%d",
		req.Category, templateName, model, reasoning, timeout, len(rendered))

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	chatResp, err := h.aiClient.Complete(ctx, rendered, req.Question,
		ai.WithModel(model),
		ai.WithReasoning(reasoning))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("AI request failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, askResponse{
		ModelID:      id,
		Category:     req.Category,
		Question:     req.Question,
		Answer:       chatResp.Content,
		FinishReason: chatResp.FinishReason,
		Configured:   true,
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
			"changeset_id": csID,
			"explanation":  "AI advisor is not configured. Set the UNM_OPENAI_API_KEY environment variable to enable AI features.",
			"ai_configured": false,
		})
		return
	}

	// Compute the impact
	impact, err := h.impactAnalyzer.Analyze(stored.Model, storedCS.Changeset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("impact analysis failed: %v", err))
		return
	}

	// Serialize impact and changeset for the prompt
	actions := make([]string, len(storedCS.Changeset.Actions))
	for i, a := range storedCS.Changeset.Actions {
		actions[i] = fmt.Sprintf("- %s: %+v", a.Type, a)
	}
	actionsText := ""
	for _, a := range actions {
		actionsText += a + "\n"
	}

	deltasText := ""
	for _, d := range impact.Deltas {
		deltasText += fmt.Sprintf("- %s: %s → %s (%s)\n", d.Dimension, d.Before, d.After, d.Change)
		if d.Detail != "" {
			deltasText += fmt.Sprintf("  Detail: %s\n", d.Detail)
		}
	}

	data := map[string]any{
		"SystemName":      stored.Model.System.Name,
		"ChangesetID":     csID,
		"ChangesetActions": actionsText,
		"ImpactDeltas":    deltasText,
	}

	lib, err := ai.NewPromptLibrary()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load prompt library")
		return
	}
	renderer := ai.NewPromptRenderer(lib)

	rendered, err := renderer.Render("whatif/impact-assessment", data)
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

// teamSummary is used as prompt template data for teams.
type teamSummary struct {
	Name            string
	TeamType        string
	Size            int
	CognitiveLoad   string
	ServiceCount    int
	CapabilityCount int
	Services        []string
	Capabilities    []string
}

// serviceSummary is used as prompt template data for services.
type serviceSummary struct {
	Name            string
	OwnerTeam       string
	DependencyCount int
	DependsOn       []string
	Capabilities    []string
}

// capSummary is used as prompt template data for capabilities.
type capSummary struct {
	Name              string
	Visibility        string
	OwnerTeams        []string
	RealizingServices []string
}

// needSummary is used as prompt template data for needs.
type needSummary struct {
	Name                   string
	Actor                  string
	SupportedByCapabilities []string
	TeamSpan               int
	AtRisk                 bool
}

// interactionSummary is used as prompt template data for interactions.
type interactionSummary struct {
	From        string
	To          string
	Mode        string
	Via         string
	Description string
}

// signalSummary is used as prompt template data for signals.
type signalSummary struct {
	Category         string
	Severity         string
	Description      string
	AffectedEntities string
	Evidence         string
}

// bottleneckSummary is used as prompt template data for service bottlenecks.
type bottleneckSummary struct {
	Service    string
	FanIn      int
	FanOut     int
	IsCritical bool
}

// couplingSummary is used as prompt template data for data-asset coupling.
type couplingSummary struct {
	DataAsset   string
	Type        string
	Services    []string
	IsCrossteam bool
}

// gapSummary holds gap analysis findings.
type gapSummary struct {
	UnmappedNeeds          []string
	UnrealizedCapabilities []string
	UnownedServices        []string
	UnneededCapabilities   []string
}

// complexitySummary is used as prompt template data for service complexity.
type complexitySummary struct {
	Service         string
	DependencyScore int
	CapabilityScore int
	DataAssetScore  int
	TotalComplexity int
}

// valueStreamSummary is used as prompt template data for value stream coherence.
type valueStreamSummary struct {
	TeamName       string
	NeedsServed    []string
	CoherenceScore float64
	LowCoherence   bool
}

// externalDepSummary is used as prompt template data for external dependencies.
type externalDepSummary struct {
	Name    string
	UsedBy  []string
}

// unlinkedCapSummary is used as prompt template data for unlinked capabilities.
type unlinkedCapSummary struct {
	Name       string
	Visibility string
}

// buildAIPromptData assembles all model data and ALL analysis results into a map
// suitable for rendering AI prompt templates. Every analyzer is run so the AI
// has the complete picture: cognitive load, bottlenecks, gaps, coupling,
// complexity, dependency cycles, value stream coherence, and more.
func buildAIPromptData(m *entity.UNMModel, userQuestion string, h *Handler) map[string]any {
	// ── Run ALL analyzers ───────────────────────────────────────────────
	clReport := h.cognitiveLoad.Analyze(m)
	vcReport := h.valueChain.Analyze(m)
	fragReport := h.fragmentation.Analyze(m)
	depReport := h.dependency.Analyze(m)
	gapReport := h.gap.Analyze(m)
	bnReport := h.bottleneck.Analyze(m)
	cpReport := h.coupling.Analyze(m)
	cxReport := h.complexity.Analyze(m)
	ixDivReport := h.interactions.Analyze(m)
	unlReport := h.unlinked.Analyze(m)
	vsReport := h.valueStream.Analyze(m)

	// Cognitive load by team for quick lookup
	loadByTeam := make(map[string]string)
	for _, tl := range clReport.TeamLoads {
		loadByTeam[tl.Team.Name] = string(tl.OverallLevel)
	}

	// Value chain by need for quick lookup
	vcByNeed := make(map[string]analyzer.NeedDeliveryRisk)
	for _, ndr := range vcReport.NeedRisks {
		vcByNeed[ndr.NeedName] = ndr
	}

	// Service-to-capabilities reverse index
	svcToCaps := make(map[string][]string)
	for _, cap := range m.Capabilities {
		for _, rel := range cap.RealizedBy {
			svcName := rel.TargetID.String()
			svcToCaps[svcName] = append(svcToCaps[svcName], cap.Name)
		}
	}

	// ── Teams ───────────────────────────────────────────────────────────
	teams := make([]teamSummary, 0, len(m.Teams))
	for _, t := range m.Teams {
		var svcNames []string
		for _, svc := range m.Services {
			if svc.OwnerTeamName == t.Name {
				svcNames = append(svcNames, svc.Name)
			}
		}
		capNames := make([]string, 0, len(t.Owns))
		for _, rel := range t.Owns {
			capNames = append(capNames, rel.TargetID.String())
		}
		teams = append(teams, teamSummary{
			Name:            t.Name,
			TeamType:        string(t.TeamType),
			Size:            t.EffectiveSize(),
			CognitiveLoad:   loadByTeam[t.Name],
			ServiceCount:    len(svcNames),
			CapabilityCount: len(capNames),
			Services:        svcNames,
			Capabilities:    capNames,
		})
	}
	sort.Slice(teams, func(i, j int) bool { return teams[i].Name < teams[j].Name })

	// ── Services ────────────────────────────────────────────────────────
	services := make([]serviceSummary, 0, len(m.Services))
	for _, svc := range m.Services {
		depTargets := make([]string, 0, len(svc.DependsOn))
		for _, rel := range svc.DependsOn {
			depTargets = append(depTargets, rel.TargetID.String())
		}
		services = append(services, serviceSummary{
			Name:            svc.Name,
			OwnerTeam:       svc.OwnerTeamName,
			DependencyCount: len(svc.DependsOn),
			DependsOn:       depTargets,
			Capabilities:    svcToCaps[svc.Name],
		})
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })

	// ── Capabilities ────────────────────────────────────────────────────
	caps := make([]capSummary, 0, len(m.Capabilities))
	for _, cap := range m.Capabilities {
		ownerTeams := make([]string, 0)
		for _, t := range m.Teams {
			for _, rel := range t.Owns {
				if rel.TargetID.String() == cap.Name {
					ownerTeams = append(ownerTeams, t.Name)
				}
			}
		}
		realizingServices := make([]string, 0, len(cap.RealizedBy))
		for _, rel := range cap.RealizedBy {
			realizingServices = append(realizingServices, rel.TargetID.String())
		}
		caps = append(caps, capSummary{
			Name:              cap.Name,
			Visibility:        cap.Visibility,
			OwnerTeams:        ownerTeams,
			RealizingServices: realizingServices,
		})
	}
	sort.Slice(caps, func(i, j int) bool { return caps[i].Name < caps[j].Name })

	// ── Needs ───────────────────────────────────────────────────────────
	needs := make([]needSummary, 0, len(m.Needs))
	for _, n := range m.Needs {
		suppBy := make([]string, 0, len(n.SupportedBy))
		for _, rel := range n.SupportedBy {
			suppBy = append(suppBy, rel.TargetID.String())
		}
		ndr := vcByNeed[n.Name]
		needs = append(needs, needSummary{
			Name:                    n.Name,
			Actor:                   n.ActorName,
			SupportedByCapabilities: suppBy,
			TeamSpan:                ndr.TeamSpan,
			AtRisk:                  ndr.AtRisk,
		})
	}
	sort.Slice(needs, func(i, j int) bool { return needs[i].Name < needs[j].Name })

	// ── Interactions (with via/description) ──────────────────────────────
	interactions := make([]interactionSummary, 0, len(m.Interactions))
	for _, ix := range m.Interactions {
		interactions = append(interactions, interactionSummary{
			From:        ix.FromTeamName,
			To:          ix.ToTeamName,
			Mode:        string(ix.Mode),
			Via:         ix.Via,
			Description: ix.Description,
		})
	}

	// ── Signals (with evidence) ─────────────────────────────────────────
	signals := make([]signalSummary, 0, len(m.Signals))
	for _, s := range m.Signals {
		affectedStr := strings.Join(s.AffectedEntities, ", ")
		signals = append(signals, signalSummary{
			Category:         string(s.Category),
			Severity:         string(s.Severity),
			Description:      s.Description,
			AffectedEntities: affectedStr,
			Evidence:         s.Evidence,
		})
	}

	// ── Value chains ────────────────────────────────────────────────────
	type valueChainEntry struct {
		Actor        string
		Need         string
		Capabilities []string
		Services     []string
		Teams        []string
	}
	var valueChains []valueChainEntry
	for _, n := range m.Needs {
		vc := valueChainEntry{Actor: n.ActorName, Need: n.Name}
		teamSet := make(map[string]bool)
		for _, rel := range n.SupportedBy {
			capName := rel.TargetID.String()
			vc.Capabilities = append(vc.Capabilities, capName)
			for _, cap := range m.Capabilities {
				if cap.Name == capName {
					for _, rRel := range cap.RealizedBy {
						svcName := rRel.TargetID.String()
						vc.Services = append(vc.Services, svcName)
						for _, svc := range m.Services {
							if svc.Name == svcName {
								teamSet[svc.OwnerTeamName] = true
							}
						}
					}
				}
			}
		}
		for t := range teamSet {
			vc.Teams = append(vc.Teams, t)
		}
		valueChains = append(valueChains, vc)
	}

	// ── Fragmented capabilities (from analyzer, not just ownership count) ──
	type fragmentedCap struct {
		Name       string
		Teams      []string
		Visibility string
	}
	var fragmentedCaps []fragmentedCap
	for _, fc := range fragReport.FragmentedCapabilities {
		teamNames := make([]string, 0, len(fc.Teams))
		for _, t := range fc.Teams {
			teamNames = append(teamNames, t.Name)
		}
		fragmentedCaps = append(fragmentedCaps, fragmentedCap{
			Name:       fc.Capability.Name,
			Teams:      teamNames,
			Visibility: fc.Capability.Visibility,
		})
	}

	// ── Cognitive load detail per team ───────────────────────────────────
	type cogLoadDetail struct {
		Team               string
		OverallLevel       string
		DomainSpread       string
		DomainSpreadVal    int
		ServiceLoad        string
		ServiceLoadVal     int
		InteractionLoad    string
		InteractionLoadVal int
		DependencyLoad     string
		DependencyLoadVal  int
		ServiceCount       int
		CapabilityCount    int
		TeamSize           int
	}
	var cogLoadDetails []cogLoadDetail
	for _, tl := range clReport.TeamLoads {
		cogLoadDetails = append(cogLoadDetails, cogLoadDetail{
			Team:               tl.Team.Name,
			OverallLevel:       string(tl.OverallLevel),
			DomainSpread:       string(tl.DomainSpread.Level),
			DomainSpreadVal:    tl.DomainSpread.Value,
			ServiceLoad:        string(tl.ServiceLoad.Level),
			ServiceLoadVal:     tl.ServiceLoad.Value,
			InteractionLoad:    string(tl.InteractionLoad.Level),
			InteractionLoadVal: tl.InteractionLoad.Value,
			DependencyLoad:     string(tl.DependencyLoad.Level),
			DependencyLoadVal:  tl.DependencyLoad.Value,
			ServiceCount:       tl.ServiceCount,
			CapabilityCount:    tl.CapabilityCount,
			TeamSize:           tl.TeamSize,
		})
	}

	// ── Bottleneck analysis ─────────────────────────────────────────────
	bottlenecks := make([]bottleneckSummary, 0)
	for _, b := range bnReport.ServiceBottlenecks {
		if b.IsCritical || b.IsWarning {
			bottlenecks = append(bottlenecks, bottleneckSummary{
				Service:    b.Service.Name,
				FanIn:      b.FanIn,
				FanOut:     b.FanOut,
				IsCritical: b.IsCritical,
			})
		}
	}

	// ── Coupling analysis ───────────────────────────────────────────────
	couplings := make([]couplingSummary, 0, len(cpReport.DataAssetCouplings))
	for _, c := range cpReport.DataAssetCouplings {
		assetType := ""
		if c.DataAsset != nil {
			assetType = c.DataAsset.Type
		}
		couplings = append(couplings, couplingSummary{
			DataAsset:   c.DataAsset.Name,
			Type:        assetType,
			Services:    c.Services,
			IsCrossteam: c.IsCrossteam,
		})
	}

	// ── Gap analysis ────────────────────────────────────────────────────
	gaps := gapSummary{}
	for _, n := range gapReport.UnmappedNeeds {
		gaps.UnmappedNeeds = append(gaps.UnmappedNeeds, n.Name+" (actor: "+n.ActorName+")")
	}
	for _, c := range gapReport.UnrealizedCapabilities {
		gaps.UnrealizedCapabilities = append(gaps.UnrealizedCapabilities, c.Name)
	}
	for _, s := range gapReport.UnownedServices {
		gaps.UnownedServices = append(gaps.UnownedServices, s.Name)
	}
	for _, c := range gapReport.UnneededCapabilities {
		gaps.UnneededCapabilities = append(gaps.UnneededCapabilities, c.Name)
	}

	// ── Complexity analysis (top services) ──────────────────────────────
	complexities := make([]complexitySummary, 0, len(cxReport.Services))
	for _, s := range cxReport.Services {
		complexities = append(complexities, complexitySummary{
			Service:         s.Service.Name,
			DependencyScore: s.DependencyScore,
			CapabilityScore: s.CapabilityScore,
			DataAssetScore:  s.DataAssetScore,
			TotalComplexity: s.TotalComplexity,
		})
	}

	// ── Dependency analysis ─────────────────────────────────────────────
	svcCycles := make([][]string, 0)
	for _, c := range depReport.ServiceCycles {
		svcCycles = append(svcCycles, c.Path)
	}
	capCycles := make([][]string, 0)
	for _, c := range depReport.CapabilityCycles {
		capCycles = append(capCycles, c.Path)
	}
	critPath := depReport.CriticalServicePath
	if critPath == nil {
		critPath = []string{}
	}

	// ── Value stream coherence ──────────────────────────────────────────
	valueStreams := make([]valueStreamSummary, 0, len(vsReport.TeamCoherences))
	for _, tc := range vsReport.TeamCoherences {
		valueStreams = append(valueStreams, valueStreamSummary{
			TeamName:       tc.TeamName,
			NeedsServed:    tc.NeedsServed,
			CoherenceScore: tc.CoherenceScore,
			LowCoherence:   tc.LowCoherence,
		})
	}

	// ── Interaction diversity ───────────────────────────────────────────
	modeDist := make(map[string]int, len(ixDivReport.ModeDistribution))
	for mode, count := range ixDivReport.ModeDistribution {
		modeDist[string(mode)] = count
	}
	type overReliantTeam struct {
		TeamName string
		Mode     string
		Count    int
	}
	overReliant := make([]overReliantTeam, 0, len(ixDivReport.OverReliantTeams))
	for _, or_ := range ixDivReport.OverReliantTeams {
		overReliant = append(overReliant, overReliantTeam{
			TeamName: or_.TeamName,
			Mode:     string(or_.Mode),
			Count:    or_.Count,
		})
	}
	isolatedTeams := ixDivReport.IsolatedTeams
	if isolatedTeams == nil {
		isolatedTeams = []string{}
	}

	// ── Unlinked capabilities ───────────────────────────────────────────
	unlinkedCaps := make([]unlinkedCapSummary, 0)
	for _, uc := range unlReport.UnlinkedLeafCapabilities {
		if !uc.IsExpected {
			unlinkedCaps = append(unlinkedCaps, unlinkedCapSummary{
				Name:       uc.Capability.Name,
				Visibility: uc.Visibility,
			})
		}
	}

	// ── External dependencies ───────────────────────────────────────────
	externalDeps := make([]externalDepSummary, 0, len(m.ExternalDependencies))
	for _, ed := range m.ExternalDependencies {
		usedBy := make([]string, 0, len(ed.UsedBy))
		for _, u := range ed.UsedBy {
			usedBy = append(usedBy, u.ServiceName)
		}
		externalDeps = append(externalDeps, externalDepSummary{
			Name:   ed.Name,
			UsedBy: usedBy,
		})
	}

	// ── Value chain aggregate counts ────────────────────────────────────
	type vcCounts struct {
		CrossTeamNeeds int
		AtRiskNeeds    int
		UnbackedNeeds  int
	}
	vcAgg := vcCounts{
		CrossTeamNeeds: vcReport.CrossTeamNeedCount,
		AtRiskNeeds:    vcReport.AtRiskNeedCount,
		UnbackedNeeds:  vcReport.UnbackedNeedCount,
	}

	return map[string]any{
		"SystemName":             m.System.Name,
		"ModelDescription":       m.System.Description,
		"Teams":                  teams,
		"Services":               services,
		"Capabilities":           caps,
		"Needs":                  needs,
		"Interactions":           interactions,
		"Signals":                signals,
		"ValueChains":            valueChains,
		"FragmentedCapabilities": fragmentedCaps,
		"CognitiveLoadDetails":   cogLoadDetails,
		"UserQuestion":           userQuestion,
		"Question":               userQuestion,
		// NEW: full analysis data
		"Bottlenecks":            bottlenecks,
		"Couplings":              couplings,
		"Gaps":                   gaps,
		"Complexities":           complexities,
		"ServiceCycles":          svcCycles,
		"CapabilityCycles":       capCycles,
		"CriticalServicePath":    critPath,
		"MaxServiceDepth":        depReport.MaxServiceDepth,
		"MaxCapabilityDepth":     depReport.MaxCapabilityDepth,
		"ValueStreamCoherence":   valueStreams,
		"LowCoherenceCount":      vsReport.LowCoherenceCount,
		"ModeDistribution":       modeDist,
		"OverReliantTeams":       overReliant,
		"IsolatedTeams":          isolatedTeams,
		"AllModesSame":           ixDivReport.AllModesSame,
		"UnlinkedCapabilities":   unlinkedCaps,
		"ExternalDependencies":   externalDeps,
		"ValueChainCounts":       vcAgg,
	}
}
