// Package bootstrap provides test infrastructure for E2E testing.
package bootstrap

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	httpAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	httpMiddleware "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/routing"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// TestServer wraps production app with HTTP test interface.
// This is a THIN WRAPPER that provides test-friendly methods while using
// PRODUCTION app and routes exactly as deployed.
//
// Key design principles:
// - Uses PRODUCTION app from app.Builder.Build()
// - Uses httptest.Server for HTTP testing (standard Go approach)
// - No test-specific routing logic
// - Provides authenticated HTTP methods (GET/POST with Principal injection)
//
// Architecture:
// - Wraps httptest.Server with production route setup
// - Provides convenience methods for authenticated requests
// - Handles principal injection via X-Remote-User header
type TestServer struct {
	app    *app.App
	server *httptest.Server
	logger *slog.Logger
	client *http.Client
}

// TestServerConfig holds parameters for building a TestServer that needs URL alignment.
// This is used when creating a test server where the app's PublicURL must match
// the actual test server's listening address.
type TestServerConfig struct {
	Config *ports.Config
	Logger *slog.Logger
}

// ServerType indicates what type of routes the test server should serve.
type ServerType int

const (
	// ServerTypeEndUser serves end-user routes (OAuth2, consent UI, etc.).
	ServerTypeEndUser ServerType = iota
	// ServerTypeAdmin serves admin routes (agent/service management).
	ServerTypeAdmin

	// DevEndUserPort is the fixed port used for end-user servers in dev mode.
	// This matches the Vite proxy configuration for local development.
	DevEndUserPort = 8000
)

// TestServerOption configures a TestServer.
type TestServerOption func(*testServerOptions)

// testServerOptions holds configuration for building a TestServer.
type testServerOptions struct {
	serverType ServerType
	port       int  // 0 for random port, non-zero for fixed port
	fixedPort  bool // true to use fixed port
}

// validate checks that the options are internally consistent.
func (o *testServerOptions) validate() error {
	if o.fixedPort && o.port == 0 {
		return fmt.Errorf("fixedPort is true but port is 0: specify a port number")
	}
	if !o.fixedPort && o.port != 0 {
		return fmt.Errorf("fixedPort is false but port is %d: use WithFixedPort() to set a fixed port", o.port)
	}
	return nil
}

// WithServerType sets the server type (EndUser or Admin).
func WithServerType(st ServerType) TestServerOption {
	return func(o *testServerOptions) {
		o.serverType = st
	}
}

// WithRandomPort configures the server to use a random port (default).
func WithRandomPort() TestServerOption {
	return func(o *testServerOptions) {
		o.port = 0
		o.fixedPort = false
	}
}

// WithFixedPort configures the server to use a specific port.
// Use this for development mode where the Vite proxy expects port 8000.
func WithFixedPort(port int) TestServerOption {
	return func(o *testServerOptions) {
		o.port = port
		o.fixedPort = true
	}
}

// WithDevMode configures the server for development mode.
// For end-user servers, this uses fixed port 8000 (for Vite proxy).
// For admin servers, this uses a random port.
func WithDevMode() TestServerOption {
	return func(o *testServerOptions) {
		// Only use fixed port for end-user server in dev mode
		if o.serverType == ServerTypeEndUser {
			o.port = DevEndUserPort
			o.fixedPort = true
		} else {
			o.port = 0
			o.fixedPort = false
		}
	}
}

