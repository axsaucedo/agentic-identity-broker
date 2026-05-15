package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/matchers"
)

// Hybrid Mode E2E Tests
//
// Maps to US2 (scenarios 1–3, 7–11) and US3 (scenarios 3–4) from
// specs/030-hybrid-oauth-modes/spec.md. US1 and US2 scenarios 4–6 are
// covered by config unit tests; US2.7–8 are covered in mode_configuration_e2e_test.go.

var _ = Describe("US2: Hybrid Mode Agent Coexistence", func() {
	var (
		logger         *slog.Logger
		storageFactory *bootstrap.StorageFactory
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		testStorage    *storageadapter.Adapter
		server         *bootstrap.TestServer
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		storageFactory = bootstrap.NewStorageFactory(logger)
		mockUpstream = helpers.NewMockUpstreamOAuth2Server()

		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		config := fixtures.HybridConfig(mockUpstream.Server.URL)
		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())

		server, err = bootstrap.NewEndUserTestServer(app, logger)
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

	// Scenario US2.1 from specs/030-hybrid-oauth-modes/spec.md
	It("accepts proxy agent requests (proxy path) in hybrid mode", func() {
		now := time.Now()
		proxyAgent := &storage.Agent{
			ID:           id.NewAgentID(),
			ClientID:     ptr.To(id.ClientID("upstream-client-abc")),
			DisplayName:  "Proxy Agent",
			Description:  "Agent with upstream ClientID — proxy path",
			RedirectURIs: []string{"https://example.com/cb"},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		Expect(testStorage.Agents().Create(context.Background(), proxyAgent)).To(Succeed())

		// Without a grant, proxy mode redirects to consent — proves the proxy agent
		// was resolved and accepted by the hybrid mode strategy.
		resp, err := server.AuthenticatedGET(
			fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://example.com/cb&response_type=code&state=xyz", proxyAgent.ID.String()),
			fixtures.DefaultPrincipal().String(),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		Expect(resp.StatusCode).To(Equal(http.StatusFound))
		Expect(resp.Header.Get("Location")).To(ContainSubstring("/consent/agent/" + proxyAgent.ID.String()))
	})

	// Scenario US2.2 from specs/030-hybrid-oauth-modes/spec.md
	It("accepts local agent requests (local path) in hybrid mode", func() {
		localAgent := fixtures.LocalAgent()
		localAgent.RedirectURIs = []string{"https://example.com/cb"}
		Expect(testStorage.Agents().Create(context.Background(), localAgent)).To(Succeed())

		resp, err := server.AuthenticatedGET(
			fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://example.com/cb&response_type=code&state=xyz", localAgent.ID.String()),
			fixtures.DefaultPrincipal().String(),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		Expect(resp.StatusCode).To(Equal(http.StatusFound))
		Expect(resp.Header.Get("Location")).To(ContainSubstring("/consent/agent/" + localAgent.ID.String()))
	})

	// Scenario US2.10 from specs/030-hybrid-oauth-modes/spec.md
	It("serves JWKS endpoint in hybrid mode", func() {
		resp, err := http.Get(server.BaseURL() + "/oauth2/jwks.json")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var jwks map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&jwks)).ToNot(HaveOccurred())
		keys, ok := jwks["keys"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(keys)).To(BeNumerically(">=", 1))
	})

	// Scenario US2.11 from specs/030-hybrid-oauth-modes/spec.md
	It("exposes discovery metadata with JWKS URI in hybrid mode", func() {
		resp, err := http.Get(server.BaseURL() + "/.well-known/oauth-authorization-server")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var body map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&body)).ToNot(HaveOccurred())
		Expect(body).To(HaveKey("token_endpoint"))
		Expect(body).To(HaveKey("jwks_uri"))
	})
})

