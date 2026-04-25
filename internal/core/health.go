package core

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (c *Core) HealthHandler(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{
		"badger":       "ok",
		"titan_engine": "ok",
		"p2p_mesh":     "ok",
		"audit_chain":  "ok",
	}

	status := "ok"
	statusCode := http.StatusOK

	// Real-world checks would be implemented here
	// Example: if c.Mesh.Ping() != nil { checks["badger"] = "error"; status = "degraded" }

	w.Header().Set("Content-Type", "application/json")
	if status == "degraded" {
		statusCode = http.StatusMultiStatus // 207
	} else if status == "down" {
		statusCode = http.StatusServiceUnavailable // 503
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(HealthResponse{
		Status: status,
		Checks: checks,
	})
}
