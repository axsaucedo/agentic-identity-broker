// Package postgres implements PostgreSQL storage adapters.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// ThirdpartyServiceRepository implements ports.ThirdpartyOAuth2ServiceRepository using PostgreSQL.
type ThirdpartyServiceRepository struct {
	adapter        *Adapter
	encryptionPort ports.EncryptionPort
}

// NewThirdpartyServiceRepository creates a new PostgreSQL third-party service repository.
// The adapter must be initialized before use.
// encryptionPort is used to encrypt/decrypt client secrets.
func NewThirdpartyServiceRepository(adapter *Adapter, encryptionPort ports.EncryptionPort) *ThirdpartyServiceRepository {
	return &ThirdpartyServiceRepository{
		adapter:        adapter,
		encryptionPort: encryptionPort,
	}
}

// Create creates a new OAuth2 service configuration in PostgreSQL.
// Generates a UUID for the service if ID is empty.
// Encrypts client_secret using the configured EncryptionPort.
// Returns StorageError with Kind=Conflict if service ID already exists.
func (r *ThirdpartyServiceRepository) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if service == nil {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			nil,
			"service cannot be nil",
		)
	}

	// Validate before storing
	if err := service.ValidateForCreate(); err != nil {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			err,
			"service validation failed",
		)
	}

	// Generate ID if not provided
	if service.ID == "" {
		service.ID = uuid.New().String()
	}

	// Encrypt client secret
	encryptionContext := map[string]string{
		"service_id": service.ID,
	}
	encryptedSecret, err := r.encryptionPort.Encrypt(ctx, []byte(service.ClientSecret), encryptionContext)
	if err != nil {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to encrypt client secret",
		)
	}

	// Marshal scopes to JSON
	scopesJSON, err := json.Marshal(service.Scopes)
	if err != nil {
		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			err,
			"failed to marshal scopes to JSON",
		)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `
		INSERT INTO thirdparty_oauth2_services (
			id, display_name, client_id, client_secret_encrypted, issuer_uri,
			enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
			scopes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = r.adapter.db.ExecContext(
		execCtx,
		query,
		service.ID,
		service.DisplayName,
		service.ClientID,
		encryptedSecret,
		service.IssuerURI,
		service.Discovery.EnableDiscovery,
		service.Discovery.MetadataURL,
		service.Endpoints.TokenEndpoint,
		service.Endpoints.AuthorizeEndpoint,
		scopesJSON,
		service.CreatedAt,
		service.UpdatedAt,
	)

	if err != nil {
		// Check for timeout
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError(
				"CreateThirdpartyOAuth2Service",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}

		// Check for duplicate key violation (PostgreSQL error code 23505)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return storage.NewStorageError(
				"CreateThirdpartyOAuth2Service",
				storage.ErrorKindConflict,
				err,
				"service with this ID already exists",
			)
		}

		return storage.NewStorageError(
			"CreateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to create service",
		)
	}

	return nil
}

// Get retrieves an OAuth2 service configuration by ID from PostgreSQL.
// Decrypts client_secret using the configured EncryptionPort.
// Returns StorageError with Kind=NotFound if service not found.
func (r *ThirdpartyServiceRepository) Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"GetThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if id == "" {
		return nil, storage.NewStorageError(
			"GetThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			nil,
			"service ID cannot be empty",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, display_name, client_id, client_secret_encrypted, issuer_uri,
		       enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
		       scopes, created_at, updated_at
		FROM thirdparty_oauth2_services
		WHERE id = $1
	`

	var (
		service         storage.ThirdpartyOAuth2Service
		encryptedSecret []byte
		scopesJSON      []byte
	)

	err := r.adapter.db.QueryRowContext(queryCtx, query, id).Scan(
		&service.ID,
		&service.DisplayName,
		&service.ClientID,
		&encryptedSecret,
		&service.IssuerURI,
		&service.Discovery.EnableDiscovery,
		&service.Discovery.MetadataURL,
		&service.Endpoints.TokenEndpoint,
		&service.Endpoints.AuthorizeEndpoint,
		&scopesJSON,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.NewStorageError(
				"GetThirdpartyOAuth2Service",
				storage.ErrorKindNotFound,
				ports.ErrNotFound,
				"service not found",
			)
		}
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError(
				"GetThirdpartyOAuth2Service",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return nil, storage.NewStorageError(
			"GetThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to get service",
		)
	}

	// Unmarshal scopes
	if err := json.Unmarshal(scopesJSON, &service.Scopes); err != nil {
		return nil, storage.NewStorageError(
			"GetThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to unmarshal scopes",
		)
	}

	// Decrypt client secret
	encryptionContext := map[string]string{
		"service_id": service.ID,
	}
	decryptedSecret, err := r.encryptionPort.Decrypt(ctx, encryptedSecret, encryptionContext)
	if err != nil {
		return nil, storage.NewStorageError(
			"GetThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to decrypt client secret",
		)
	}
	service.ClientSecret = string(decryptedSecret)

	// Return deep copy to prevent external mutation
	return service.Copy(), nil
}

