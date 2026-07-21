package e2e_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/matchers"
)

var _ = Describe("Provider authorization parameters", func() {
	var (
		adminServer   *bootstrap.TestServer
		enduserServer *bootstrap.TestServer
		storage       *storageadapter.Adapter
		factory       *bootstrap.StorageFactory
		principal     string
	)

	serviceRequest := func(params map[string]string) map[string]interface{} {
		request := map[string]interface{}{
			"display_name":  "Zalando Platform",
			"client_id":     "zalando-client",
			"client_secret": "zalando-secret",
			"issuer_uri":    "https://provider.example.com",
			"discovery":     map[string]interface{}{"enable_discovery": false},
			"endpoints": map[string]interface{}{
				"token_endpoint":     "https://provider.example.com/oauth/token",
				"authorize_endpoint": "https://provider.example.com/oauth/authorize",
			},
			"scopes": []map[string]interface{}{{"scope_value": "profile", "description": "Profile"}},
		}
		if params != nil {
			request["authorization_params"] = params
		}
		return request
	}

	create := func(request map[string]interface{}) map[string]interface{} {
		body, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		response, err := adminServer.AuthenticatedPOST("/api/services", principal, "application/json", bytes.NewReader(body))
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(matchers.HaveStatusCode(http.StatusCreated))
		defer func() { _ = response.Body.Close() }()
		var decoded map[string]interface{}
		Expect(json.NewDecoder(response.Body).Decode(&decoded)).To(Succeed())
		return decoded
	}

	BeforeEach(func() {
		logger := slog.New(slog.NewTextHandler(GinkgoWriter, nil))
		factory = bootstrap.NewStorageFactory(logger)
		var err error
		storage, err = factory.NewTestStorage()
		Expect(err).NotTo(HaveOccurred())
		app, err := bootstrap.NewServerFactory(fixtures.DefaultOAuth2Config(), logger).BuildApp(storage)
		Expect(err).NotTo(HaveOccurred())
		adminServer, err = bootstrap.NewAdminTestServer(app, logger)
		Expect(err).NotTo(HaveOccurred())
		enduserServer, err = bootstrap.NewEndUserTestServer(app, logger)
		Expect(err).NotTo(HaveOccurred())
		principal = fixtures.DefaultPrincipal().String()
	})

	AfterEach(func() {
		if adminServer != nil {
			adminServer.Close()
		}
		if enduserServer != nil {
			enduserServer.Close()
		}
		if factory != nil && storage != nil {
			_ = factory.CloseStorage(storage)
		}
	})

	// Scenario 1.1 from specs/034-provider-auth-params/spec.md
	It("returns configured parameters from create, get, and list", func() {
		created := create(serviceRequest(map[string]string{"business_partner_id": "12345"}))
		Expect(created["authorization_params"]).To(Equal(map[string]interface{}{"business_partner_id": "12345"}))
		serviceID := created["id"].(string)

		response, err := adminServer.AuthenticatedGET("/api/services/"+serviceID, principal)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		var got map[string]interface{}
		Expect(json.NewDecoder(response.Body).Decode(&got)).To(Succeed())
		Expect(got["authorization_params"]).To(Equal(map[string]interface{}{"business_partner_id": "12345"}))

		list, err := adminServer.AuthenticatedGET("/api/services", principal)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = list.Body.Close() }()
		var services []map[string]interface{}
		Expect(json.NewDecoder(list.Body).Decode(&services)).To(Succeed())
		Expect(services).To(ContainElement(HaveKeyWithValue("authorization_params", map[string]interface{}{"business_partner_id": "12345"})))
	})

	// Scenario 1.2 from specs/034-provider-auth-params/spec.md
	It("preserves parameters when update omits them", func() {
		created := create(serviceRequest(map[string]string{"business_partner_id": "12345"}))
		request := serviceRequest(nil)
		request["display_name"] = "Updated Zalando"
		body, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		response, err := adminServer.DirectRequest(http.MethodPut, "/api/services/"+created["id"].(string), principal, map[string]string{"Content-Type": "application/json"}, bytes.NewReader(body))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response).To(matchers.HaveStatusCode(http.StatusOK))
		var updated map[string]interface{}
		Expect(json.NewDecoder(response.Body).Decode(&updated)).To(Succeed())
		Expect(updated["authorization_params"]).To(Equal(map[string]interface{}{"business_partner_id": "12345"}))
	})

	// Scenario 1.3 from specs/034-provider-auth-params/spec.md
	It("allows an empty map to clear parameters", func() {
		created := create(serviceRequest(map[string]string{"business_partner_id": "12345"}))
		request := serviceRequest(map[string]string{})
		body, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		response, err := adminServer.DirectRequest(http.MethodPut, "/api/services/"+created["id"].(string), principal, map[string]string{"Content-Type": "application/json"}, bytes.NewReader(body))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response).To(matchers.HaveStatusCode(http.StatusOK))
		var updated map[string]interface{}
		Expect(json.NewDecoder(response.Body).Decode(&updated)).To(Succeed())
		Expect(updated).NotTo(HaveKey("authorization_params"))
	})

	// Scenario 2.1 from specs/034-provider-auth-params/spec.md
	It("adds configured parameters to the upstream authorization URL", func() {
		created := create(serviceRequest(map[string]string{"business_partner_id": "12345"}))
		response, err := enduserServer.AuthenticatedGET("/api/third-party/"+created["id"].(string)+"/oauth2/authorize?redirect_uri="+url.QueryEscape("http://localhost:8000/done"), principal)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response).To(matchers.HaveStatusCode(http.StatusFound))
		upstream, err := url.Parse(response.Header.Get("Location"))
		Expect(err).NotTo(HaveOccurred())
		Expect(upstream.Query().Get("business_partner_id")).To(Equal("12345"))
		Expect(upstream.Query().Get("state")).NotTo(BeEmpty())
	})

	// Scenario 2.2 from specs/034-provider-auth-params/spec.md
	It("adds no provider parameters for an unconfigured service", func() {
		created := create(serviceRequest(nil))
		response, err := enduserServer.AuthenticatedGET("/api/third-party/"+created["id"].(string)+"/oauth2/authorize?redirect_uri="+url.QueryEscape("http://localhost:8000/done"), principal)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		upstream, err := url.Parse(response.Header.Get("Location"))
		Expect(err).NotTo(HaveOccurred())
		Expect(upstream.Query().Get("business_partner_id")).To(BeEmpty())
	})

	// Scenario 2.3 from specs/034-provider-auth-params/spec.md
	It("ignores conflicting browser parameters", func() {
		created := create(serviceRequest(map[string]string{"business_partner_id": "12345"}))
		path := "/api/third-party/" + created["id"].(string) + "/oauth2/authorize?redirect_uri=" + url.QueryEscape("http://localhost:8000/done") + "&business_partner_id=untrusted"
		response, err := enduserServer.AuthenticatedGET(path, principal)
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		upstream, err := url.Parse(response.Header.Get("Location"))
		Expect(err).NotTo(HaveOccurred())
		Expect(upstream.Query().Get("business_partner_id")).To(Equal("12345"))
	})

	// Scenario 3.1 from specs/034-provider-auth-params/spec.md
	It("rejects blank parameter names and values", func() {
		request := serviceRequest(map[string]string{" ": "value"})
		body, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		response, err := adminServer.AuthenticatedPOST("/api/services", principal, "application/json", bytes.NewReader(body))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response).To(matchers.HaveStatusCode(http.StatusBadRequest))
	})

	// Scenario 3.2 from specs/034-provider-auth-params/spec.md
	It("rejects broker-owned parameter names", func() {
		request := serviceRequest(map[string]string{"state": "unsafe"})
		body, err := json.Marshal(request)
		Expect(err).NotTo(HaveOccurred())
		response, err := adminServer.AuthenticatedPOST("/api/services", principal, "application/json", bytes.NewReader(body))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = response.Body.Close() }()
		Expect(response).To(matchers.HaveStatusCode(http.StatusBadRequest))
	})
})
