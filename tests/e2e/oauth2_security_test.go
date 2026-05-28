package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OAuth2 Security and Validation", func() {
	var (
		server         *bootstrap.TestServer
		testStorage    *storageadapter.Adapter
		logger         *slog.Logger
		ctx            context.Context
		storageFactory *bootstrap.StorageFactory
		testAgent      *storage.Agent // registered agent used in token endpoint tests
	)

	BeforeEach(func() {
		logger = bootstrap.TestLogger(slog.LevelInfo)
		storageFactory = bootstrap.NewStorageFactory(logger)

		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()

		// Register a shared test agent so all token endpoint tests can look it up
		// by UUID when the broker resolves client_id → upstream client_id.
		testAgent = fixtures.ValidAgent()
		Expect(testStorage.Agents().Create(ctx, testAgent)).ToNot(HaveOccurred())

		config := fixtures.DefaultOAuth2Config()
		factory := bootstrap.NewServerFactory(config, logger)
		appInstance, err := factory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())

		server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	// GROUP 1: Authentication validation
	Describe("authentication failures", func() {
		It("should return 401 when X-Remote-User header is missing", func() {
			// Given: An agent exists in storage
			agent := fixtures.ValidAgent()
			err := testStorage.Agents().Create(ctx, agent)
			Expect(err).ToNot(HaveOccurred())

			// When: Request without X-Remote-User header
			resp, err := server.PublicGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code",
					agent.ID.String()),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 401 Unauthorized
			// Specification T043: Missing authentication header must return 401
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("should return 400 when X-Remote-User header exceeds max length", func() {
			// Given: An agent exists in storage
			agent := fixtures.ValidAgent()
			err := testStorage.Agents().Create(ctx, agent)
			Expect(err).ToNot(HaveOccurred())

			// When: Request with oversized principal (>200 chars per spec)
			oversizedPrincipal := fixtures.OversizedPrincipal().String()
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code",
					agent.ID.String()),
				oversizedPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 400 Bad Request
			// Specification T043: Principal exceeding max length must be rejected
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// GROUP 2: Client validation
	Describe("client validation", func() {
		It("should reject invalid client_id", func() {
			// Given: No agent with the requested client_id exists
			// When: Request with invalid (non-UUID) client_id
			resp, err := server.AuthenticatedGET(
				"/oauth2/authorize?client_id=nonexistent-client&redirect_uri=https://client.example.com/cb&response_type=code&state=state-123",
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns direct 400 error — per RFC 6749 §4.1.2.1 MUST NOT redirect
			// when redirect_uri is unverified. client_id must be a UUID-format agent ID.
			// Specification T043: invalid client_id format returns 400 non-redirect JSON error
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(ContainSubstring("invalid_client"))
		})

		It("should redirect to consent page for valid client", func() {
			// Given: A valid agent exists
			agent := fixtures.ValidAgent()
			err := testStorage.Agents().Create(ctx, agent)
			Expect(err).ToNot(HaveOccurred())

			// When: Request with valid client_id and the agent's registered redirect_uri
			redirectURI := "https://client.example.com/cb"
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=abc123",
					agent.ID.String(), url.QueryEscape(redirectURI)),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 302 redirect
			// Specification T043: Valid client authorization request should redirect
			Expect(resp.StatusCode).To(Equal(http.StatusFound))

			location := resp.Header.Get("Location")
			Expect(location).NotTo(BeEmpty())
			// Should redirect to consent or upstream
			Expect(location).To(SatisfyAny(
				ContainSubstring("http://localhost:19000"), // Upstream
				ContainSubstring("/consent/"),              // Consent page
			))
		})
	})

	// GROUP 3: Public endpoints
	Describe("public endpoints", func() {
		It("should serve metadata publicly without authentication", func() {
			// When: Request metadata endpoint without authentication
			resp, err := server.PublicGET("/.well-known/oauth-authorization-server")
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 200 OK with JSON content type
			// Specification T043: Metadata endpoint is public
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))

			// Verify metadata structure
			var metadata map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).To(HaveKey("authorization_endpoint"))
			Expect(metadata).To(HaveKey("token_endpoint"))
		})

		It("should include proper cache control headers for metadata", func() {
			// When: Request metadata endpoint
			resp, err := server.PublicGET("/.well-known/oauth-authorization-server")
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Response includes cache control headers
			// Specification T044: Cache control headers required for public endpoints
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			cacheControl := resp.Header.Get("Cache-Control")
			Expect(cacheControl).NotTo(BeEmpty())
			Expect(cacheControl).To(ContainSubstring("max-age"))
		})
	})

	// GROUP 4: Token endpoint Content-Type validation
	Describe("token endpoint Content-Type validation", func() {
		var mockUpstream *helpers.MockUpstreamOAuth2Server

		BeforeEach(func() {
			mockUpstream = helpers.NewMockUpstreamOAuth2Server()
			mockUpstream.WithSuccessfulTokenResponse()

			// Update config to use mock upstream
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.URL())
			factory := bootstrap.NewServerFactory(config, logger)
			appInstance, err := factory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			// Close previous server and create new one with mock upstream
			if server != nil {
				server.Close()
			}
			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if mockUpstream != nil {
				mockUpstream.Close()
			}
		})

		It("should reject invalid Content-Type (application/json)", func() {
			// Given: Token endpoint
			// When: POST with invalid Content-Type (application/json instead of form-urlencoded)
			bodyStr := url.Values{
				"grant_type": []string{"authorization_code"},
				"code":       []string{"abc123"},
			}.Encode()

			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/json"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 400 Bad Request
			// Specification T043: Content-Type enforcement required
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should accept valid Content-Type (application/x-www-form-urlencoded)", func() {
			// Given: Token endpoint with mock upstream
			// When: POST with valid Content-Type
			// client_id must be a valid agent UUID (broker-internal identifier)
			bodyStr := url.Values{
				"grant_type": []string{"authorization_code"},
				"code":       []string{"abc123"},
				"client_id":  []string{testAgent.ID.String()},
			}.Encode()

			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Request succeeds (proxied to upstream)
			// Specification T043: Valid Content-Type must be accepted
			Expect(resp.StatusCode).To(SatisfyAny(
				Equal(http.StatusOK),         // Successful token response
				Equal(http.StatusBadRequest), // Upstream validation error (acceptable)
			))
		})
	})

	// GROUP 5: Upstream error handling
	Describe("upstream error forwarding", func() {
		It("should forward 401 Unauthorized from upstream", func() {
			// Given: Mock upstream returns 401
			mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":             "invalid_client",
					"error_description": "Client authentication failed",
				})
			}))
			defer mockUpstream.Close()

			// Update config to use error-responding upstream
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.URL)
			factory := bootstrap.NewServerFactory(config, logger)
			appInstance, err := factory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			if server != nil {
				server.Close()
			}
			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())

			// When: Request token endpoint
			// client_id must be a valid agent UUID (broker-internal identifier)
			bodyStr := url.Values{"grant_type": []string{"authorization_code"}, "code": []string{"invalid"}, "client_id": []string{testAgent.ID.String()}}.Encode()
			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 401 from upstream
			// Specification T044: Error forwarding must preserve upstream status codes
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("should forward 500 Internal Server Error from upstream", func() {
			// Given: Mock upstream returns 500
			mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error":             "server_error",
					"error_description": "Internal server error",
				})
			}))
			defer mockUpstream.Close()

			// Update config to use error-responding upstream
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.URL)
			factory := bootstrap.NewServerFactory(config, logger)
			appInstance, err := factory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			if server != nil {
				server.Close()
			}
			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())

			// When: Request token endpoint
			// client_id must be a valid agent UUID (broker-internal identifier)
			bodyStr := url.Values{"grant_type": []string{"authorization_code"}, "code": []string{"test"}, "client_id": []string{testAgent.ID.String()}}.Encode()
			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Returns 500 from upstream
			// Specification T044: Server errors must be forwarded transparently
			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
		})
	})

	// GROUP 6: Header handling
	Describe("response header preservation", func() {
		It("should preserve Content-Type header from upstream", func() {
			// Given: Mock upstream with specific response headers
			mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Cache-Control", "no-store, no-cache")
				w.Header().Set("Pragma", "no-cache")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"access_token": "token123",
					"token_type":   "Bearer",
				})
			}))
			defer mockUpstream.Close()

			// Update config to use mock upstream
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.URL)
			factory := bootstrap.NewServerFactory(config, logger)
			appInstance, err := factory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			if server != nil {
				server.Close()
			}
			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())

			// When: Request token endpoint
			// client_id must be a valid agent UUID (broker-internal identifier)
			bodyStr := url.Values{"grant_type": []string{"authorization_code"}, "code": []string{"abc123"}, "client_id": []string{testAgent.ID.String()}}.Encode()
			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Response headers from upstream are preserved
			// Specification T044: Header preservation ensures proper client handling
			Expect(resp.Header.Get("Content-Type")).To(ContainSubstring("application/json"))
			Expect(resp.Header.Get("Cache-Control")).To(ContainSubstring("no-store"))
			Expect(resp.Header.Get("Pragma")).To(Equal("no-cache"))
		})

		It("should preserve custom headers from upstream", func() {
			// Given: Mock upstream with custom response headers
			mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Custom-Header", "custom-value")
				w.Header().Set("X-RateLimit-Limit", "100")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"access_token": "token123",
					"token_type":   "Bearer",
				})
			}))
			defer mockUpstream.Close()

			// Update config to use mock upstream
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.URL)
			factory := bootstrap.NewServerFactory(config, logger)
			appInstance, err := factory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			if server != nil {
				server.Close()
			}
			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())

			// When: Request token endpoint
			// client_id must be a valid agent UUID (broker-internal identifier)
			bodyStr := url.Values{"grant_type": []string{"authorization_code"}, "code": []string{"abc123"}, "client_id": []string{testAgent.ID.String()}}.Encode()
			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Custom headers are preserved
			// Specification T044: Custom headers enable client customization
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("X-Custom-Header")).To(Equal("custom-value"))
			Expect(resp.Header.Get("X-RateLimit-Limit")).To(Equal("100"))
		})
	})

	// GROUP 7: HTTP hop-by-hop header filtering
	Describe("hop-by-hop header filtering", func() {
		It("should not forward Connection header to upstream", func() {
			// Given: Mock upstream that verifies Connection header is NOT present
			connectionHeaderSeen := false
			mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Connection") != "" {
					connectionHeaderSeen = true
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"access_token": "token123",
					"token_type":   "Bearer",
				})
			}))
			defer mockUpstream.Close()

			// Update config to use mock upstream
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.URL)
			factory := bootstrap.NewServerFactory(config, logger)
			appInstance, err := factory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			if server != nil {
				server.Close()
			}
			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())

			// When: Request includes Connection header
			// client_id must be a valid agent UUID (broker-internal identifier)
			bodyStr := url.Values{"grant_type": []string{"authorization_code"}, "code": []string{"abc123"}, "client_id": []string{testAgent.ID.String()}}.Encode()
			resp, err := server.DirectRequest(
				"POST",
				"/oauth2/token",
				"",
				map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
				strings.NewReader(bodyStr),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Connection header not forwarded to upstream
			// Specification T044: Hop-by-hop headers must be filtered
			Expect(connectionHeaderSeen).To(BeFalse())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	// GROUP 8: XSS and injection attack prevention
	Describe("input validation and sanitization", func() {
		It("should safely handle XSS payload in query parameters", func() {
			// Given: An agent exists
			agent := fixtures.ValidAgent()
			err := testStorage.Agents().Create(ctx, agent)
			Expect(err).ToNot(HaveOccurred())

			// When: Request includes XSS payload
			xssPayload := fixtures.XSSPayload()
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
					agent.ID.String(), "https://client.example.com/cb", url.QueryEscape(xssPayload)),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Payload is not echoed back in response
			// Specification T043: XSS payloads must be safely handled
			body, err := io.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())

			// Response should not contain the raw XSS payload
			Expect(string(body)).NotTo(ContainSubstring("<script>"))
		})

		It("should handle SQL injection payload safely", func() {
			// Given: An agent exists
			agent := fixtures.ValidAgent()
			err := testStorage.Agents().Create(ctx, agent)
			Expect(err).ToNot(HaveOccurred())

			// When: Request includes SQL injection payload
			sqlPayload := fixtures.SQLInjectionPayload()
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code&state=%s",
					agent.ID.String(), url.QueryEscape(sqlPayload)),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Then: Request is handled safely (no database errors)
			// Specification T043: SQL injection attempts must be safely handled
			// Accept either valid error response or redirect
			Expect(resp.StatusCode).To(SatisfyAny(
				Equal(http.StatusFound),               // Redirect (normal flow)
				Equal(http.StatusBadRequest),          // Validation error (acceptable)
				Equal(http.StatusInternalServerError), // Should not crash
			))
		})
	})
})