// Update updates an existing OAuth2 service configuration in PostgreSQL.
// Encrypts client_secret using the configured EncryptionPort.
// Returns StorageError with Kind=NotFound if service ID not found.
func (r *ThirdpartyServiceRepository) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if service == nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			nil,
			"service cannot be nil",
		)
	}

	// Validate before updating
	if err := service.Validate(); err != nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			err,
			"service validation failed",
		)
	}

	// Encrypt client secret
	encryptionContext := map[string]string{
		"service_id": service.ID,
	}
	encryptedSecret, err := r.encryptionPort.Encrypt(ctx, []byte(service.ClientSecret), encryptionContext)
	if err != nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to encrypt client secret",
		)
	}

	// Marshal scopes to JSON
	scopesJSON, err := json.Marshal(service.Scopes)
	if err != nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			err,
			"failed to marshal scopes to JSON",
		)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `
		UPDATE thirdparty_oauth2_services
		SET display_name = $2,
		    client_id = $3,
		    client_secret_encrypted = $4,
		    issuer_uri = $5,
		    enable_discovery = $6,
		    metadata_url = $7,
		    token_endpoint = $8,
		    authorize_endpoint = $9,
		    scopes = $10,
		    updated_at = $11
		WHERE id = $1
	`

	result, err := r.adapter.db.ExecContext(
		execCtx,
		query,
		service.ID,
		service.DisplayName,
		service.ClientID,
		encryptedSecret,
		service.IssuerURI,
		service.Discovery.EnableDiscovery,
		service.Discovery.MetadataURL,
		service.Endpoints.TokenEndpoint,
		service.Endpoints.AuthorizeEndpoint,
		scopesJSON,
		service.UpdatedAt,
	)

	if err != nil {
		// Check for timeout
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError(
				"UpdateThirdpartyOAuth2Service",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}

		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to update service",
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to get affected rows",
		)
	}

	if rows == 0 {
		return storage.NewStorageError(
			"UpdateThirdpartyOAuth2Service",
			storage.ErrorKindNotFound,
			ports.ErrNotFound,
			"service not found",
		)
	}

	return nil
}

