package enduser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ory/fosite"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
)

// mockTokenMintingStrategy is a configurable test double for ports.TokenMintingStrategy.
type mockTokenMintingStrategy struct {
	clientCredentialsFn         func(context.Context, id.AgentID, string, string) (*ports.TokenResponse, error)
	authorizationCodeExchangeFn func(context.Context, id.AgentID, string, string, string, string) (*ports.TokenResponse, error)
}

func (m *mockTokenMintingStrategy) HandleClientCredentials(ctx context.Context, agentID id.AgentID, clientSecret, scope string) (*ports.TokenResponse, error) {
	return m.clientCredentialsFn(ctx, agentID, clientSecret, scope)
}

func (m *mockTokenMintingStrategy) HandleAuthorizationCodeExchange(ctx context.Context, agentID id.AgentID, clientSecret, code, redirectURI, codeVerifier string) (*ports.TokenResponse, error) {
	return m.authorizationCodeExchangeFn(ctx, agentID, clientSecret, code, redirectURI, codeVerifier)
}

// fixedMinting returns a mock strategy that always returns the given response/error for both grant types.
func fixedMinting(resp *ports.TokenResponse, err error) *mockTokenMintingStrategy {
	return &mockTokenMintingStrategy{
		clientCredentialsFn: func(_ context.Context, _ id.AgentID, _, _ string) (*ports.TokenResponse, error) {
			return resp, err
		},
		authorizationCodeExchangeFn: func(_ context.Context, _ id.AgentID, _, _, _, _ string) (*ports.TokenResponse, error) {
			return resp, err
		},
	}
}

// mockMultiAgentVerifier is a test double for MultiAgentVerifier
type mockMultiAgentVerifier struct {
	verifyFn func(ctx context.Context, responseBody []byte, expectedAgentID id.AgentID) error
}

func (m *mockMultiAgentVerifier) VerifyAgentIDClaim(ctx context.Context, responseBody []byte, expectedAgentID id.AgentID) error {
	return m.verifyFn(ctx, responseBody, expectedAgentID)
}

// stubAgentRepo is a minimal agent repository stub for unit testing.
// Get always returns the configured agent (or error), ignoring the agentID argument.
// This is intentional: unit tests in this package focus on HTTP handler behaviour
// (client_id replacement, error propagation) rather than repository routing. The
// correct agent is selected by configuring the stub with the expected agent.
// Storage-layer routing (fetching the correct agent by ID) is tested in the storage adapter tests.
type stubAgentRepo struct {
	agent *storage.Agent
	err   error
}

func newStubAgentRepo(agentID id.AgentID, upstreamClientID string) *stubAgentRepo {
	return &stubAgentRepo{
		agent: &storage.Agent{
			ID:       agentID,
			ClientID: id.ClientID(upstreamClientID),
		},
	}
}

func (r *stubAgentRepo) Get(_ context.Context, _ id.AgentID) (*storage.Agent, error) {
	return r.agent, r.err
}

func (r *stubAgentRepo) Create(_ context.Context, _ *storage.Agent) error { return nil }
func (r *stubAgentRepo) Update(_ context.Context, _ *storage.Agent) error { return nil }
func (r *stubAgentRepo) Delete(_ context.Context, _ id.AgentID) error     { return nil }
func (r *stubAgentRepo) List(_ context.Context) ([]*storage.Agent, error) { return nil, nil }
func (r *stubAgentRepo) GetByClientID(_ context.Context, _ id.ClientID) (*storage.Agent, error) {
	return nil, nil
}

// TestOAuth2TokenHandler_ServeHTTP_ContentTypeValidation tests Content-Type validation
func TestOAuth2TokenHandler_ServeHTTP_ContentTypeValidation(t *testing.T) {
	agentID := id.NewAgentID()
	agentRepo := newStubAgentRepo(agentID, "test-upstream-client")

	// Mock upstream server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "token123", "token_type": "Bearer"}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	tests := []struct {
		name           string
		contentType    string
		wantStatusCode int
	}{
		{
			name:           "valid content-type application/x-www-form-urlencoded",
			contentType:    "application/x-www-form-urlencoded",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid content-type application/json",
			contentType:    "application/json",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid content-type text/plain",
			contentType:    "text/plain",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := "grant_type=authorization_code&code=abc123&client_id=" + agentID.String()
			req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", strings.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
		})
	}
}

