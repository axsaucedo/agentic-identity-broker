// Package bootstrap provides test infrastructure for E2E testing.
package bootstrap

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

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
}

// TestServerConfig holds parameters for building a TestServer that needs URL alignment.
// This is used when creating a test server where the app's PublicURL must match
// the actual test server's listening address.
type TestServerConfig struct {
	Config *ports.Config
	Logger *slog.Logger
}

// NewTestServer creates a test server from a production app.
// This wraps httptest.Server with production routes and a test-friendly interface.
//
// CRITICAL: This function creates the httptest server which listens on a random port.
// However, if the app.Config.Server.EndUser.PublicURL is already set (e.g., from a test fixture),
// that value is preserved and used by the OAuth2 services. The app was already initialized with
// this PublicURL during app.Builder.Build(), so services have copies of that URL internally.
//
// If you need the test server to use its actual listening address in OAuth2 metadata/redirects,
// you MUST use NewTestServerWithURLUpdate() instead, which rebuilds the app with the correct URL.
//
// Parameters:
//   - app: Fully-wired application from app.Builder.Build()
//   - logger: Structured logger
//
// Returns:
//   - *TestServer: Ready to make authenticated requests
//   - error: If server setup fails
//
// Postconditions:
//   - Server is created and listening (httptest.Server starts automatically)
//   - Call Close() to shut down (typically in defer)
//
// Example:
//
//	server, err := NewTestServer(app, logger)
//	require.NoError(t, err)
//	defer server.Close()
//	// server is ready to make authenticated requests
//
// NOTE: If the app's PublicURL needs to match the test server's actual address,
// use NewTestServerWithURLUpdate() instead.
func NewTestServer(app *app.App, logger *slog.Logger) (*TestServer, error) {
	if app == nil {
		return nil, fmt.Errorf("app is required")
	}

	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Create chi router with production middleware setup
	router := chi.NewRouter()

	// Add standard middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Health endpoint (public)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// Register production routes using production routing setup
	// This ensures proper middleware application: some routes are public (metadata, token),
	// while others require authentication (authorize, consent API)
	routing.SetupEnduserRoutes(router, app.EnduserHandlers, routing.EnduserRouteConfig{
		Authentication: app.Config.Server.EndUser.Authentication,
		Logger:         logger,
	})

	// Create httptest server with production router
	server := httptest.NewServer(router)

	return &TestServer{
		app:    app,
		server: server,
		logger: logger,
	}, nil
}

// Close gracefully shuts down the server.
// Safe to call multiple times (httptest.Server.Close is idempotent).
func (ts *TestServer) Close() {
	if ts.server != nil {
		ts.server.Close()
	}
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
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	resp, err := client.Do(req)
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
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	resp, err := client.Do(req)
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
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	resp, err := client.Do(req)
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
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	resp, err := client.Do(req)
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

	// Add health endpoint (available immediately)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

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
		Authentication: appInstance.Config.Server.EndUser.Authentication,
		Logger:         b.logger,
	})

	b.logger.Info("Test server created and configured", "url", testServer.URL)

	// Step 7: Return the TestServer with aligned URL
	return &TestServer{
		app:    appInstance,
		server: testServer,
		logger: b.logger,
	}, nil
}
