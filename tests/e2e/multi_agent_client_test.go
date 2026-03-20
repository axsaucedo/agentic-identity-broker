package e2e_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	storagedomain "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/matchers"
)

var _ = Describe("Multi-Agent Client Delegation", func() {
	// Shared infrastructure for all multi-agent tests.
	// Each test group (US1, US2, US3) builds its own server with appropriate config.
	var (
		logger         *slog.Logger
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		mockUpstream = helpers.NewMockUpstreamOAuth2Server()
		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if mockUpstream != nil {
			mockUpstream.Close()
		}
		if storageFactory != nil && testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	// ──────────────────────────────────────────────────────────────────
	// User Story 1: Multiple Agents Share One Upstream OAuth2 Client
	// ──────────────────────────────────────────────────────────────────
	Describe("US1: Multiple Agents Share One Upstream OAuth2 Client", func() {
		var (
			alpha *storagedomain.Agent
			beta  *storagedomain.Agent
		)

		BeforeEach(func() {
			alpha = fixtures.MultiAgentAlpha()
			beta = fixtures.MultiAgentBeta()
			ctx := context.Background()
			Expect(testStorage.Agents().Create(ctx, alpha)).ToNot(HaveOccurred())
			Expect(testStorage.Agents().Create(ctx, beta)).ToNot(HaveOccurred())
		})

		Context("when multi-agent client sharing is enabled", func() {
			var server *bootstrap.TestServer

			BeforeEach(func() {
				ctx := context.Background()
				// Register GitHub service and create an active session for the default principal.
				// US1 S4 (claim mismatch) requires an active session to reach claim validation.
				githubService := fixtures.GitHubService()
				Expect(testStorage.Services().Create(ctx, githubService)).ToNot(HaveOccurred())
				session := fixtures.GitHubSessionForPrincipal(fixtures.DefaultPrincipal().String())
				Expect(testStorage.UserSessions().Create(ctx, session)).ToNot(HaveOccurred())

				config := fixtures.MultiAgentEnabledConfig(mockUpstream.URL())
				sf := bootstrap.NewServerFactory(config, logger)
				appInstance, err := sf.BuildApp(testStorage)
				Expect(err).ToNot(HaveOccurred())
				server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				if server != nil {
					server.Close()
				}
			})

			// US1 Scenario 1 from specs/021-multi-agent-clientid/spec.md
			It("should append agent ID param to upstream authorize URL", Label("US1"), func() {
				// Given: Feature enabled, agent has active grant
				grant := fixtures.ActiveGrant(
					fixtures.DefaultPrincipal().String(),
					alpha.ID.String(),
					fixtures.GitHubService().ID.String(),
					[]string{"read"},
				)
				Expect(testStorage.UserGrants().Create(context.Background(), grant)).ToNot(HaveOccurred())

				// When: Authorization request using agent internal ID as client_id
				resp, err := server.AuthenticatedGET(
					"/oauth2/authorize?client_id="+alpha.ID.String()+
						"&redirect_uri=https://client.example.com/cb&response_type=code&state=xyz",
					fixtures.DefaultPrincipal().String(),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Upstream redirect URL contains agent ID param (x_agent_id=<agent.id>)
				redirectURL, err := helpers.ExtractRedirectURL(resp)
				Expect(err).ToNot(HaveOccurred())
				Expect(redirectURL.Query().Get("x_agent_id")).To(Equal(alpha.ID.String()))
			})

			// US1 Scenario 2 from specs/021-multi-agent-clientid/spec.md
			It("should resolve client_id against agent.id not agent.client_id", Label("US1"), func() {
				// Given: Agent registered with client_id="shared-upstream" but internal ID is a UUID
				// Active grant for the default principal
				grant := fixtures.ActiveGrant(
					fixtures.DefaultPrincipal().String(),
					alpha.ID.String(),
					fixtures.GitHubService().ID.String(),
					[]string{"read"},
				)
				Expect(testStorage.UserGrants().Create(context.Background(), grant)).ToNot(HaveOccurred())

				// When: Authorization request uses agent.ID (UUID) as client_id
				resp, err := server.AuthenticatedGET(
					"/oauth2/authorize?client_id="+alpha.ID.String()+
						"&redirect_uri=https://client.example.com/cb&response_type=code&state=xyz",
					fixtures.DefaultPrincipal().String(),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Agent is found and request proceeds (proxied to upstream, not 400 invalid_client)
				Expect(resp.StatusCode).ToNot(Equal(http.StatusBadRequest))
				Expect(resp.StatusCode).To(SatisfyAny(
					Equal(http.StatusFound),
					Equal(http.StatusSeeOther),
				))
				location := resp.Header.Get("Location")
				Expect(location).To(ContainSubstring(mockUpstream.URL()))
			})

			// US1 Scenario 3 from specs/021-multi-agent-clientid/spec.md
			It("should proxy token after verifying agent ID claim", Label("US1"), func() {
				// Given: Mock upstream returns JWT token containing x_agent_id=alpha.ID
				mockUpstream.WithSuccessfulTokenResponse().
					ReturnTokenWithClaim("x_agent_id", alpha.ID.String())

				// When: Token request with alpha agent's ID as client_id
				formData := url.Values{
					"grant_type":   []string{"authorization_code"},
					"code":         []string{"mock-auth-code-123"},
					"client_id":    []string{alpha.ID.String()},
					"redirect_uri": []string{"https://client.example.com/cb"},
				}
				resp, err := server.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Token is returned (broker verified the claim and forwarded the token)
				Expect(resp).To(matchers.HaveStatusCode(http.StatusOK))
				Expect(resp).To(matchers.ContainAgentIDClaim("x_agent_id", alpha.ID.String()))
			})

			// US1 Scenario 4 from specs/021-multi-agent-clientid/spec.md
			It("should verify claim value matches initiating agent ID", Label("US1"), func() {
				// Given: Mock upstream returns JWT with x_agent_id=BETA (wrong agent!)
				mockUpstream.WithSuccessfulTokenResponse().
					ReturnTokenWithClaim("x_agent_id", beta.ID.String())

				// When: Token request using ALPHA agent's ID as client_id (alpha initiated)
				formData := url.Values{
					"grant_type":   []string{"authorization_code"},
					"code":         []string{"mock-auth-code-123"},
					"client_id":    []string{alpha.ID.String()},
					"redirect_uri": []string{"https://client.example.com/cb"},
				}
				resp, err := server.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Claim mismatch — broker returns OAuth2 error and withholds token
				Expect(resp).To(matchers.HaveOAuth2Error("server_error"))
			})

			// US1 Scenario 5 from specs/021-multi-agent-clientid/spec.md
			It("should return error and withhold token when agent ID claim absent", Label("US1"), func() {
				// Given: Mock upstream returns a plain non-JWT token (no x_agent_id claim)
				mockUpstream.WithSuccessfulTokenResponse().
					WithAccessToken("plain-opaque-token-without-any-claim")

				// When: Token request — broker expects claim verification
				formData := url.Values{
					"grant_type":   []string{"authorization_code"},
					"code":         []string{"mock-auth-code-123"},
					"client_id":    []string{alpha.ID.String()},
					"redirect_uri": []string{"https://client.example.com/cb"},
				}
				resp, err := server.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Claim absent — broker fails closed and does NOT forward the token
				Expect(resp.StatusCode).ToNot(Equal(http.StatusOK))
			})
		})

		Context("when multi-agent client sharing is disabled", func() {
			var server *bootstrap.TestServer

			BeforeEach(func() {
				config := fixtures.MultiAgentDisabledConfig(mockUpstream.URL())
				sf := bootstrap.NewServerFactory(config, logger)
				appInstance, err := sf.BuildApp(testStorage)
				Expect(err).ToNot(HaveOccurred())
				server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				if server != nil {
					server.Close()
				}
			})

			// US1 Scenario 6 from specs/021-multi-agent-clientid/spec.md
			It("should not inject param or verify claim when feature is disabled", Label("US1"), func() {
				// Given: Feature disabled; alpha has an active grant
				grant := fixtures.ActiveGrant(
					fixtures.DefaultPrincipal().String(),
					alpha.ID.String(),
					fixtures.GitHubService().ID.String(),
					[]string{"read"},
				)
				Expect(testStorage.UserGrants().Create(context.Background(), grant)).ToNot(HaveOccurred())

				// When: Authorization request using agent.ID as client_id (still UUID-based)
				resp, err := server.AuthenticatedGET(
					"/oauth2/authorize?client_id="+alpha.ID.String()+
						"&redirect_uri=https://client.example.com/cb&response_type=code&state=xyz",
					fixtures.DefaultPrincipal().String(),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Proxied to upstream without x_agent_id param (feature is disabled)
				redirectURL, err := helpers.ExtractRedirectURL(resp)
				Expect(err).ToNot(HaveOccurred())
				Expect(redirectURL.Query().Get("x_agent_id")).To(BeEmpty())
				// And: Redirect goes to upstream normally
				Expect(redirectURL.String()).To(ContainSubstring(mockUpstream.URL()))
			})
		})
	})

	// ──────────────────────────────────────────────────────────────────
	// User Story 2: Token Exchange Resolves Agent from Token Claim
	// ──────────────────────────────────────────────────────────────────
	Describe("US2: Token Exchange Resolves Agent from Token Claim", func() {
		var (
			alpha         *storagedomain.Agent
			enduserServer *bootstrap.TestServer
		)

		AfterEach(func() {
			if enduserServer != nil {
				enduserServer.Close()
			}
		})

		Context("when multi-agent client sharing is enabled", func() {
			var clientAssertion string

			BeforeEach(func() {
				alpha = fixtures.MultiAgentAlpha()
				ctx := context.Background()
				Expect(testStorage.Agents().Create(ctx, alpha)).ToNot(HaveOccurred())

				// Register GitHub as the third-party service for resource-based lookup.
				// Token endpoint is pointed at the mock upstream so exchange requests reach it.
				githubService := fixtures.GitHubService()
				githubService.Endpoints.TokenEndpoint = mockUpstream.URL() + "/oauth/token"
				Expect(testStorage.Services().Create(ctx, githubService)).ToNot(HaveOccurred())

				// Grant alpha agent access to GitHub service (required by token exchange flow).
				grant := fixtures.ActiveGrant(
					fixtures.DefaultPrincipal().String(),
					alpha.ID.String(),
					githubService.ID.String(),
					[]string{"repo", "user"},
				)
				Expect(testStorage.UserGrants().Create(ctx, grant)).ToNot(HaveOccurred())

				// Create active session for default principal — token exchange requires an active session.
				session := fixtures.GitHubSessionForPrincipal(fixtures.DefaultPrincipal().String())
				Expect(testStorage.UserSessions().Create(ctx, session)).ToNot(HaveOccurred())

				// Generate client assertion for alpha agent — authenticates the requesting agent.
				// sub = alpha.ID (broker UUID) since in multi-agent mode each agent has its own UUID.
				now := time.Now()
				assertionClaims := map[string]interface{}{
					"sub": alpha.ID.String(),
					"iss": mockUpstream.URL(),
					"aud": "token-exchange-broker",
					"exp": now.Add(1 * time.Hour).Unix(),
					"iat": now.Unix(),
				}
				var err error
				clientAssertion, err = helpers.SignTestJWT(assertionClaims, mockUpstream.GetPrivateKeyPEM())
				Expect(err).ToNot(HaveOccurred())

				config := fixtures.MultiAgentEnabledConfig(mockUpstream.URL())
				sf := bootstrap.NewServerFactory(config, logger)
				appInstance, err := sf.BuildApp(testStorage)
				Expect(err).ToNot(HaveOccurred())
				enduserServer, err = bootstrap.NewEndUserTestServer(appInstance, logger)
				Expect(err).ToNot(HaveOccurred())
			})

			// US2 Scenario 1 from specs/021-multi-agent-clientid/spec.md
			It("should resolve agent from token claim when feature enabled", Label("US2"), func() {
				// Given: Subject token contains x_agent_id=alpha.ID, CEL reads it directly.
				// Full RFC 8693 claims required: iss must match upstream for issuer validation.
				now := time.Now()
				subjectTokenClaims := map[string]interface{}{
					"sub":        fixtures.DefaultPrincipal().String(),
					"x_agent_id": alpha.ID.String(),
					"azp":        string(alpha.ClientID),
					"iss":        mockUpstream.URL(),
					"aud":        "token-exchange-broker",
					"exp":        now.Add(1 * time.Hour).Unix(),
					"iat":        now.Unix(),
				}
				subjectToken, err := helpers.SignTestJWT(subjectTokenClaims, mockUpstream.GetPrivateKeyPEM())
				Expect(err).ToNot(HaveOccurred())

				// When: Token exchange with subject token containing agent ID claim
				formData := url.Values{
					"grant_type":            []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
					"subject_token":         []string{subjectToken},
					"subject_token_type":    []string{"urn:ietf:params:oauth:token-type:access_token"},
					"requested_token_type":  []string{"urn:ietf:params:oauth:token-type:access_token"},
					"resource":              []string{"https://api.github.com"},
					"client_assertion_type": []string{"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
					"client_assertion":      []string{clientAssertion},
				}
				mockUpstream.WithSuccessfulTokenResponse()
				resp, err := enduserServer.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Token exchange succeeds — agent resolved from claim
				Expect(resp).To(matchers.HaveStatusCode(http.StatusOK))
			})

			// US2 Scenario 2 from specs/021-multi-agent-clientid/spec.md
			It("should fail token exchange when agent ID claim absent from subject token", Label("US2"), func() {
				// Given: Subject token WITHOUT x_agent_id claim; CEL policy expects it.
				// Full RFC 8693 claims required to pass issuer validation before claim check.
				now := time.Now()
				subjectTokenClaims := map[string]interface{}{
					"sub": fixtures.DefaultPrincipal().String(),
					"azp": string(alpha.ClientID),
					"iss": mockUpstream.URL(),
					"aud": "token-exchange-broker",
					"exp": now.Add(1 * time.Hour).Unix(),
					"iat": now.Unix(),
				}
				subjectToken, err := helpers.SignTestJWT(subjectTokenClaims, mockUpstream.GetPrivateKeyPEM())
				Expect(err).ToNot(HaveOccurred())

				// When: Token exchange with subject token missing the agent ID claim
				formData := url.Values{
					"grant_type":            []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
					"subject_token":         []string{subjectToken},
					"subject_token_type":    []string{"urn:ietf:params:oauth:token-type:access_token"},
					"requested_token_type":  []string{"urn:ietf:params:oauth:token-type:access_token"},
					"resource":              []string{"https://api.github.com"},
					"client_assertion_type": []string{"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
					"client_assertion":      []string{clientAssertion},
				}
				resp, err := enduserServer.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: CEL expression fails (claim absent) — token exchange returns error
				Expect(resp.StatusCode).ToNot(Equal(http.StatusOK))
			})

			// US2 Scenario 4 from specs/021-multi-agent-clientid/spec.md
			It("should reject token exchange when agent ID claim does not match any registered agent", Label("US2"), func() {
				// Given: Subject token has x_agent_id claim with UUID not matching any registered agent.
				// Full RFC 8693 claims required to pass issuer validation before agent lookup.
				unknownAgentID := "00000000-0000-0000-0000-000000000099"
				now := time.Now()
				subjectTokenClaims := map[string]interface{}{
					"sub":        fixtures.DefaultPrincipal().String(),
					"x_agent_id": unknownAgentID,
					"iss":        mockUpstream.URL(),
					"aud":        "token-exchange-broker",
					"exp":        now.Add(1 * time.Hour).Unix(),
					"iat":        now.Unix(),
				}
				subjectToken, err := helpers.SignTestJWT(subjectTokenClaims, mockUpstream.GetPrivateKeyPEM())
				Expect(err).ToNot(HaveOccurred())

				// When: Token exchange with subject token referencing an unregistered agent
				formData := url.Values{
					"grant_type":            []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
					"subject_token":         []string{subjectToken},
					"subject_token_type":    []string{"urn:ietf:params:oauth:token-type:access_token"},
					"requested_token_type":  []string{"urn:ietf:params:oauth:token-type:access_token"},
					"resource":              []string{"https://api.github.com"},
					"client_assertion_type": []string{"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
					"client_assertion":      []string{clientAssertion},
				}
				resp, err := enduserServer.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: Unknown agent ID — exchange is rejected
				Expect(resp.StatusCode).ToNot(Equal(http.StatusOK))
			})
		})

		Context("when multi-agent client sharing is disabled", func() {
			var clientAssertionDisabled string

			BeforeEach(func() {
				alpha = fixtures.MultiAgentAlpha()
				ctx := context.Background()
				Expect(testStorage.Agents().Create(ctx, alpha)).ToNot(HaveOccurred())

				// Register GitHub as the third-party service for resource-based lookup.
				githubService := fixtures.GitHubService()
				githubService.Endpoints.TokenEndpoint = mockUpstream.URL() + "/oauth/token"
				Expect(testStorage.Services().Create(ctx, githubService)).ToNot(HaveOccurred())

				// Grant alpha agent access to GitHub service.
				grant := fixtures.ActiveGrant(
					fixtures.DefaultPrincipal().String(),
					alpha.ID.String(),
					githubService.ID.String(),
					[]string{"repo", "user"},
				)
				Expect(testStorage.UserGrants().Create(ctx, grant)).ToNot(HaveOccurred())

				// Create active session for default principal — token exchange requires an active session.
				session := fixtures.GitHubSessionForPrincipal(fixtures.DefaultPrincipal().String())
				Expect(testStorage.UserSessions().Create(ctx, session)).ToNot(HaveOccurred())

				// Generate client assertion for alpha agent.
				now := time.Now()
				assertionClaims := map[string]interface{}{
					"sub": alpha.ID.String(),
					"iss": mockUpstream.URL(),
					"aud": "token-exchange-broker",
					"exp": now.Add(1 * time.Hour).Unix(),
					"iat": now.Unix(),
				}
				var err error
				clientAssertionDisabled, err = helpers.SignTestJWT(assertionClaims, mockUpstream.GetPrivateKeyPEM())
				Expect(err).ToNot(HaveOccurred())

				config := fixtures.MultiAgentDisabledConfig(mockUpstream.URL())
				sf := bootstrap.NewServerFactory(config, logger)
				appInstance, err := sf.BuildApp(testStorage)
				Expect(err).ToNot(HaveOccurred())
				enduserServer, err = bootstrap.NewEndUserTestServer(appInstance, logger)
				Expect(err).ToNot(HaveOccurred())
			})

			// US2 Scenario 3 from specs/021-multi-agent-clientid/spec.md
			It("should resolve agent via resolveAgentIdByClientId CEL function when disabled", Label("US2"), func() {
				// Given: Feature disabled; CEL policy uses resolveAgentIdByClientId(subject_token.azp)
				// Subject token has azp=alpha.ClientID plus full RFC 8693 claims (iss, aud, exp, iat).
				now := time.Now()
				claims := map[string]interface{}{
					"sub": fixtures.DefaultPrincipal().String(),
					"azp": string(alpha.ClientID),
					"iss": mockUpstream.URL(),
					"aud": "token-exchange-broker",
					"exp": now.Add(1 * time.Hour).Unix(),
					"iat": now.Unix(),
				}
				subjectToken, err := helpers.SignTestJWT(claims, mockUpstream.GetPrivateKeyPEM())
				Expect(err).ToNot(HaveOccurred())

				// When: Token exchange where CEL calls resolveAgentIdByClientId(azp)
				formData := url.Values{
					"grant_type":            []string{"urn:ietf:params:oauth:grant-type:token-exchange"},
					"subject_token":         []string{subjectToken},
					"subject_token_type":    []string{"urn:ietf:params:oauth:token-type:access_token"},
					"requested_token_type":  []string{"urn:ietf:params:oauth:token-type:access_token"},
					"resource":              []string{"https://api.github.com"},
					"client_assertion_type": []string{"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
					"client_assertion":      []string{clientAssertionDisabled},
				}
				mockUpstream.WithSuccessfulTokenResponse()
				resp, err := enduserServer.DirectRequest(
					"POST",
					"/oauth2/token",
					"",
					map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
					strings.NewReader(formData.Encode()),
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() { _ = resp.Body.Close() }()

				// Then: resolveAgentIdByClientId maps upstream client_id to agent.id → exchange succeeds
				Expect(resp).To(matchers.HaveStatusCode(http.StatusOK))
			})
		})
	})

	// ──────────────────────────────────────────────────────────────────
	// User Story 3: Operator Configures Multi-Agent Client Sharing
	// ──────────────────────────────────────────────────────────────────
	// US3 tests validate server startup behavior — they create their own
	// server instances inside each It() block to test config validation.
	Describe("US3: Operator Configures Multi-Agent Client Sharing", func() {
		// US3 Scenario 1 from specs/021-multi-agent-clientid/spec.md
		It("should start successfully with valid multi_agent_client configuration", Label("US3"), func() {
			// Given: Config with feature enabled and both required fields present
			config := fixtures.MultiAgentEnabledConfig(mockUpstream.URL())

			// When: Building the application
			sf := bootstrap.NewServerFactory(config, logger)
			appInstance, err := sf.BuildApp(testStorage)

			// Then: Startup succeeds — no validation error
			Expect(err).ToNot(HaveOccurred())

			// Cleanup
			srv, err := bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()
		})

		// US3 Scenario 2 from specs/021-multi-agent-clientid/spec.md
		It("should fail to start when enabled but agent_id_param_name is absent", Label("US3"), func() {
			// Given: Config with feature enabled but agent_id_param_name missing
			config := fixtures.MultiAgentEnabledConfig(mockUpstream.URL())
			config.OAuth2AuthServer.MultiAgentClient.AgentIDParamName = ""

			// When: Building the application
			sf := bootstrap.NewServerFactory(config, logger)
			_, err := sf.BuildApp(testStorage)

			// Then: Startup fails with a clear error about the missing param name
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("agent_id_param_name"))
		})

		// US3 Scenario 3 from specs/021-multi-agent-clientid/spec.md
		It("should fail to start when enabled but agent_id_claim_name is absent", Label("US3"), func() {
			// Given: Config with feature enabled but agent_id_claim_name missing
			config := fixtures.MultiAgentEnabledConfig(mockUpstream.URL())
			config.OAuth2AuthServer.MultiAgentClient.AgentIDClaimName = ""

			// When: Building the application
			sf := bootstrap.NewServerFactory(config, logger)
			_, err := sf.BuildApp(testStorage)

			// Then: Startup fails with a clear error about the missing claim name
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("agent_id_claim_name"))
		})

		// US3 Scenario 4 from specs/021-multi-agent-clientid/spec.md
		It("should operate in single-agent mode when multi_agent_client is disabled", Label("US3"), func() {
			// Given: Config with feature explicitly disabled
			config := fixtures.MultiAgentDisabledConfig(mockUpstream.URL())

			// When: Building and starting the application
			sf := bootstrap.NewServerFactory(config, logger)
			appInstance, err := sf.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())

			srv, err := bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
			defer srv.Close()

			// Then: Broker starts normally and serves requests (single-agent mode, no behavioral change)
			agent := fixtures.ValidAgent()
			Expect(testStorage.Agents().Create(context.Background(), agent)).ToNot(HaveOccurred())

			grant := fixtures.ActiveGrant(
				fixtures.DefaultPrincipal().String(),
				agent.ID.String(),
				fixtures.GitHubService().ID.String(),
				[]string{"read"},
			)
			Expect(testStorage.UserGrants().Create(context.Background(), grant)).ToNot(HaveOccurred())

			// Authorize with agent.ID (UUID) — should proxy to upstream normally
			resp, err := srv.AuthenticatedGET(
				"/oauth2/authorize?client_id="+agent.ID.String()+
					"&redirect_uri=https://client.example.com/cb&response_type=code&state=xyz",
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Single-agent mode: proxies normally, no x_agent_id param injected
			Expect(resp.StatusCode).To(SatisfyAny(
				Equal(http.StatusFound),
				Equal(http.StatusSeeOther),
			))
			redirectURL, err := helpers.ExtractRedirectURL(resp)
			Expect(err).ToNot(HaveOccurred())
			Expect(redirectURL.Query().Get("x_agent_id")).To(BeEmpty())
		})
	})
})
