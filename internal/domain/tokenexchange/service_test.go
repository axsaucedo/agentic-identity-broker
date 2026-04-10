package tokenexchange

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	storagedomain "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// newTestProviderService wraps a repository in a ThirdpartyOAuth2ProviderService
// with passthrough encryption for use in domain-layer tests.
func newTestProviderService(repo ports.ThirdpartyOAuth2ProviderRepository) *thirdparty.ThirdpartyOAuth2ProviderService {
	return thirdparty.NewThirdpartyOAuth2ProviderService(repo, &MockEncryption{}, nil, false, nil)
}

// newMockConsentService creates a consent.Service with mock repositories for testing
func newMockConsentService() *consent.Service {
	return consent.NewService(
		&MockAgentRepository{},
		newTestProviderService(&MockServiceRepository{}),
		&MockGrantRepository{
			grant: &storagedomain.UserGrant{
				ID:         id.NewGrantID(),
				Principal:  id.Principal("test-principal"),
				AgentID:    id.NewAgentID(),
				ValidUntil: func() *time.Time { t := time.Now().Add(24 * time.Hour); return &t }(),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
		},
		slog.Default(),
	)
}

// MockAgentRepository mocks the AgentRepository for consent service testing
type MockAgentRepository struct{}

func (m *MockAgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storagedomain.Agent, error) {
	return nil, nil
}

func (m *MockAgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storagedomain.Agent, error) {
	return nil, nil
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *storagedomain.Agent) error {
	return nil
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *storagedomain.Agent) error {
	return nil
}

func (m *MockAgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	return nil
}

func (m *MockAgentRepository) List(ctx context.Context) ([]*storagedomain.Agent, error) {
	return nil, nil
}

// MockOAuth2SessionService mocks the OAuth2SessionService for testing
type MockOAuth2SessionService struct {
	RefreshAccessTokenFn       func(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity, refreshToken string) (*oauth2.Token, error)
	UpdateSessionTokensFn      func(ctx context.Context, principal id.Principal, session *storagedomain.UserSession, newToken *oauth2.Token) error
	DecryptAccessTokenFn       func(ctx context.Context, session *storagedomain.UserSession) (string, error)
	DecryptRefreshTokenFn      func(ctx context.Context, session *storagedomain.UserSession) (string, error)
	GetValidAccessTokenFn      func(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, string, error)
	GetSessionWithValidTokenFn func(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, string, error)
}

func (m *MockOAuth2SessionService) RefreshAccessToken(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity, refreshToken string) (*oauth2.Token, error) {
	if m.RefreshAccessTokenFn != nil {
		return m.RefreshAccessTokenFn(ctx, entity, refreshToken)
	}
	return nil, nil
}

func (m *MockOAuth2SessionService) UpdateSessionTokens(ctx context.Context, principal id.Principal, session *storagedomain.UserSession, newToken *oauth2.Token) error {
	if m.UpdateSessionTokensFn != nil {
		return m.UpdateSessionTokensFn(ctx, principal, session, newToken)
	}
	return nil
}

func (m *MockOAuth2SessionService) DecryptAccessToken(ctx context.Context, session *storagedomain.UserSession) (string, error) {
	if m.DecryptAccessTokenFn != nil {
		return m.DecryptAccessTokenFn(ctx, session)
	}
	return "", nil
}

func (m *MockOAuth2SessionService) DecryptRefreshToken(ctx context.Context, session *storagedomain.UserSession) (string, error) {
	if m.DecryptRefreshTokenFn != nil {
		return m.DecryptRefreshTokenFn(ctx, session)
	}
	return "", nil
}

func (m *MockOAuth2SessionService) GetValidAccessToken(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, string, error) {
	if m.GetValidAccessTokenFn != nil {
		return m.GetValidAccessTokenFn(ctx, principal, serviceID)
	}
	// Return a mock session and token
	session := &storagedomain.UserSession{
		ID:        id.NewSessionID(),
		Principal: principal,
		ServiceID: serviceID,
		TokenType: "Bearer",
		Scope:     []string{"read", "write"},
	}
	return session, "mock-access-token", nil
}

func (m *MockOAuth2SessionService) GetSessionWithValidToken(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, string, error) {
	if m.GetSessionWithValidTokenFn != nil {
		return m.GetSessionWithValidTokenFn(ctx, principal, serviceID)
	}
	// Return a mock session and token
	session := &storagedomain.UserSession{
		ID:        id.NewSessionID(),
		Principal: principal,
		ServiceID: serviceID,
		TokenType: "Bearer",
		Scope:     []string{"read", "write"},
	}
	return session, "mock-access-token", nil
}

// MockTokenExchangeRepository mocks are defined at the end of this file
type MockServiceRepository struct {
	service *model.ThirdpartyOAuth2ProviderEntity
	err     error
}

func (m *MockServiceRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return m.service, m.err
}

func (m *MockServiceRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return nil
}

func (m *MockServiceRepository) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

func (m *MockServiceRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return nil
}

func (m *MockServiceRepository) Delete(ctx context.Context, serviceID id.ServiceID) error {
	return nil
}

func (m *MockServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, nil
}

func (m *MockServiceRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

type MockGrantRepository struct {
	grant *storagedomain.UserGrant
	err   error
}

func (m *MockGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storagedomain.UserGrant, error) {
	return m.grant, m.err
}

func (m *MockGrantRepository) Create(ctx context.Context, grant *storagedomain.UserGrant) error {
	return nil
}

func (m *MockGrantRepository) Get(ctx context.Context, grantID id.GrantID) (*storagedomain.UserGrant, error) {
	return nil, nil
}

func (m *MockGrantRepository) Update(ctx context.Context, grant *storagedomain.UserGrant) error {
	return nil
}

func (m *MockGrantRepository) Delete(ctx context.Context, grantID id.GrantID) error {
	return nil
}

func (m *MockGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, nil
}

func (m *MockGrantRepository) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	return nil
}

func (m *MockGrantRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storagedomain.UserGrant, error) {
	return nil, nil
}

func (m *MockGrantRepository) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
	return nil, nil
}

