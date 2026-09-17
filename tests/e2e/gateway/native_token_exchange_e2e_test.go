package gateway_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v4/jwt"
	"github.com/mark3labs/mcp-go/mcp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	storagedomain "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/gateway/support"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

const (
	directGatewayClientID = "https://agentgateway-direct-e2e.example.test"
	brokerAudience        = "token-exchange-broker"
	protectedResource     = "https://api.github.com"
	exchangedCredential   = "github-token-xyz"
)

type gatewayCall struct {
	response string
	status   int
	err      error
}

type nativeGatewayEnvironmentOptions struct {
	directReferenceYAML            []byte
	resource                       string
	omitResource                   bool
	clientID                       string
	assertionAudience              string
	assertionIssuer                string
	subjectIssuer                  string
	subjectAudience                string
	clientAuthAlgorithm            string
	untrustedGatewaySigningKey     bool
	withoutGrant                   bool
	withoutSession                 bool
	requireGitHubScope             bool
	sessionScopes                  []string
	unavailableClientAssertionJWKS bool
	unavailableBroker              bool
	skipRouteLiveness              bool
	expectGatewayStartupFailure    bool
}

type nativeGatewayEnvironment struct {
	ctx                     context.Context
	cancel                  context.CancelFunc
	storageFactory          *bootstrap.StorageFactory
	storage                 *storageadapter.Adapter
	broker                  *bootstrap.TestServer
	upstream                *helpers.MockUpstreamOAuth2Server
	tokenEndpoint           *support.TokenEndpointRecorder
	backend                 *support.DownstreamBackend
	gateway                 *support.Agentgateway
	gatewayStartupErr       error
	gatewaySigningKey       *support.GatewaySigningKey
	clientAssertionTrustKey *support.GatewaySigningKey
	clientAssertionJWKS     *support.HTTPSJWKSServer
	restoreDefaultTransport func()
	principal               string
	agent                   *storagedomain.Agent
}

type rejectedExchangeCase struct {
	name             string
	brokerError      string
	expectedCode     int
	options          nativeGatewayEnvironmentOptions
	configFailure    bool
	assertNoResource bool
}