// TestOAuth2TokenHandler_ServeHTTP_HeaderFiltering tests hop-by-hop header filtering
func TestOAuth2TokenHandler_ServeHTTP_HeaderFiltering(t *testing.T) {
	agentID := id.NewAgentID()
	agentRepo := newStubAgentRepo(agentID, "test-upstream-client")

	// Mock upstream server that echoes back request headers
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Verify hop-by-hop headers were not forwarded
		for _, header := range []string{"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailers", "Transfer-Encoding", "Upgrade"} {
			if r.Header.Get(header) != "" {
				_, _ = w.Write([]byte(`{"error": "hop-by-hop header forwarded: ` + header + `"}`))
				return
			}
		}

		_, _ = w.Write([]byte(`{"access_token": "token123", "token_type": "Bearer"}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	reqBody := strings.NewReader("grant_type=authorization_code&code=abc123&client_id=" + agentID.String())
	req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", reqBody)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add hop-by-hop headers that should be filtered
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Keep-Alive", "timeout=5")
	req.Header.Set("Transfer-Encoding", "chunked")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	respBody, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(respBody), "access_token")
}

// TestOAuth2TokenHandler_ServeHTTP_SuccessfulProxy tests successful token request proxy
func TestOAuth2TokenHandler_ServeHTTP_SuccessfulProxy(t *testing.T) {
	agentID := id.NewAgentID()
	agentRepo := newStubAgentRepo(agentID, "test-upstream-client")

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "token123", "token_type": "Bearer", "expires_in": 3600}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	reqBody := strings.NewReader("grant_type=authorization_code&code=abc123&client_id=" + agentID.String() + "&redirect_uri=https://client.example.com/callback")
	req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", reqBody)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))

	respBody, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(respBody), "access_token")
	assert.Contains(t, string(respBody), "token123")
}

// TestOAuth2TokenHandler_ServeHTTP_UpstreamError tests upstream errors are proxied
func TestOAuth2TokenHandler_ServeHTTP_UpstreamError(t *testing.T) {
	agentID := id.NewAgentID()
	agentRepo := newStubAgentRepo(agentID, "test-upstream-client")

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid_grant", "error_description": "Authorization code expired"}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	reqBody := strings.NewReader("grant_type=authorization_code&code=expired&client_id=" + agentID.String())
	req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", reqBody)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Upstream error status is preserved
	assert.Equal(t, http.StatusBadRequest, w.Code)
	respBody, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(respBody), "invalid_grant")
}

// TestIsHopByHopHeader tests hop-by-hop header identification
func TestIsHopByHopHeader(t *testing.T) {
	tests := []struct {
		name       string
		headerName string
		isHopByHop bool
	}{
		// Hop-by-hop headers (RFC 7230)
		{"Connection", "Connection", true},
		{"Keep-Alive", "Keep-Alive", true},
		{"Proxy-Authenticate", "Proxy-Authenticate", true},
		{"Proxy-Authorization", "Proxy-Authorization", true},
		{"Te", "Te", true},
		{"Trailers", "Trailers", true},
		{"Transfer-Encoding", "Transfer-Encoding", true},
		{"Upgrade", "Upgrade", true},

		// Non-hop-by-hop headers
		{"Content-Type", "Content-Type", false},
		{"Content-Length", "Content-Length", false},
		{"Authorization", "Authorization", false},
		{"Accept", "Accept", false},
		{"User-Agent", "User-Agent", false},
		{"Cache-Control", "Cache-Control", false},
		{"Pragma", "Pragma", false},
		{"X-Custom-Header", "X-Custom-Header", false},

		// Case insensitivity
		{"connection (lowercase)", "connection", true},
		{"TRANSFER-ENCODING (uppercase)", "TRANSFER-ENCODING", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHopByHopHeader(tt.headerName)
			assert.Equal(t, tt.isHopByHop, result)
		})
	}
}

// TestOAuth2TokenHandler_ProxyToUpstream_MultiAgentVerifier tests claim verification
// for multi-agent client sharing (Feature 021 US1).
func TestOAuth2TokenHandler_ProxyToUpstream_MultiAgentVerifier(t *testing.T) {
	agentID := id.NewAgentID()
	agentRepo := newStubAgentRepo(agentID, "test-upstream-client")
	upstreamResponseBody := `{"access_token":"tok123","token_type":"Bearer"}`

	tests := []struct {
		name             string
		verifier         ports.MultiAgentVerifier
		clientID         string // form body client_id
		wantStatusCode   int
		wantBodyContains string
	}{
		{
			name:             "nil verifier (feature disabled) passes response through unchanged",
			verifier:         nil,
			clientID:         agentID.String(),
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "access_token",
		},
		{
			name: "verifier returns nil (claim matches) passes response through",
			verifier: &mockMultiAgentVerifier{verifyFn: func(_ context.Context, _ []byte, _ id.AgentID) error {
				return nil
			}},
			clientID:         agentID.String(),
			wantStatusCode:   http.StatusOK,
			wantBodyContains: "access_token",
		},
		{
			name: "verifier returns claim-absent error returns 500 server_error",
			verifier: &mockMultiAgentVerifier{verifyFn: func(_ context.Context, _ []byte, _ id.AgentID) error {
				return errors.New("agent ID claim absent from upstream token")
			}},
			clientID:         agentID.String(),
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "server_error",
		},
		{
			name: "verifier returns claim-mismatch error returns 500 server_error",
			verifier: &mockMultiAgentVerifier{verifyFn: func(_ context.Context, _ []byte, _ id.AgentID) error {
				return errors.New("agent ID claim mismatch")
			}},
			clientID:         agentID.String(),
			wantStatusCode:   http.StatusInternalServerError,
			wantBodyContains: "server_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(upstreamResponseBody))
			}))
			defer mockUpstream.Close()

			handler := &OAuth2TokenHandler{
				UpstreamTokenURL:   mockUpstream.URL,
				MultiAgentVerifier: tt.verifier,
				AgentRepository:    agentRepo,
			}

			body := "grant_type=authorization_code&code=abc123&client_id=" + tt.clientID
			req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatusCode, w.Code)
			respBody, _ := io.ReadAll(w.Body)
			assert.Contains(t, string(respBody), tt.wantBodyContains)
		})
	}
}

// TestOAuth2TokenHandler_ServeHTTP_ResponseStreaming tests response is streamed properly
func TestOAuth2TokenHandler_ServeHTTP_ResponseStreaming(t *testing.T) {
	agentID := id.NewAgentID()
	agentRepo := newStubAgentRepo(agentID, "test-upstream-client")

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Response-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
		// Return large response to test streaming
		_, _ = w.Write([]byte(`{"access_token": "verylongtoken123456789", "token_type": "Bearer", "expires_in": 3600, "scope": "openid profile email"}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	reqBody := strings.NewReader("grant_type=authorization_code&code=abc123&client_id=" + agentID.String())
	req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", reqBody)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Response-Header"))

	respBody, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(respBody), "access_token")
	assert.Contains(t, string(respBody), "verylongtoken123456789")
}