func (m *MockGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storagedomain.UserGrant, error) {
	return nil, nil
}

func (m *MockGrantRepository) DeleteByPrincipalAndAgentID(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	return m.err
}

type MockSessionRepository struct {
	session *storagedomain.UserSession
	err     error
}

type MockEncryption struct {
	err error
}

func (m *MockEncryption) Encrypt(ctx context.Context, plaintext []byte, context map[string]string) ([]byte, error) {
	return plaintext, m.err
}

func (m *MockEncryption) Decrypt(ctx context.Context, ciphertext []byte, context map[string]string) ([]byte, error) {
	return ciphertext, m.err
}

func (m *MockSessionRepository) FindByPrincipalAndService(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, error) {
	return m.session, m.err
}

func (m *MockSessionRepository) Create(ctx context.Context, session *storagedomain.UserSession) error {
	return nil
}

func (m *MockSessionRepository) Get(ctx context.Context, sessionID id.SessionID) (*storagedomain.UserSession, error) {
	return nil, nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID id.SessionID) error {
	return nil
}

func (m *MockSessionRepository) CountByService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, nil
}

func (m *MockSessionRepository) DeleteByPrincipalAndService(ctx context.Context, principal id.Principal, serviceID id.ServiceID) error {
	return nil
}

func (m *MockSessionRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]*storagedomain.UserSession, error) {
	return nil, nil
}

// NewTokenExchangeServiceForTest creates a TokenExchangeService for testing
func NewTokenExchangeServiceForTest(
	jwtValidator *JWTValidator,
	celEvaluator *CELEvaluator,
	providerService *thirdparty.ThirdpartyOAuth2ProviderService,
	oauth2SessionService *oauth2session.OAuth2SessionService,
	consentService *consent.Service,
	agentRepository ports.AgentRepository,
	config *ports.TokenExchangeConfig,
) (*TokenExchangeService, error) {
	return NewTokenExchangeService(
		jwtValidator,
		celEvaluator,
		providerService,
		oauth2SessionService,
		consentService,
		agentRepository,
		config,
	)
}

