package postgres

import (
	"context"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/jmoiron/sqlx"
)

type oauth2TransactionKey struct{}

func (a *Adapter) BeginTX(ctx context.Context) (context.Context, error) {
	if a.db == nil {
		return nil, storage.NewStorageError("OAuth2Transaction.BeginTX", storage.ErrorKindConnection, nil, "database not initialized")
	}
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, storage.NewStorageError("OAuth2Transaction.BeginTX", storage.ErrorKindConnection, err, "failed to begin transaction")
	}
	return context.WithValue(ctx, oauth2TransactionKey{}, tx), nil
}

func (a *Adapter) Commit(ctx context.Context) error {
	tx, ok := oauth2Transaction(ctx)
	if !ok {
		return fmt.Errorf("OAuth2Transaction.Commit: transaction missing from context")
	}
	if err := tx.Commit(); err != nil {
		return storage.NewStorageError("OAuth2Transaction.Commit", storage.ErrorKindConnection, err, "failed to commit transaction")
	}
	return nil
}

func (a *Adapter) Rollback(ctx context.Context) error {
	tx, ok := oauth2Transaction(ctx)
	if !ok {
		return fmt.Errorf("OAuth2Transaction.Rollback: transaction missing from context")
	}
	if err := tx.Rollback(); err != nil {
		return storage.NewStorageError("OAuth2Transaction.Rollback", storage.ErrorKindConnection, err, "failed to roll back transaction")
	}
	return nil
}

func oauth2Transaction(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(oauth2TransactionKey{}).(*sqlx.Tx)
	return tx, ok
}

func (a *Adapter) oauth2Executor(ctx context.Context) sqlx.ExtContext {
	if tx, ok := oauth2Transaction(ctx); ok {
		return tx
	}
	return a.db
}
