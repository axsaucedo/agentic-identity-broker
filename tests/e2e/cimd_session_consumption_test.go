package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/bootstrap"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/helpers"
)

// CIMD Session Consumption E2E Tests
//
// Covers FR-029 and SR-014: session-based consent submission with single-use
// session consumption and anti-replay enforcement.

var _ = Describe("CIMD Session Consumption on Grant Submission", func() {
	var (
		logger         *slog.Logger
		mockUpstream   *helpers.MockUpstreamOAuth2Server
		storageFactory *bootstrap.StorageFactory
		testStorage    *storageadapter.Adapter
		server         *bootstrap.TestServer
		principalStr   string
	)

	BeforeEach(func() {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
		mockUpstream = helpers.NewMockUpstreamOAuth2Server()
		storageFactory = bootstrap.NewStorageFactory(logger)
		var err error
		testStorage, err = storageFactory.NewTestStorage()
		Expect(err).ToNot(HaveOccurred())

		config := fixtures.OAuth2ConfigWithCIMD(mockUpstream.Server.URL)
		serverFactory := bootstrap.NewServerFactory(config, logger)
		appInstance, err := serverFactory.BuildApp(testStorage)
		Expect(err).ToNot(HaveOccurred())
		server, err = bootstrap.NewEndUserTestServer(appInstance, logger)
		Expect(err).ToNot(HaveOccurred())

		principalStr = fixtures.DefaultPrincipal().String()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
		if mockUpstream != nil {
			mockUpstream.Close()
		}
		if testStorage != nil {
			_ = storageFactory.CloseStorage(testStorage)
		}
	})

	createAgentAndSession := func(principal string) (*domstorage.Agent, *domstorage.AuthorizationSession) {
		now := time.Now()
		agent := &domstorage.Agent{
			ID:          id.NewAgentID(),
			ClientID:    id.ClientID("https://agent.example.com/client"),
			DisplayName: "Session Consumption Agent",
			Description: "E2E test agent for session consumption",
			ClientURIs:  []string{"https://agent.example.com/client"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

		session, err := domstorage.NewAuthorizationSession(
			agent.ID,
			id.Principal(principal),
			"https://agent.example.com/client",
			"/oauth2/authorize?client_id=https://agent.example.com/client&redirect_uri=https://agent.example.com/callback&scope=repo&response_type=code&state=xyz",
			"https://agent.example.com/callback",
			"repo",
			"xyz",
			"challenge123",
			"S256",
			&domstorage.CIMDMetadataSnapshot{
				ClientID:     "https://agent.example.com/client",
				ClientName:   "Test CIMD Agent",
				RedirectURIs: []string{"https://agent.example.com/callback"},
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(testStorage.AuthorizationSessions().Create(context.Background(), session)).To(Succeed())

		return agent, session
	}

	emptyGrantBody := func() *bytes.Reader {
		body, _ := json.Marshal(map[string]interface{}{
			"delegated_oauth2_tokens": []map[string]interface{}{},
		})
		return bytes.NewReader(body)
	}

	// FR-029 from specs/028-cimd-support/spec.md
	Describe("when consent is submitted with a valid session_id", func() {
		It("creates the grant and returns redirect_url from server-side OriginalURL", func() {
			agent, session := createAgentAndSession(principalStr)

			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=%s", agent.ID, session.SessionID)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			var body map[string]any
			Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())

			Expect(body).To(HaveKey("redirect_url"))
			Expect(body["redirect_url"]).To(Equal(session.OriginalURL))

			Expect(body).To(HaveKey("data"))
		})
	})

	// SR-014 from specs/028-cimd-support/spec.md
	Describe("when the same session_id is submitted a second time (replay)", func() {
		It("rejects the second submission", func() {
			agent, session := createAgentAndSession(principalStr)

			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=%s", agent.ID, session.SessionID)

			resp1, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp1.Body.Close() }()
			Expect(resp1.StatusCode).To(Equal(http.StatusCreated))

			resp2, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp2.Body.Close() }()

			Expect(resp2.StatusCode).To(Equal(http.StatusBadRequest))

			var body map[string]any
			Expect(json.NewDecoder(resp2.Body).Decode(&body)).To(Succeed())
			Expect(body["message"]).To(ContainSubstring("already been used"))
		})
	})

	// SR-014 from specs/028-cimd-support/spec.md
	Describe("when consent is submitted with an expired session_id", func() {
		It("rejects with 400 Bad Request", func() {
			now := time.Now()
			agent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID("https://agent.example.com/client"),
				DisplayName: "Expired Session Agent",
				Description: "E2E test for expired session grant rejection",
				ClientURIs:  []string{"https://agent.example.com/client"},
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), agent)).To(Succeed())

			session, err := domstorage.NewAuthorizationSession(
				agent.ID,
				id.Principal(principalStr),
				"https://agent.example.com/client",
				"/oauth2/authorize?client_id=https://agent.example.com/client",
				"https://agent.example.com/callback",
				"repo",
				"xyz",
				"challenge123",
				"S256",
				nil,
			)
			Expect(err).ToNot(HaveOccurred())
			session.ExpiresAt = time.Now().Add(-1 * time.Hour)
			Expect(testStorage.AuthorizationSessions().Create(context.Background(), session)).To(Succeed())

			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=%s", agent.ID, session.SessionID)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// SR-014 from specs/028-cimd-support/spec.md
	Describe("when consent is submitted with a pre-consumed session_id", func() {
		It("rejects with 400 Bad Request", func() {
			agent, session := createAgentAndSession(principalStr)
			Expect(testStorage.AuthorizationSessions().ConsumeIf(context.Background(), session.SessionID, nil)).To(Succeed())

			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=%s", agent.ID, session.SessionID)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// SR-014 from specs/028-cimd-support/spec.md
	Describe("when consent is submitted with a non-existent session_id", func() {
		It("rejects with 400 Bad Request", func() {
			agent, _ := createAgentAndSession(principalStr)

			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=nonexistent0000000000000000000000", agent.ID)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	// SR-014 principal isolation: session belongs to a different user
	Describe("when session_id belongs to a different principal", func() {
		It("rejects with 403 Forbidden", func() {
			agent, session := createAgentAndSession(principalStr)

			otherPrincipal := "other-user@example.com"
			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=%s", agent.ID, session.SessionID)
			resp, err := server.AuthenticatedPOST(path, otherPrincipal, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})
	})

	// SR-014 agent binding: session was created for a different agent
	Describe("when session_id was created for a different agent", func() {
		It("rejects with 400 Bad Request", func() {
			_, session := createAgentAndSession(principalStr)

			now := time.Now()
			otherAgent := &domstorage.Agent{
				ID:          id.NewAgentID(),
				ClientID:    id.ClientID(id.NewAgentID().String()),
				DisplayName: "Other Agent",
				Description: "Different agent",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			Expect(testStorage.Agents().Create(context.Background(), otherAgent)).To(Succeed())

			path := fmt.Sprintf("/api/consent/agent/%s/grants?session_id=%s", otherAgent.ID, session.SessionID)
			resp, err := server.AuthenticatedPOST(path, principalStr, "application/json", emptyGrantBody())
			Expect(err).ToNot(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})