func startNativeGatewayEnvironment(options nativeGatewayEnvironmentOptions) *nativeGatewayEnvironment {
	if options.clientID == "" {
		options.clientID = directGatewayClientID
	}
	if options.assertionAudience == "" {
		options.assertionAudience = brokerAudience
	}
	if options.assertionIssuer == "" {
		options.assertionIssuer = directGatewayClientID
	}
	if !options.omitResource && options.resource == "" {
		options.resource = protectedResource
	}

	environment := &nativeGatewayEnvironment{}
	environment.ctx, environment.cancel = context.WithTimeout(context.Background(), 120*time.Second)
	DeferCleanup(environment.cancel)

	logger := bootstrap.TestLogger(slog.LevelWarn)
	environment.upstream = helpers.NewMockUpstreamOAuth2Server()
	DeferCleanup(environment.upstream.Close)

	var err error
	environment.gatewaySigningKey, err = support.NewGatewaySigningKey()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		Expect(environment.gatewaySigningKey.Close()).To(Succeed())
	})

	environment.clientAssertionTrustKey = environment.gatewaySigningKey
	if options.untrustedGatewaySigningKey {
		environment.clientAssertionTrustKey, err = support.NewGatewaySigningKey()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(environment.clientAssertionTrustKey.Close()).To(Succeed())
		})
	}

	environment.clientAssertionJWKS, err = support.NewHTTPSJWKSServer(environment.clientAssertionTrustKey.PublicJWKSet())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(environment.clientAssertionJWKS.Close)
	environment.restoreDefaultTransport = environment.clientAssertionJWKS.InstallDefaultTransport()
	DeferCleanup(environment.restoreDefaultTransport)

	environment.storageFactory = bootstrap.NewStorageFactory(logger)
	environment.storage, err = environment.storageFactory.NewTestStorage()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		Expect(environment.storageFactory.CloseStorage(environment.storage)).To(Succeed())
	})

	environment.principal = fixtures.DefaultPrincipal().String()
	environment.agent = fixtures.ValidAgent()
	Expect(environment.storage.Agents().Create(environment.ctx, environment.agent)).To(Succeed())

	githubService := fixtures.GitHubService()
	githubService.Endpoints.TokenEndpoint = environment.upstream.URL() + "/oauth/token"
	githubService.Endpoints.AuthorizeEndpoint = environment.upstream.URL() + "/oauth/authorize"
	Expect(environment.storage.Services().Create(environment.ctx, githubService)).To(Succeed())
	Expect(fixtures.SeedPlaceholderGrantData(environment.ctx, environment.storage, githubService.ID)).To(Succeed())
	if options.requireGitHubScope {
		permissionSet, getErr := environment.storage.PermissionSets().Get(environment.ctx, fixtures.PlaceholderPermissionSetID)
		Expect(getErr).NotTo(HaveOccurred())
		githubScopeConfigured := false
		for index := range permissionSet.ServiceScopes {
			if permissionSet.ServiceScopes[index].ServiceID == githubService.ID {
				permissionSet.ServiceScopes[index].Scopes = []string{"repo"}
				githubScopeConfigured = true
			}
		}
		Expect(githubScopeConfigured).To(BeTrue(), "the fixture permission set must require the GitHub repo scope")
		Expect(environment.storage.PermissionSets().Update(environment.ctx, permissionSet)).To(Succeed())
	}
	if !options.withoutGrant {
		grant := fixtures.ActiveGrant(environment.principal, environment.agent.ID.String(), githubService.ID.String(), []string{"repo", "user"})
		Expect(environment.storage.UserGrants().Create(environment.ctx, grant)).To(Succeed())
	}
	if !options.withoutSession {
		session := fixtures.GitHubSessionForPrincipal(environment.principal)
		if options.sessionScopes != nil {
			session.Scope = options.sessionScopes
		}
		Expect(environment.storage.UserSessions().Create(environment.ctx, session)).To(Succeed())
	}

	config := fixtures.OAuth2ConfigWithTokenExchange(environment.upstream.URL())
	config.Security.SkipThirdpartyHTTPSValidation = true
	config.TokenExchange.ExpectedAudience = brokerAudience
	config.TokenExchange.ClientAssertion = ports.ClientAssertionTrustConfig{
		IssuerURI: options.assertionIssuer,
		JWKSURI:   environment.clientAssertionJWKS.URL(),
	}
	config.TokenExchange.Authorization.CEL.Expression = `client_assertion.iss == "` + options.assertionIssuer + `"`
	if options.unavailableClientAssertionJWKS {
		environment.clientAssertionJWKS.Close()
	}
	serverFactory := bootstrap.NewServerFactory(config, logger)
	serverBuilder, err := bootstrap.NewTestServerBuilder(config, environment.storage, serverFactory, logger)
	Expect(err).NotTo(HaveOccurred())
	environment.tokenEndpoint = &support.TokenEndpointRecorder{}
	environment.broker, err = serverBuilder.WithContainerReachability(environment.tokenEndpoint.Wrap).Build()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(environment.broker.Close)

	environment.backend, err = support.StartDownstreamBackend()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() {
		Expect(environment.backend.Close()).To(Succeed())
	})

	environment.gateway, environment.gatewayStartupErr = support.StartAgentgateway(environment.ctx, support.AgentgatewayOptions{
		DirectReferenceYAML: options.directReferenceYAML,
		BrokerURL:           environment.broker.ContainerURL(),
		BackendURL:          environment.backend.URL(),
		SigningKeyPath:      environment.gatewaySigningKey.PrivateKeyPath(),
		KeyID:               environment.gatewaySigningKey.KeyID(),
		Resource:            options.resource,
		ClientID:            options.clientID,
		AssertionAudience:   options.assertionAudience,
		ClientAuthAlgorithm: options.clientAuthAlgorithm,
		OmitResource:        options.omitResource,
		SkipRouteLiveness:   options.skipRouteLiveness,
		DisableCache:        true,
	})
	if options.expectGatewayStartupFailure {
		Expect(environment.gatewayStartupErr).To(HaveOccurred())
		Expect(environment.gateway).To(BeNil())
		return environment
	}
	Expect(environment.gatewayStartupErr).NotTo(HaveOccurred())
	DeferCleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cleanupCancel()
		Expect(environment.gateway.Close(cleanupCtx)).To(Succeed())
	})

	if options.unavailableBroker {
		environment.broker.Close()
	}
	return environment
}

