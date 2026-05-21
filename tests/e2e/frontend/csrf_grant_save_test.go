// Package e2e_test provides end-to-end tests for the frontend UI using Playwright.
// This file tests the complete consent grant save flow including CSRF protection,
// verifying that the full agent → broker → consent → save cycle works correctly.
package e2e_test

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/pages"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Consent Grant Save Flow tests verify that users can navigate to the consent screen
// and successfully save grants, including CSRF token handling across parallel API requests.
// This covers the full agent → broker → consent → save cycle.
var _ = Describe("Consent Grant Save Flow", func() {
	var (
		ctx         context.Context
		consentPage *pages.ConsentPage
		testAgentID string
	)

	// Well-known IDs for this test's fixtures
	var (
		csrfTestServiceID  = id.MustParseServiceID("d0000000-0000-0000-0000-000000000001")
		csrfTestService2ID = id.MustParseServiceID("d0000000-0000-0000-0000-000000000002")
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create two third-party services (the consent page fetches service details)
		svc := &model.ThirdpartyOAuth2ProviderEntity{
			ID:          csrfTestServiceID,
			DisplayName: "CSRF Test Service",
			ClientID:    id.ClientID("csrf-test-client"),
			Secret:      fixtures.EncryptedSecret(csrfTestServiceID.String(), "csrf-test-secret"),
			IssuerURI:   "https://csrf-test.example.com",
			Endpoints: model.OAuth2Endpoints{
				AuthorizeEndpoint: "https://csrf-test.example.com/authorize",
				TokenEndpoint:     "https://csrf-test.example.com/token",
			},
			Scopes: []model.OAuthScope{
				{ScopeValue: "read", Description: "Read access"},
				{ScopeValue: "write", Description: "Write access"},
			},
		}
		err := GetTestStorage().Services().Create(ctx, svc)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test service")

		svc2 := &model.ThirdpartyOAuth2ProviderEntity{
			ID:          csrfTestService2ID,
			DisplayName: "CSRF Test Service 2",
			ClientID:    id.ClientID("csrf-test-client-2"),
			Secret:      fixtures.EncryptedSecret(csrfTestService2ID.String(), "csrf-test-secret-2"),
			IssuerURI:   "https://csrf-test2.example.com",
			Endpoints: model.OAuth2Endpoints{
				AuthorizeEndpoint: "https://csrf-test2.example.com/authorize",
				TokenEndpoint:     "https://csrf-test2.example.com/token",
			},
			Scopes: []model.OAuthScope{
				{ScopeValue: "email", Description: "Email access"},
			},
		}
		err = GetTestStorage().Services().Create(ctx, svc2)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test service 2")

		// Create agent with OPTIONAL service requirements only (no permission sets).
		// This ensures the "Approve & Delegate" button is always enabled without
		// needing active service sessions.
		agent := fixtures.AgentWithClientID("csrf-flow-test-client")
		testAgentID = agent.ID.String()
		agent.ServiceRequirements = []storage.ServiceRequirement{
			{
				ServiceID:       csrfTestServiceID,
				RequirementType: storage.RequirementTypeOptional,
				RequiredScopes:  []string{"read"},
			},
			{
				ServiceID:       csrfTestService2ID,
				RequirementType: storage.RequirementTypeOptional,
				RequiredScopes:  []string{"email"},
			},
		}
		err = GetTestStorage().Agents().Create(ctx, agent)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test agent")

		consentPage = pages.NewConsentPage(GetTestPage(), GetFrontendURL())
	})

	AfterEach(func() {
		if consentPage != nil {
			_ = consentPage.Close()
		}
	})

	// This test exercises the full CSRF-protected consent grant save flow.
	// It reproduces the production bug where parallel GET requests during page load
	// could race on CSRF token generation, causing the subsequent POST to fail with 403.
	It("should successfully save a grant through the consent screen with CSRF protection", func() {
		// specs/007-consent-frontend/spec.md — User Story 3, Scenario 7:
		// "Given a user clicks Approve & Delegate, When the request is processed successfully,
		// Then the system creates or updates the grant and displays a success message."
		// Regression: CSRF token race on parallel GETs must not cause 403 on subsequent POST.
		err := consentPage.NavigateToAgent(ctx, testAgentID)
		Expect(err).NotTo(HaveOccurred(), "Failed to navigate to consent page")

		// Verify the page loaded with the agent name
		agentName, err := consentPage.GetAgentName(ctx)
		Expect(err).NotTo(HaveOccurred(), "Failed to get agent name")
		Expect(agentName).NotTo(BeEmpty(), "Agent name should be displayed")

		// Submit consent — this sends POST /api/consent/agents/{id}/grants with the
		// CSRF token from the cookie in the X-CSRF-Token header.
		err = consentPage.SubmitConsent(ctx)
		Expect(err).NotTo(HaveOccurred(), "Failed to click Approve & Delegate button")

		// Wait for the success toast — if CSRF is broken, we'd get an error toast
		// with "permission" / "Forbidden" content instead.
		err = consentPage.WaitForGrantSuccess(ctx, 5000)
		Expect(err).NotTo(HaveOccurred(), "Grant save should succeed (CSRF token must be valid)")
	})

	// Verify CSRF works correctly when navigating with redirect_uri (OAuth2 authorization flow).
	It("should save grant without CSRF error when redirect_uri parameter is present", func() {
		// specs/011-agent-permission-requirements/spec.md — User Story 6, Scenario 1:
		// "Given a user is on the consent screen with a redirect_uri query parameter,
		// When the user clicks the Approve button, Then the backend issues an HTTP redirect
		// to the URL specified in redirect_uri."
		// Regression: CSRF token must remain valid when the page is loaded with a redirect_uri param.

		// Navigate with redirect_uri simulating the OAuth2 authorization redirect
		err := consentPage.NavigateToAgentWithRedirectURI(ctx, testAgentID, "/oauth2/callback?code=abc")
		Expect(err).NotTo(HaveOccurred(), "Failed to navigate with redirect_uri")

		// Verify page loaded
		agentName, err := consentPage.GetAgentName(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(agentName).NotTo(BeEmpty())

		// Submit consent — the POST includes redirect_uri and the CSRF token must be valid.
		// If CSRF is broken, the POST returns 403 and the frontend shows an error toast.
		err = consentPage.SubmitConsent(ctx)
		Expect(err).NotTo(HaveOccurred(), "Failed to submit consent with redirect_uri")

		// The backend returns a redirect_url in the JSON body which the frontend follows.
		// Verify no CSRF/permission error toast appeared — success is proven by the absence
		// of an error after submitting.
		Expect(consentPage.WaitForNoValidationError(ctx, 3000)).To(Succeed(),
			"No CSRF/permission error should appear after grant save with redirect_uri")
	})
})
