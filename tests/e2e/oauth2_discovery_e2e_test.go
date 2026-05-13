package e2e_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
)

var _ = Describe("US6: Discovery and JWKS (local mode)", func() {
	var (
		enduserServer  *bootstrap.TestServer
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		logger         *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		config := fixtures.LocalConfig()
		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		enduserServer, err = bootstrap.NewEndUserTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if enduserServer != nil {
			enduserServer.Close()
		}
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	It("discovery endpoint returns metadata in local mode", func() {
		resp, err := http.Get(enduserServer.BaseURL() + "/.well-known/oauth-authorization-server")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var body map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&body)).ToNot(HaveOccurred())
		Expect(body).To(HaveKey("issuer"))
		Expect(body).To(HaveKey("authorization_endpoint"))
		Expect(body).To(HaveKey("token_endpoint"))
		Expect(body).To(HaveKey("jwks_uri"))
		Expect(body).To(HaveKey("response_types_supported"))
		Expect(body).To(HaveKey("grant_types_supported"))
		Expect(body).To(HaveKey("code_challenge_methods_supported"))
	})

	It("JWKS endpoint returns valid key set", func() {
		resp, err := http.Get(enduserServer.BaseURL() + "/oauth2/jwks.json")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(resp.Header.Get("Cache-Control")).To(ContainSubstring("max-age=300"))

		var jwks map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&jwks)).ToNot(HaveOccurred())
		Expect(jwks).To(HaveKey("keys"))

		keys := jwks["keys"].([]interface{})
		Expect(len(keys)).To(BeNumerically(">=", 1))

		// Verify no private key material
		for _, k := range keys {
			keyMap := k.(map[string]interface{})
			Expect(keyMap).ToNot(HaveKey("d"))
		}
	})

	It("discovery 404 in proxy mode", func() {
		// Create a proxy-mode server
		proxyConfig := fixtures.OAuth2ConfigWithUpstream("http://localhost:19999")
		proxyStorage, err := storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = storageFactory.CloseStorage(proxyStorage) }()

		proxyFactory := bootstrap.NewServerFactory(proxyConfig, logger)
		proxyApp, err := proxyFactory.BuildApp(proxyStorage)
		Expect(err).ToNot(HaveOccurred())

		proxyServer, err := bootstrap.NewEndUserTestServer(proxyApp, logger)
		Expect(err).ToNot(HaveOccurred())
		defer proxyServer.Close()

		// JWKS should not be available in proxy mode (conditional routing)
		resp, err := http.Get(proxyServer.BaseURL() + "/oauth2/jwks.json")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		// In proxy mode, JWKS route is not registered → 404 or 405
		Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusNotFound), Equal(http.StatusMethodNotAllowed)))
	})
})