func (environment *nativeGatewayEnvironment) mintSubjectToken(issuer, audience string) string {
	if issuer == "" {
		issuer = environment.upstream.URL()
	}
	if audience == "" {
		audience = brokerAudience
	}
	now := time.Now()
	token, err := helpers.SignTestJWT(map[string]interface{}{
		"sub": environment.principal,
		"azp": environment.agent.ID.String(),
		"iss": issuer,
		"aud": audience,
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
	}, environment.upstream.GetPrivateKeyPEM())
	Expect(err).NotTo(HaveOccurred())
	return token
}

func (environment *nativeGatewayEnvironment) closeNow() error {
	if environment == nil {
		return nil
	}

	var closeErrs []error
	if environment.gateway != nil {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
		closeErrs = append(closeErrs, environment.gateway.Close(cleanupCtx))
		cleanupCancel()
	}
	if environment.broker != nil {
		environment.broker.Close()
	}
	if environment.backend != nil {
		closeErrs = append(closeErrs, environment.backend.Close())
	}
	if environment.clientAssertionJWKS != nil {
		environment.clientAssertionJWKS.Close()
	}
	if environment.upstream != nil {
		environment.upstream.Close()
	}
	if environment.restoreDefaultTransport != nil {
		environment.restoreDefaultTransport()
	}
	if environment.gatewaySigningKey != nil {
		closeErrs = append(closeErrs, environment.gatewaySigningKey.Close())
	}
	if environment.clientAssertionTrustKey != nil && environment.clientAssertionTrustKey != environment.gatewaySigningKey {
		closeErrs = append(closeErrs, environment.clientAssertionTrustKey.Close())
	}
	if environment.storageFactory != nil && environment.storage != nil {
		closeErrs = append(closeErrs, environment.storageFactory.CloseStorage(environment.storage))
	}
	if environment.cancel != nil {
		environment.cancel()
	}
	return errors.Join(closeErrs...)
}

func callGatewayWhoAmI(ctx context.Context, gateway *support.Agentgateway, subjectToken string) gatewayCall {
	response, err := gateway.CallWhoAmI(ctx, subjectToken)
	return gatewayCall{
		response: response,
		status:   http.StatusOK,
		err:      err,
	}
}

func initializeGatewayMCP(ctx context.Context, gateway *support.Agentgateway, subjectToken string) gatewayCall {
	requestBody := fmt.Sprintf(
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":%q,"clientInfo":{"name":"gateway-direct-e2e","version":"1.0"},"capabilities":{}}}`,
		mcp.LATEST_PROTOCOL_VERSION,
	)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, gateway.BaseURL+"/mcp", strings.NewReader(requestBody))
	if err != nil {
		return gatewayCall{err: fmt.Errorf("create MCP initialization request through agentgateway: %w", err)}
	}
	request.Header.Set("Accept", "application/json, text/event-stream")
	request.Header.Set("Authorization", "Bearer "+subjectToken)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Mcp-Protocol-Version", mcp.LATEST_PROTOCOL_VERSION)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return gatewayCall{err: fmt.Errorf("send MCP initialization request through agentgateway: %w", err)}
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return gatewayCall{status: response.StatusCode}
	}
	return gatewayCall{
		status: response.StatusCode,
		err:    fmt.Errorf("agentgateway MCP initialization returned HTTP %s", response.Status),
	}
}

