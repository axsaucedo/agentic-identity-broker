package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

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

	It("returns server_error from /oauth2/token before signing key provisioning", func() {
		config := fixtures.LocalConfig()
		testStorage, err := storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = storageFactory.CloseStorage(testStorage) }()

		agent := fixtures.LocalAgent()
		Expect(testStorage.Agents().Create(context.Background(), agent)).ToNot(HaveOccurred())

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())

		adminServer, err := bootstrap.NewAdminTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
		defer adminServer.Close()

		enduserServer, err := bootstrap.NewEndUserTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
		defer enduserServer.Close()

		credentialsResp, err := http.Post(
			adminServer.BaseURL()+"/api/agents/"+agent.ID.String()+"/client-credentials",
			"application/json",
			nil,
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = credentialsResp.Body.Close() }()
		Expect(credentialsResp.StatusCode).To(Equal(http.StatusCreated))

		var credentials map[string]interface{}
		Expect(json.NewDecoder(credentialsResp.Body).Decode(&credentials)).ToNot(HaveOccurred())
		clientSecret, ok := credentials["client_secret"].(string)
		Expect(ok).To(BeTrue())
		Expect(clientSecret).ToNot(BeEmpty())

		form := url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {agent.ID.String()},
			"client_secret": {clientSecret},
		}
		resp, err := enduserServer.PublicPOST(
			"/oauth2/token",
			"application/x-www-form-urlencoded",
			strings.NewReader(form.Encode()),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))

		var body map[string]string
		Expect(json.NewDecoder(resp.Body).Decode(&body)).ToNot(HaveOccurred())
		Expect(body["error"]).To(Equal("server_error"))
	})
})