// TestOAuth2TokenHandler_ClientIDValidation tests that client_id is always validated as an
// agent UUID, regardless of whether MultiAgentVerifier is set.  The client_id is the
// broker-internal agent UUID from the perspective of every OAuth2 client (spec: Feature 021).
func TestOAuth2TokenHandler_ClientIDValidation(t *testing.T) {
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"Bearer"}`))
	}))
	defer mockUpstream.Close()

	verifiers := []struct {
		name     string
		verifier ports.MultiAgentVerifier
	}{
		{"without verifier", nil},
		{"with verifier", &mockMultiAgentVerifier{verifyFn: func(_ context.Context, _ []byte, _ id.AgentID) error { return nil }}},
	}

	tests := []struct {
		name           string
		body           string
		wantStatusCode int
		wantError      string
	}{
		{
			name:           "missing client_id returns 401 invalid_client",
			body:           "grant_type=authorization_code&code=abc123",
			wantStatusCode: http.StatusUnauthorized,
			wantError:      "invalid_client",
		},
		{
			name:           "non-UUID client_id returns 401 invalid_client",
			body:           "grant_type=authorization_code&code=abc123&client_id=not-a-uuid",
			wantStatusCode: http.StatusUnauthorized,
			wantError:      "invalid_client",
		},
		{
			name:           "legacy non-UUID client_id returns 401 invalid_client",
			body:           "grant_type=authorization_code&code=abc123&client_id=legacy-client-id",
			wantStatusCode: http.StatusUnauthorized,
			wantError:      "invalid_client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Behaviour must be identical regardless of whether MultiAgentVerifier is set.
			for _, v := range verifiers {
				t.Run(v.name, func(t *testing.T) {
					handler := &OAuth2TokenHandler{
						UpstreamTokenURL:   mockUpstream.URL,
						MultiAgentVerifier: v.verifier,
					}

					req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", strings.NewReader(tt.body))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					w := httptest.NewRecorder()

					handler.ServeHTTP(w, req)

					assert.Equal(t, tt.wantStatusCode, w.Code)
					respBody, _ := io.ReadAll(w.Body)
					assert.Contains(t, string(respBody), tt.wantError)
				})
			}
		})
	}
}

// TestOAuth2TokenHandler_ProxyToUpstream_ClientIDReplacement verifies that the broker
// replaces the agent UUID in the token request body with the agent's upstream client_id
// before forwarding to the upstream OAuth2 server (review comment r2995280734).
func TestOAuth2TokenHandler_ProxyToUpstream_ClientIDReplacement(t *testing.T) {
	agentID := id.NewAgentID()
	const upstreamClientID = "shared-upstream-oauth2-client"

	var receivedClientID string
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the client_id that upstream received
		if err := r.ParseForm(); err == nil {
			receivedClientID = r.FormValue("client_id")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"Bearer"}`))
	}))
	defer mockUpstream.Close()

	agentRepo := newStubAgentRepo(agentID, upstreamClientID)
	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	body := "grant_type=authorization_code&code=abc&client_id=" + agentID.String()
	req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Upstream must receive the configured upstream client_id, NOT the broker agent UUID.
	assert.Equal(t, upstreamClientID, receivedClientID)
	assert.NotEqual(t, agentID.String(), receivedClientID)
}

