// Package server wires the MCP mock HTTP server using the mcp-go library.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/mock-mcp-server/internal/config"
	"github.com/agentic-identity-broker/mock-mcp-server/internal/handlers"
	"github.com/mark3labs/mcp-go/server"
)

// Server is the MCP mock HTTP server.
type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

// New creates and configures a new MCP mock server using the mcp-go library.
func New(cfg *config.Config) *Server {
	mcpServer := handlers.NewMCPServer()
	streamable := server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithHTTPContextFunc(handlers.AuthHeaderContextFunc),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.Health)
	mux.Handle("/mcp", streamable)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Bind, cfg.Server.Port)
	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start listens and serves HTTP requests (blocks until error or shutdown).
func (s *Server) Start() error {
	addr := fmt.Sprintf("http://%s:%d", s.cfg.Server.Bind, s.cfg.Server.Port)
	slog.Info("MCP mock server starting", "addr", addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
