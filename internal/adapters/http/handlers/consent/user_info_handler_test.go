package consent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetUserInfo_Success tests successful retrieval of user info.
func TestGetUserInfo_Success(t *testing.T) {
	principalValue := "user@example.com"

	handler := NewUserInfoHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetUserInfo(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp GetUserInfoResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, principalValue, resp.Data.Principal)
	assert.Equal(t, principalValue, resp.Data.DisplayName) // For now, display name = principal
	assert.Nil(t, resp.Data.PictureURL)
}

// TestGetUserInfo_DifferentPrincipals tests with different principal values.
func TestGetUserInfo_DifferentPrincipals(t *testing.T) {
	testCases := []struct {
		name      string
		principal string
	}{
		{
			name:      "email principal",
			principal: "user@example.com",
		},
		{
			name:      "UUID principal",
			principal: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "username principal",
			principal: "john.doe",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewUserInfoHandler(nil)

			req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
			ctx := principal.WithPrincipal(req.Context(), tc.principal)
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			handler.GetUserInfo(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)

			var resp GetUserInfoResponse
			err := json.NewDecoder(rec.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tc.principal, resp.Data.Principal)
			assert.Equal(t, tc.principal, resp.Data.DisplayName)
		})
	}
}

// TestGetUserInfo_MissingPrincipal tests error when principal is not in context.
func TestGetUserInfo_MissingPrincipal(t *testing.T) {
	handler := NewUserInfoHandler(nil)

	// Request without principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)

	rec := httptest.NewRecorder()
	handler.GetUserInfo(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error)
	assert.Equal(t, "authentication required", resp.Message)
}

// TestGetUserInfo_EmptyPrincipal tests error when principal is empty string.
func TestGetUserInfo_EmptyPrincipal(t *testing.T) {
	handler := NewUserInfoHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	// Add empty principal to context (simulating a bug or misconfiguration)
	ctx := principal.WithPrincipal(req.Context(), "")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetUserInfo(rec, req)

	// Should return unauthorized since empty principal is treated as missing
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TestGetUserInfo_ContentType tests that response has correct content type.
func TestGetUserInfo_ContentType(t *testing.T) {
	principalValue := "user@example.com"

	handler := NewUserInfoHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetUserInfo(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

// TestGetUserInfo_ResponseStructure tests the response structure.
func TestGetUserInfo_ResponseStructure(t *testing.T) {
	principalValue := "test@example.com"

	handler := NewUserInfoHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetUserInfo(rec, req)

	var resp GetUserInfoResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify response envelope structure
	assert.NotNil(t, resp.Data)

	// Verify required fields
	assert.NotEmpty(t, resp.Data.Principal)
	assert.NotEmpty(t, resp.Data.DisplayName)

	// Optional field should be nil for now
	assert.Nil(t, resp.Data.PictureURL)
}