// TestOAuth2TokenHandler_ProxyToUpstream_AgentNotFound verifies that when the agent UUID
// is valid but the agent does not exist in the repository, the handler fails closed with
// 401 invalid_client without forwarding the request upstream (per SR-001).
func TestOAuth2TokenHandler_ProxyToUpstream_AgentNotFound(t *testing.T) {
	agentID := id.NewAgentID()

	upstreamCalled := false
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer mockUpstream.Close()

	// Repo returns an error for any agent lookup (agent not found)
	agentRepo := &stubAgentRepo{
		agent: nil,
		err:   errors.New("agent not found"),
	}
	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
		AgentRepository:  agentRepo,
	}

	body := "grant_type=authorization_code&code=abc&client_id=" + agentID.String()
	req := httptest.NewRequest("POST", "https://broker.example.com/oauth2/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	respBody, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(respBody), "invalid_client")
	assert.False(t, upstreamCalled, "upstream must not be called when agent is not found")
}

func TestWriteTokenResponse(t *testing.T) {
	t.Run("success returns 200 with complete JSON body", func(t *testing.T) {
		h := &OAuth2TokenHandler{}
		w := httptest.NewRecorder()

		h.writeTokenResponse(w, &ports.TokenResponse{
			AccessToken: "tok123",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))

		var body map[string]interface{}
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
		assert.Equal(t, "tok123", body["access_token"])
		assert.Equal(t, "Bearer", body["token_type"])
		assert.EqualValues(t, 3600, body["expires_in"])
		assert.NotContains(t, body, "scope")
	})

	t.Run("scope included when non-empty", func(t *testing.T) {
		h := &OAuth2TokenHandler{}
		w := httptest.NewRecorder()

		h.writeTokenResponse(w, &ports.TokenResponse{
			AccessToken: "tok456",
			TokenType:   "Bearer",
			ExpiresIn:   900,
			Scope:       "read write",
		})

		assert.Equal(t, http.StatusOK, w.Code)
		var body map[string]interface{}
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&body))
		assert.Equal(t, "read write", body["scope"])
	})
}