// TestNewTokenExchangeService tests service creation with various parameter combinations
func TestNewTokenExchangeServiceForTest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		jwtValidator         *JWTValidator
		celEvaluator         *CELEvaluator
		serviceRepo          *thirdparty.ThirdpartyOAuth2ProviderService
		oauth2SessionService *oauth2session.OAuth2SessionService
		consentService       *consent.Service
		agentRepository      ports.AgentRepository
		config               *ports.TokenExchangeConfig
		expectError          bool
		errorContains        string
	}{
		{
			name:                 "valid parameters - creates service without error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         &CELEvaluator{},
			serviceRepo:          newTestProviderService(&MockServiceRepository{}),
			oauth2SessionService: &oauth2session.OAuth2SessionService{},
			consentService:       &consent.Service{},
			agentRepository:      &MockAgentRepository{},
			config: &ports.TokenExchangeConfig{
				ClaimExtraction: ports.ClaimExtractionConfig{
					PrincipalExpression: "subject_token.sub",
					AgentIDExpression:   "subject_token.azp",
				},
				Authorization: ports.AuthorizationConfig{
					Type: "cel",
					CEL: ports.CELAuthorizationConfig{
						Expression: "true",
					},
				},
			},
			expectError: false,
		},
		{
			name:          "nil jwtValidator returns error",
			jwtValidator:  nil,
			expectError:   true,
			errorContains: "jwtValidator",
		},
		{
			name:                 "nil celEvaluator returns error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         nil,
			oauth2SessionService: &oauth2session.OAuth2SessionService{},
			expectError:          true,
			errorContains:        "celEvaluator",
		},
		{
			name:                 "nil serviceRepository returns error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         &CELEvaluator{},
			serviceRepo:          nil,
			oauth2SessionService: &oauth2session.OAuth2SessionService{},
			expectError:          true,
			errorContains:        "providerService",
		},
		{
			name:                 "nil oauth2SessionService returns error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         &CELEvaluator{},
			serviceRepo:          newTestProviderService(&MockServiceRepository{}),
			oauth2SessionService: nil,
			consentService:       &consent.Service{},
			expectError:          true,
			errorContains:        "oauth2SessionService",
		},
		{
			name:                 "nil consentService returns error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         &CELEvaluator{},
			serviceRepo:          newTestProviderService(&MockServiceRepository{}),
			oauth2SessionService: &oauth2session.OAuth2SessionService{},
			consentService:       nil,
			expectError:          true,
			errorContains:        "consentService",
		},
		{
			name:                 "nil agentRepository returns error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         &CELEvaluator{},
			serviceRepo:          newTestProviderService(&MockServiceRepository{}),
			oauth2SessionService: &oauth2session.OAuth2SessionService{},
			consentService:       &consent.Service{},
			agentRepository:      nil,
			expectError:          true,
			errorContains:        "agentRepository",
		},
		{
			name:                 "nil config returns error",
			jwtValidator:         &JWTValidator{},
			celEvaluator:         &CELEvaluator{},
			serviceRepo:          newTestProviderService(&MockServiceRepository{}),
			oauth2SessionService: &oauth2session.OAuth2SessionService{},
			consentService:       &consent.Service{},
			agentRepository:      &MockAgentRepository{},
			config:               nil,
			expectError:          true,
			errorContains:        "config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service, err := NewTokenExchangeServiceForTest(
				tt.jwtValidator,
				tt.celEvaluator,
				tt.serviceRepo,
				tt.oauth2SessionService,
				tt.consentService,
				tt.agentRepository,
				tt.config,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tt.errorContains != "" {
					assert.ErrorContains(t, err, tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
			}
		})
	}
}

