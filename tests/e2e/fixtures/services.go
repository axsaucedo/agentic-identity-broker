package fixtures

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
)

// GitHubService returns a fixture for GitHub OAuth2 service.
// This is used for token exchange testing - represents GitHub as a third-party service.
// Includes protected_resources for resource-based service lookup (US2).
func GitHubService() *model.ThirdpartyOAuth2ProviderEntity {
	now := time.Now()
	metadataURL := "https://github.com/.well-known/oauth-authorization-server"

	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "github-service",
		DisplayName: "GitHub",
		ClientID:    "github-client-id",
		Secret:      EncryptedSecret("github-service", "github-client-secret"),
		IssuerURI:   "https://github.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: true,
			MetadataURL:     &metadataURL,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "repo", Description: "Access repository"},
			{ScopeValue: "user", Description: "Access user information"},
			{ScopeValue: "read:org", Description: "Read organization information"},
		},
		ProtectedResources: []string{"https://api.github.com"},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// GoogleService returns a fixture for Google OAuth2 service.
// This is used for token exchange testing - represents Google as a third-party service.
func GoogleService() *model.ThirdpartyOAuth2ProviderEntity {
	now := time.Now()
	metadataURL := "https://accounts.google.com/.well-known/openid-configuration"

	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "google-service",
		DisplayName: "Google",
		ClientID:    "google-client-id",
		Secret:      EncryptedSecret("google-service", "google-client-secret"),
		IssuerURI:   "https://accounts.google.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: true,
			MetadataURL:     &metadataURL,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth2.googleapis.com/token",
			AuthorizeEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "calendar", Description: "Access calendar"},
			{ScopeValue: "drive", Description: "Access Google Drive"},
			{ScopeValue: "userinfo.email", Description: "Access email"},
		},
		ProtectedResources: []string{"https://www.googleapis.com"},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// MicrosoftService returns a fixture for Microsoft Azure OAuth2 service.
// This is used for token exchange testing - represents Microsoft as a third-party service.
func MicrosoftService() *model.ThirdpartyOAuth2ProviderEntity {
	now := time.Now()
	metadataURL := "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration"

	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "microsoft-service",
		DisplayName: "Microsoft Azure",
		ClientID:    "microsoft-client-id",
		Secret:      EncryptedSecret("microsoft-service", "microsoft-client-secret"),
		IssuerURI:   "https://login.microsoftonline.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: true,
			MetadataURL:     &metadataURL,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			AuthorizeEndpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "mail.read", Description: "Read mail"},
			{ScopeValue: "calendar.read", Description: "Read calendar"},
			{ScopeValue: "user.read", Description: "Read user profile"},
		},
		ProtectedResources: []string{"https://graph.microsoft.com"},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// ServiceWithID returns a service with the specified ID.
// Base service can be customized for testing specific scenarios.
func ServiceWithID(id string) *model.ThirdpartyOAuth2ProviderEntity {
	now := time.Now()

	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          id,
		DisplayName: "Test Service " + id,
		ClientID:    "test-client-" + id,
		Secret:      EncryptedSecret(id, "test-secret-"+id),
		IssuerURI:   "https://test-issuer.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://test-issuer.example.com/token",
			AuthorizeEndpoint: "https://test-issuer.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
			{ScopeValue: "write", Description: "Write access"},
		},
		ProtectedResources: []string{"https://test-issuer.example.com/api"},
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}
