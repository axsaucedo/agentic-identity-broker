package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

// PostgresThirdpartyOAuth2ProviderRepository implements ports.ThirdpartyOAuth2ProviderRepository
// using PostgreSQL. All entities must have Secret in encrypted state before Create/Update.
// Retrieval returns entities with Secret in encrypted state for the domain service to decrypt.
type PostgresThirdpartyOAuth2ProviderRepository struct {
	adapter *Adapter
}

// NewPostgresThirdpartyOAuth2ProviderRepository creates a new PostgreSQL provider repository.
func NewPostgresThirdpartyOAuth2ProviderRepository(adapter *Adapter) *PostgresThirdpartyOAuth2ProviderRepository {
	return &PostgresThirdpartyOAuth2ProviderRepository{adapter: adapter}
}

// Create stores a new provider entity in PostgreSQL.
// Entity.ID must be set by the caller before storing (the domain service generates it before encrypting).
// Entity.Secret must be in encrypted state.
func (r *PostgresThirdpartyOAuth2ProviderRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindConnection, nil, "database not initialized")
	}
	if entity == nil {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindValidation, nil, "entity cannot be nil")
	}
	if entity.ID.IsZero() {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider",
			storage.ErrorKindValidation, nil,
			"provider ID cannot be empty: caller must set ID before storing")
	}

	record, err := entityToRecord(entity)
	if err != nil {
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindValidation, err, "failed to convert entity to record")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `
		INSERT INTO thirdparty_oauth2_services (
			id, display_name, client_id, client_secret_encrypted, oauth2_flavor, issuer_uri,
			enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
			scopes, protected_resources, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = r.adapter.db.ExecContext(
		execCtx, query,
		record.ID, record.DisplayName, record.ClientID, record.SecretCiphertext,
		record.Flavor, record.IssuerURI, record.EnableDiscovery, record.MetadataURL,
		record.TokenEndpoint, record.AuthorizeEndpoint,
		record.Scopes, pq.Array(record.ProtectedResources),
		record.CreatedAt, record.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindConflict, err, "provider with this ID already exists")
		}
		return storage.NewStorageError("CreateThirdpartyOAuth2Provider", storage.ErrorKindConnection, err, fmt.Sprintf("failed to create provider: %v", err))
	}

	return nil
}

// Get retrieves a provider entity by ID from PostgreSQL.
// Returns entity with Secret in encrypted state.
func (r *PostgresThirdpartyOAuth2ProviderRepository) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("GetThirdpartyOAuth2Provider", storage.ErrorKindConnection, nil, "database not initialized")
	}
	if serviceID.IsZero() {
		return nil, storage.NewStorageError("GetThirdpartyOAuth2Provider", storage.ErrorKindValidation, nil, "provider ID cannot be empty")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, display_name, client_id, client_secret_encrypted, oauth2_flavor, issuer_uri,
		       enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
		       scopes, protected_resources, created_at, updated_at
		FROM thirdparty_oauth2_services
		WHERE id = $1
	`

	var record ThirdpartyOAuth2ProviderRecord
	err := r.adapter.db.QueryRowContext(queryCtx, query, serviceID).Scan(
		&record.ID, &record.DisplayName, &record.ClientID, &record.SecretCiphertext,
		&record.Flavor, &record.IssuerURI, &record.EnableDiscovery, &record.MetadataURL,
		&record.TokenEndpoint, &record.AuthorizeEndpoint,
		&record.Scopes, pq.Array(&record.ProtectedResources),
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.NewStorageError("GetThirdpartyOAuth2Provider", storage.ErrorKindNotFound, ports.ErrNotFound, "provider not found")
		}
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError("GetThirdpartyOAuth2Provider", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("GetThirdpartyOAuth2Provider", storage.ErrorKindConnection, err, "failed to get provider")
	}

	return recordToEntity(&record)
}