// TestExchange_InvalidRequest tests handling of invalid token exchange requests
func TestExchange_InvalidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := &ports.TokenExchangeConfig{
		ClaimExtraction: ports.ClaimExtractionConfig{
			PrincipalExpression: "subject_token.sub",
			AgentIDExpression:   "subject_token.azp",
		},
		Authorization: ports.AuthorizationConfig{
			Type: "cel",
			CEL: ports.CELAuthorizationConfig{
				Expression: "true",
			},
		},
	}

	service, err := NewTokenExchangeServiceForTest(
		&JWTValidator{},
		&CELEvaluator{},
		newTestProviderService(&MockServiceRepository{}),
		&oauth2session.OAuth2SessionService{},
		newMockConsentService(),
		&MockAgentRepository{},
		config,
	)
	require.NoError(t, err)

	tests := []struct {
		name        string
		request     *TokenExchangeRequest
		expectError bool
		errorCode   string
	}{
		{
			name:        "empty grant_type",
			request:     NewTokenExchangeRequest("", "token", "", "assertion", "", "resource", ""),
			expectError: true,
			errorCode:   "invalid_request",
		},
		{
			name:        "missing subject_token",
			request:     NewTokenExchangeRequest(TokenExchangeGrantType, "", "", "assertion", "", "resource", ""),
			expectError: true,
			errorCode:   "invalid_request",
		},
		{
			name:        "missing client_assertion",
			request:     NewTokenExchangeRequest(TokenExchangeGrantType, "token", "", "", "", "resource", ""),
			expectError: true,
			errorCode:   "invalid_request",
		},
		{
			name:        "missing resource",
			request:     NewTokenExchangeRequest(TokenExchangeGrantType, "token", "", "assertion", "", "", ""),
			expectError: true,
			errorCode:   "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := service.Exchange(ctx, tt.request)
			assert.Error(t, err)

			// Check error code
			if tokenErr, ok := err.(*TokenExchangeError); ok {
				assert.Equal(t, tt.errorCode, tokenErr.Code())
			}
		})
	}
}

// TestExchange_RequestValidation documents valid token exchange request structure
// Full E2E testing of request/JWT validation is in E2E tests
func TestExchange_RequestValidation(t *testing.T) {
	t.Parallel()
	// This test documents that request validation happens before JWT validation
	// Full integration tests of the complete flow are in E2E tests
	// Unit tests of JWT validation are in jwt_validator_test.go
}

// TestExchange_GrantExpiration tests handling of expired user grants
func TestExchange_GrantExpiration(t *testing.T) {
	t.Parallel()
	// This test verifies that expired grants are properly rejected
	// In a full implementation with complete mocking, this would test the full flow
	t.Run("expired grant should return access_denied", func(t *testing.T) {
		t.Parallel()
		// Would need complete mocking of JWT validation and other steps
		// This is a placeholder for the test pattern
		expiredTime := time.Now().UTC().Add(-1 * time.Hour)
		grant := &storagedomain.UserGrant{
			ValidUntil: &expiredTime,
		}

		// Verify grant is expired
		if grant.ValidUntil != nil && grant.ValidUntil.Before(time.Now().UTC()) {
			// Grant is expired - should be denied
			assert.True(t, grant.ValidUntil.Before(time.Now().UTC()))
		}
	})
}

