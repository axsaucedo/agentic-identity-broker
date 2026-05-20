package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domotp2 "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	domcimd "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/cimd"
	domstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

// CIMD Session Consumption E2E Tests
//
// Covers FR-029 and SR-014: session-based consent submission using stateless JWE tokens.
// Anti-replay is enforced by the token TTL — there is no server-side consumed flag.

var _ = Describe("CIMD Session Consumption on Grant Submission", func() {
	var (
		logger         *slog.Logger
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		server         *bootstrap.TestServer
		appInstance    *app.App
		principalStr   string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
		mockUpstream = helpers.NewMockUpstreamOAuth2Server()
		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
		serverFactory := bootstrap.NewServerFactory(config, logger)
		appInstance, err = serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
		Expect(err).ToNot(HaveOccurred())

		principalStr = fixtures.DefaultPrincipal().String()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
		if mockUpstream != nil {
			mockUpstream.Close()
		}
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	// buildToken creates a valid JWE session token for the given agent and principal.
	buildToken := func(agentID id.AgentID, principal, redirectURI, originalURL string) string {
		svc, ok := appInstance.OAuth2Service.(*domotp2.Service)
		Expect(ok).To(BeTrue(), "OAuth2Service must be *domotp2.Service")
		claims, err := domotp2.NewAuthorizationSessionClaims(
			agentID,
			id.Principal(principal),
			originalURL,
			&domcimd.ClientIDMetadataDocument{
				ClientID:     "https://agent.example.com/client",
				ClientName:   "Test CIMD Agent",
				RedirectURIs: []string{redirectURI},
			},
		)
		Expect(err).NotTo(HaveOccurred())
		token, err := svc.CreateAuthorizationSessionToken(claims)
		Expect(err).ToNot(HaveOccurred())
		return token
	}

	// buildExpiredToken creates a JWE token with a past ExpiresAt.
	buildExpiredToken := func(agentID id.AgentID, principal string) string {
		svc, ok := appInstance.OAuth2Service.(*domotp2.Service)
		Expect(ok).To(BeTrue(), "OAuth2Service must be *domotp2.Service")
		past := time.Now().Add(-1 * time.Hour)
		claims := &domotp2.AuthorizationSessionClaims{
			AgentID:   agentID,
			Principal: id.Principal(principal),
			IssuedAt:  past,
			ExpiresAt: past,
		}
		token, err := svc.CreateAuthorizationSessionToken(claims)
		Expect(err).ToNot(HaveOccurred())
		return token
	}

	createAgent := func() *domstorage.Agent {
		now := time.Now()
		agent := &domstorage.Agent{
			ID:          id.NewAgentID(),
			DisplayName: "Session Consumption Agent",
			Description: "E2E test agent for session consumption",
			ClientURIs:  []string{"https://agent.example.com/client"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())
		return agent
	}

	emptyGrantBody := func() *bytes.Reader {
		body, _ := json.Marshal(map[string]interface{}{
			"delegated_oauth2_tokens": []map[string]interface{}{},
		})
		return bytes.NewReader(body)
	}

	// FR-029 from specs/028-cimd-support/spec.md
	Describe("when consent is submitted with a valid session_token", func() {
		It("creates the grant and returns redirect_url from the JWE claims OriginalURL", func() {
			agent := createAgent()
			originalURL := "/oauth2/authorize?client_id=https://agent.example.com/client&redirect_uri=https://agent.example.com/callback&scope=repo&response_type=code&state=xyz"
			token := buildToken(agent.ID, principalStr, "https://agent.example.com/callback", originalURL)

			path := fmt.Sprintf("/api/consent/agents/%s/grants?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			Expect(body).To(HaveKey("redirect_url"))
			Expect(body["redirect_url"]).To(Equal(originalURL))
			Expect(body).To(HaveKey("data"))
		})
	})

	// SR-014 from specs/028-cimd-support/spec.md — expired token is rejected
	Describe("when consent is submitted with an expired session_token", func() {
		It("rejects with 400 Bad Request", func() {
			agent := createAgent()
			token := buildExpiredToken(agent.ID, principalStr)

			path := fmt.Sprintf("/api/consent/agents/%s/grants?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// SR-014 from specs/028-cimd-support/spec.md — malformed token is rejected
	Describe("when consent is submitted with a malformed session_token", func() {
		It("rejects with 400 Bad Request", func() {
			agent := createAgent()

			path := fmt.Sprintf("/api/consent/agents/%s/grants?session_token=not-a-valid-jwe", agent.ID)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// SR-014 principal isolation: session token issued for a different user
	Describe("when session_token belongs to a different principal", func() {
		It("rejects with 403 Forbidden", func() {
			agent := createAgent()
			// Token issued for a different principal
			token := buildToken(agent.ID, "other-user@example.com",
				"https://agent.example.com/callback",
				"/oauth2/authorize?client_id=https://agent.example.com/client",
			)

			path := fmt.Sprintf("/api/consent/agents/%s/grants?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})
	})

	// SR-014 agent binding: session token was created for a different agent
	Describe("when session_token was created for a different agent", func() {
		It("rejects with 400 Bad Request", func() {
			agent := createAgent()

			// Create token bound to a different agent ID
			differentAgentID := id.NewAgentID()
			token := buildToken(differentAgentID, principalStr,
				"https://agent.example.com/callback",
				"/oauth2/authorize?client_id=https://agent.example.com/client",
			)

			path := fmt.Sprintf("/api/consent/agents/%s/grants?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})
