package handler

import (
	"net"
	"net/http"
	"strings"
)

// configResponse is the safe (no secrets) config payload returned by GET /api/config.
type configResponse struct {
	AI       aiConfigResponse       `json:"ai"`
	Analysis analysisConfigResponse `json:"analysis"`
	Features featuresConfigResponse `json:"features"`
}

type featuresConfigResponse struct {
	DebugRoutes bool `json:"debug_routes"`
}

type aiConfigResponse struct {
	Enabled bool `json:"enabled"`
}


type analysisConfigResponse struct {
	DefaultTeamSize               int                        `json:"default_team_size"`
	OverloadedCapabilityThreshold int                        `json:"overloaded_capability_threshold"`
	Bottleneck                    bottleneckConfigResponse   `json:"bottleneck"`
	Signals                       signalsConfigResponse      `json:"signals"`
	ValueChain                    valueChainConfigResponse   `json:"value_chain"`
}

type bottleneckConfigResponse struct {
	FanInWarning  int `json:"fan_in_warning"`
	FanInCritical int `json:"fan_in_critical"`
}

type signalsConfigResponse struct {
	NeedTeamSpanWarning      int `json:"need_team_span_warning"`
	NeedTeamSpanCritical     int `json:"need_team_span_critical"`
	HighSpanServiceThreshold int `json:"high_span_service_threshold"`
	InteractionOverReliance  int `json:"interaction_over_reliance"`
	DepthChainThreshold      int `json:"depth_chain_threshold"`
}

type valueChainConfigResponse struct {
	AtRiskTeamSpan int `json:"at_risk_team_span"`
}

// handleGetConfig returns safe (non-secret) configuration values.
func (h *Handler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	aiEnabled := h.aiClient != nil && h.aiClient.IsConfigured()

	if aiEnabled && len(h.cfg.AI.AllowedIPs) > 0 {
		clientIP := extractClientIP(r)
		aiEnabled = isIPAllowed(clientIP, h.cfg.AI.AllowedIPs)
	}

	resp := configResponse{
		AI: aiConfigResponse{
			Enabled: aiEnabled,
		},
		Features: featuresConfigResponse{
			DebugRoutes: h.cfg.Features.DebugRoutes,
		},
		Analysis: analysisConfigResponse{
			DefaultTeamSize:               h.cfg.Analysis.DefaultTeamSize,
			OverloadedCapabilityThreshold: h.cfg.Analysis.OverloadedCapabilityThreshold,
			Bottleneck: bottleneckConfigResponse{
				FanInWarning:  h.cfg.Analysis.Bottleneck.FanInWarning,
				FanInCritical: h.cfg.Analysis.Bottleneck.FanInCritical,
			},
			Signals: signalsConfigResponse{
				NeedTeamSpanWarning:      h.cfg.Analysis.Signals.NeedTeamSpanWarning,
				NeedTeamSpanCritical:     h.cfg.Analysis.Signals.NeedTeamSpanCritical,
				HighSpanServiceThreshold: h.cfg.Analysis.Signals.HighSpanServiceThreshold,
				InteractionOverReliance:  h.cfg.Analysis.Signals.InteractionOverReliance,
				DepthChainThreshold:      h.cfg.Analysis.Signals.DepthChainThreshold,
			},
			ValueChain: valueChainConfigResponse{
				AtRiskTeamSpan: h.cfg.Analysis.ValueChain.AtRiskTeamSpan,
			},
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

// extractClientIP returns the real client IP, respecting X-Forwarded-For and X-Real-IP proxy headers.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can be a comma-separated list; take the first (original client).
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// isIPAllowed returns true if ip matches any entry in the allowlist.
// Supports exact IPs and CIDR ranges (e.g. "10.0.0.0/8").
func isIPAllowed(ip string, allowlist []string) bool {
	parsed := net.ParseIP(ip)
	for _, entry := range allowlist {
		if strings.Contains(entry, "/") {
			_, network, err := net.ParseCIDR(entry)
			if err == nil && parsed != nil && network.Contains(parsed) {
				return true
			}
		} else if entry == ip {
			return true
		}
	}
	return false
}