// TestHandleLocalMinting_ClientCredentials covers the client_credentials path through
// handleLocalMinting: success, input validation failures, and strategy error mapping.
func TestHandleLocalMinting_ClientCredentials(t *testing.T) {
	successResp := &ports.TokenResponse{AccessToken: "tok123", TokenType: "Bearer", ExpiresIn: 3600, Scope: "read"}

	tests := []struct {
		name          string
		body          string
		mintingErr    error // nil = return successResp
		wantStatus    int
		wantErrorCode string // empty = expect success
	}{
		{
			name:       "success returns 200 with all token fields",
			body:       "grant_type=client_credentials&client_id=550e8400-e29b-41d4-a716-446655440000&client_secret=secret&scope=read",
			wantStatus: http.StatusOK,
		},
		{
			name:          "missing client_id returns 400 invalid_request",
			body:          "grant_type=client_credentials&client_secret=secret",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_request",
		},
		{
			name:          "non-UUID client_id returns 401 invalid_client",
			body:          "grant_type=client_credentials&client_id=broker_abc&client_secret=secret",
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: "invalid_client",
		},
		{
			name:          "missing client_secret returns 400 invalid_request",
			body:          "grant_type=client_credentials&client_id=550e8400-e29b-41d4-a716-446655440000",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_request",
		},
		{
			name:          "strategy ErrInvalidClient returns 401 invalid_client",
			body:          "grant_type=client_credentials&client_id=550e8400-e29b-41d4-a716-446655440000&client_secret=wrong",
			mintingErr:    fosite.ErrInvalidClient,
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: "invalid_client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minting := &mockTokenMintingStrategy{
				clientCredentialsFn: func(_ context.Context, _ id.AgentID, _, _ string) (*ports.TokenResponse, error) {
					if tt.mintingErr != nil {
						return nil, tt.mintingErr
					}
					return successResp, nil
				},
			}
			handler := &OAuth2TokenHandler{TokenMinting: minting}
			req := httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			var body map[string]interface{}
			_ = json.NewDecoder(w.Body).Decode(&body)
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, body["error"])
			} else {
				assert.Equal(t, "tok123", body["access_token"])
				assert.Equal(t, "Bearer", body["token_type"])
				assert.EqualValues(t, 3600, body["expires_in"])
				assert.Equal(t, "read", body["scope"])
			}
		})
	}
}

// TestHandleLocalMinting_AuthorizationCode covers the authorization_code path through
// handleLocalMinting: success, input validation failures, and strategy error mapping.
func TestHandleLocalMinting_AuthorizationCode(t *testing.T) {
	successResp := &ports.TokenResponse{AccessToken: "tok456", TokenType: "Bearer", ExpiresIn: 900}

	tests := []struct {
		name          string
		body          string
		mintingErr    error
		wantStatus    int
		wantErrorCode string
	}{
		{
			name:       "success returns 200 with all token fields",
			body:       "grant_type=authorization_code&client_id=550e8400-e29b-41d4-a716-446655440000&client_secret=secret&code=authcode123&redirect_uri=https://example.com/cb&code_verifier=verifier",
			wantStatus: http.StatusOK,
		},
		{
			name:          "missing client_id returns 400 invalid_request",
			body:          "grant_type=authorization_code&client_secret=secret&code=abc",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_request",
		},
		{
			name:          "non-UUID client_id returns 401 invalid_client",
			body:          "grant_type=authorization_code&client_id=broker_abc&client_secret=secret&code=abc",
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: "invalid_client",
		},
		{
			name:          "missing client_secret returns 400 invalid_request",
			body:          "grant_type=authorization_code&client_id=550e8400-e29b-41d4-a716-446655440000&code=abc",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_request",
		},
		{
			name:          "missing code returns 400 invalid_request",
			body:          "grant_type=authorization_code&client_id=550e8400-e29b-41d4-a716-446655440000&client_secret=secret",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_request",
		},
		{
			name:          "strategy ErrInvalidGrant returns 400 invalid_grant",
			body:          "grant_type=authorization_code&client_id=550e8400-e29b-41d4-a716-446655440000&client_secret=secret&code=expired",
			mintingErr:    fosite.ErrInvalidGrant,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "invalid_grant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minting := &mockTokenMintingStrategy{
				authorizationCodeExchangeFn: func(_ context.Context, _ id.AgentID, _, _, _, _ string) (*ports.TokenResponse, error) {
					if tt.mintingErr != nil {
						return nil, tt.mintingErr
					}
					return successResp, nil
				},
			}
			handler := &OAuth2TokenHandler{TokenMinting: minting}
			req := httptest.NewRequest("POST", "/oauth2/token", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			var body map[string]interface{}
			_ = json.NewDecoder(w.Body).Decode(&body)
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, body["error"])
			} else {
				assert.Equal(t, "tok456", body["access_token"])
				assert.Equal(t, "Bearer", body["token_type"])
				assert.EqualValues(t, 900, body["expires_in"])
			}
		})
	}
}

