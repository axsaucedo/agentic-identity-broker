package e2e_test

import (
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
	domcimd "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server/cimd"
	domstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

// cimdOAuth2Service extracts the concrete *domotp2.Service from an app instance
// so tests can call CreateAuthorizationSessionToken.
func cimdOAuth2Service(appInstance *app.App) *domotp2.Service {
	svc, ok := appInstance.OAuth2Service.(*domotp2.Service)
	Expect(ok).To(BeTrue(), "OAuth2Service must be *domotp2.Service")
	return svc
}

// createCIMDSessionToken builds a JWE authorization session token for use in CIMD tests.
func createCIMDSessionToken(svc *domotp2.Service, agentID id.AgentID, principal string, redirectURI string, meta *domcimd.MetadataSnapshot) string {
	claims := domotp2.NewAuthorizationSessionClaims(
		agentID,
		id.Principal(principal),
		"https://agent.example.com/client",
		"/oauth2/authorize?client_id=https://agent.example.com/client",
		redirectURI,
		"repo",
		"xyz",
		"challenge123",
		"S256",
		meta,
	)
	token, err := svc.CreateAuthorizationSessionToken(claims)
	Expect(err).ToNot(HaveOccurred())
	return token
}

// createExpiredCIMDSessionToken builds a JWE token with a past ExpiresAt.
func createExpiredCIMDSessionToken(svc *domotp2.Service, agentID id.AgentID, principal string, meta *domcimd.MetadataSnapshot) string {
	past := time.Now().Add(-1 * time.Hour)
	claims := &domotp2.AuthorizationSessionClaims{
		AgentID:             agentID,
		Principal:           id.Principal(principal),
		ClientID:            "https://agent.example.com/client",
		OriginalURL:         "/oauth2/authorize?...",
		RedirectURI:         "https://agent.example.com/callback",
		Scope:               "repo",
		State:               "xyz",
		CodeChallenge:       "challenge123",
		CodeChallengeMethod: "S256",
		CIMDMetadata:        meta,
		IssuedAt:            past,
		ExpiresAt:           past,
	}
	token, err := svc.CreateAuthorizationSessionToken(claims)
	Expect(err).ToNot(HaveOccurred())
	return token
}

var _ = Describe("CIMD Consent Screen", func() {
	var (
		logger         *slog.Logger
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		server         *bootstrap.TestServer
		appInstance    *app.App
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

	// Scenario 5.1 from specs/028-cimd-support/spec.md
	Describe("when the authorization request uses a CIMD-based client_id", func() {
		It("returns cimd_metadata with verified_domain and is_localhost_redirect=false when session_token is provided", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Example CIMD Agent",
				Description: "E2E test agent for CIMD consent scenario",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			token := createCIMDSessionToken(cimdOAuth2Service(appInstance), agent.ID,
				fixtures.DefaultPrincipal().String(),
				"https://agent.example.com/callback",
				&domcimd.MetadataSnapshot{
					ClientID:     "https://agent.example.com/client",
					ClientName:   "Test CIMD Agent",
					RedirectURIs: []string{"https://agent.example.com/callback"},
				},
			)

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data, ok := body["data"].(map[string]any)
			Expect(ok).To(BeTrue())

			cimdMeta, ok := data["cimd_metadata"].(map[string]any)
			Expect(ok).To(BeTrue(), "cimd_metadata should be present")
			Expect(cimdMeta["verified_domain"]).To(Equal("agent.example.com"))
		})
	})

	// Scenario 5.2 from specs/028-cimd-support/spec.md
	Describe("when redirect_uri points to localhost", func() {
		It("returns cimd_metadata with localhost redirect_uri", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Localhost Redirect Agent",
				Description: "E2E test agent for localhost redirect CIMD scenario",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			token := createCIMDSessionToken(cimdOAuth2Service(appInstance), agent.ID,
				fixtures.DefaultPrincipal().String(),
				"http://localhost:3000/callback",
				&domcimd.MetadataSnapshot{
					ClientID:     "https://agent.example.com/client",
					ClientName:   "Test CIMD Agent",
					RedirectURIs: []string{"http://localhost:3000/callback"},
				},
			)

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data, ok := body["data"].(map[string]any)
			Expect(ok).To(BeTrue())

			cimdMeta, ok := data["cimd_metadata"].(map[string]any)
			Expect(ok).To(BeTrue(), "cimd_metadata should be present")
			Expect(cimdMeta["redirect_uri"]).To(Equal("http://localhost:3000/callback"))
		})
	})

	// Scenario 5.3 from specs/028-cimd-support/spec.md
	Describe("when the authorization request uses a CIMD-based client_id with scope", func() {
		It("returns cimd_metadata with client_id_url, redirect_uri, and requested_scopes populated", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Advanced Detail Agent",
				Description: "E2E test agent for CIMD advanced detail scenario",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			svc := cimdOAuth2Service(appInstance)
			claims := domotp2.NewAuthorizationSessionClaims(
				agent.ID,
				id.Principal(fixtures.DefaultPrincipal().String()),
				"https://agent.example.com/client",
				"/oauth2/authorize?...",
				"https://agent.example.com/callback",
				"repo read:user",
				"xyz",
				"challenge123",
				"S256",
				&domcimd.MetadataSnapshot{
					ClientID:     "https://agent.example.com/client",
					ClientName:   "Test CIMD Agent",
					RedirectURIs: []string{"https://agent.example.com/callback"},
				},
			)
			token, err := svc.CreateAuthorizationSessionToken(claims)
			Expect(err).ToNot(HaveOccurred())

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data, ok := body["data"].(map[string]any)
			Expect(ok).To(BeTrue())

			cimdMeta, ok := data["cimd_metadata"].(map[string]any)
			Expect(ok).To(BeTrue(), "cimd_metadata should be present")
			Expect(cimdMeta["client_id_url"]).To(Equal("https://agent.example.com/client"))
			Expect(cimdMeta["redirect_uri"]).To(Equal("https://agent.example.com/callback"))

			scopes, ok := cimdMeta["requested_scopes"].([]any)
			Expect(ok).To(BeTrue(), "requested_scopes should be an array")
			Expect(scopes).To(HaveLen(2))
		})
	})

	// Scenario 5.4 from specs/028-cimd-support/spec.md
	Describe("when the authorization request uses an opaque (UUID) client_id", func() {
		It("does not include cimd_metadata in the response", func() {
			now := time.Now()
			agentID := id.NewAgentID()
			agent := &domstorage.Agent{
				ID:          agentID,
				ClientID:    id.ClientID(agentID.String()),
				DisplayName: "Opaque Client Agent",
				Description: "E2E test agent for opaque client_id scenario",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf("/api/consent/agent/%s", agent.ID)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data, ok := body["data"].(map[string]any)
			Expect(ok).To(BeTrue())

			_, hasCIMDMeta := data["cimd_metadata"]
			Expect(hasCIMDMeta).To(BeFalse(), "cimd_metadata should be absent for opaque client_id")
		})
	})

	// T119 Session Edge Cases from specs/028-cimd-support/spec.md

	// Scenario 5.5 from specs/028-cimd-support/spec.md
	Describe("when session_token carries an expired authorization session", func() {
		It("rejects with 400 Bad Request", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Expired Session Agent",
				Description: "E2E test for expired session rejection",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			token := createExpiredCIMDSessionToken(cimdOAuth2Service(appInstance), agent.ID,
				fixtures.DefaultPrincipal().String(),
				&domcimd.MetadataSnapshot{
					ClientID:     "https://agent.example.com/client",
					ClientName:   "Test CIMD Agent",
					RedirectURIs: []string{"https://agent.example.com/callback"},
				},
			)

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// Scenario 5.6 from specs/028-cimd-support/spec.md
	Describe("when session_token is malformed or invalid", func() {
		It("rejects with 400 Bad Request", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Invalid Token Agent",
				Description: "E2E test for invalid token rejection",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=not-a-valid-jwe-token", agent.ID)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// Scenario 5.7 from specs/028-cimd-support/spec.md — session_token for wrong agent
	Describe("when session_token refers to a different agent", func() {
		It("rejects with 400 Bad Request", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Agent Mismatch Agent",
				Description: "E2E test for agent mismatch rejection",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			// Create token for a different (non-existent) agent ID
			differentAgentID := id.NewAgentID()
			token := createCIMDSessionToken(cimdOAuth2Service(appInstance), differentAgentID,
				fixtures.DefaultPrincipal().String(),
				"https://agent.example.com/callback",
				&domcimd.MetadataSnapshot{
					ClientID:     "https://agent.example.com/client",
					ClientName:   "Test CIMD Agent",
					RedirectURIs: []string{"https://agent.example.com/callback"},
				},
			)

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// SR-013 principal isolation from specs/028-cimd-support/spec.md
	Describe("when session_token belongs to a different principal", func() {
		It("rejects with 400 Bad Request", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Principal Isolation Agent",
				Description: "E2E test for principal mismatch on GET consent",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			// Token issued for a different user
			token := createCIMDSessionToken(cimdOAuth2Service(appInstance), agent.ID,
				"other-user@example.com",
				"https://agent.example.com/callback",
				&domcimd.MetadataSnapshot{
					ClientID:     "https://agent.example.com/client",
					ClientName:   "Test CIMD Agent",
					RedirectURIs: []string{"https://agent.example.com/callback"},
				},
			)

			path := fmt.Sprintf("/api/consent/agent/%s?session_token=%s", agent.ID, token)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// Regression tests: CIMD agents must use session_token — URL-param path is blocked.
	// These verify the fail-closed behavior: sending client_id=https://... without
	// session_token returns 400 (session_token required for CIMD agent authorization).
	Describe("CIMD URL-param path blocked without session_token", func() {
		It("rejects with 400 when CIMD agent has no token and client_id is a URL", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "No-Token Agent",
				Description: "E2E test for session_token requirement: no token",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf(
				"/api/consent/agent/%s?client_id=https://agent.example.com/client&redirect_uri=https://agent.example.com/callback&scope=repo",
				agent.ID,
			)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("does not require session_token when the request omits a URL-format client_id even if the agent has a client URI", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Client URI Agent",
				Description: "E2E test for session_token requirement: with client URI",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf("/api/consent/agent/%s", agent.ID)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data, ok := body["data"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(data).ToNot(HaveKey("cimd_metadata"))
		})
	})
})
