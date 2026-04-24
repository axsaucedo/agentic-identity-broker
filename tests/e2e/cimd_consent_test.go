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
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

var _ = Describe("CIMD Consent Screen", func() {
	var (
		logger         *slog.Logger
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		server         *bootstrap.TestServer
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
		appInstance, err := serverFactory.BuildApp(testStorage)
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
		It("returns cimd_metadata with verified_domain and is_localhost_redirect=false", func() {
			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Example CIMD Agent",
				Description: "E2E test agent for CIMD consent scenario",
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

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data, ok := body["data"].(map[string]any)
			Expect(ok).To(BeTrue())

			cimdMeta, ok := data["cimd_metadata"].(map[string]any)
			Expect(ok).To(BeTrue(), "cimd_metadata should be present")
			Expect(cimdMeta["verified_domain"]).To(Equal("agent.example.com"))
			Expect(cimdMeta["is_localhost_redirect"]).To(BeFalse())
		})
	})

	// Scenario 5.2 from specs/028-cimd-support/spec.md
	Describe("when redirect_uri points to localhost", func() {
		It("returns cimd_metadata with is_localhost_redirect=true", func() {
			now := time.Now()
			agent := &storage.Agent{
				ID:               id.NewAgentID(),
				ClientID:         id.ClientID("https://agent.example.com/client"),
				DisplayName:      "Localhost Redirect Agent",
				Description:      "E2E test agent for localhost redirect CIMD scenario",
				ClientURIs:       []string{"https://agent.example.com/client"},
				CIMDRedirectURIs: []string{"http://localhost:3000/callback"},
				CreatedAt:        now,
				UpdatedAt:        now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf(
				"/api/consent/agent/%s?client_id=https://agent.example.com/client&redirect_uri=http://localhost:3000/callback&scope=repo",
				agent.ID,
			)
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
			Expect(cimdMeta["is_localhost_redirect"]).To(BeTrue())
		})
	})

	// Scenario 5.3 from specs/028-cimd-support/spec.md
	Describe("when the authorization request uses a CIMD-based client_id with scope", func() {
		It("returns cimd_metadata with client_id_url, redirect_uri, and requested_scopes populated", func() {
			now := time.Now()
			agent := &storage.Agent{
				ID:               id.NewAgentID(),
				ClientID:         id.ClientID("https://agent.example.com/client"),
				DisplayName:      "Advanced Detail Agent",
				Description:      "E2E test agent for CIMD advanced detail scenario",
				ClientURIs:       []string{"https://agent.example.com/client"},
				CIMDRedirectURIs: []string{"https://agent.example.com/callback"},
				CreatedAt:        now,
				UpdatedAt:        now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf(
				"/api/consent/agent/%s?client_id=https://agent.example.com/client&redirect_uri=https://agent.example.com/callback&scope=repo+read:user",
				agent.ID,
			)
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

	// Regression test for fail-closed redirect_uri behavior in buildCIMDMetadata:
	// redirect_uri must be empty when the CIMD snapshot has no registered URIs
	// or when the requested redirect_uri is not in the snapshot.
	Describe("redirect_uri fail-closed behavior", func() {
		It("clears redirect_uri when agent has no CIMD snapshot (CIMDRedirectURIs is empty)", func() {
			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "No-Snapshot Agent",
				Description: "E2E test for redirect_uri fail-closed: no snapshot",
				ClientURIs:  []string{"https://agent.example.com/client"},
				// CIMDRedirectURIs intentionally omitted — snapshot not yet populated
				CreatedAt: now,
				UpdatedAt: now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf(
				"/api/consent/agent/%s?client_id=https://agent.example.com/client&redirect_uri=https://agent.example.com/callback&scope=repo",
				agent.ID,
			)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data := body["data"].(map[string]any)
			cimdMeta := data["cimd_metadata"].(map[string]any)
			Expect(cimdMeta["redirect_uri"]).To(BeEmpty(), "redirect_uri must be empty when snapshot is not populated")
		})

		It("clears redirect_uri when requested URI is not in the CIMD snapshot", func() {
			now := time.Now()
			agent := &storage.Agent{
				ID:               id.NewAgentID(),
				ClientID:         id.ClientID("https://agent.example.com/client"),
				DisplayName:      "Mismatch Agent",
				Description:      "E2E test for redirect_uri fail-closed: URI mismatch",
				ClientURIs:       []string{"https://agent.example.com/client"},
				CIMDRedirectURIs: []string{"https://agent.example.com/registered-callback"},
				CreatedAt:        now,
				UpdatedAt:        now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			path := fmt.Sprintf(
				"/api/consent/agent/%s?client_id=https://agent.example.com/client&redirect_uri=https://attacker.example.com/steal&scope=repo",
				agent.ID,
			)
			resp, err := server.AuthenticatedGET(path, fixtures.DefaultPrincipal().String())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			data := body["data"].(map[string]any)
			cimdMeta := data["cimd_metadata"].(map[string]any)
			Expect(cimdMeta["redirect_uri"]).To(BeEmpty(), "redirect_uri must be empty when not in snapshot")
		})
	})

	// Scenario 5.4 from specs/028-cimd-support/spec.md
	Describe("when the authorization request uses an opaque (UUID) client_id", func() {
		It("does not include cimd_metadata in the response", func() {
			now := time.Now()
			agentID := id.NewAgentID()
			agent := &storage.Agent{
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
})
