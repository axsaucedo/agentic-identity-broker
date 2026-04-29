package e2e_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// cimdTestHTTPClient returns an HTTP client that redirects all connections to
// fakeHostname to the given test server. Uses InsecureSkipVerify since the test
// server cert covers 127.0.0.1, not the fake hostname. Safe for test use only.
func cimdTestHTTPClient(server *httptest.Server, fakeHostname string) *http.Client {
	parsed, _ := url.Parse(server.URL)
	serverAddr := parsed.Host
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, _ := net.SplitHostPort(addr)
				if host == fakeHostname {
					addr = serverAddr
				}
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
		},
	}
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
			cimdServer  *httptest.Server
			clientURL   string
			redirectURI string
			server      *bootstrap.TestServer
		)

		BeforeEach(func() {
			// Use a fake public hostname to satisfy validateClientURI (port 443 or absent required).
			// The custom HTTP client redirects connections to this hostname to the test server.
			const fakeHost = "cimd-e2e.test.invalid"

			// Start TLS mock CIMD server — URL known only after server starts
			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Cache-Control", "max-age=300")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(cimdDocument(clientURL, []string{redirectURI}))
			}))

			clientURL = "https://" + fakeHost + "/client"
			redirectURI = "https://" + fakeHost + "/callback"

			// Register agent with pre-registered CIMD client URI (FR-026)
			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				ClientURIs:  []string{clientURL},
				DisplayName: "CIMD Test Agent",
				Description: "E2E test agent for CIMD authorization",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			// Wire CIMD-enabled config with custom test client that redirects fakeHost → test server
			config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
			serverFactory = bootstrap.NewServerFactory(config, logger)

			bl, err := domaincimd.NewSSRFBlocklist(nil)
			Expect(err).ToNot(HaveOccurred())
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdTestHTTPClient(cimdServer, fakeHost), bl, 5120)

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

		It("resolves the agent and redirects to the consent page", func() {
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=xyz",
					clientURL, redirectURI,
				),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// Successful CIMD resolution with no existing grant → redirect to consent page
			Expect(resp.StatusCode).To(Equal(http.StatusFound))
			Expect(resp.Header.Get("Location")).To(ContainSubstring("/consent/agent/"))
		})
	})

	// Scenario US1.2 from specs/028-cimd-support/spec.md
	Describe("when CIMD document client_id does not match the request URL", func() {
		var (
			cimdServer *httptest.Server
			clientURL  string
			server     *bootstrap.TestServer
		)

		BeforeEach(func() {
			const fakeHost = "cimd-e2e-mismatch.test.invalid"

			// CIMD server returns a document with a DIFFERENT client_id
			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Intentionally wrong client_id in the document
				_, _ = w.Write(cimdDocument("https://other.example.com/different", []string{clientURL + "/callback"}))
			}))

			clientURL = "https://" + fakeHost + "/client"

			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				ClientURIs:  []string{clientURL},
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
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdTestHTTPClient(cimdServer, fakeHost), bl, 5120)

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
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s/callback&response_type=code&state=xyz",
					clientURL, clientURL,
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
			clientURL  string
			server     *bootstrap.TestServer
		)

		BeforeEach(func() {
			const fakeHost = "cimd-e2e-404.test.invalid"

			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))

			clientURL = "https://" + fakeHost + "/client"

			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				ClientURIs:  []string{clientURL},
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
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdTestHTTPClient(cimdServer, fakeHost), bl, 5120)

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
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s/callback&response_type=code&state=xyz",
					clientURL, clientURL,
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
			const fakeHost = "cimd-e2e-redirect.test.invalid"

			cimdServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// Document only allows /registered-callback, not /other-callback
				_, _ = w.Write(cimdDocument(clientURL, []string{clientURL + "/registered-callback"}))
			}))

			clientURL = "https://" + fakeHost + "/client"

			now := time.Now()
			agent := &storage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(clientURL),
				ClientURIs:  []string{clientURL},
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
			cimdFetcher := adaptercmd.NewFetcherWithClient(cimdTestHTTPClient(cimdServer, fakeHost), bl, 5120)

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

		It("rejects the request with invalid_request", func() {
			// Use a redirect URI that is NOT registered in the CIMD document
			resp, err := server.AuthenticatedGET(
				fmt.Sprintf(
					"/oauth2/authorize?client_id=%s&redirect_uri=%s/other-callback&response_type=code&state=xyz",
					clientURL, clientURL,
				),
				fixtures.DefaultPrincipal().String(),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
			Expect(body["error"]).To(Equal("invalid_request"))
		})
	})
})
