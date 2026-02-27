package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
)

// ThirdpartyOAuth2ProviderRecord is the adapter-local database record struct for
// third-party OAuth2 providers. It maps directly to the thirdparty_oauth2_services
// table schema using db tags.
//
// Unlike the legacy storage struct, this record holds SecretCiphertext only —
// there is no plaintext client_secret field. Plaintext secrets never touch this struct.
//
// This type is used exclusively within the postgres adapter; it is never exposed to
// domain code. Conversion to/from domain entities uses entityToRecord and recordToEntity.
type ThirdpartyOAuth2ProviderRecord struct {
	ID                 string             `db:"id"`
	DisplayName        string             `db:"display_name"`
	ClientID           string             `db:"client_id"`
	SecretCiphertext   []byte             `db:"client_secret_encrypted"`
	IssuerURI          string             `db:"issuer_uri"`
	EnableDiscovery    bool               `db:"enable_discovery"`
	MetadataURL        *string            `db:"metadata_url"`
	TokenEndpoint      string             `db:"token_endpoint"`
	AuthorizeEndpoint  string             `db:"authorize_endpoint"`
	Scopes             providerScopeArray `db:"scopes"`
	ProtectedResources []string           `db:"-"` // scanned via pq.Array in query methods
	CreatedAt          time.Time          `db:"created_at"`
	UpdatedAt          time.Time          `db:"updated_at"`
}

// providerScopeArray handles JSONB serialization of model.OAuthScope slices for PostgreSQL.
// This is an adapter-local type; model.OAuthScope has no db/json tags by design.
type providerScopeArray []model.OAuthScope

// scopeJSON is an intermediate struct for JSON serialization with snake_case keys,
// since model.OAuthScope intentionally has no json tags.
type scopeJSON struct {
	ScopeValue  string `json:"scope_value"`
	Description string `json:"description"`
}

// Scan implements sql.Scanner for JSONB deserialization of OAuthScope arrays.
func (a *providerScopeArray) Scan(value any) error {
	if value == nil {
		*a = []model.OAuthScope{}
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan providerScopeArray: expected []byte")
	}

	var raw []scopeJSON
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal providerScopeArray: %w", err)
	}

	scopes := make([]model.OAuthScope, len(raw))
	for i, r := range raw {
		scopes[i] = model.OAuthScope{
			ScopeValue:  r.ScopeValue,
			Description: r.Description,
		}
	}
	*a = scopes
	return nil
}

// Value implements driver.Valuer for JSONB serialization of OAuthScope arrays.
func (a providerScopeArray) Value() (driver.Value, error) {
	if a == nil {
		return json.Marshal([]scopeJSON{})
	}
	raw := make([]scopeJSON, len(a))
	for i, scope := range a {
		raw[i] = scopeJSON{
			ScopeValue:  scope.ScopeValue,
			Description: scope.Description,
		}
	}
	return json.Marshal(raw)
}

// entityToRecord converts a ThirdpartyOAuth2ProviderEntity to a ThirdpartyOAuth2ProviderRecord.
// The entity's Secret must be in encrypted state — GetCiphertext() must succeed.
// Returns an error if the entity is nil or the secret is not yet encrypted.
func entityToRecord(entity *model.ThirdpartyOAuth2ProviderEntity) (*ThirdpartyOAuth2ProviderRecord, error) {
	if entity == nil {
		return nil, errors.New("entity cannot be nil")
	}

	ciphertext, err := entity.Secret.GetCiphertext()
	if err != nil {
		return nil, fmt.Errorf("entity secret must be encrypted before storing: %w", err)
	}

	record := &ThirdpartyOAuth2ProviderRecord{
		ID:                entity.ID,
		DisplayName:       entity.DisplayName,
		ClientID:          entity.ClientID,
		SecretCiphertext:  ciphertext,
		IssuerURI:         entity.IssuerURI,
		EnableDiscovery:   entity.Discovery.EnableDiscovery,
		TokenEndpoint:     entity.Endpoints.TokenEndpoint,
		AuthorizeEndpoint: entity.Endpoints.AuthorizeEndpoint,
		Scopes:            providerScopeArray(entity.Scopes),
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
	}

	if entity.Discovery.MetadataURL != nil {
		s := *entity.Discovery.MetadataURL
		record.MetadataURL = &s
	}

	if entity.ProtectedResources != nil {
		record.ProtectedResources = make([]string, len(entity.ProtectedResources))
		copy(record.ProtectedResources, entity.ProtectedResources)
	}

	return record, nil
}

// recordToEntity converts a ThirdpartyOAuth2ProviderRecord to a ThirdpartyOAuth2ProviderEntity.
// The resulting entity's Secret is in encrypted state. The domain service must decrypt it
// before the entity can be used for OAuth2 operations.
func recordToEntity(record *ThirdpartyOAuth2ProviderRecord) *model.ThirdpartyOAuth2ProviderEntity {
	if record == nil {
		return nil
	}

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          record.ID,
		DisplayName: record.DisplayName,
		ClientID:    record.ClientID,
		Secret:      model.NewEncryptedSecret(record.SecretCiphertext),
		IssuerURI:   record.IssuerURI,
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: record.EnableDiscovery,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     record.TokenEndpoint,
			AuthorizeEndpoint: record.AuthorizeEndpoint,
		},
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}

	if record.MetadataURL != nil {
		s := *record.MetadataURL
		entity.Discovery.MetadataURL = &s
	}

	if record.Scopes != nil {
		entity.Scopes = make([]model.OAuthScope, len(record.Scopes))
		copy(entity.Scopes, record.Scopes)
	}

	if record.ProtectedResources != nil {
		entity.ProtectedResources = make([]string, len(record.ProtectedResources))
		copy(entity.ProtectedResources, record.ProtectedResources)
	}

	return entity
}