// Update updates an existing provider entity in PostgreSQL.
// Entity.Secret must be in encrypted state.
func (r *PostgresThirdpartyOAuth2ProviderRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindConnection, nil, "database not initialized")
	}
	if entity == nil {
		return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindValidation, nil, "entity cannot be nil")
	}

	record, err := entityToRecord(entity)
	if err != nil {
		return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindValidation, err, "failed to convert entity to record")
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	query := `
		UPDATE thirdparty_oauth2_services
		SET display_name = $2,
		    client_id = $3,
		    client_secret_encrypted = $4,
		    oauth2_flavor = $5,
		    issuer_uri = $6,
		    enable_discovery = $7,
		    metadata_url = $8,
		    token_endpoint = $9,
		    authorize_endpoint = $10,
		    scopes = $11,
		    protected_resources = $12,
		    updated_at = $13
		WHERE id = $1
		RETURNING created_at
	`

	var createdAt time.Time
	err = r.adapter.db.QueryRowContext(
		execCtx, query,
		record.ID, record.DisplayName, record.ClientID, record.SecretCiphertext,
		record.Flavor, record.IssuerURI, record.EnableDiscovery, record.MetadataURL,
		record.TokenEndpoint, record.AuthorizeEndpoint,
		record.Scopes, pq.Array(record.ProtectedResources),
		record.UpdatedAt,
	).Scan(&createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindNotFound, ports.ErrNotFound, "provider not found")
		}
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return storage.NewStorageError("UpdateThirdpartyOAuth2Provider", storage.ErrorKindConnection, err, "failed to update provider")
	}

	entity.CreatedAt = createdAt
	return nil
}

// Delete removes a provider entity by ID from PostgreSQL.
// Returns Conflict error if grants reference this provider (FR-022).
func (r *PostgresThirdpartyOAuth2ProviderRepository) Delete(ctx context.Context, serviceID id.ServiceID) error {
	if r.adapter.db == nil {
		return storage.NewStorageError("DeleteThirdpartyOAuth2Provider", storage.ErrorKindConnection, nil, "database not initialized")
	}
	if serviceID.IsZero() {
		return storage.NewStorageError("DeleteThirdpartyOAuth2Provider", storage.ErrorKindValidation, nil, "provider ID cannot be empty")
	}

	count, err := r.CountGrantsReferencingService(ctx, serviceID)
	if err != nil {
		return err
	}
	if count > 0 {
		return storage.NewStorageError(
			"DeleteThirdpartyOAuth2Provider",
			storage.ErrorKindConflict,
			nil,
			fmt.Sprintf("cannot delete provider: %d grants reference it", count),
		)
	}

	execCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Write)
	defer cancel()

	_, err = r.adapter.db.ExecContext(execCtx, `DELETE FROM thirdparty_oauth2_services WHERE id = $1`, serviceID)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return storage.NewStorageError("DeleteThirdpartyOAuth2Provider", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return storage.NewStorageError("DeleteThirdpartyOAuth2Provider", storage.ErrorKindConnection, err, "failed to delete provider")
	}

	return nil
}

