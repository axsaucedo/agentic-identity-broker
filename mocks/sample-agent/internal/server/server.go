package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/sample-agent/internal/config"
	"github.com/agentic-identity-broker/sample-agent/internal/handlers"
	"golang.org/x/oauth2"
)

// Server holds the HTTP server and OAuth2 configuration
type Server struct {
	router       *http.ServeMux
	httpServer   *http.Server
	cfg          *config.Config
	oauth2Config *oauth2.Config
}

// New creates and configures a new OAuth2 client server
func New(cfg *config.Config) (*Server, error) {
	// Create OAuth2 configuration
	oauth2Cfg := &oauth2.Config{
		ClientID:     cfg.OAuth2.ClientID,
		ClientSecret: cfg.OAuth2.ClientSecret,
		RedirectURL:  cfg.OAuth2.RedirectURI,
		Scopes:       cfg.OAuth2.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.BrokerInfo.AuthorizeEndpoint,
			TokenURL: cfg.BrokerInfo.TokenEndpoint,
		},
	}

	slog.Info("Sample OAuth2 client configured",
		"client_id", cfg.OAuth2.ClientID,
		"redirect_uri", cfg.OAuth2.RedirectURI,
		"broker_auth_endpoint", cfg.BrokerInfo.AuthorizeEndpoint,
		"broker_token_endpoint", cfg.BrokerInfo.TokenEndpoint)

	// Create router
	router := http.NewServeMux()

	// Create handlers
	h := handlers.New(oauth2Cfg, cfg)

	// Register routes
	router.HandleFunc("/health", h.Health)
	router.HandleFunc("/", h.Home)
	router.HandleFunc("/login", h.Login)
	router.HandleFunc("/oauth2/callback", h.Callback)
	router.HandleFunc("/logout", h.Logout)
	router.HandleFunc("/call-mcp", h.CallMCP)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Bind, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		router:       router,
		httpServer:   httpServer,
		cfg:          cfg,
		oauth2Config: oauth2Cfg,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("http://%s:%d", s.cfg.Server.Bind, s.cfg.Server.Port)
	slog.Info("Starting sample OAuth2 client server", "addr", addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
