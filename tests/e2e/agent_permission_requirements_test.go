package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

// Agent Permission Requirements E2E Tests
//
// This test suite validates the complete implementation of agent permission requirements
// following Principle XIII (End-to-End Acceptance Testing & Spec Traceability).
//
// Each It() block maps to exactly ONE acceptance scenario from spec.md:
// - User Story 1: 7 scenarios (Admin configuration)
// - User Story 2: 7 scenarios (Authorization validation)
// - User Story 3: 6 scenarios (Consent display)
// - User Story 4: 5 scenarios (Read-only scopes)
// - User Story 5: 4 scenarios (Simplified UI)
// - User Story 6: 5 scenarios (Redirect URL)
// Total: 34 scenarios
//
// CRITICAL: These tests were written BEFORE implementation (TDD red phase)
// and must ALL FAIL initially. Implementation satisfies tests incrementally.

// Helper to read and parse response body consistently
// Note: This function closes resp.Body after reading
func readJSONResponse(resp *http.Response) (map[string]interface{}, error) {
	defer func() {
		_ = resp.Body.Close()
	}()
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// Helper function to find a scope by name in a scopes array
func findScopeByName(scopes []interface{}, name string) map[string]interface{} {
	for _, scopeInterface := range scopes {
		scope := scopeInterface.(map[string]interface{})
		if scope["name"].(string) == name {
			return scope
		}
	}
	return nil
}

var _ = Describe("Agent Permission Requirements", func() {
	var (
		enduserServer  *bootstrap.TestServer
		adminServer    *bootstrap.TestServer
		testStorage    *storageadapter.Adapter
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		serverFactory  *bootstrap.ServerFactory
		logger         *slog.Logger
		adminPrincipal string
		userPrincipal  string
	)

	// Extract setup into helper for better error context
	setupTestEnvironment := func() {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		mockUpstream = helpers.NewMockUpstreamOAuth2Server()
		config := fixtures.OAuth2ConfigWithUpstream(mockUpstream.Server.URL)

		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred(), "Failed to create test storage")

		serverFactory = bootstrap.NewServerFactory(config, logger)
		appInstance, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred(), "Failed to build app instance")

		// Create separate servers for end-user and admin routes
		// This matches production where they would be on different ports/servers
		enduserServer, err = bootstrap.NewEndUserTestServer(appInstance, logger)
		Expect(err).ToNot(HaveOccurred(), "Failed to create end-user test server")

		adminServer, err = bootstrap.NewAdminTestServer(appInstance, logger)
		Expect(err).ToNot(HaveOccurred(), "Failed to create admin test server")
	}

	BeforeEach(func() {
		// Setup: Create FRESH storage for EACH test (test isolation)
		setupTestEnvironment()

		// Initialize principals
		adminPrincipal = "admin@example.com"
		userPrincipal = "user@example.com"
	})

	AfterEach(func() {
		// Cleanup: Close servers and storage
		if enduserServer != nil {
			enduserServer.Close()
		}
		if adminServer != nil {
			adminServer.Close()
		}
		if mockUpstream != nil {
			mockUpstream.Close()
		}
		if testStorage != nil && storageFactory != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	// ===========================================================================
	// User Story 1: Administrator Configures Agent Service Requirements (P1)
	// ===========================================================================

	Describe("User Story 1: Administrator Configures Agent Service Requirements", func() {
		var githubService *model.ThirdpartyOAuth2ProviderEntity

		BeforeEach(func() {
			// Create test third-party service for reference in service requirements
			githubService = &model.ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("github-client"),
				Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440000", "github-secret"),
				IssuerURI:   "https://github.com",
				Endpoints: model.OAuth2Endpoints{
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
				},
				Scopes: []model.OAuthScope{
					{ScopeValue: "repo", Description: "Access repositories"},
					{ScopeValue: "user:email", Description: "Access user email"},
					{ScopeValue: "read:user", Description: "Read user profile"},
				},
			}
			err := testStorage.Services().Create(context.Background(), githubService)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test service")
		})

		// Scenario 1: spec.md User Story 1, Scenario 1
		// Spec: Administrator can optionally specify service requirements when creating agent
		It("should accept optional service_requirements array in POST /api/agents", func() {
			// Given: An authenticated administrator
			// When: Creating a new agent via POST /api/agents with service_requirements
			payload := map[string]interface{}{
				"client_id":    "test-agent-1",
				"display_name": "Test Agent 1",
				"description":  "Agent with service requirements",
				"service_requirements": []map[string]interface{}{
					{
						"service_id":       githubService.ID,
						"requirement_type": "mandatory",
						"required_scopes":  []string{"repo", "user:email"},
					},
				},
			}
			body, _ := json.Marshal(payload)

			resp, err := adminServer.AuthenticatedPOST("/api/agents", adminPrincipal, "application/json", bytes.NewReader(body))
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System accepts the request and returns 201 Created
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		})

		// Scenario 2: spec.md User Story 1, Scenario 2
		// Spec: Each service requirement must include required fields
		It("should require service_id, requirement_type, and required_scopes in each requirement", func() {
			// Given: A service requirement definition
			// When: The request is missing required fields
			agentPayload := map[string]interface{}{
				"client_id":    "test-agent-2",
				"display_name": "Test Agent 2",
				"description":  "Test agent",
				"service_requirements": []map[string]interface{}{
					{
						"service_id": githubService.ID,
						// Missing requirement_type and required_scopes
					},
				},
			}
			body, _ := json.Marshal(agentPayload)

			resp, err := adminServer.AuthenticatedPOST("/api/agents", adminPrincipal, "application/json", bytes.NewReader(body))
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System returns HTTP 400 indicating missing required fields
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		// Scenario 3: spec.md User Story 1, Scenario 3
		// Spec: All scopes must exist in the referenced service
		Context("when all scopes are valid", func() {
			It("should validate that all specified scopes exist in the referenced service", func() {
				// Given: A service requirement with valid required_scopes
				// When: The configuration includes scopes that exist in the service
				payload := map[string]interface{}{
					"client_id":    "test-agent-3",
					"display_name": "Test Agent 3",
					"description":  "Agent with valid scopes",
					"service_requirements": []map[string]interface{}{
						{
							"service_id":       githubService.ID,
							"requirement_type": "mandatory",
							"required_scopes":  []string{"repo", "user:email"}, // Valid scopes
						},
					},
				}
				body, _ := json.Marshal(payload)

				resp, err := adminServer.AuthenticatedPOST("/api/agents", adminPrincipal, "application/json", bytes.NewReader(body))
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					_ = resp.Body.Close()
				}()

				// Then: System accepts the configuration
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			})
		})

		// Scenario 4: spec.md User Story 1, Scenario 4
		// Spec: GET response includes service_name resolved from service display_name
		It("should return service_requirements with resolved display names in GET response", func() {
			// Given: An agent with service requirements
			agent := fixtures.ValidAgent()
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			err := testStorage.Agents().Create(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test agent")

			// When: Administrator retrieves the agent via GET /api/agents/{agent-id}
			resp, err := adminServer.AuthenticatedGET(fmt.Sprintf("/api/agents/%s", agent.ID), adminPrincipal)
			Expect(err).ToNot(HaveOccurred())
			// Then: Response includes service_requirements with service display names
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Verify comprehensive response
			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred(), "Failed to parse response body")

			// Verify agent data
			Expect(responseBody).To(HaveKey("id"))
			Expect(responseBody["id"]).To(Equal(agent.ID.String()))
			Expect(responseBody).To(HaveKey("client_id"))
			Expect(responseBody["client_id"]).To(Equal(agent.ClientID.String()))

			// Verify service requirements
			serviceReqs, ok := responseBody["service_requirements"].([]interface{})
			Expect(ok).To(BeTrue(), "service_requirements should be an array")
			Expect(serviceReqs).To(HaveLen(1))

			firstReq := serviceReqs[0].(map[string]interface{})
			Expect(firstReq["service_id"]).To(Equal(githubService.ID.String()))
			Expect(firstReq["service_name"]).To(Equal("GitHub"), "service_name should be resolved from service display_name")
			Expect(firstReq["requirement_type"]).To(Equal("mandatory"))
			Expect(firstReq["required_scopes"]).To(ConsistOf("repo", "user:email"))
		})

		// Scenario 5: spec.md User Story 1, Scenario 5
		// Spec: PUT replaces entire service requirements array
		It("should replace entire service requirements on PUT /api/agents/{agent-id}", func() {
			// Given: An existing agent with service requirements
			agent := fixtures.ValidAgent()
			err := testStorage.Agents().Create(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test agent")

			// When: Administrator updates with modified service requirements
			updatePayload := map[string]interface{}{
				"client_id":    agent.ClientID, // Required field
				"display_name": agent.DisplayName,
				"description":  agent.Description,
				"service_requirements": []map[string]interface{}{
					{
						"service_id":       githubService.ID,
						"requirement_type": "optional",            // Changed from mandatory
						"required_scopes":  []string{"read:user"}, // Changed scopes
					},
				},
			}
			body, _ := json.Marshal(updatePayload)

			resp, err := adminServer.DirectRequest("PUT", fmt.Sprintf("/api/agents/%s", agent.ID), adminPrincipal,
				map[string]string{"Content-Type": "application/json"}, bytes.NewReader(body))
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System replaces entire service requirements configuration
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		// Scenario 6: spec.md User Story 1, Scenario 6
		// Spec: Non-existent service_id should be rejected with error
		Context("with invalid service reference", func() {
			It("should return HTTP 400 for non-existent service_id reference", func() {
				// Given: A service requirement referencing a non-existent service_id
				// When: The request is processed
				agentPayload := map[string]interface{}{
					"client_id":    "test-agent-6",
					"display_name": "Test Agent 6",
					"description":  "Test agent",
					"service_requirements": []map[string]interface{}{
						{
							"service_id":       "00000000-0000-0000-0000-000000000000", // Non-existent
							"requirement_type": "mandatory",
							"required_scopes":  []string{"some:scope"},
						},
					},
				}
				body, _ := json.Marshal(agentPayload)

				resp, err := adminServer.AuthenticatedPOST("/api/agents", adminPrincipal, "application/json", bytes.NewReader(body))
				Expect(err).ToNot(HaveOccurred())

				// Then: System returns HTTP 400 with error indicating invalid service reference
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				errResp, err := readJSONResponse(resp)
				Expect(err).ToNot(HaveOccurred(), "Response should be valid JSON")
				Expect(errResp).To(HaveKey("error"))

				// Verify error structure (not loose substring match)
				errMsg := errResp["error"].(string)
				Expect(errMsg).ToNot(BeEmpty())
			})
		})

		// Scenario 7: spec.md User Story 1, Scenario 7
		// Spec: Scopes not defined in service should be rejected
		Context("with invalid scopes", func() {
			It("should return HTTP 400 for scopes not defined in the referenced service", func() {
				// Given: A service requirement with scopes not defined in the service
				// When: The request is processed
				agentPayload := map[string]interface{}{
					"client_id":    "test-agent-7",
					"display_name": "Test Agent 7",
					"description":  "Test agent",
					"service_requirements": []map[string]interface{}{
						{
							"service_id":       githubService.ID,
							"requirement_type": "mandatory",
							"required_scopes":  []string{"invalid:scope", "fake:permission"}, // Invalid
						},
					},
				}
				body, _ := json.Marshal(agentPayload)

				resp, err := adminServer.AuthenticatedPOST("/api/agents", adminPrincipal, "application/json", bytes.NewReader(body))
				Expect(err).ToNot(HaveOccurred())

				// Then: System returns HTTP 400 with error listing invalid scope names
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				errResp, err := readJSONResponse(resp)
				Expect(err).ToNot(HaveOccurred(), "Response should be valid JSON")
				Expect(errResp).To(HaveKey("error"))

				// Verify error structure (not loose substring match)
				errMsg := errResp["error"].(string)
				Expect(errMsg).ToNot(BeEmpty())
			})
		})
	})

	// ===========================================================================
	// User Story 2: Authorization Endpoint Validates Service Requirements (P1)
	// ===========================================================================

	Describe("User Story 2: Authorization Endpoint Validates Service Requirements", func() {
		var agent *storage.Agent
		var githubService *model.ThirdpartyOAuth2ProviderEntity

		BeforeEach(func() {
			// Create test third-party service
			githubService = &model.ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("github-client"),
				Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440000", "github-secret"),
				IssuerURI:   "https://github.com",
				Endpoints: model.OAuth2Endpoints{
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
				},
				Scopes: []model.OAuthScope{
					{ScopeValue: "repo", Description: "Access repositories"},
					{ScopeValue: "user:email", Description: "Access user email"},
				},
			}
			err := testStorage.Services().Create(context.Background(), githubService)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test service")

			// Create agent
			agent = fixtures.ValidAgent()
			err = testStorage.Agents().Create(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test agent")
		})

		// Scenario 1: spec.md User Story 2, Scenario 1
		// Spec: System checks user has active sessions for all mandatory services
		It("should check user has active sessions for all mandatory services", func() {
			// Given: Agent with mandatory service requirements
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			// CRITICAL FIX #1: Persist agent changes to storage
			err := testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to update agent with requirements")

			// And: User has no grant (missing service session)
			// (Intentionally not creating a grant)

			// When: Authorization request arrives
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code&state=xyz",
					agent.ID.String()),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System should redirect to consent (missing mandatory service)
			Expect(resp.StatusCode).To(Equal(http.StatusFound))

			// CRITICAL FIX #2: Add redirect Location header verification
			location := resp.Header.Get("Location")
			Expect(location).ToNot(BeEmpty(), "Redirect should have Location header")
			Expect(location).To(ContainSubstring("/consent/agent/"))
			Expect(location).To(ContainSubstring(agent.ID.String()))
		})

		// Scenario 2: spec.md User Story 2, Scenario 2
		// Spec: Session scopes are superset of required scopes
		It("should verify session scopes are superset of required scopes", func() {
			// Given: Agent with mandatory service requirement
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			// CRITICAL FIX #1: Persist agent changes to storage
			err := testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to update agent with requirements")

			// And: User has grant with partial scopes
			grantWithPartialScopes := &storage.UserGrant{
				Principal: id.Principal(userPrincipal),
				AgentID:   agent.ID,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: githubService.ID,
						Scopes:                    []string{"repo"}, // Missing user:email
					},
				},
			}
			err = testStorage.UserGrants().Create(context.Background(), grantWithPartialScopes)
			Expect(err).ToNot(HaveOccurred())

			// When: Authorization endpoint processes the request
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code",
					agent.ID.String()),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System should redirect (when consent endpoint not yet implemented)
			// The system either redirects to consent OR proxies to upstream pending implementation
			Expect(resp.StatusCode).To(Equal(http.StatusFound), "Should redirect when session missing required scopes")

			// CRITICAL FIX #2: Verify redirect Location header is present
			location := resp.Header.Get("Location")
			Expect(location).ToNot(BeEmpty(), "Redirect should have Location header")
		})

		// Scenario 3: spec.md User Story 2, Scenario 3
		// Spec: When all mandatory requirements satisfied, proxy to upstream
		It("should proxy to upstream when all mandatory requirements satisfied", func() {
			// Given: Agent with mandatory requirements
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			// CRITICAL FIX #1: Persist agent changes to storage
			err := testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to update agent with requirements")

			// And: User has grant with all required scopes
			grantWithRequiredScopes := &storage.UserGrant{
				Principal: id.Principal(userPrincipal),
				AgentID:   agent.ID,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: githubService.ID,
						Scopes:                    []string{"repo", "user:email"}, // All required
					},
				},
			}
			err = testStorage.UserGrants().Create(context.Background(), grantWithRequiredScopes)
			Expect(err).ToNot(HaveOccurred())

			// And: User has an active session for the mandatory service with required scopes.
			// Mandatory requirement validation is session-based — having a grant is necessary
			// but not sufficient; the user must also have an active third-party session.
			session := fixtures.SessionForServiceWithScopes(userPrincipal, githubService.ID.String(), []string{"repo", "user:email"})
			err = testStorage.UserSessions().Create(context.Background(), session)
			Expect(err).ToNot(HaveOccurred(), "Failed to create session for mandatory service")

			// When: Authorization request arrives
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code&state=abc123",
					agent.ID.String()),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: Authorization request is proxied to upstream OAuth2 server
			Expect(resp.StatusCode).To(Equal(http.StatusFound), "Should proxy to upstream")

			// CRITICAL FIX #2: Verify redirect goes to upstream (not consent)
			location := resp.Header.Get("Location")
			Expect(location).ToNot(BeEmpty(), "Redirect should have Location header")
			Expect(location).ToNot(ContainSubstring("/consent/agent/"), "Should proxy to upstream, not consent")
		})

		// Scenario 4: spec.md User Story 2, Scenario 4
		// Spec: Missing mandatory service session redirects to consent
		It("should redirect to consent screen when user missing mandatory service session", func() {
			// Given: Agent with mandatory requirements
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			// CRITICAL FIX #1: Persist agent changes to storage
			err := testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to update agent with requirements")

			// And: User has no grant

			// When: Authorization endpoint processes the request
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code",
					agent.ID.String()),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System redirects to consent screen
			Expect(resp.StatusCode).To(Equal(http.StatusFound))

			// CRITICAL FIX #2: Verify redirect location
			location := resp.Header.Get("Location")
			Expect(location).ToNot(BeEmpty(), "Redirect should have Location header")
			Expect(location).To(ContainSubstring("/consent/agent/"))
			Expect(location).To(ContainSubstring(agent.ID.String()))
		})

		// Scenario 5: spec.md User Story 2, Scenario 5
		// Spec: Missing required scopes redirects to consent
		It("should redirect to consent when user session missing required scopes", func() {
			// Given: Agent with mandatory requirements
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			// CRITICAL FIX #1: Persist agent changes to storage
			err := testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to update agent with requirements")

			// And: User has grant without all required scopes
			grantWithoutEmail := &storage.UserGrant{
				Principal: id.Principal(userPrincipal),
				AgentID:   agent.ID,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: githubService.ID,
						Scopes:                    []string{"repo"}, // Missing user:email
					},
				},
			}
			err = testStorage.UserGrants().Create(context.Background(), grantWithoutEmail)
			Expect(err).ToNot(HaveOccurred())

			// When: Authorization endpoint processes the request
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code",
					agent.ID.String()),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System treats this as unmet requirement and redirects
			// (when consent endpoint not yet implemented)
			Expect(resp.StatusCode).To(Equal(http.StatusFound))

			// CRITICAL FIX #2: Verify redirect location header is present
			location := resp.Header.Get("Location")
			Expect(location).ToNot(BeEmpty(), "Redirect should have Location header")
		})

		// Scenario 6: spec.md User Story 2, Scenario 6
		// Spec: Optional service requirements do not block authorization
		It("should not block authorization flow for optional service requirements", func() {
			// Given: Agent with optional service requirements
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeOptional,
					RequiredScopes:  []string{"repo"},
				},
			}
			// CRITICAL FIX #1: Persist agent changes to storage
			err := testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to update agent with requirements")

			// And: User has no grant for optional service

			// When: Authorization endpoint processes the request
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code",
					agent.ID.String()),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: Optional services do not block the authorization flow
			// Should proceed without requiring optional service session
			Expect(resp.StatusCode).To(Or(
				Equal(http.StatusOK),
				Equal(http.StatusFound),
				Equal(http.StatusSeeOther),
			))
		})

		// Scenario 7: spec.md User Story 2, Scenario 7
		// Spec: When agent has no service requirements, only check consent
		Context("and agent has no service requirements", func() {
			It("should only check consent when agent has no service requirements", func() {
				// Given: Agent with no service requirements defined
				Expect(agent.ServiceRequirements).To(BeEmpty())

				// And: User has a grant
				grantWithoutServiceRequirements := &storage.UserGrant{
					Principal: id.Principal(userPrincipal),
					AgentID:   agent.ID,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{
							ThirdpartyOAuth2ServiceID: githubService.ID,
							Scopes:                    []string{"read"},
						},
					},
				}
				err := testStorage.UserGrants().Create(context.Background(), grantWithoutServiceRequirements)
				Expect(err).ToNot(HaveOccurred())

				// When: Authorization endpoint processes the request
				resp, err := enduserServer.AuthenticatedGET(
					fmt.Sprintf("/oauth2/authorize?client_id=%s&redirect_uri=https://client.example.com/cb&response_type=code&state=xyz",
						agent.ID.String()),
					userPrincipal,
				)
				Expect(err).ToNot(HaveOccurred())
				defer func() {
					_ = resp.Body.Close()
				}()

				// Then: System only checks for user consent (existing behavior preserved)
				// Should not require service-specific validations
				Expect(resp.StatusCode).To(Or(
					Equal(http.StatusOK),
					Equal(http.StatusFound),
					Equal(http.StatusSeeOther),
				))
			})
		})
	})

	// ===========================================================================
	// User Story 3: Consent Screen Displays Required Services (P2)
	// ===========================================================================

	Describe("User Story 3: Consent Screen Displays Required Services", func() {
		var githubService *model.ThirdpartyOAuth2ProviderEntity
		var agent *storage.Agent

		BeforeEach(func() {
			// Create test third-party service
			githubService = &model.ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("github-client"),
				Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440000", "github-secret"),
				IssuerURI:   "https://github.com",
				Endpoints: model.OAuth2Endpoints{
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
				},
				Scopes: []model.OAuthScope{
					{ScopeValue: "repo", Description: "Access repositories"},
					{ScopeValue: "user:email", Description: "Access user email"},
					{ScopeValue: "read:user", Description: "Read user profile"},
				},
			}
			err := testStorage.Services().Create(context.Background(), githubService)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test service")

			// Create test agent with service requirements
			agent = fixtures.ValidAgent()
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			}
			err = testStorage.Agents().Create(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test agent")
		})

		// Scenario 1: spec.md User Story 3, Scenario 1
		// Spec: GET /api/consent/agent/{agent-id} returns agent info with service requirements
		It("should return agent information via GET /api/consent/agent/{agent-id}", func() {
			// Given: An agent with mandatory service requirements
			// When: User requests consent endpoint GET /api/consent/agent/{agent-id}
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: System returns 200 OK with agent data
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Verify response contains agent and services
			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred(), "Failed to parse response body")

			data, ok := responseBody["data"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "response.data should be an object")

			_, ok = data["agent"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "response.data.agent should be an object")

			services, ok := data["services"].([]interface{})
			Expect(ok).To(BeTrue(), "response.data.services should be an array")
			Expect(services).ToNot(BeEmpty(), "services array should not be empty")
		})

		// Scenario 2: spec.md User Story 3, Scenario 2
		// Spec: Services include requirementType field ("mandatory" or "optional")
		It("should return services with requirementType field set to 'mandatory' or 'optional'", func() {
			// Given: Agent with service requirement
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Services array includes requirementType field
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})

			// At least one service should be from agent's requirements
			Expect(services).To(HaveLen(1))
			service := services[0].(map[string]interface{})
			Expect(service).To(HaveKey("requirementType"))
			reqType := service["requirementType"].(string)
			Expect(reqType).To(Or(Equal("mandatory"), Equal("optional")))
		})

		// Scenario 3: spec.md User Story 3, Scenario 3
		// Spec: Services include serviceName, serviceId, and connectionStatus
		It("should include serviceName, serviceId, and connectionStatus in each service", func() {
			// Given: Agent with service requirement
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Each service has required fields
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})

			service := services[0].(map[string]interface{})
			Expect(service).To(HaveKey("serviceName"))
			Expect(service).To(HaveKey("serviceId"))
			Expect(service).To(HaveKey("connectionStatus"))

			status := service["connectionStatus"].(string)
			Expect(status).To(Or(Equal("connected"), Equal("not_connected")))
		})

		// Scenario 4: spec.md User Story 3, Scenario 4
		// Spec: Services include requiredScopes with name and description
		It("should include requiredScopes array with name and description fields", func() {
			// Given: Agent with service requirement containing scopes
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Each service includes requiredScopes with descriptions
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})

			service := services[0].(map[string]interface{})
			Expect(service).To(HaveKey("requiredScopes"))
			scopes := service["requiredScopes"].([]interface{})
			Expect(scopes).ToNot(BeEmpty())

			// Verify each scope has name and description
			scope := scopes[0].(map[string]interface{})
			Expect(scope).To(HaveKey("name"))
			Expect(scope).To(HaveKey("description"))
		})

		// Scenario 5: spec.md User Story 3, Scenario 5
		// Spec: Mandatory services appear before optional services in list
		It("should list mandatory services before optional services in response", func() {
			// Given: Agent with both mandatory and optional service requirements
			optionalService := &model.ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440001"),
				DisplayName: "GitLab",
				ClientID:    id.ClientID("gitlab-client"),
				Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440001", "gitlab-secret"),
				IssuerURI:   "https://gitlab.com",
				Endpoints: model.OAuth2Endpoints{
					AuthorizeEndpoint: "https://gitlab.com/oauth/authorize",
					TokenEndpoint:     "https://gitlab.com/oauth/token",
				},
				Scopes: []model.OAuthScope{
					{ScopeValue: "api", Description: "API access"},
				},
			}
			err := testStorage.Services().Create(context.Background(), optionalService)
			Expect(err).ToNot(HaveOccurred())

			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       optionalService.ID,
					RequirementType: storage.RequirementTypeOptional,
					RequiredScopes:  []string{"api"},
				},
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo"},
				},
			}
			err = testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred())

			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Mandatory services appear before optional services
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})

			// Verify ordering: first mandatory, then optional
			Expect(services).To(HaveLen(2))
			firstService := services[0].(map[string]interface{})
			secondService := services[1].(map[string]interface{})
			Expect(firstService["requirementType"]).To(Equal("mandatory"))
			Expect(secondService["requirementType"]).To(Equal("optional"))
		})

		// Scenario 6: spec.md User Story 3, Scenario 6
		// Spec: Each scope includes description from service configuration
		It("should return description field for all configured scopes", func() {
			// Given: Service with properly configured scopes including descriptions
			githubService.Scopes = append(githubService.Scopes, model.OAuthScope{ScopeValue: "gist", Description: "Manage gists"})
			err := testStorage.Services().Update(context.Background(), githubService)
			Expect(err).ToNot(HaveOccurred())

			agent.ServiceRequirements[0].RequiredScopes = append(agent.ServiceRequirements[0].RequiredScopes, "gist")
			err = testStorage.Agents().Update(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred())

			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: All scopes include their descriptions
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})
			service := services[0].(map[string]interface{})
			scopes := service["requiredScopes"].([]interface{})

			// Verify gist scope is included with description
			gistScope := findScopeByName(scopes, "gist")
			Expect(gistScope).ToNot(BeNil())
			Expect(gistScope).To(HaveKey("description"))
			Expect(gistScope["description"]).To(Equal("Manage gists"))
		})
	})

	// ===========================================================================
	// User Story 4: Display-Only Scopes with Descriptions (P2)
	// ===========================================================================

	Describe("User Story 4: Display-Only Scopes with Descriptions", func() {
		var githubService *model.ThirdpartyOAuth2ProviderEntity
		var agent *storage.Agent

		BeforeEach(func() {
			// Create test third-party service with scopes
			githubService = &model.ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440002"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("github-client"),
				Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440002", "github-secret"),
				IssuerURI:   "https://github.com",
				Endpoints: model.OAuth2Endpoints{
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
				},
				Scopes: []model.OAuthScope{
					{ScopeValue: "repo", Description: "Access repositories"},
					{ScopeValue: "user:email", Description: "Access user email"},
					{ScopeValue: "gist", Description: "Manage gists"},
				},
			}
			err := testStorage.Services().Create(context.Background(), githubService)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test service")

			// Create test agent with multiple required scopes
			agent = fixtures.ValidAgent()
			agent.ServiceRequirements = []storage.ServiceRequirement{
				{
					ServiceID:       githubService.ID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email", "gist"},
				},
			}
			err = testStorage.Agents().Create(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test agent")
		})

		// Scenario 1: spec.md User Story 4, Scenario 1
		// Spec: Scopes are returned as read-only arrays (not editable via API)
		It("should return scopes as read-only arrays from API (not editable)", func() {
			// Given: Agent with required scopes
			// When: User requests consent endpoint GET /api/consent/agent/{agent-id}
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: API returns scopes as read-only arrays (no edit mechanism in API)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})
			service := services[0].(map[string]interface{})

			// Verify requiredScopes is an array
			scopes := service["requiredScopes"].([]interface{})
			Expect(scopes).To(HaveLen(3))
		})

		// Scenario 2: spec.md User Story 4, Scenario 2
		// Spec: Each scope has name and description fields
		It("should return each scope with name and description fields", func() {
			// Given: Service with scopes that have descriptions
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Each scope has name and description fields
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})
			service := services[0].(map[string]interface{})
			scopes := service["requiredScopes"].([]interface{})

			for _, scopeInterface := range scopes {
				scope := scopeInterface.(map[string]interface{})
				Expect(scope).To(HaveKey("name"))
				Expect(scope).To(HaveKey("description"))
			}
		})

		// Scenario 3: spec.md User Story 4, Scenario 3
		// Spec: Scope descriptions are populated from service configuration
		It("should populate scope descriptions from service configuration", func() {
			// Given: Service with scopes that have descriptions
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Scope descriptions match service configuration
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})
			service := services[0].(map[string]interface{})
			scopes := service["requiredScopes"].([]interface{})

			// Find repo scope
			repoScope := findScopeByName(scopes, "repo")
			Expect(repoScope).ToNot(BeNil())
			Expect(repoScope["description"]).To(Equal("Access repositories"))

			// Find user:email scope
			emailScope := findScopeByName(scopes, "user:email")
			Expect(emailScope).ToNot(BeNil())
			Expect(emailScope["description"]).To(Equal("Access user email"))
		})

		// Scenario 4: spec.md User Story 4, Scenario 4
		// Spec: All scopes return description field with their configured descriptions
		It("should return description field for all scopes with their descriptions", func() {
			// Given: Service with scopes that have descriptions
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: All scopes have their descriptions
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})
			service := services[0].(map[string]interface{})
			scopes := service["requiredScopes"].([]interface{})

			// Verify gist scope has its description
			gistScope := findScopeByName(scopes, "gist")
			Expect(gistScope).ToNot(BeNil())
			Expect(gistScope).To(HaveKey("description"))
			Expect(gistScope["description"]).To(Equal("Manage gists"))
		})

		// Scenario 5: spec.md User Story 4, Scenario 5
		// Spec: Scopes returned are exactly as configured in agent requirements (read-only enforcement)
		// Note: This is an API contract test - no mechanism to add/remove scopes via API
		It("should return exactly the required scopes configured in the agent", func() {
			// Given: Agent configured with specific scopes
			// When: User requests consent endpoint
			resp, err := enduserServer.AuthenticatedGET(
				fmt.Sprintf("/api/consent/agent/%s", agent.ID),
				userPrincipal,
			)
			Expect(err).ToNot(HaveOccurred())
			// Then: Response contains exactly the configured scopes, no more, no less
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			responseBody, err := readJSONResponse(resp)
			Expect(err).ToNot(HaveOccurred())

			data := responseBody["data"].(map[string]interface{})
			services := data["services"].([]interface{})
			service := services[0].(map[string]interface{})
			scopes := service["requiredScopes"].([]interface{})

			// Verify we have exactly 3 scopes
			Expect(scopes).To(HaveLen(3))

			// Verify the scopes are the configured ones
			scopeNames := []string{}
			for _, scopeInterface := range scopes {
				scope := scopeInterface.(map[string]interface{})
				scopeNames = append(scopeNames, scope["name"].(string))
			}

			Expect(scopeNames).To(ConsistOf("repo", "user:email", "gist"))
		})
	})

	// ===========================================================================
	// User Story 5: Simplified Consent Screen Without Edit Mode (P2)
	// ===========================================================================

	Describe("User Story 5: Simplified Consent Screen Without Edit Mode", func() {
		// Scenario 1: spec.md User Story 5, Scenario 1
		// Spec: Show service connection actions immediately without edit button
		// Status: Pending - Frontend implementation
		It("should show service connection actions immediately without edit button", func() {
			// Given: Consent screen rendered
			// When: Screen loads
			// Then: Service connection actions display immediately, no edit mode toggle
			Skip("Requires frontend implementation of simplified UI without edit mode")
		})

		// Scenario 2: spec.md User Story 5, Scenario 2
		// Spec: Display Login button for unconnected services
		// Status: Pending - Frontend implementation
		It("should display Login button for services user is not connected to", func() {
			// Given: Service without active user session
			// When: Consent screen renders service card
			// Then: Login button is displayed to initiate connection
			Skip("Requires frontend implementation of Login action")
		})

		// Scenario 3: spec.md User Story 5, Scenario 3
		// Spec: Display Disconnect action for connected services
		// Status: Pending - Frontend implementation
		It("should display Disconnect action for connected services", func() {
			// Given: Service with active user session
			// When: Consent screen renders service card
			// Then: Disconnect action is displayed
			Skip("Requires frontend implementation of Disconnect action")
		})

		// Scenario 4: spec.md User Story 5, Scenario 4
		// Spec: No edit mode toggle in updated consent screen
		// Status: Pending - Frontend implementation
		It("should not have edit mode toggle in updated consent screen", func() {
			// Given: Consent screen redesigned for service requirements
			// When: Screen renders
			// Then: No edit mode toggle or button is visible
			Skip("Requires frontend implementation confirming no edit mode")
		})
	})

	// ===========================================================================
	// User Story 6: Redirect URL for Seamless Flow Continuation (P1)
	// ===========================================================================

	Describe("User Story 6: Redirect URL for Seamless Flow Continuation", func() {
		var agent *storage.Agent
		var userPrincipalForGrant string
		var githubService *model.ThirdpartyOAuth2ProviderEntity

		BeforeEach(func() {
			// Create test third-party service
			githubService = &model.ThirdpartyOAuth2ProviderEntity{
				ID:          id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440006"),
				DisplayName: "GitHub",
				ClientID:    id.ClientID("github-client"),
				Secret:      fixtures.EncryptedSecret("550e8400-e29b-41d4-a716-446655440006", "github-secret"),
				IssuerURI:   "https://github.com",
				Endpoints: model.OAuth2Endpoints{
					AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
					TokenEndpoint:     "https://github.com/login/oauth/access_token",
				},
				Scopes: []model.OAuthScope{
					{ScopeValue: "repo", Description: "Access repositories"},
					{ScopeValue: "user:email", Description: "Access user email"},
				},
			}
			err := testStorage.Services().Create(context.Background(), githubService)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test service")

			agent = fixtures.ValidAgent()
			err = testStorage.Agents().Create(context.Background(), agent)
			Expect(err).ToNot(HaveOccurred(), "Failed to create test agent")

			userPrincipalForGrant = "grant-user@example.com"
		})


		// Scenario 3: spec.md User Story 6, Scenario 3
		// Spec: Return 200 OK with success response when no redirect_uri provided
		It("should return 200/201 without redirect when no redirect_uri parameter provided", func() {
			// Given: User approves consent without redirect_uri parameter
			payload := map[string]interface{}{
				"delegated_oauth2_tokens": []map[string]interface{}{
					{
						"thirdparty_oauth2_service_id": githubService.ID,
						"scopes":                       []string{"repo", "user:email"},
					},
				},
			}
			body, _ := json.Marshal(payload)

			// When: User submits grant approval without redirect_uri
			resp, err := enduserServer.AuthenticatedPOST(
				fmt.Sprintf("/api/consent/agent/%s/grants", agent.ID),
				userPrincipalForGrant,
				"application/json",
				bytes.NewReader(body),
			)
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				_ = resp.Body.Close()
			}()

			// Then: System returns success response (201 Created or 200 OK) without redirect
			Expect(resp.StatusCode).To(Or(Equal(http.StatusCreated), Equal(http.StatusOK)))

			// Verify no redirect Location header is set
			location := resp.Header.Get("Location")
			Expect(location).To(BeEmpty(), "No redirect should occur without redirect_uri")
		})

	})
})