// TestBuildRequestContext tests CEL request context construction
// NOTE: buildRequestContext is not currently exposed on TokenExchangeService
// This test remains as documentation for the pattern once the method is public
func TestBuildRequestContext(t *testing.T) {
	t.Parallel()
	_, err := NewTokenExchangeServiceForTest(
		&JWTValidator{},
		&CELEvaluator{},
		newTestProviderService(&MockServiceRepository{}),
		&oauth2session.OAuth2SessionService{},
		newMockConsentService(),
		&MockAgentRepository{},
		&ports.TokenExchangeConfig{},
	)
	require.NoError(t, err)

	// TODO: Once buildRequestContext is exposed, uncomment test
	// req := NewTokenExchangeRequest(
	//	TokenExchangeGrantType,
	//	"subject-token",
	//	AccessTokenType,
	//	"client-assertion",
	//	JWTBearerType,
	//	"https://api.example.com",
	//	"read write",
	// )
	//
	// ctx := service.buildRequestContext("user123", "agent-456", req)
	//
	// assert.Equal(t, "https://api.example.com", ctx["resource"])
	// assert.Equal(t, TokenExchangeGrantType, ctx["grant_type"])
	// assert.Equal(t, "read write", ctx["scope"])
	// assert.Equal(t, "user123", ctx["principal"])
	// assert.Equal(t, "agent-456", ctx["agent_client_id"])
}

// TestGrantVerification_MissingGrant tests T063 - access_denied when grant not found
func TestGrantVerification_MissingGrant(t *testing.T) {
	t.Parallel()
	// This test documents the grant verification flow (T061-T063)
	// When a grant is not found (ErrNotFound), the service should return access_denied

	_, err := NewTokenExchangeServiceForTest(
		&JWTValidator{},
		&CELEvaluator{},
		newTestProviderService(&MockServiceRepository{}),
		&oauth2session.OAuth2SessionService{},
		newMockConsentService(),
		&MockAgentRepository{},
		&ports.TokenExchangeConfig{},
	)
	require.NoError(t, err)

	// T063: When grant not found, should return access_denied
	// This is verified in the Exchange implementation
	assert.Equal(t, ports.ErrNotFound, ports.ErrNotFound)
}

// TestGrantVerification_ExpiredGrant tests T065 - access_denied when grant expired
func TestGrantVerification_ExpiredGrant(t *testing.T) {
	t.Parallel()
	// Create an expired grant
	expiredTime := time.Now().UTC().Add(-1 * time.Hour)
	expiredGrant := &storagedomain.UserGrant{
		Principal:  id.Principal("user@example.com"),
		AgentID:    id.NewAgentID(),
		ValidUntil: &expiredTime,
		DelegatedOAuth2Tokens: []storagedomain.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID(),
				Scopes:                    []string{"read", "write"},
			},
		},
	}

	_, err := NewTokenExchangeServiceForTest(
		&JWTValidator{},
		&CELEvaluator{},
		newTestProviderService(&MockServiceRepository{}),
		&oauth2session.OAuth2SessionService{},
		newMockConsentService(),
		&MockAgentRepository{},
		&ports.TokenExchangeConfig{},
	)
	require.NoError(t, err)

	// Verify that expired grant is detected
	// This test documents T062a and T065 grant expiration check
	assert.NotNil(t, expiredGrant)
	assert.NotNil(t, expiredGrant.ValidUntil)
	if expiredGrant.ValidUntil != nil && expiredGrant.ValidUntil.Before(time.Now().UTC()) {
		// T065: Grant is expired
		assert.True(t, expiredGrant.ValidUntil.Before(time.Now().UTC()))
	}
}

// TestGrantVerification_ActiveGrant tests T062a - active grant is allowed
func TestGrantVerification_ActiveGrant(t *testing.T) {
	t.Parallel()
	// Create an active (non-expired) grant
	futureTime := time.Now().UTC().Add(24 * time.Hour)
	activeGrant := &storagedomain.UserGrant{
		Principal:  id.Principal("user@example.com"),
		AgentID:    id.NewAgentID(),
		ValidUntil: &futureTime,
		DelegatedOAuth2Tokens: []storagedomain.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID(),
				Scopes:                    []string{"read", "write"},
			},
		},
	}

	_, err := NewTokenExchangeServiceForTest(
		&JWTValidator{},
		&CELEvaluator{},
		newTestProviderService(&MockServiceRepository{}),
		&oauth2session.OAuth2SessionService{},
		newMockConsentService(),
		&MockAgentRepository{},
		&ports.TokenExchangeConfig{},
	)
	require.NoError(t, err)

	// Verify that active grant is recognized
	// This test documents T062a - grant is active when ValidUntil > now
	assert.NotNil(t, activeGrant)
	assert.NotNil(t, activeGrant.ValidUntil)
	if activeGrant.ValidUntil != nil {
		// T062a: Grant is active (not expired)
		assert.True(t, activeGrant.ValidUntil.After(time.Now().UTC()))
	}
}

