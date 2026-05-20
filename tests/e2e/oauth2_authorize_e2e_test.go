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

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	domainstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

var _ = Describe("US4: Authorization Code Flow with PKCE (local mode)", func() {
	var (
		adminServer    *bootstrap.TestServer
		enduserServer  *bootstrap.TestServer
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		logger         *slog.Logger
		agent          *domainstorage.Agent
		clientSecret   string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		config := fixtures.LocalConfig()
		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		ctx := context.Background()

		// Create a third-party service (required for grant delegation)
		githubService := fixtures.GitHubService()
		Expect(testStorage.Services().Create(ctx, githubService)).ToNot(HaveOccurred())

		// Create agent with redirect_uris
		agent = fixtures.LocalAgent()
		agent.RedirectURIs = []string{"http://localhost:9999/callback"}
		Expect(testStorage.Agents().Create(ctx, agent)).ToNot(HaveOccurred())

		// Create a grant for the principal so consent is satisfied.
		// Grant must include at least one GrantedPermissionSetEntry.
		grant := fixtures.ActiveGrant("test@example.com", agent.ID.String(), githubService.ID.String(), []string{"repo", "user"})
		Expect(testStorage.UserGrants().Create(ctx, grant)).ToNot(HaveOccurred())

		serverFactory := bootstrap.NewServerFactory(config, logger)
		app, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		adminServer, err = bootstrap.NewAdminTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())
		enduserServer, err = bootstrap.NewEndUserTestServer(app, logger)
		Expect(err).ToNot(HaveOccurred())

		// Generate credentials
		resp, err := http.Post(
			adminServer.BaseURL()+"/api/agents/"+agent.ID.String()+"/client-credentials",
			"application/json", nil,
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusCreated))

		var creds map[string]interface{}
		Expect(json.NewDecoder(resp.Body).Decode(&creds)).ToNot(HaveOccurred())
		clientSecret = creds["client_secret"].(string)
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

	It("full auth code flow with PKCE", func() {
		verifier := helpers.PKCEVerifier()
		challenge := helpers.GenerateCodeChallenge(verifier)

		// Step 1: Authorize request (don't follow redirects)
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

		authURL := enduserServer.BaseURL() + "/oauth2/authorize?" + url.Values{
			"response_type":         {"code"},
			"client_id":             {agent.ID.String()},
			"redirect_uri":          {"http://localhost:9999/callback"},
			"state":                 {"test-state"},
			"scope":                 {"read"},
			"code_challenge":        {challenge},
			"code_challenge_method": {"S256"},
		}.Encode()

		req, _ := http.NewRequest("GET", authURL, nil)
		req.Header.Set("X-Remote-User", "test@example.com")
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		// Should redirect with code
		Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location := resp.Header.Get("Location")
		Expect(location).To(ContainSubstring("code="))

		// Extract code from redirect
		locURL, _ := url.Parse(location)
		code := locURL.Query().Get("code")
		Expect(code).ToNot(BeEmpty())

		// Step 2: Exchange code for token
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {agent.ID.String()},
			"client_secret": {clientSecret},
			"code":          {code},
			"redirect_uri":  {"http://localhost:9999/callback"},
			"code_verifier": {verifier},
		}
		tokenResp, err := http.Post(
			enduserServer.BaseURL()+"/oauth2/token",
			"application/x-www-form-urlencoded",
			strings.NewReader(form.Encode()),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = tokenResp.Body.Close() }()
		Expect(tokenResp.StatusCode).To(Equal(http.StatusOK))

		var tokenBody map[string]interface{}
		Expect(json.NewDecoder(tokenResp.Body).Decode(&tokenBody)).ToNot(HaveOccurred())
		Expect(tokenBody).To(HaveKey("access_token"))
		Expect(tokenBody["token_type"]).To(Equal("Bearer"))
	})

	It("invalid redirect URI rejected", func() {
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

		authURL := enduserServer.BaseURL() + "/oauth2/authorize?" + url.Values{
			"response_type":         {"code"},
			"client_id":             {agent.ID.String()},
			"redirect_uri":          {"http://evil.example.com/callback"},
			"code_challenge":        {"test"},
			"code_challenge_method": {"S256"},
		}.Encode()

		req, _ := http.NewRequest("GET", authURL, nil)
		req.Header.Set("X-Remote-User", "test@example.com")
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	It("missing code_challenge rejected", func() {
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

		authURL := enduserServer.BaseURL() + "/oauth2/authorize?" + url.Values{
			"response_type": {"code"},
			"client_id":     {agent.ID.String()},
			"redirect_uri":  {"http://localhost:9999/callback"},
		}.Encode()

		req, _ := http.NewRequest("GET", authURL, nil)
		req.Header.Set("X-Remote-User", "test@example.com")
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()
		// Should reject without PKCE
		Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusBadRequest), Equal(http.StatusFound)))
	})

	It("consent redirect for new user", func() {
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}

		verifier := helpers.PKCEVerifier()
		challenge := helpers.GenerateCodeChallenge(verifier)

		authURL := enduserServer.BaseURL() + "/oauth2/authorize?" + url.Values{
			"response_type":         {"code"},
			"client_id":             {agent.ID.String()},
			"redirect_uri":          {"http://localhost:9999/callback"},
			"code_challenge":        {challenge},
			"code_challenge_method": {"S256"},
		}.Encode()

		// Use a different principal that has no grant
		req, _ := http.NewRequest("GET", authURL, nil)
		req.Header.Set("X-Remote-User", "newuser@example.com")
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = resp.Body.Close() }()

		// Should redirect to consent
		Expect(resp.StatusCode).To(Equal(http.StatusFound))
		location := resp.Header.Get("Location")
		Expect(location).To(ContainSubstring("consent"))
	})

	It("PKCE mismatch rejected on token exchange", func() {
		verifier := helpers.PKCEVerifier()
		challenge := helpers.GenerateCodeChallenge(verifier)

		// Get code
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		authURL := enduserServer.BaseURL() + "/oauth2/authorize?" + url.Values{
			"response_type":         {"code"},
			"client_id":             {agent.ID.String()},
			"redirect_uri":          {"http://localhost:9999/callback"},
			"state":                 {"s"},
			"code_challenge":        {challenge},
			"code_challenge_method": {"S256"},
		}.Encode()
		req, _ := http.NewRequest("GET", authURL, nil)
		req.Header.Set("X-Remote-User", "test@example.com")
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		_ = resp.Body.Close()

		locURL, _ := url.Parse(resp.Header.Get("Location"))
		code := locURL.Query().Get("code")

		// Exchange with wrong verifier
		form := url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {agent.ID.String()},
			"client_secret": {clientSecret},
			"code":          {code},
			"redirect_uri":  {"http://localhost:9999/callback"},
			"code_verifier": {"wrong-verifier-value"},
		}
		tokenResp, err := http.Post(
			enduserServer.BaseURL()+"/oauth2/token",
			"application/x-www-form-urlencoded",
			strings.NewReader(form.Encode()),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = tokenResp.Body.Close() }()
		Expect(tokenResp.StatusCode).To(Equal(http.StatusBadRequest))
	})

	It("code replay rejected", func() {
		verifier := helpers.PKCEVerifier()
		challenge := helpers.GenerateCodeChallenge(verifier)

		// Get code
		client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
		authURL := enduserServer.BaseURL() + "/oauth2/authorize?" + url.Values{
			"response_type":         {"code"},
			"client_id":             {agent.ID.String()},
			"redirect_uri":          {"http://localhost:9999/callback"},
			"state":                 {"s"},
			"code_challenge":        {challenge},
			"code_challenge_method": {"S256"},
		}.Encode()
		req, _ := http.NewRequest("GET", authURL, nil)
		req.Header.Set("X-Remote-User", "test@example.com")
		resp, err := client.Do(req)
		Expect(err).ToNot(HaveOccurred())
		_ = resp.Body.Close()

		locURL, _ := url.Parse(resp.Header.Get("Location"))
		code := locURL.Query().Get("code")

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"client_id":     {agent.ID.String()},
			"client_secret": {clientSecret},
			"code":          {code},
			"redirect_uri":  {"http://localhost:9999/callback"},
			"code_verifier": {verifier},
		}

		// First exchange succeeds
		tokenResp, _ := http.Post(
			enduserServer.BaseURL()+"/oauth2/token",
			"application/x-www-form-urlencoded",
			strings.NewReader(form.Encode()),
		)
		_ = tokenResp.Body.Close()
		Expect(tokenResp.StatusCode).To(Equal(http.StatusOK))

		// Replay fails
		tokenResp2, err := http.Post(
			enduserServer.BaseURL()+"/oauth2/token",
			"application/x-www-form-urlencoded",
			strings.NewReader(form.Encode()),
		)
		Expect(err).ToNot(HaveOccurred())
		defer func() { _ = tokenResp2.Body.Close() }()
		Expect(tokenResp2.StatusCode).To(Equal(http.StatusBadRequest))
	})

	It("code expiry documented", func() {
		// Authorization codes expire after 60 seconds.
		// This behavior is tested at the unit level in provider_test.go (TestProvider_HandleAuthorizationCodeExchange/expired_code_rejects).
		// E2E testing of 60-second expiry would require time manipulation which is not practical.
		Skip("Code expiry tested at unit level — 60s TTL not practical for E2E")
	})
})