func assertNoInboundCredentialReachedBackend(environment *nativeGatewayEnvironment, inboundCredential string) {
	Expect(environment.backend.RequestCount()).To(Equal(0), "a failed exchange must stop before the protected backend")
	Expect(environment.backend.AuthorizationHeaders()).ToNot(
		ContainElement("Bearer "+inboundCredential),
		"a failed exchange must not fall back to the inbound credential",
	)
	Expect(environment.gateway.ExtProc.ConnectionCount()).To(Equal(0), "a direct route must not contact the ExtProc stand-in")
}

func lastBackendAuthorization(backend *support.DownstreamBackend) string {
	headers := backend.AuthorizationHeaders()
	Expect(headers).NotTo(BeEmpty())
	return headers[len(headers)-1]
}

func assertDocumentedNativePolicy(renderedConfig []byte) {
	config := string(renderedConfig)
	Expect(config).To(ContainSubstring("oauthTokenExchange"))
	Expect(config).To(ContainSubstring("path: /oauth2/token"))
	Expect(config).To(ContainSubstring("resources:"))
	Expect(config).To(ContainSubstring(protectedResource))
	Expect(config).To(ContainSubstring("method: privateKeyJwt"))
	Expect(config).To(ContainSubstring("clientId: " + directGatewayClientID))
	Expect(config).To(ContainSubstring("assertionAudience: " + brokerAudience))
	Expect(config).To(ContainSubstring("signingKey:"))
	Expect(config).To(ContainSubstring("file:"))
	Expect(config).To(ContainSubstring("maxEntries: 0"))
	Expect(config).ToNot(ContainSubstring("extProc"))
}

func assertDocumentedTokenExchangeForm(form url.Values, expectedSubjectToken string) string {
	Expect(form.Get("grant_type")).To(Equal("urn:ietf:params:oauth:grant-type:token-exchange"))
	Expect(form.Get("subject_token")).To(Equal(expectedSubjectToken))
	Expect(form.Get("subject_token_type")).To(Equal("urn:ietf:params:oauth:token-type:access_token"))
	Expect(form.Get("resource")).To(Equal(protectedResource))
	Expect(form.Get("client_id")).To(Equal(directGatewayClientID))
	Expect(form.Get("client_assertion_type")).To(Equal("urn:ietf:params:oauth:client-assertion-type:jwt-bearer"))

	clientAssertion := form.Get("client_assertion")
	Expect(clientAssertion).ToNot(BeEmpty())
	assertion, err := jwt.ParseInsecure([]byte(clientAssertion))
	Expect(err).ToNot(HaveOccurred())

	issuer, ok := assertion.Issuer()
	Expect(ok).To(BeTrue())
	Expect(issuer).To(Equal(directGatewayClientID))
	subject, ok := assertion.Subject()
	Expect(ok).To(BeTrue())
	Expect(subject).To(Equal(directGatewayClientID))
	audience, ok := assertion.Audience()
	Expect(ok).To(BeTrue())
	Expect(audience).To(ConsistOf(brokerAudience))
	jti, ok := assertion.JwtID()
	Expect(ok).To(BeTrue())
	Expect(jti).ToNot(BeEmpty())

	return jti
}

