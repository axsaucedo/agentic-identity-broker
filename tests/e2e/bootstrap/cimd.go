package bootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/go-chi/chi/v5"

	adaptercmd "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/cimd"
	httpAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/routing"
	domaincimd "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/cimd"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// CIMDTestHTTPClient returns an *http.Client that routes requests for fakeHostname
// to server. Uses InsecureSkipVerify since the test server cert covers 127.0.0.1,
// not the fake hostname. Safe for test use only.
//
// Use NewCIMDTestFetcher for the common case. Use this when you need to configure
// the client before building a fetcher (e.g., set Timeout for timeout tests).
func CIMDTestHTTPClient(server *httptest.Server, fakeHostname string) *http.Client {
	parsed, _ := url.Parse(server.URL)
	serverAddr := parsed.Host
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, _ := net.SplitHostPort(addr)
				if host == fakeHostname {
					addr = serverAddr
				}
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
		},
	}
}

// NewCIMDTestFetcher returns a ports.CIMDFetcher wired to route requests for
// fakeHostname to server.
func NewCIMDTestFetcher(server *httptest.Server, fakeHostname string, maxResponseBytes int64) (ports.CIMDFetcher, error) {
	return NewCIMDTestFetcherFromClient(CIMDTestHTTPClient(server, fakeHostname), maxResponseBytes)
}

// NewCIMDTestFetcherFromClient builds a CIMDFetcher from a pre-configured *http.Client.
// Use with CIMDTestHTTPClient when custom client configuration is needed (e.g., Timeout).
func NewCIMDTestFetcherFromClient(client *http.Client, maxResponseBytes int64) (ports.CIMDFetcher, error) {
	bl, err := domaincimd.NewSSRFBlocklist(nil)
	if err != nil {
		return nil, err
	}
	return adaptercmd.NewFetcherWithClient(client, bl, maxResponseBytes), nil
}

// NewCIMDEndUserTestServer creates an end-user test server with CIMD fetcher injection and
// URL alignment. URL alignment ensures that the OAuth2 service's consent redirect URL
// (derived from config.Server.EndUser.PublicURL) matches the actual httptest server port.
//
// This uses a two-phase approach:
//  1. Start a minimal httptest.Server to claim a random port.
//  2. Update sf.config.Server.EndUser.PublicURL with that port, then build the app.
//  3. Register end-user routes on the already-started router.
func NewCIMDEndUserTestServer(storage interface{}, sf *ServerFactory, cimdFetcher ports.CIMDFetcher, logger *slog.Logger) (*TestServer, error) {
	serverCfg := httpAdapter.ServerConfig{
		Authentication:   sf.config.Server.EndUser.Authentication,
		JWTAuthenticator: nil,
	}

	// Phase 1: start a bare router (with production middleware) to claim a random port.
	router := httpAdapter.NewHandler(serverCfg, func(r chi.Router) {}, logger)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})
	testServer := httptest.NewServer(router)

	// Phase 2: align PublicURL with the actual listening address, then build the app.
	sf.config.Server.EndUser.PublicURL = testServer.URL
	appInstance, err := sf.BuildAppWithCIMDFetcher(storage, cimdFetcher)
	if err != nil {
		testServer.Close()
		return nil, fmt.Errorf("failed to build CIMD app: %w", err)
	}

	// Phase 3: register end-user routes on the already-started router.
	routing.SetupEnduserRoutes(router, appInstance.EnduserHandlers, routing.EnduserRouteConfig{
		Authentication:   appInstance.Config.Server.EndUser.Authentication,
		JWTAuthenticator: appInstance.JWTAuthenticator,
		Logger:           logger,
		CORS:             appInstance.Config.Server.EndUser.CORS,
		Telemetry:        appInstance.Config.Telemetry,
	})
	if appInstance.EnduserHandlers.SPA != nil {
		router.Handle("/*", appInstance.EnduserHandlers.SPA)
	}

	logger.Info("CIMD test server listening on random port", "url", testServer.URL, "type", "end-user")

	return &TestServer{app: appInstance, server: testServer, logger: logger}, nil
}