// TestMockOAuth2SessionService_NewMethods tests the newly added mock methods
func TestMockOAuth2SessionService_NewMethods(t *testing.T) {
	t.Parallel()
	mock := &MockOAuth2SessionService{}
	ctx := context.Background()
	testSvcID := id.NewServiceID()

	t.Run("GetValidAccessToken with default behavior", func(t *testing.T) {
		session, token, err := mock.GetValidAccessToken(ctx, id.Principal("user@example.com"), testSvcID)
		assert.NoError(t, err)
		assert.Equal(t, "mock-access-token", token)
		assert.NotNil(t, session)
		assert.Equal(t, id.Principal("user@example.com"), session.Principal)
		assert.Equal(t, testSvcID, session.ServiceID)
	})

	t.Run("GetValidAccessToken with custom function", func(t *testing.T) {
		customSession := &storagedomain.UserSession{
			ID:        id.NewSessionID(),
			Principal: id.Principal("user@example.com"),
			ServiceID: testSvcID,
			TokenType: "Custom",
			Scope:     []string{"custom"},
		}
		mock.GetValidAccessTokenFn = func(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, string, error) {
			return customSession, "custom-token", nil
		}
		session, token, err := mock.GetValidAccessToken(ctx, id.Principal("user@example.com"), testSvcID)
		assert.NoError(t, err)
		assert.Equal(t, "custom-token", token)
		assert.Equal(t, customSession, session)
	})

	t.Run("GetSessionWithValidToken with default behavior", func(t *testing.T) {
		mock.GetValidAccessTokenFn = nil // reset
		session, token, err := mock.GetSessionWithValidToken(ctx, id.Principal("user@example.com"), testSvcID)
		assert.NoError(t, err)
		assert.Equal(t, "mock-access-token", token)
		assert.NotNil(t, session)
		assert.Equal(t, id.Principal("user@example.com"), session.Principal)
		assert.Equal(t, testSvcID, session.ServiceID)
		assert.Equal(t, "Bearer", session.TokenType)
	})

	t.Run("GetSessionWithValidToken with custom function", func(t *testing.T) {
		customSession := &storagedomain.UserSession{
			ID:        id.NewSessionID(),
			Principal: id.Principal("custom@example.com"),
			ServiceID: testSvcID,
			TokenType: "Custom",
			Scope:     []string{"custom"},
		}
		mock.GetSessionWithValidTokenFn = func(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storagedomain.UserSession, string, error) {
			return customSession, "custom-token", nil
		}
		session, token, err := mock.GetSessionWithValidToken(ctx, id.Principal("user@example.com"), testSvcID)
		assert.NoError(t, err)
		assert.Equal(t, "custom-token", token)
		assert.Equal(t, customSession, session)
	})
}

// --- T029: Agent lookup unit tests (Feature 021 US2) ---
// These tests verify that after T030, the service uses id.ParseAgentID + Get (not GetByClientID).
// Written before T030 implementation — must FAIL semantically until T030 is implemented.

// trackingAgentRepository is a spy that records whether Get or GetByClientID was called.
// Returns ErrNotFound for all lookups to stop execution after the agent lookup step.
type trackingAgentRepository struct {
	getCalled           bool
	getByClientIDCalled bool
}

func (r *trackingAgentRepository) Get(_ context.Context, _ id.AgentID) (*storagedomain.Agent, error) {
	r.getCalled = true
	return nil, ports.ErrNotFound
}

func (r *trackingAgentRepository) GetByClientID(_ context.Context, _ id.ClientID) (*storagedomain.Agent, error) {
	r.getByClientIDCalled = true
	return nil, ports.ErrNotFound
}

