package e2e_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
)

var _ = Describe("US2: Server Mode Configuration (local mode)", func() {
	var (
		logger         *slog.Logger
		storageFactory *bootstrap.StorageFactory
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		storageFactory = bootstrap.NewStorageFactory(logger)
	})

	It("starts in local mode", func() {
		config := fixtures.LocalConfig()
		testStorage, err := storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = storageFactory.CloseStorage(testStorage) }()

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		Expect(app).ToNot(BeNil())
	})

	It("no upstream URI needed in local mode", func() {
		config := fixtures.LocalConfig()
		Expect(config.OAuth2AuthServer.Proxy.UpstreamIssuerURI).To(BeEmpty())
		Expect(config.OAuth2AuthServer.Proxy.UpstreamTokenEndpoint).To(BeEmpty())

		testStorage, err := storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = storageFactory.CloseStorage(testStorage) }()

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		Expect(app).ToNot(BeNil())
	})

	It("default is proxy mode", func() {
		config := fixtures.DefaultOAuth2Config()
		Expect(config.OAuth2AuthServer.Mode).To(Equal("proxy"))
	})

	It("auto-generates signing key on startup", func() {
		config := fixtures.LocalConfig()
		testStorage, err := storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = storageFactory.CloseStorage(testStorage) }()

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())

		adminServer, err := bootstrap.NewAdminTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
		defer adminServer.Close()

		resp, err := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var body map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&body)).ToNot(HaveOccurred())
		items, ok := body["items"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(items)).To(BeNumerically(">=", 1))
	})
})
