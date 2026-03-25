package handler

import (
	"encoding/json"
	"net/http"
)

// handleHealth responds with {"status":"ok"} for liveness checks.
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
}