// Delete deletes an OAuth2 service configuration by ID from PostgreSQL.
// Returns error with Kind=Conflict if active grants reference this service (FR-022).
// Idempotent: returns nil if service doesn't exist.
func (r *ThirdpartyServiceRepository) Delete(ctx context.Context, id string) error {
	if r.adapter.db == nil {
		return storage.NewStorageError(
			"DeleteThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	if id == "" {
		return storage.NewStorageError(
			"DeleteThirdpartyOAuth2Service",
			storage.ErrorKindValidation,
			nil,
			"service ID cannot be empty",
		)
	}

	// Check if any grants reference this service (FR-022)
	count, err := r.CountGrantsReferencingService(ctx, id)
	if err != nil {
		return err
	}

	if count > 0 {
		return storage.NewStorageError(
			"DeleteThirdpartyOAuth2Service",
			storage.ErrorKindConflict,
			nil,
			fmt.Sprintf("cannot delete service: %d grants reference it", count),
		)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `DELETE FROM thirdparty_oauth2_services WHERE id = $1`

	_, err = r.adapter.db.ExecContext(execCtx, query, id)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError(
				"DeleteThirdpartyOAuth2Service",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return storage.NewStorageError(
			"DeleteThirdpartyOAuth2Service",
			storage.ErrorKindConnection,
			err,
			"failed to delete service",
		)
	}

	return nil
}

// List retrieves all OAuth2 service configurations from PostgreSQL.
// Decrypts client secrets for each service.
// Returns empty slice if no services exist (not an error).
func (r *ThirdpartyServiceRepository) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError(
			"ListThirdpartyOAuth2Services",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, display_name, client_id, client_secret_encrypted, issuer_uri,
		       enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
		       scopes, created_at, updated_at
		FROM thirdparty_oauth2_services
		ORDER BY created_at DESC
	`

	rows, err := r.adapter.db.QueryContext(queryCtx, query)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError(
				"ListThirdpartyOAuth2Services",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return nil, storage.NewStorageError(
			"ListThirdpartyOAuth2Services",
			storage.ErrorKindConnection,
			err,
			"failed to list services",
		)
	}
	defer rows.Close()

	var services []*storage.ThirdpartyOAuth2Service

	for rows.Next() {
		var (
			service         storage.ThirdpartyOAuth2Service
			encryptedSecret []byte
			scopesJSON      []byte
		)

		err := rows.Scan(
			&service.ID,
			&service.DisplayName,
			&service.ClientID,
			&encryptedSecret,
			&service.IssuerURI,
			&service.Discovery.EnableDiscovery,
			&service.Discovery.MetadataURL,
			&service.Endpoints.TokenEndpoint,
			&service.Endpoints.AuthorizeEndpoint,
			&scopesJSON,
			&service.CreatedAt,
			&service.UpdatedAt,
		)

		if err != nil {
			return nil, storage.NewStorageError(
				"ListThirdpartyOAuth2Services",
				storage.ErrorKindConnection,
				err,
				"failed to scan service row",
			)
		}

		// Unmarshal scopes
		if err := json.Unmarshal(scopesJSON, &service.Scopes); err != nil {
			return nil, storage.NewStorageError(
				"ListThirdpartyOAuth2Services",
				storage.ErrorKindConnection,
				err,
				"failed to unmarshal scopes",
			)
		}

		// Decrypt client secret
		encryptionContext := map[string]string{
			"service_id": service.ID,
		}
		decryptedSecret, err := r.encryptionPort.Decrypt(ctx, encryptedSecret, encryptionContext)
		if err != nil {
			return nil, storage.NewStorageError(
				"ListThirdpartyOAuth2Services",
				storage.ErrorKindConnection,
				err,
				"failed to decrypt client secret",
			)
		}
		service.ClientSecret = string(decryptedSecret)

		services = append(services, service.Copy())
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError(
			"ListThirdpartyOAuth2Services",
			storage.ErrorKindConnection,
			err,
			"error iterating service rows",
		)
	}

	return services, nil
}

// CountGrantsReferencingService returns the number of active grants that reference this service.
// Used to enforce FR-022 (block service deletion if grants exist).
// Queries the user_grants table for grants that have this service in their delegated_oauth2_tokens.
func (r *ThirdpartyServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	if r.adapter.db == nil {
		return 0, storage.NewStorageError(
			"CountGrantsReferencingService",
			storage.ErrorKindConnection,
			nil,
			"database not initialized",
		)
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	// Query to count grants that reference this service in their delegated_oauth2_tokens
	// Using JSONB containment operator @>
	query := `
		SELECT COUNT(*)
		FROM user_grants
		WHERE delegated_oauth2_tokens @> $1::jsonb
	`

	// Construct JSONB search pattern: [{"thirdparty_oauth2_service_id": "<serviceID>"}]
	searchPattern := fmt.Sprintf(`[{"thirdparty_oauth2_service_id": "%s"}]`, serviceID)

	var count int
	err := r.adapter.db.QueryRowContext(queryCtx, query, searchPattern).Scan(&count)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return 0, storage.NewStorageError(
				"CountGrantsReferencingService",
				storage.ErrorKindTimeout,
				err,
				"operation exceeded timeout",
			)
		}
		return 0, storage.NewStorageError(
			"CountGrantsReferencingService",
			storage.ErrorKindConnection,
			err,
			"failed to count grants",
		)
	}

	return count, nil
}