// NewTestServerV2 creates a test server with the specified options.
// This replaces NewTestServer() and provides separate end-user and admin servers.
//
// Example usage:
//
//	// End-user server with random port
//	server, err := NewTestServerV2(app, logger, WithServerType(ServerTypeEndUser))
//
//	// Admin server with random port
//	server, err := NewTestServerV2(app, logger, WithServerType(ServerTypeAdmin))
//
//	// End-user server for dev mode (fixed port 8000)
//	if os.Getenv("E2E_FRONTEND_MODE") == "dev" {
//		server, err := NewTestServerV2(app, logger, WithServerType(ServerTypeEndUser), WithDevMode())
//	}
//
// Parameters:
//   - app: Fully-wired application from app.Builder.Build()
//   - logger: Structured logger
//   - opts: Configuration options
//
// Returns:
//   - *TestServer: Ready to make authenticated requests
//   - error: If server setup fails
func NewTestServerV2(app *app.App, logger *slog.Logger, opts ...TestServerOption) (*TestServer, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Apply options
	options := &testServerOptions{
		serverType: ServerTypeEndUser, // Default to end-user
		port:       0,                 // Default to random port
		fixedPort:  false,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Validate options
	if err := options.validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Determine route setup and server config based on server type.
	// Use production NewHandler to align bootstrap with production server path.
	var (
		routeSetup       func(chi.Router)
		healthComponents func() map[string]string
	)
	serverCfg := httpAdapter.ServerConfig{
		Authentication:   app.Config.Server.EndUser.Authentication,
		JWTAuthenticator: app.JWTAuthenticator,
	}

	switch options.serverType {
	case ServerTypeEndUser:
		// Temporarily disable SPA so SetupEnduserRoutes skips /consent/* registration,
		// then add a catch-all /* instead for test flexibility.
		spaSaved := app.EnduserHandlers.SPA
		app.EnduserHandlers.SPA = nil
		defer func() { app.EnduserHandlers.SPA = spaSaved }()
		healthComponents = app.EnduserHealthComponents
		routeSetup = func(r chi.Router) {
			routing.SetupEnduserRoutes(r, app.EnduserHandlers, routing.EnduserRouteConfig{
				Authentication:   app.Config.Server.EndUser.Authentication,
				JWTAuthenticator: app.JWTAuthenticator,
				Logger:           logger,
				CORS:             app.Config.Server.EndUser.CORS,
				Telemetry:        app.Config.Telemetry,
			})
			if spaSaved != nil {
				r.Handle("/*", spaSaved)
			}
		}

	case ServerTypeAdmin:
		if app.AdminHandlers == nil {
			return nil, fmt.Errorf("admin handlers not available: ensure app was built with admin handlers enabled")
		}
		routeSetup = func(r chi.Router) {
			routing.SetupAdminRoutes(r, app.AdminHandlers, routing.AdminRouteConfig{
				CORS:      app.Config.Server.Admin.CORS,
				Telemetry: app.Config.Telemetry,
			})
		}

	default:
		return nil, fmt.Errorf("unknown server type: %d", options.serverType)
	}

	// Build router using the production NewHandler (same middleware stack as production).
	router := httpAdapter.NewHandler(serverCfg, routeSetup, logger)

	router.Get("/health", httpAdapter.NewHealthHandler(
		func() ports.HealthState { return ports.HealthStateHealthy },
		time.Now(),
		healthComponents,
		logger,
	))

	// Create httptest server with appropriate port configuration
	var server *httptest.Server
	if options.fixedPort {
		// Fixed port mode (for dev mode with Vite proxy)
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", options.port))
		if err != nil {
			return nil, fmt.Errorf("failed to listen on port %d: %w", options.port, err)
		}
		server = &httptest.Server{
			Listener: listener,
			Config:   &http.Server{Handler: router},
		}
		server.Start()
		logger.Info("Test server listening on fixed port",
			"port", options.port,
			"url", server.URL,
			"type", serverTypeName(options.serverType))
	} else {
		// Random port mode (default for test isolation)
		server = httptest.NewServer(router)
		logger.Info("Test server listening on random port",
			"url", server.URL,
			"type", serverTypeName(options.serverType))
	}

	return &TestServer{
		app:    app,
		server: server,
		logger: logger,
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// serverTypeName returns a human-readable name for the server type.
func serverTypeName(st ServerType) string {
	switch st {
	case ServerTypeEndUser:
		return "end-user"
	case ServerTypeAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

// NewEndUserTestServer creates an end-user test server.
// This is a convenience function that wraps NewTestServerV2 with ServerTypeEndUser.
//
// The server will use:
//   - Fixed port 8000 if E2E_FRONTEND_MODE=dev (for Vite proxy)
//   - Random port otherwise (for test isolation)
//
// Example:
//
//	server, err := NewEndUserTestServer(app, logger)
//	require.NoError(t, err)
//	defer server.Close()
func NewEndUserTestServer(app *app.App, logger *slog.Logger) (*TestServer, error) {
	opts := []TestServerOption{WithServerType(ServerTypeEndUser)}

	// In dev mode, use fixed port for Vite proxy
	if os.Getenv("E2E_FRONTEND_MODE") == "dev" {
		opts = append(opts, WithDevMode())
	}

	return NewTestServerV2(app, logger, opts...)
}

// NewAdminTestServer creates an admin test server.
// This is a convenience function that wraps NewTestServerV2 with ServerTypeAdmin.
//
// The server always uses a random port for test isolation.
//
// Example:
//
//	server, err := NewAdminTestServer(app, logger)
//	require.NoError(t, err)
//	defer server.Close()
func NewAdminTestServer(app *app.App, logger *slog.Logger) (*TestServer, error) {
	return NewTestServerV2(app, logger, WithServerType(ServerTypeAdmin))
}

// Close gracefully shuts down the server.
// Safe to call multiple times (httptest.Server.Close is idempotent).
func (ts *TestServer) Close() {
	if ts.server != nil {
		ts.server.Close()
	}
}

// App returns the underlying App instance, giving tests access to domain services
// for building test fixtures (e.g. JWE tokens via OAuth2Service).
func (ts *TestServer) App() *app.App {
	return ts.app
}

// BaseURL returns the server's base URL for requests.
// Returns: "http://127.0.0.1:PORT" format string
func (ts *TestServer) BaseURL() string {
	if ts.server == nil {
		return ""
	}
	return ts.server.URL
}

// AuthenticatedGET makes an authenticated GET request with Principal injection.
// This is a convenience method that adds X-Remote-User header from Principal.
//
// Parameters:
//   - path: Request path (e.g., "/api/agents")
//   - principal: Principal value to inject (e.g., "user@example.com")
//
// Returns:
//   - *http.Response: Response from server (caller must close Body)
//   - error: If request fails
//
// Header injection:
// - Adds X-Remote-User: {principal} header
// - This simulates authentication from reverse proxy
// - Follows production middleware setup exactly
//
// Example:
//
//	resp, err := server.AuthenticatedGET("/api/agents", "user@example.com")
//	require.NoError(t, err)
//	defer resp.Body.Close()
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
func (ts *TestServer) AuthenticatedGET(path string, principal string) (*http.Response, error) {
	if principal == "" {
		return nil, fmt.Errorf("principal is required")
	}

	req, err := http.NewRequest("GET", ts.BaseURL()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Inject principal via header (production middleware configuration)
	req.Header.Set("X-Remote-User", principal)

	// Make request using HTTP client that does NOT follow redirects
	// E2E tests need to verify redirect responses themselves
	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// AuthenticatedPOST makes an authenticated POST request with Principal injection.
// This is a convenience method that adds X-Remote-User header from Principal.
//
// Parameters:
//   - path: Request path (e.g., "/api/agents")
//   - principal: Principal value to inject (e.g., "user@example.com")
//   - contentType: Content-Type header (e.g., "application/json")
//   - body: Request body reader (e.g., strings.NewReader(jsonData))
//
// Returns:
//   - *http.Response: Response from server (caller must close Body)
//   - error: If request fails
//
// Header injection:
// - Adds X-Remote-User: {principal} header
// - Sets Content-Type header as specified
// - Follows production middleware setup exactly
//
// Example:
//
//	jsonBody := strings.NewReader(`{"id":"agent-1"}`)
//	resp, err := server.AuthenticatedPOST("/api/agents", "user@example.com", "application/json", jsonBody)
//	require.NoError(t, err)
//	defer resp.Body.Close()
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
func (ts *TestServer) AuthenticatedPOST(path string, principal string, contentType string, body io.Reader) (*http.Response, error) {
	if principal == "" {
		return nil, fmt.Errorf("principal is required")
	}

	req, err := http.NewRequest("POST", ts.BaseURL()+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Inject principal via header (production middleware configuration)
	req.Header.Set("X-Remote-User", principal)

	// Set content type if provided
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Make request using HTTP client that does NOT follow redirects
	// E2E tests need to verify redirect responses themselves
	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// PublicGET makes an unauthenticated GET request (no Principal).
// Use this for testing public endpoints (health, metadata, etc.).
//
// Parameters:
//   - path: Request path (e.g., "/health")
//
// Returns:
//   - *http.Response: Response from server (caller must close Body)
//   - error: If request fails
//
// Example:
//
//	resp, err := server.PublicGET("/health")
//	require.NoError(t, err)
//	defer resp.Body.Close()
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
func (ts *TestServer) PublicGET(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", ts.BaseURL()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// No authentication header (public endpoint)

	// Make request using HTTP client that does NOT follow redirects
	// E2E tests need to verify redirect responses themselves
	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// PublicPOST makes an unauthenticated POST request (no Principal).
// Use this for public endpoints like /oauth2/token that authenticate via request body (JWTs).
// Do NOT use X-Remote-User header as authentication comes from subject_token and client_assertion.
//
// Parameters:
//   - path: Request path (e.g., "/oauth2/token")
//   - contentType: Content-Type header (e.g., "application/x-www-form-urlencoded")
//   - body: Request body reader (e.g., strings.NewReader(urlEncodedData))
//
// Returns:
//   - *http.Response: Response from server (caller must close Body)
//   - error: If request fails
//
// Example:
//
//	data := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:token-exchange"}, ...}
//	resp, err := server.PublicPOST("/oauth2/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
//	require.NoError(t, err)
//	defer resp.Body.Close()
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
func (ts *TestServer) PublicPOST(path string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", ts.BaseURL()+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type if provided
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// No authentication header (public endpoint, authenticates via request body JWTs)

	// Make request using HTTP client that does NOT follow redirects
	// E2E tests need to verify redirect responses themselves
	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// DirectRequest makes a raw HTTP request (for advanced testing).
// Use AuthenticatedGET/POST for most tests.
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path
//   - principal: Principal to inject (empty for unauthenticated)
//   - headers: Additional headers to set
//   - body: Request body (can be nil)
//
// Returns:
//   - *http.Response: Response from server (caller must close Body)
//   - error: If request fails
func (ts *TestServer) DirectRequest(method string, path string, principal string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, ts.BaseURL()+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add principal if provided
	if principal != "" {
		req.Header.Set("X-Remote-User", principal)
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Make request using HTTP client that does NOT follow redirects
	// E2E tests need to verify redirect responses themselves
	resp, err := ts.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// NewTestServerBuilder creates a test server with proper URL alignment.
// This builder handles the critical flow where the test server's actual listening URL
// must be known BEFORE building the app, because services are initialized with the
// PublicURL during app.Builder.Build().
//
// Flow:
// 1. Accept config and factory
// 2. Create a placeholder httptest server to determine its listening address
// 3. Update config.Server.EndUser.PublicURL with the actual test server URL
// 4. Build the app with the updated config
// 5. Register routes and start the actual server
// 6. Return fully configured TestServer
//
// Parameters:
//   - config: Application configuration (will be modified with actual server URL)
//   - storage: Storage adapter to pass to app builder
//   - factory: ServerFactory to build the app with updated config
//   - logger: Structured logger
//
// Returns:
//   - *TestServer: TestServer with app configured to use its actual listening URL
//   - error: If setup fails
//
// Example:
//
//	config := fixtures.OAuth2ConfigWithPublicURL("https://broker.example.com")
//	builder := NewTestServerBuilder(config, storage, factory, logger)
//	server, err := builder.Build()
//	require.NoError(t, err)
//	defer server.Close()
//	// server.BaseURL() will match what OAuth2 services use internally
func NewTestServerBuilder(
	config *ports.Config,
	storage interface{},
	factory *ServerFactory,
	logger *slog.Logger,
) (*TestServerBuilderImpl, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if storage == nil {
		return nil, fmt.Errorf("storage is required")
	}
	if factory == nil {
		return nil, fmt.Errorf("factory is required")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return &TestServerBuilderImpl{
		config:  config,
		storage: storage,
		factory: factory,
		logger:  logger,
	}, nil
}

// TestServerBuilderImpl implements the builder pattern for TestServer with URL alignment.
type TestServerBuilderImpl struct {
	config  *ports.Config
	storage interface{}
	factory *ServerFactory
	logger  *slog.Logger
}

// Build constructs and returns a fully configured TestServer with URL alignment.
// This uses the elegant approach: create mux first, start server to get port,
// then build app with correct URL and register routes on the existing mux.
//
// Returns:
//   - *TestServer: Fully configured and ready to use
//   - error: If any step fails
func (b *TestServerBuilderImpl) Build() (*TestServer, error) {
	// Step 1: Create chi.Mux with standard middleware
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Apply optional principal middleware to all routes.
	// Must be added before any route registration (chi requires Use() before Get()/Post()/etc.).
	// TestServerBuilderImpl is used only for tests without JWT config, so nil JWTAuthenticator is safe.
	router.Use(httpMiddleware.OptionalPrincipalMiddleware(b.config.Server.EndUser.Authentication, nil, b.logger))

	// Add health endpoint (available immediately)
	router.Get("/health", httpAdapter.NewHealthHandler(
		func() ports.HealthState { return ports.HealthStateHealthy },
		time.Now(),
		nil,
		b.logger,
	))

	// Step 2: Create httptest server with the mux (this assigns a random port)
	testServer := httptest.NewServer(router)

	// Step 3: Get the actual server URL (includes the random port)
	actualURL := testServer.URL

	// Step 4: Update config with the actual server URL
	b.logger.Info(
		"Updating PublicURL for test server alignment",
		"configured_url", b.config.Server.EndUser.PublicURL,
		"actual_url", actualURL,
	)
	b.config.Server.EndUser.PublicURL = actualURL

	// Step 5: Build the app with the correct URL
	// Now services (OAuth2Service, etc.) will be initialized with the correct PublicURL
	appInstance, err := b.factory.BuildApp(b.storage)
	if err != nil {
		testServer.Close()
		return nil, fmt.Errorf("failed to build app: %w", err)
	}

	// Step 6: Register production routes on the existing mux
	// This adds all the actual endpoints while keeping the same httptest server
	routing.SetupEnduserRoutes(router, appInstance.EnduserHandlers, routing.EnduserRouteConfig{
		Authentication:   appInstance.Config.Server.EndUser.Authentication,
		JWTAuthenticator: appInstance.JWTAuthenticator,
		Logger:           b.logger,
		Telemetry:        appInstance.Config.Telemetry,
	})

	b.logger.Info("Test server created and configured", "url", testServer.URL)

	// Step 7: Return the TestServer with aligned URL
	return &TestServer{
		app:    appInstance,
		server: testServer,
		logger: b.logger,
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}
