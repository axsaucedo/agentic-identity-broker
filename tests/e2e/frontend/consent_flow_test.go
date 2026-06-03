// Package e2e_test provides end-to-end tests for the frontend UI using Playwright.
// This file contains tests for the OAuth2 consent flow.
package e2e_test

import (
	"context"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/pages"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ConsentFlow tests verify the OAuth2 consent flow where users grant scopes to agents.
var _ = Describe("Consent Flow", func() {
	var (
		ctx         context.Context
		consentPage *pages.ConsentPage
		testAgentID string
	)

	// Setup per-test resources in BeforeEach
	BeforeEach(func() {
		ctx = context.Background()

		// Step 1: Create a third-party OAuth2 service with scopes
		// This provides the scopes that can be delegated by the user
		service := &model.ThirdpartyOAuth2ProviderEntity{
			ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
			DisplayName: "GitHub",
			ClientID:    id.ClientID("github-client-id"),
			Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440000", "github-client-secret"),
			IssuerURI:   "https://github.com",
			Endpoints: model.OAuth2Endpoints{
				AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
				TokenEndpoint:     "https://github.com/login/oauth/access_token",
			},
			Scopes: []model.OAuthScope{
				{ScopeValue: "repo", Description: "Repository access"},
				{ScopeValue: "user", Description: "User profile access"},
			},
		}
		err := GetTestStorage().Services().Create(ctx, service)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test service")

		// Step 2: Create a test agent with service requirements
		// The agent's service requirements define what services can be delegated to it
		agent := fixtures.ValidAgent()
		testAgentID = agent.ID.String()
		agent.ServiceRequirements = []storage.ServiceRequirement{
			{
				ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"repo", "user"},
			},
		}
		err = GetTestStorage().Agents().Create(ctx, agent)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test agent")

		// Step 3: Create a grant linking the agent to the current user
		// This makes the agent appear in the user's consent page with delegated services/scopes
		principal := fixtures.DefaultPrincipal().String()
		grant := fixtures.IndefiniteGrant(principal, testAgentID, "550e8400-e29b-41d4-a716-446655440000", []string{"repo", "user"})
		err = GetTestStorage().UserGrants().Create(ctx, grant)
		Expect(err).NotTo(HaveOccurred(), "Failed to create test grant")

		// Initialize the page object
		consentPage = pages.NewConsentPage(GetTestPage(), GetFrontendURL())
	})

	// Cleanup after each test
	AfterEach(func() {
		if consentPage != nil {
			_ = consentPage.Close()
		}
	})

	Context("when agent has only optional service requirements", func() {
		BeforeEach(func() {
			// Create two optional services using fixtures; customize only the IDs to avoid
			// conflicts with the service created in the outer BeforeEach (…440000).
			githubService := fixtures.ServiceWithID("550e8400-e29b-41d4-a716-446655440001")
			err := GetTestStorage().Services().Create(ctx, githubService)
			Expect(err).NotTo(HaveOccurred(), "Failed to create optional GitHub service")

			gitlabService := fixtures.ServiceWithID("550e8400-e29b-41d4-a716-446655440002")
			err = GetTestStorage().Services().Create(ctx, gitlabService)
			Expect(err).NotTo(HaveOccurred(), "Failed to create optional GitLab service")

			// Create agent with both services as OPTIONAL requirements.
			// Use AnotherAgent() to avoid client_id conflict with the outer BeforeEach
			// which creates a ValidAgent() with ClientID "test-client-valid".
			agent := fixtures.AnotherAgent()
			testAgentID = agent.ID.String()
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440001"),
					RequirementType: storage.RequirementTypeOptional,
					RequiredScopes:  []string{"read"},
				},
				{
					ServiceID:       id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440002"),
					RequirementType: storage.RequirementTypeOptional,
					RequiredScopes:  []string{"write"},
				},
			}
			err = GetTestStorage().Agents().Create(ctx, agent)
			Expect(err).NotTo(HaveOccurred(), "Failed to create optional-only agent")
		})

		It("should allow approving consent without delegating any optional service", func() {
			// specs/011-agent-permission-requirements/spec.md — User Story 2, Acceptance Scenario 6:
			// "Given optional service requirements, When the authorization endpoint processes the
			// request, Then optional services do not block the authorization flow."

			err := consentPage.NavigateToAgent(ctx, testAgentID)
			Expect(err).NotTo(HaveOccurred(), "Failed to navigate to consent page")

			// Verify the "Approve & Delegate" button is enabled:
			// optional-only agents must never disable the button
			enabled, err := consentPage.IsConsentButtonEnabled(ctx)
			Expect(err).NotTo(HaveOccurred(), "Failed to check consent button state")
			Expect(enabled).To(BeTrue(), "Approve & Delegate button should be enabled for optional-only agent")

			// Click "Approve & Delegate" WITHOUT connecting any services
			err = consentPage.SubmitConsent(ctx)
			Expect(err).NotTo(HaveOccurred(), "SubmitConsent should succeed when button is enabled")

			// Regression guard: no validation error must appear after clicking Approve.
			// WaitForNoValidationError waits for an alert to become visible within the timeout
			// window and treats a timeout as the expected "no error" outcome.
			Expect(consentPage.WaitForNoValidationError(ctx, 2000)).To(Succeed(), "Expected no validation error for optional-only agent")
		})
	})

	// Minimal test: Verify Playwright works and frontend renders
	It("should load consent page and display basic UI elements", func() {
		// When: Navigate to consent page for the test agent
		err := consentPage.NavigateToAgent(ctx, testAgentID)
		Expect(err).NotTo(HaveOccurred(), "Failed to navigate to consent page")

		// Then: Agent name heading should be visible
		agentName, err := consentPage.GetAgentName(ctx)
		Expect(err).NotTo(HaveOccurred(), "Failed to get agent name")
		Expect(agentName).NotTo(BeEmpty(), "Agent name should not be empty")

		// And: Available scopes should be present
		scopes, err := consentPage.GetAvailableScopes(ctx)
		Expect(err).NotTo(HaveOccurred(), "Failed to get available scopes")
		Expect(scopes).NotTo(BeEmpty(), "Should have at least one scope available")

		// And: Take screenshot for verification
		err = consentPage.TakeScreenshot(ctx, "consent_page_loaded")
		Expect(err).NotTo(HaveOccurred(), "Failed to take screenshot")

		GetLogger().Info("Test passed: Consent page renders correctly with UI elements visible")
	})

	// Test: Verify scopes are displayed correctly
	// Note: In the current UI (Phase 8), scopes are displayed as read-only badges based on service requirements.
	// Service delegation happens at the service level (Login/Delegate buttons), not at individual scope level.
	It("should display service scopes as read-only badges based on requirements", func() {
		// Given: User navigates to consent page
		err := consentPage.NavigateToAgent(ctx, testAgentID)
		Expect(err).NotTo(HaveOccurred(), "Failed to navigate to consent page")

		// When: Get available scopes displayed on the page
		scopes, err := consentPage.GetAvailableScopes(ctx)
		Expect(err).NotTo(HaveOccurred(), "Failed to get scopes")
		Expect(scopes).NotTo(BeEmpty(), "Must have at least one scope")

		// Then: Verify expected scopes from service definition are present
		Expect(scopes).To(ContainElement("repo"), "Service should display 'repo' scope")
		Expect(scopes).To(ContainElement("user"), "Service should display 'user' scope")

		// And: Take screenshot for verification
		err = consentPage.TakeScreenshot(ctx, "service_scopes_displayed")
		Expect(err).NotTo(HaveOccurred(), "Failed to take screenshot")

		GetLogger().Info("Test passed: Service scopes are displayed correctly as read-only badges",
			"scopes_count", len(scopes),
			"scopes", scopes,
		)
	})

	// SC-005 from specs/028b-portless-registration/spec.md
	// Verify the loopback warning is shown when the registered redirect URI is portless
	// (http://localhost/callback) and the runtime redirect_uri includes an ephemeral port.
	// FR-006: the localhost warning applies regardless of whether the registered entry specified a port.
	Context("when registered redirect URI is portless loopback and runtime URI carries an ephemeral port", func() {
		var (
			sc005AgentID string
			sc005Token   string
		)

		BeforeEach(func() {
			now := time.Now()
			sc005Agent := &storage.Agent{
				ID:           id.NewAgentID(),
				DisplayName:  "SC-005 Portless Loopback Agent",
				Description:  "Test agent for SC-005: portless loopback registered redirect URI",
				ClientURIs:   []string{"https://sc005-example.com/client_metadata.json"},
				RedirectURIs: []string{"http://localhost/callback"},
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			err := GetTestStorage().Agents().Create(ctx, sc005Agent)
			Expect(err).NotTo(HaveOccurred(), "Failed to create SC-005 CIMD test agent")
			sc005AgentID = sc005Agent.ID.String()

			// The session token carries the runtime redirect_uri with ephemeral port 52341.
			// The CIMD metadata lists the portless registered URI — matching is port-agnostic after the fix.
			sc005Token = newCIMDSessionToken(
				sc005Agent.ID,
				"http://localhost:52341/callback",
				&ports.SessionCIMDMetadata{
					ClientID:     "https://sc005-example.com/client_metadata.json",
					ClientName:   "SC-005 Test Client",
					RedirectURIs: []string{"http://localhost/callback"},
				},
			)
		})

		It("should display the localhost redirect warning alert", func() {
			sc005Page := pages.NewConsentPage(GetTestPage(), GetFrontendURL())
			defer func() { _ = sc005Page.Close() }()

			err := sc005Page.NavigateToAgentWithSessionToken(ctx, sc005AgentID, sc005Token)
			Expect(err).NotTo(HaveOccurred(), "Failed to navigate to SC-005 consent page")

			has, err := sc005Page.HasCIMDLocalhostWarning(ctx)
			Expect(err).NotTo(HaveOccurred(), "Failed to check localhost warning visibility")
			Expect(has).To(BeTrue(), "CIMDLocalhostWarning should be visible for portless-registered loopback URI")

			err = sc005Page.TakeScreenshot(ctx, "consent_loopback_warning_portless")
			Expect(err).NotTo(HaveOccurred(), "Failed to save SC-005 screenshot")
		})
	})
})
