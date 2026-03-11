// Package routing provides HTTP route configuration for different server types.
package routing

import (
	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// AdminRouteConfig provides optional configuration for admin route setup.
type AdminRouteConfig struct {
	// CORS configuration for API routes
	CORS ports.CORSConfig
}

// SetupAdminRoutes registers all administrative API routes.
// Routes include agent and service management endpoints.
//
// Route structure:
//
//	POST   /api/agents                 - Create agent
//	GET    /api/agents                 - List agents
//	GET    /api/agents/{agent-id}      - Get agent details
//	PUT    /api/agents/{agent-id}      - Update agent
//	DELETE /api/agents/{agent-id}      - Delete agent
//
//	POST   /api/services               - Create service
//	GET    /api/services               - List services
//	GET    /api/services/{service-id}  - Get service details
//	PUT    /api/services/{service-id}  - Update service
//	DELETE /api/services/{service-id}  - Delete service
func SetupAdminRoutes(r chi.Router, h *app.AdminHandlers, cfg AdminRouteConfig) {
	r.Route("/api", func(r chi.Router) {
		// Apply CORS middleware (no-op if AllowedOrigins empty)
		r.Use(middleware.CORSMiddleware(cfg.CORS))

		// Agent management routes
		r.Route("/agents", func(r chi.Router) {
			r.Post("/", h.Agents.CreateAgent)             // POST /api/agents
			r.Get("/", h.Agents.ListAgents)               // GET /api/agents
			r.Get("/{agent-id}", h.Agents.GetAgent)       // GET /api/agents/:agent-id
			r.Put("/{agent-id}", h.Agents.UpdateAgent)    // PUT /api/agents/:agent-id
			r.Delete("/{agent-id}", h.Agents.DeleteAgent) // DELETE /api/agents/:agent-id
		})

		// Services management routes
		r.Route("/services", func(r chi.Router) {
			r.Post("/", h.Services.CreateService)               // POST /api/services
			r.Get("/", h.Services.ListServices)                 // GET /api/services
			r.Get("/{service-id}", h.Services.GetService)       // GET /api/services/:service-id
			r.Put("/{service-id}", h.Services.UpdateService)    // PUT /api/services/:service-id
			r.Delete("/{service-id}", h.Services.DeleteService) // DELETE /api/services/:service-id
		})
	})
}