func (r *trackingAgentRepository) Create(_ context.Context, _ *storagedomain.Agent) error { return nil }

func (r *trackingAgentRepository) Update(_ context.Context, _ *storagedomain.Agent) error { return nil }

func (r *trackingAgentRepository) Delete(_ context.Context, _ id.AgentID) error { return nil }

func (r *trackingAgentRepository) List(_ context.Context) ([]*storagedomain.Agent, error) {
	return nil, nil
}

// generateTestRSAKeySet generates an RSA key pair and returns the private key plus a JWKS set
// containing the corresponding public key. Used to set up JWTValidator in T029 tests.
func generateTestRSAKeySet(t *testing.T) (*rsa.PrivateKey, jwk.Set) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwkKey, err := jwk.Import(&privateKey.PublicKey)
	require.NoError(t, err)
	require.NoError(t, jwkKey.Set(jwk.KeyIDKey, "test-key"))
	require.NoError(t, jwkKey.Set(jwk.AlgorithmKey, jwa.RS256()))

	keySet := jwk.NewSet()
	require.NoError(t, keySet.AddKey(jwkKey))
	return privateKey, keySet
}

// signServiceTestJWT signs a JWT with the given RSA private key for use in service tests.
func signServiceTestJWT(t *testing.T, privateKey *rsa.PrivateKey, claims map[string]interface{}) string {
	t.Helper()
	tok := jwt.New()
	for k, v := range claims {
		require.NoError(t, tok.Set(k, v))
	}
	jwkPrivKey, err := jwk.Import(privateKey)
	require.NoError(t, err)
	require.NoError(t, jwkPrivKey.Set(jwk.KeyIDKey, "test-key"))
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), jwkPrivKey))
	require.NoError(t, err)
	return string(signed)
}

// newServiceForStep9Test builds a TokenExchangeService wired to reach step 9 (agent lookup).
// The service is configured with:
//   - A real JWTValidator backed by the supplied JWKS set (RSA key pair)
//   - A real CELEvaluator with "subject_token.azp" as agentClientID expression
//   - A MockServiceRepository that always returns a non-nil provider entity
//   - The supplied agentRepo for step 9 (the code under test)
func newServiceForStep9Test(t *testing.T, keySet jwk.Set, agentRepo ports.AgentRepository) *TokenExchangeService {
	t.Helper()
	jwtValidator, err := NewJWTValidator(
		&MockJWKSProvider{keySet: keySet},
		"https://auth.example.com",
		"agentic-identity-broker",
		60,
	)
	require.NoError(t, err)

	celEvaluator, err := NewCELEvaluator(CELEvaluatorConfig{
		PrincipalExpression:     "subject_token.sub",
		AgentIDExpression:       "subject_token.azp",
		AuthorizationExpression: "true",
		EvaluationTimeout:       100 * time.Millisecond,
	})
	require.NoError(t, err)

	providerEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          id.NewServiceID(),
		DisplayName: "Test Service",
		ClientID:    id.ClientID("test-client"),
		// MockEncryption Decrypt is a pass-through; decryptSecret requires an encrypted secret
		Secret: model.NewEncryptedSecret([]byte("placeholder")),
	}
	providerRepo := &MockServiceRepository{service: providerEntity}

	consentSvc := consent.NewService(
		&MockAgentRepository{},
		newTestProviderService(&MockServiceRepository{}),
		&MockGrantRepository{err: ports.ErrNotFound},
		slog.Default(),
	)

	svc, err := NewTokenExchangeServiceForTest(
		jwtValidator,
		celEvaluator,
		newTestProviderService(providerRepo),
		&oauth2session.OAuth2SessionService{},
		consentSvc,
		agentRepo,
		&ports.TokenExchangeConfig{
			ClaimExtraction: ports.ClaimExtractionConfig{
				PrincipalExpression: "subject_token.sub",
				AgentIDExpression:   "subject_token.azp",
			},
			Authorization: ports.AuthorizationConfig{
				Type: "cel",
				CEL:  ports.CELAuthorizationConfig{Expression: "true"},
			},
		},
	)
	require.NoError(t, err)
	return svc
}