// TestHandleLocalMinting_UnsupportedGrantType verifies the default branch returns
// 400 unsupported_grant_type for any grant type other than client_credentials or
// authorization_code (e.g. password, implicit, device_code).
func TestHandleLocalMinting_UnsupportedGrantType(t *testing.T) {
	handler := &OAuth2TokenHandler{TokenMinting: fixedMinting(nil, nil)}
	req := httptest.NewRequest("POST", "/oauth2/token",
		strings.NewReader("grant_type=password&username=user&password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var body map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&body)
	assert.Equal(t, "unsupported_grant_type", body["error"])
}

// TestHandleMintingError_RFC6749StatusCodes verifies the complete RFC 6749 error code →
// HTTP status mapping in handleMintingError.
func TestHandleMintingError_RFC6749StatusCodes(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantStatus    int
		wantErrorCode string
	}{
		{"ErrInvalidClient → 401 invalid_client", fosite.ErrInvalidClient, http.StatusUnauthorized, "invalid_client"},
		{"ErrInvalidScope → 400 invalid_scope", fosite.ErrInvalidScope, http.StatusBadRequest, "invalid_scope"},
		{"ErrInvalidGrant → 400 invalid_grant", fosite.ErrInvalidGrant, http.StatusBadRequest, "invalid_grant"},
		{"unknown error → 500 server_error", errors.New("unexpected db failure"), http.StatusInternalServerError, "server_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &OAuth2TokenHandler{}
			w := httptest.NewRecorder()

			h.handleMintingError(w, tt.err, "client_credentials", "broker_test")

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			var body map[string]interface{}
			_ = json.NewDecoder(w.Body).Decode(&body)
			assert.Equal(t, tt.wantErrorCode, body["error"])
		})
	}
}

func TestHandleMintingError_OpaqueDescriptions(t *testing.T) {
	internalDetail := "scope \"read:admin\" not allowed for this agent"

	t.Run("invalid_scope does not leak internal detail", func(t *testing.T) {
		h := &OAuth2TokenHandler{}
		w := httptest.NewRecorder()

		h.handleMintingError(w, fmt.Errorf("%s: %w", internalDetail, fosite.ErrInvalidScope), "client_credentials", "broker_test")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "invalid_scope")
		assert.NotContains(t, string(body), internalDetail)
		assert.NotContains(t, string(body), "read:admin")
	})

	t.Run("invalid_grant does not leak internal detail", func(t *testing.T) {
		h := &OAuth2TokenHandler{}
		w := httptest.NewRecorder()

		h.handleMintingError(w, fmt.Errorf("%s: %w", internalDetail, fosite.ErrInvalidGrant), "authorization_code", "broker_test")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		body, _ := io.ReadAll(w.Body)
		assert.Contains(t, string(body), "invalid_grant")
		assert.NotContains(t, string(body), internalDetail)
		assert.NotContains(t, string(body), "read:admin")
	})
}