// List retrieves all provider entities from PostgreSQL ordered by created_at DESC.
// Returns entities with Secret in encrypted state.
func (r *PostgresThirdpartyOAuth2ProviderRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("ListThirdpartyOAuth2Providers", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, display_name, client_id, client_secret_encrypted, oauth2_flavor, issuer_uri,
		       enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
		       scopes, protected_resources, created_at, updated_at
		FROM thirdparty_oauth2_services
		ORDER BY created_at DESC
	`

	rows, err := r.adapter.db.QueryContext(queryCtx, query)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError("ListThirdpartyOAuth2Providers", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("ListThirdpartyOAuth2Providers", storage.ErrorKindConnection, err, "failed to list providers")
	}
	defer func() { _ = rows.Close() }()

	var entities []*model.ThirdpartyOAuth2ProviderEntity

	for rows.Next() {
		var record ThirdpartyOAuth2ProviderRecord
		err := rows.Scan(
			&record.ID, &record.DisplayName, &record.ClientID, &record.SecretCiphertext,
			&record.Flavor, &record.IssuerURI, &record.EnableDiscovery, &record.MetadataURL,
			&record.TokenEndpoint, &record.AuthorizeEndpoint,
			&record.Scopes, pq.Array(&record.ProtectedResources),
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, storage.NewStorageError(
				"ListThirdpartyOAuth2Providers",
				storage.ErrorKindConnection,
				err,
				fmt.Sprintf("failed to scan provider row: %v", err),
			)
		}
		entity, err := recordToEntity(&record)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError("ListThirdpartyOAuth2Providers", storage.ErrorKindConnection, err, "error iterating provider rows")
	}

	return entities, nil
}

// CountGrantsReferencingService returns the count of grants referencing this provider.
// Used to enforce FR-022 (block deletion if grants exist).
func (r *PostgresThirdpartyOAuth2ProviderRepository) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	if r.adapter.db == nil {
		return 0, storage.NewStorageError("CountGrantsReferencingProvider", storage.ErrorKindConnection, nil, "database not initialized")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `SELECT COUNT(*) FROM user_grants WHERE delegated_oauth2_tokens @> $1::jsonb`
	searchPattern := fmt.Sprintf(`[{"thirdparty_oauth2_service_id": "%s"}]`, serviceID)

	var count int
	err := r.adapter.db.QueryRowContext(queryCtx, query, searchPattern).Scan(&count)
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return 0, storage.NewStorageError("CountGrantsReferencingProvider", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return 0, storage.NewStorageError("CountGrantsReferencingProvider", storage.ErrorKindConnection, err, "failed to count grants")
	}

	return count, nil
}

// FindByProtectedResource retrieves a provider by matching resource URI against
// protected_resources using PostgreSQL array containment with GIN index.
// The resourceURI must be normalized before calling.
func (r *PostgresThirdpartyOAuth2ProviderRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if r.adapter.db == nil {
		return nil, storage.NewStorageError("FindByProtectedResource", storage.ErrorKindConnection, nil, "database not initialized")
	}
	if resourceURI == "" {
		return nil, storage.NewStorageError("FindByProtectedResource", storage.ErrorKindValidation, nil, "resource URI cannot be empty")
	}

	queryCtx, cancel := context.WithTimeout(ctx, r.adapter.timeouts.Read)
	defer cancel()

	query := `
		SELECT id, display_name, client_id, client_secret_encrypted, oauth2_flavor, issuer_uri,
		       enable_discovery, metadata_url, token_endpoint, authorize_endpoint,
		       scopes, protected_resources, created_at, updated_at
		FROM thirdparty_oauth2_services
		WHERE protected_resources @> $1::text[]
	`

	rows, err := r.adapter.db.QueryContext(queryCtx, query, pq.Array([]string{resourceURI}))
	if err != nil {
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, storage.NewStorageError("FindByProtectedResource", storage.ErrorKindTimeout, err, "operation exceeded timeout")
		}
		return nil, storage.NewStorageError("FindByProtectedResource", storage.ErrorKindConnection, err, "failed to find provider by protected resource")
	}
	defer func() { _ = rows.Close() }()

	var entities []*model.ThirdpartyOAuth2ProviderEntity

	for rows.Next() {
		var record ThirdpartyOAuth2ProviderRecord
		err := rows.Scan(
			&record.ID, &record.DisplayName, &record.ClientID, &record.SecretCiphertext,
			&record.Flavor, &record.IssuerURI, &record.EnableDiscovery, &record.MetadataURL,
			&record.TokenEndpoint, &record.AuthorizeEndpoint,
			&record.Scopes, pq.Array(&record.ProtectedResources),
			&record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, storage.NewStorageError(
				"FindByProtectedResource",
				storage.ErrorKindConnection,
				err,
				fmt.Sprintf("failed to scan provider row: %v", err),
			)
		}
		entity, err := recordToEntity(&record)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError("FindByProtectedResource", storage.ErrorKindConnection, err, "error iterating provider rows")
	}

	if len(entities) == 0 {
		return nil, tokenexchange.NewInvalidTargetError("no service configured for the requested resource")
	}
	if len(entities) > 1 {
		return nil, tokenexchange.NewInvalidTargetError("multiple services configured for the same resource")
	}

	return entities[0], nil
}
