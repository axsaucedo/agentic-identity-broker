package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	adaptercmd "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/cimd"
	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	domaincimd "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/cimd"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

// cimdDocument builds a CIMD JSON document with the given clientID and redirectURIs.
func cimdDocument(clientID string, redirectURIs []string) []byte {
	doc := map[string]any{
		"client_id":     clientID,
		"client_name":   "Test CIMD Agent",
		"redirect_uris": redirectURIs,
	}
	b, _ := json.Marshal(doc)
	return b
}

var _ = Describe("CIMD Authorization", func() {
	var (
		logger         *slog.Logger
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		serverFactory  *bootstrap.ServerFactory
		testStorage    *storageadapter.Adapter
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
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
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	// Scenario US1.5 from specs/028-cimd-support/spec.md
	Describe("when CIMD is disabled", func() {
		It("rejects URL-format client_id with invalid_client without attempting a fetch", func() {
			config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.Server.URL)
			// CIMD disabled by default (Enabled: false)
			serverFactory = bootstrap.NewServerFactory(config, logger)
			appInstance, err := serverFactory.BuildApp(testStorage)
			Expect(err).ToNot(HaveOccurred())
			server, err := bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
			defer server.Close()

			resp, err := server.AuthenticatedGET(
				"/oauth2/authorize?client_id=https://agent.example.com/client&redirect_uri=https://agent.example.com/cb&response_type=code&state=xyz",
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
			Expect(body["error"]).To(Equal("invalid_client"))
		})
	})

	// Scenario US1.1 from specs/028-cimd-support/spec.md
	Describe("when CIMD is enabled and the document is valid", func() {
		var (
			cimdServer *httptest.Server
			clientURL  string
			agent      *storage.Agent
			server     *bootstrap.TestServer
		)

		BeforeEach(func() {
			var redirectURI string

			// Start TLS mock CIMD server — URL known only after server starts
			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Cache-Control", "max-age=300")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(cimdDocument(clientURL, []string{redirectURI}))
			}))

			clientURL = cimdServer.URL + "/client"
			redirectURI = cimdServer.URL + "/callback"

			// Register agent with CIMD client URI
			now := time.Now()
			agent = &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				DisplayName: "CIMD Test Agent",
				Description: "E2E test agent for CIMD authorization",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			// Wire CIMD-enabled config with test TLS client injected
			config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
			serverFactory = bootstrap.NewServerFactory(config, logger)

			bl, err := domaincimd.NewSSRFBlocklist(nil)
			Expect(err).ToNot(HaveOccurred())
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdServer.Client(), bl, 5120, nil)

			appInstance, err := serverFactory.BuildAppWithCIMDFetcher(testStorage, cimdFetcher)
			Expect(err).ToNot(HaveOccurred())

			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if cimdServer != nil {
				cimdServer.Close()
			}
			if server != nil {
				server.Close()
			}
		})

		It("resolves the agent and proceeds past invalid_client", func() {
			redirectURI := cimdServer.URL + "/callback"
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=xyz",
					clientURL, redirectURI,
				),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// The CIMD client resolved successfully — not a 400 invalid_client
			Expect(resp.StatusCode).ToNot(Equal(http.StatusBadRequest))
		})
	})

	// Scenario US1.2 from specs/028-cimd-support/spec.md
	Describe("when CIMD document client_id does not match the request URL", func() {
		var (
			cimdServer *httptest.Server
			server     *bootstrap.TestServer
		)

		BeforeEach(func() {
			var clientURL string

			// CIMD server returns a document with a DIFFERENT client_id
			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Intentionally wrong client_id in the document
				_, _ = w.Write(cimdDocument("https://other.example.com/different", []string{clientURL + "/callback"}))
			}))

			clientURL = cimdServer.URL + "/client"

			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				DisplayName: "Mismatch Agent",
				Description: "E2E test agent for CIMD mismatch scenario",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
			serverFactory = bootstrap.NewServerFactory(config, logger)

			bl, err := domaincimd.NewSSRFBlocklist(nil)
			Expect(err).ToNot(HaveOccurred())
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdServer.Client(), bl, 5120, nil)

			appInstance, err := serverFactory.BuildAppWithCIMDFetcher(testStorage, cimdFetcher)
			Expect(err).ToNot(HaveOccurred())

			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if cimdServer != nil {
				cimdServer.Close()
			}
			if server != nil {
				server.Close()
			}
		})

		It("rejects the request with invalid_client", func() {
			clientURL := cimdServer.URL + "/client"
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s/callback&response_type=code&state=xyz",
					clientURL, cimdServer.URL,
				),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
			Expect(body["error"]).To(Equal("invalid_client"))
		})
	})

	// Scenario US1.3 from specs/028-cimd-support/spec.md
	Describe("when the CIMD endpoint returns a non-200 status", func() {
		var (
			cimdServer *httptest.Server
			server     *bootstrap.TestServer
		)

		BeforeEach(func() {
			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))

			clientURL := cimdServer.URL + "/client"

			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				DisplayName: "Missing CIMD Agent",
				Description: "E2E test agent for CIMD 404 scenario",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
			serverFactory = bootstrap.NewServerFactory(config, logger)

			bl, err := domaincimd.NewSSRFBlocklist(nil)
			Expect(err).ToNot(HaveOccurred())
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdServer.Client(), bl, 5120, nil)

			appInstance, err := serverFactory.BuildAppWithCIMDFetcher(testStorage, cimdFetcher)
			Expect(err).ToNot(HaveOccurred())

			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if cimdServer != nil {
				cimdServer.Close()
			}
			if server != nil {
				server.Close()
			}
		})

		It("rejects the authorization request with invalid_client", func() {
			clientURL := cimdServer.URL + "/client"
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s/callback&response_type=code&state=xyz",
					clientURL, cimdServer.URL,
				),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
			Expect(body["error"]).To(Equal("invalid_client"))
		})
	})

	// Scenario US1.4 from specs/028-cimd-support/spec.md
	Describe("when redirect_uri is not listed in the CIMD document", func() {
		var (
			cimdServer *httptest.Server
			server     *bootstrap.TestServer
			clientURL  string
		)

		BeforeEach(func() {
			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Document only allows /registered-callback, not /other-callback
				_, _ = w.Write(cimdDocument(clientURL, []string{cimdServer.URL + "/registered-callback"}))
			}))

			clientURL = cimdServer.URL + "/client"

			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				DisplayName: "Redirect Check Agent",
				Description: "E2E test agent for CIMD redirect URI validation",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
			serverFactory = bootstrap.NewServerFactory(config, logger)

			bl, err := domaincimd.NewSSRFBlocklist(nil)
			Expect(err).ToNot(HaveOccurred())
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdServer.Client(), bl, 5120, nil)

			appInstance, err := serverFactory.BuildAppWithCIMDFetcher(testStorage, cimdFetcher)
			Expect(err).ToNot(HaveOccurred())

			server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			if cimdServer != nil {
				cimdServer.Close()
			}
			if server != nil {
				server.Close()
			}
		})

		It("rejects the request with invalid_redirect_uri", func() {
			// Use a redirect URI that is NOT registered in the CIMD document
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s/other-callback&response_type=code&state=xyz",
					clientURL, cimdServer.URL,
				),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
			Expect(body["error"]).To(Equal("invalid_redirect_uri"))
		})
	})
})
