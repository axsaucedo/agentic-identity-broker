// Package http provides HTTP server adapters for the identity broker.
package http

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthResponse represents the JSON response for the health endpoint.
// This structure matches the OpenAPI specification in contracts/health-api.yaml.
type HealthResponse struct {
	Status        string    `json:"status"`         // Health status: "starting", "healthy", "shutting_down", "unhealthy"
	Server        string    `json:"server"`         // Server identifier: "enduser" or "admin"
	Timestamp     time.Time `json:"timestamp"`      // Current server time (ISO 8601)
	UptimeSeconds int64     `json:"uptime_seconds"` // Time since server started serving requests
}

// handleHealth returns an HTTP handler for the health endpoint.
// It reads the server's health state and returns a JSON response with appropriate HTTP status code.
// The handler is designed to be fast and non-blocking, using atomic operations for health state.
func (s *Server) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current health state (atomic read, non-blocking)
		healthState := s.HealthStatus()

		// Calculate uptime since server started serving
		uptime := time.Since(s.startTime)

		// Create response
		response := HealthResponse{
			Status:        healthState.String(),
			Server:        s.name,
			Timestamp:     time.Now().UTC(),
			UptimeSeconds: int64(uptime.Seconds()),
		}

		// Marshal to JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(healthState.HTTPStatus())

		if err := json.NewEncoder(w).Encode(response); err != nil {
			s.logger.Error("Failed to encode health response",
				"error", err,
				"server", s.name)
			// Don't try to write another response; headers already sent
		}
	}
}
