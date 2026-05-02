package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type txContextKey struct{}

type queryExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func contextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func (a *Adapter) queryExecutor(ctx context.Context) queryExecutor {
	if tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx); ok && tx != nil {
		return tx
	}

	return a.db
}