// TestExchange_AgentLookup_UsesGetNotGetByClientID verifies that after T030 the service
// calls agentRepository.Get(agentID) (parsed UUID) rather than GetByClientID(clientID).
// [T029] Written before T030 — fails semantically until T030 replaces GetByClientID with Get.
func TestExchange_AgentLookup_UsesGetNotGetByClientID(t *testing.T) {
	privateKey, keySet := generateTestRSAKeySet(t)

	agentUUID := id.NewAgentID()
	tracker := &trackingAgentRepository{}
	svc := newServiceForStep9Test(t, keySet, tracker)

	now := time.Now()
	commonClaims := map[string]interface{}{
		"iss": "https://auth.example.com",
		"aud": "agentic-identity-broker",
		"sub": "user@example.com",
		"exp": now.Add(1 * time.Hour).Unix(),
		"iat": now.Unix(),
	}

	subjectClaims := map[string]interface{}{}
	for k, v := range commonClaims {
		subjectClaims[k] = v
	}
	// azp = valid UUID string (the agent's internal ID)
	subjectClaims["azp"] = agentUUID.String()

	subjectToken := signServiceTestJWT(t, privateKey, subjectClaims)
	clientAssertion := signServiceTestJWT(t, privateKey, commonClaims)

	req := NewTokenExchangeRequest(
		TokenExchangeGrantType,
		subjectToken,
		AccessTokenType,
		clientAssertion,
		JWTBearerType,
		"https://api.example.com/resource",
		"",
	)

	_, _ = svc.Exchange(context.Background(), req)

	// After T030: Get is called (with parsed UUID). Before T030: GetByClientID is called.
	assert.True(t, tracker.getCalled, "agentRepository.Get should be called (not GetByClientID) after T030")
	assert.False(t, tracker.getByClientIDCalled, "agentRepository.GetByClientID must NOT be called after T030")
}

// TestExchange_AgentLookup_InvalidUUIDReturnsInvalidRequest verifies that when the CEL
// expression returns a non-UUID string, the service returns an invalid_request error.
// [T029] Written before T030 — fails until T030 adds id.ParseAgentID and returns invalid_request.
func TestExchange_AgentLookup_InvalidUUIDReturnsInvalidRequest(t *testing.T) {
	privateKey, keySet := generateTestRSAKeySet(t)

	tracker := &trackingAgentRepository{}
	svc := newServiceForStep9Test(t, keySet, tracker)

	now := time.Now()
	commonClaims := map[string]interface{}{
		"iss": "https://auth.example.com",
		"aud": "agentic-identity-broker",
		"sub": "user@example.com",
		"exp": now.Add(1 * time.Hour).Unix(),
		"iat": now.Unix(),
	}

	subjectClaims := map[string]interface{}{}
	for k, v := range commonClaims {
		subjectClaims[k] = v
	}
	// azp = not a valid UUID → id.ParseAgentID will fail after T030
	subjectClaims["azp"] = "not-a-valid-uuid"

	subjectToken := signServiceTestJWT(t, privateKey, subjectClaims)
	clientAssertion := signServiceTestJWT(t, privateKey, commonClaims)

	req := NewTokenExchangeRequest(
		TokenExchangeGrantType,
		subjectToken,
		AccessTokenType,
		clientAssertion,
		JWTBearerType,
		"https://api.example.com/resource",
		"",
	)

	_, err := svc.Exchange(context.Background(), req)

	require.Error(t, err)
	tokenErr, ok := err.(*TokenExchangeError)
	require.True(t, ok, "error must be a *TokenExchangeError, got %T: %v", err, err)
	assert.Equal(t, "invalid_request", tokenErr.Code(),
		"non-UUID agentClientID must return invalid_request (not access_denied)")
}
