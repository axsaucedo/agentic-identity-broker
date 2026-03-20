package enduser

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
)

// mockMultiAgentVerifier is a test double for MultiAgentVerifier
type mockMultiAgentVerifier struct {
	verifyFn func(ctx context.Context, responseBody []byte, expectedAgentID id.AgentID) error
}

func (m *mockMultiAgentVerifier) VerifyAgentIDClaim(ctx context.Context, responseBody []byte, expectedAgentID id.AgentID) error {
	return m.verifyFn(ctx, responseBody, expectedAgentID)
}

// TestOAuth2TokenHandler_ServeHTTP_ContentTypeValidation tests Content-Type validation
func TestOAuth2TokenHandler_ServeHTTP_ContentTypeValidation(t *testing.T) {
	agentID := id.NewAgentID()

	// Mock upstream server
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token": "token123", "token_type": "Bearer"}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
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

	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid_grant", "error_description": "Authorization code expired"}`))
	}))
	defer mockUpstream.Close()

	handler := &OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
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
	upstreamResponseBody := `{"access_token":"tok123","token_type":"Bearer"}`

	tests := []struct {
		name             string
		verifier         MultiAgentVerifier
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
		verifier MultiAgentVerifier
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
