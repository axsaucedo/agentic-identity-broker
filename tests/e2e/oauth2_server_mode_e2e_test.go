package e2e_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/servermode"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
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
		Expect(config.OAuth2AuthServer.Mode).To(Equal(servermode.Proxy))
	})

	It("starts with no signing keys until one is provisioned via the admin API", func() {
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

		// Before provisioning: no keys.
		resp, err := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		var before map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&before)).ToNot(HaveOccurred())
		Expect(before["items"].([]interface{})).To(BeEmpty())

		// Provision via admin API.
		Expect(helpers.ProvisionSigningKey(adminServer.BaseURL())).ToNot(HaveOccurred())

		// After provisioning: one key present.
		resp2, err := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp2.Body.Close() }()
		Expect(resp2.StatusCode).To(Equal(http.StatusOK))
		var after map[string]interface{}
		Expect(json.NewDecoder(resp2.Body).Decode(&after)).ToNot(HaveOccurred())
		Expect(len(after["items"].([]interface{}))).To(BeNumerically(">=", 1))
	})
})