var _ = Describe("US2+US3: Hybrid Mode with CIMD", func() {
	var (
		logger         *slog.Logger
		storageFactory *bootstrap.StorageFactory
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		cimdServer     *httptest.Server
		testStorage    *storageadapter.Adapter
		server         *bootstrap.TestServer
		clientURL      string
		redirectURI    string
	)

	const fakeHost = "cimd-hybrid-test.test.invalid"

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		storageFactory = bootstrap.NewStorageFactory(logger)
		mockUpstream = helpers.NewMockUpstreamOAuth2Server()

		clientURL = "https://" + fakeHost + "/client"
		redirectURI = "https://" + fakeHost + "/callback"

		cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "max-age=300")
			w.WriteHeader(http.StatusOK)
			doc := map[string]any{
				"client_id":     clientURL,
				"client_name":   "Hybrid CIMD Agent",
				"redirect_uris": []string{redirectURI},
			}
			_ = json.NewEncoder(w).Encode(doc)
		}))

		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		config := fixtures.HybridConfigWithCIMD(mockUpstream.Server.URL)
		serverFactory := bootstrap.NewServerFactory(config, logger)

		cimdFetcher, err := bootstrap.NewCIMDTestFetcher(cimdServer, fakeHost, 5120)
		Expect(err).ToNot(HaveOccurred())

		app, err := serverFactory.BuildAppWithCIMDFetcher(testStorage, cimdFetcher)
		Expect(err).ToNot(HaveOccurred())

		server, err = bootstrap.NewEndUserTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
		if cimdServer != nil {
			cimdServer.Close()
		}
		if mockUpstream != nil {
			mockUpstream.Close()
		}
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	// Scenario US3.3 from specs/030-hybrid-oauth-modes/spec.md
	It("accepts hybrid+CIMD configuration and starts successfully", func() {
		// Builder succeeded (server is non-nil) — hybrid+CIMD config is accepted.
		// The JWKS endpoint confirms local token signing is wired.
		resp, err := http.Get(server.BaseURL() + "/oauth2/jwks.json")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})

	// Scenario US2.3 from specs/030-hybrid-oauth-modes/spec.md
	It("resolves CIMD agent via URL-format client_id in hybrid mode", func() {
		now := time.Now()
		cimdAgent := &storage.Agent{
			ID:          id.NewAgentID(),
			ClientURIs:  []string{clientURL},
			DisplayName: "CIMD Agent",
			Description: "CIMD agent in hybrid mode",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		Expect(testStorage.Agents().Create(context.Background(), cimdAgent)).To(Succeed())

		resp, err := server.AuthenticatedGET(
			fmt.Sprintf(
				"/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=xyz",
				clientURL, redirectURI,
			),
			fixtures.DefaultPrincipal().String(),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		// CIMD agent resolved via URL → accepted by hybrid mode → consent redirect
		Expect(resp.StatusCode).To(Equal(http.StatusFound))
		Expect(resp.Header.Get("Location")).To(ContainSubstring("/consent/agent/" + cimdAgent.ID.String()))
	})

	// Scenario US3.4 from specs/030-hybrid-oauth-modes/spec.md
	It("proxy agent bypasses CIMD in hybrid mode — resolved by UUID without CIMD lookup", func() {
		now := time.Now()
		proxyAgent := &storage.Agent{
			ID:           id.NewAgentID(),
			ClientID:     ptr.To(id.ClientID("upstream-client-xyz")),
			DisplayName:  "Proxy Agent",
			Description:  "Proxy agent in hybrid+CIMD mode — resolved by UUID",
			RedirectURIs: []string{"https://example.com/cb"},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		Expect(testStorage.Agents().Create(context.Background(), proxyAgent)).To(Succeed())

		// UUID client_id → opaque path → no CIMD fetch → proxy agent accepted
		resp, err := server.AuthenticatedGET(
			fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://example.com/cb&response_type=code&state=xyz", proxyAgent.ID.String()),
			fixtures.DefaultPrincipal().String(),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		Expect(resp).To(matchers.HaveStatusCode(http.StatusFound))
		Expect(resp.Header.Get("Location")).To(ContainSubstring("/consent/agent/" + proxyAgent.ID.String()))
	})
})
