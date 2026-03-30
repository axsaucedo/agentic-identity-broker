package e2e_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
)

var _ = Describe("US5: Signing Key Management (issue_token mode)", func() {
	var (
		adminServer    *bootstrap.TestServer
		enduserServer  *bootstrap.TestServer
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		logger         *slog.Logger
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		config := fixtures.IssueTokenConfig()
		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		adminServer, err = bootstrap.NewAdminTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
		enduserServer, err = bootstrap.NewEndUserTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if adminServer != nil {
			adminServer.Close()
		}
		if enduserServer != nil {
			enduserServer.Close()
		}
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	It("adds a signing key", func() {
		resp, err := http.Post(
			adminServer.BaseURL()+"/api/oauth2-server/signing-keys",
			"application/json",
			strings.NewReader(`{"algorithm":"ES256"}`),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		var body map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&body)).ToNot(HaveOccurred())
		Expect(body).To(HaveKey("kid"))
		Expect(body["algorithm"]).To(Equal("ES256"))
		Expect(body["is_current"]).To(BeTrue())
	})

	It("lists signing keys", func() {
		resp, err := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var body map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&body)).ToNot(HaveOccurred())
		Expect(body).To(HaveKey("items"))
	})

	It("promotes a key to current", func() {
		// Add a second key (auto-current)
		resp, _ := http.Post(
			adminServer.BaseURL()+"/api/oauth2-server/signing-keys",
			"application/json",
			strings.NewReader(`{"algorithm":"ES256"}`),
		)
		var key2 map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&key2)).ToNot(HaveOccurred())
		_ = resp.Body.Close()

		// Get the auto-generated key (not current anymore)
		listResp, _ := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		var listBody map[string]interface{}
		Expect(json.NewDecoder(listResp.Body).Decode(&listBody)).ToNot(HaveOccurred())
		_ = listResp.Body.Close()

		items := listBody["items"].([]interface{})
		var nonCurrentKid string
		for _, item := range items {
			m := item.(map[string]interface{})
			if !m["is_current"].(bool) {
				nonCurrentKid = m["kid"].(string)
				break
			}
		}
		Expect(nonCurrentKid).ToNot(BeEmpty())

		// Promote non-current key
		req, _ := http.NewRequest(http.MethodPut,
			adminServer.BaseURL()+"/api/oauth2-server/signing-keys/"+nonCurrentKid+"/current", nil)
		promoteResp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = promoteResp.Body.Close() }()
		Expect(promoteResp.StatusCode).To(Equal(http.StatusOK))
	})

	It("removes a non-current key", func() {
		// Add second key (becomes current, demoting the auto-generated one)
		resp, _ := http.Post(
			adminServer.BaseURL()+"/api/oauth2-server/signing-keys",
			"application/json",
			strings.NewReader(`{"algorithm":"ES256"}`),
		)
		_ = resp.Body.Close()

		// Find non-current key
		listResp, _ := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		var listBody map[string]interface{}
		Expect(json.NewDecoder(listResp.Body).Decode(&listBody)).ToNot(HaveOccurred())
		_ = listResp.Body.Close()

		items := listBody["items"].([]interface{})
		var nonCurrentKid string
		for _, item := range items {
			m := item.(map[string]interface{})
			if !m["is_current"].(bool) {
				nonCurrentKid = m["kid"].(string)
				break
			}
		}
		Expect(nonCurrentKid).ToNot(BeEmpty())

		// Delete non-current key
		req, _ := http.NewRequest(http.MethodDelete,
			adminServer.BaseURL()+"/api/oauth2-server/signing-keys/"+nonCurrentKid, nil)
		delResp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = delResp.Body.Close() }()
		Expect(delResp.StatusCode).To(Equal(http.StatusNoContent))
	})

	It("cannot remove last key", func() {
		// Only auto-generated key exists; try to delete it
		listResp, _ := http.Get(adminServer.BaseURL() + "/api/oauth2-server/signing-keys")
		var listBody map[string]interface{}
		Expect(json.NewDecoder(listResp.Body).Decode(&listBody)).ToNot(HaveOccurred())
		_ = listResp.Body.Close()

		items := listBody["items"].([]interface{})
		Expect(len(items)).To(BeNumerically(">=", 1))
		kid := items[0].(map[string]interface{})["kid"].(string)

		req, _ := http.NewRequest(http.MethodDelete,
			adminServer.BaseURL()+"/api/oauth2-server/signing-keys/"+kid, nil)
		delResp, err := http.DefaultClient.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = delResp.Body.Close() }()
		Expect(delResp.StatusCode).To(Equal(http.StatusConflict))
	})

	It("key appears in JWKS via discovery", func() {
		// Get JWKS
		resp, err := http.Get(enduserServer.BaseURL() + "/oauth2/jwks.json")
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		var jwks map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&jwks)).ToNot(HaveOccurred())
		keys := jwks["keys"].([]interface{})
		Expect(len(keys)).To(BeNumerically(">=", 1))

		// Verify key has kid field
		key := keys[0].(map[string]interface{})
		Expect(key).To(HaveKey("kid"))
		Expect(key).To(HaveKey("kty"))
	})
})