var _ = Describe("Agentgateway Native Token Exchange", func() {
	Context("Exchange Tokens Through the Native Gateway Policy", func() {
		var (
			environment       *nativeGatewayEnvironment
			inboundCredential string
		)

		BeforeEach(func() {
			environment = startNativeGatewayEnvironment(nativeGatewayEnvironmentOptions{})
			inboundCredential = environment.mintSubjectToken("", "")
		})

		// US1-S1 from specs/045-agentgateway-token-exchange/spec.md
		It("should send distinct documented assertions for two exchanges", func() {
			mcpClient, err := environment.gateway.OpenMCPClient(environment.ctx, inboundCredential)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(mcpClient.Close()).To(Succeed()) })
			environment.tokenEndpoint.Reset()
			environment.backend.Reset()

			firstResponse, firstErr := mcpClient.CallWhoAmI(environment.ctx)
			secondResponse, secondErr := mcpClient.CallWhoAmI(environment.ctx)
			Expect(firstErr).NotTo(HaveOccurred())
			Expect(secondErr).NotTo(HaveOccurred())

			forms := environment.tokenEndpoint.Forms()
			Expect(forms).To(HaveLen(2))
			firstJTI := assertDocumentedTokenExchangeForm(forms[0], inboundCredential)
			secondJTI := assertDocumentedTokenExchangeForm(forms[1], inboundCredential)
			Expect(secondJTI).ToNot(Equal(firstJTI))
			Expect(environment.clientAssertionJWKS.RequestCount()).To(BeNumerically(">", 0), "the production Broker must validate the gateway assertion through its HTTPS JWKS")

			Expect(firstResponse).To(Equal("Bearer " + exchangedCredential))
			Expect(secondResponse).To(Equal("Bearer " + exchangedCredential))
			Expect(environment.backend.AuthorizationHeaders()).To(ContainElement("Bearer " + exchangedCredential))
			Expect(lastBackendAuthorization(environment.backend)).To(Equal("Bearer " + exchangedCredential))
			assertDocumentedNativePolicy(environment.gateway.RenderedConfig)
		})

		// US1-S2 from specs/045-agentgateway-token-exchange/spec.md
		It("should replace the inbound credential before reaching the backend", func() {
			result := callGatewayWhoAmI(environment.ctx, environment.gateway, inboundCredential)

			Expect(result.err).ToNot(HaveOccurred())
			Expect(result.status).To(Equal(http.StatusOK))
			Expect(result.response).To(Equal("Bearer " + exchangedCredential))
			Expect(lastBackendAuthorization(environment.backend)).To(Equal("Bearer " + exchangedCredential))
			Expect(environment.backend.AuthorizationHeaders()).ToNot(ContainElement("Bearer " + inboundCredential))
		})

		// US1-S3 from specs/045-agentgateway-token-exchange/spec.md
		It("should exchange directly without an ExtProc policy or contact", func() {
			result := callGatewayWhoAmI(environment.ctx, environment.gateway, inboundCredential)

			Expect(result.err).ToNot(HaveOccurred())
			Expect(result.status).To(Equal(http.StatusOK))
			Expect(result.response).To(Equal("Bearer " + exchangedCredential))
			Expect(lastBackendAuthorization(environment.backend)).To(Equal("Bearer " + exchangedCredential))
			Expect(environment.gateway.ExtProc.ConnectionCount()).To(Equal(0))
			assertDocumentedNativePolicy(environment.gateway.RenderedConfig)
		})
	})

	Context("Preserve Broker Authorization Boundaries", func() {
		// US2-S1 from specs/045-agentgateway-token-exchange/spec.md
		It("should reject invalid credentials before resource and delegation checks", func() {
			cases := []rejectedExchangeCase{
				{
					name:         "untrusted signing key",
					brokerError:  "invalid_client",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						untrustedGatewaySigningKey: true,
					},
				},
				{
					name:         "mismatched assertion issuer",
					brokerError:  "invalid_client",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						clientID: "https://untrusted-agentgateway.example.test",
					},
				},
				{
					name:         "mismatched assertion audience",
					brokerError:  "invalid_client",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						assertionAudience: "unexpected-broker-audience",
					},
				},
				{
					name:         "mismatched subject-token issuer",
					brokerError:  "invalid_grant",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						subjectIssuer: "https://untrusted-upstream.example.test",
					},
				},
				{
					name:         "mismatched subject-token audience",
					brokerError:  "invalid_grant",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						subjectAudience: "unexpected-broker-audience",
					},
				},
				{
					name:          "unsupported client assertion algorithm",
					configFailure: true,
					options: nativeGatewayEnvironmentOptions{
						clientAuthAlgorithm:         "none",
						expectGatewayStartupFailure: true,
					},
				},
			}

			for _, testCase := range cases {
				environment := startNativeGatewayEnvironment(testCase.options)
				if testCase.configFailure {
					Expect(environment.gatewayStartupErr).To(HaveOccurred(), "%s must reject the gateway configuration at load", testCase.name)
					Expect(environment.gateway).To(BeNil(), "%s must leave no serving route", testCase.name)
					Expect(environment.backend.RequestCount()).To(Equal(0))
					Expect(environment.closeNow()).To(Succeed())
					continue
				}

				inboundCredential := environment.mintSubjectToken(testCase.options.subjectIssuer, testCase.options.subjectAudience)
				result := initializeGatewayMCP(environment.ctx, environment.gateway, inboundCredential)
				Expect(result.err).To(HaveOccurred(), "%s must produce Broker %s", testCase.name, testCase.brokerError)
				Expect(result.status).To(Equal(testCase.expectedCode), "%s must map to the documented agent response", testCase.name)
				Expect(result.response).To(BeEmpty(), "%s must not complete the MCP request", testCase.name)
				assertNoInboundCredentialReachedBackend(environment, inboundCredential)
				Expect(environment.closeNow()).To(Succeed())
			}
		})

		// US2-S2 from specs/045-agentgateway-token-exchange/spec.md
		It("should deny an exchange without active delegation", func() {
			environment := startNativeGatewayEnvironment(nativeGatewayEnvironmentOptions{withoutGrant: true})
			inboundCredential := environment.mintSubjectToken("", "")

			result := initializeGatewayMCP(environment.ctx, environment.gateway, inboundCredential)

			Expect(result.err).To(HaveOccurred(), "the Broker must deny the exchange when no active UserGrant exists")
			Expect(result.status).To(Equal(http.StatusInternalServerError))
			Expect(result.response).To(BeEmpty())
			assertNoInboundCredentialReachedBackend(environment, inboundCredential)
		})

		// US2-S3 from specs/045-agentgateway-token-exchange/spec.md
		It("should fail closed for resource session JWKS and Broker failures", func() {
			cases := []rejectedExchangeCase{
				{
					name:             "missing resource",
					brokerError:      "invalid_request",
					expectedCode:     http.StatusInternalServerError,
					assertNoResource: true,
					options: nativeGatewayEnvironmentOptions{
						omitResource: true,
					},
				},
				{
					name:         "unmapped resource",
					brokerError:  "invalid_target",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						resource: "https://api.example.test/unmapped",
					},
				},
				{
					name:         "missing stored session",
					brokerError:  "invalid_grant",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						withoutSession: true,
					},
				},
				{
					name:         "insufficient stored-session scope",
					brokerError:  "invalid_grant",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						requireGitHubScope: true,
						sessionScopes:      []string{"user"},
					},
				},
				{
					name:         "unavailable client-assertion JWKS",
					brokerError:  "server_error",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						unavailableClientAssertionJWKS: true,
					},
				},
				{
					name:         "unavailable Broker endpoint",
					expectedCode: http.StatusInternalServerError,
					options: nativeGatewayEnvironmentOptions{
						unavailableBroker: true,
						skipRouteLiveness: true,
					},
				},
			}

			for _, testCase := range cases {
				environment := startNativeGatewayEnvironment(testCase.options)
				inboundCredential := environment.mintSubjectToken("", "")

				result := initializeGatewayMCP(environment.ctx, environment.gateway, inboundCredential)
				Expect(result.err).To(HaveOccurred(), "%s must fail the direct exchange", testCase.name)
				Expect(result.status).To(Equal(testCase.expectedCode), "%s must map to the documented agent response", testCase.name)
				Expect(result.response).To(BeEmpty(), "%s must not complete the MCP request", testCase.name)
				if testCase.assertNoResource {
					forms := environment.tokenEndpoint.Forms()
					Expect(forms).To(HaveLen(1), "the Broker must receive the missing-resource form")
					Expect(forms[0].Get("resource")).To(BeEmpty())
				}
				assertNoInboundCredentialReachedBackend(environment, inboundCredential)
				Expect(environment.closeNow()).To(Succeed())
			}
		})
	})
})
