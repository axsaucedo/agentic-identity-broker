package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const signingKeyRepoTestDriverName = "signing_key_repo_test_driver"

var (
	signingKeyRepoTestDriverOnce sync.Once
	signingKeyRepoTestConfigs    = struct {
		sync.Mutex
		values map[string]signingKeyRepoTestConfig
	}{values: make(map[string]signingKeyRepoTestConfig)}
)

type signingKeyRepoTestConfig struct {
	beginErr     error
	execErr      error
	execErrs     []error
	queryErr     error
	commitErr    error
	rowsAffected int64
}

type signingKeyRepoTestDriver struct{}

type signingKeyRepoTestConn struct {
	cfg       signingKeyRepoTestConfig
	execCalls int
}

type signingKeyRepoTestTx struct {
	cfg signingKeyRepoTestConfig
}

var _ driver.Conn = (*signingKeyRepoTestConn)(nil)
var _ driver.ExecerContext = (*signingKeyRepoTestConn)(nil)
var _ driver.QueryerContext = (*signingKeyRepoTestConn)(nil)
var _ driver.ConnBeginTx = (*signingKeyRepoTestConn)(nil)
var _ driver.Tx = (*signingKeyRepoTestTx)(nil)

func (d *signingKeyRepoTestDriver) Open(name string) (driver.Conn, error) {
	signingKeyRepoTestConfigs.Lock()
	cfg, ok := signingKeyRepoTestConfigs.values[name]
	signingKeyRepoTestConfigs.Unlock()
	if !ok {
		return nil, fmt.Errorf("missing signing key repo test config for %q", name)
	}
	return &signingKeyRepoTestConn{cfg: cfg}, nil
}

func (c *signingKeyRepoTestConn) Prepare(_ string) (driver.Stmt, error) {
	return nil, errors.New("Prepare not implemented")
}

func (c *signingKeyRepoTestConn) Close() error {
	return nil
}

func (c *signingKeyRepoTestConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *signingKeyRepoTestConn) BeginTx(_ context.Context, _ driver.TxOptions) (driver.Tx, error) {
	if c.cfg.beginErr != nil {
		return nil, c.cfg.beginErr
	}
	return &signingKeyRepoTestTx{cfg: c.cfg}, nil
}

func (c *signingKeyRepoTestConn) ExecContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	if len(c.cfg.execErrs) > 0 {
		var err error
		if c.execCalls < len(c.cfg.execErrs) {
			err = c.cfg.execErrs[c.execCalls]
		}
		c.execCalls++
		if err != nil {
			return nil, err
		}
		return driver.RowsAffected(c.cfg.rowsAffected), nil
	}
	if c.cfg.execErr != nil {
		return nil, c.cfg.execErr
	}
	c.execCalls++
	return driver.RowsAffected(c.cfg.rowsAffected), nil
}

func (c *signingKeyRepoTestConn) QueryContext(_ context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	if c.cfg.queryErr != nil {
		return nil, c.cfg.queryErr
	}
	return &signingKeyRepoTestRows{}, nil
}

func (t *signingKeyRepoTestTx) Commit() error {
	return t.cfg.commitErr
}

func (t *signingKeyRepoTestTx) Rollback() error {
	return nil
}

type signingKeyRepoTestRows struct{}

func (r *signingKeyRepoTestRows) Columns() []string {
	return []string{"id"}
}

func (r *signingKeyRepoTestRows) Close() error {
	return nil
}

func (r *signingKeyRepoTestRows) Next(_ []driver.Value) error {
	return io.EOF
}

func newUnitTestSigningKeyRepo(t *testing.T, cfg signingKeyRepoTestConfig) *SigningKeyRepo {
	t.Helper()

	signingKeyRepoTestDriverOnce.Do(func() {
		sql.Register(signingKeyRepoTestDriverName, &signingKeyRepoTestDriver{})
	})

	if cfg.rowsAffected == 0 {
		cfg.rowsAffected = 1
	}

	dsn := fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())
	signingKeyRepoTestConfigs.Lock()
	signingKeyRepoTestConfigs.values[dsn] = cfg
	signingKeyRepoTestConfigs.Unlock()

	db, err := sql.Open(signingKeyRepoTestDriverName, dsn)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
		signingKeyRepoTestConfigs.Lock()
		delete(signingKeyRepoTestConfigs.values, dsn)
		signingKeyRepoTestConfigs.Unlock()
	})

	return NewSigningKeyRepo(&Adapter{
		db: sqlx.NewDb(db, signingKeyRepoTestDriverName),
		timeouts: ports.StorageTimeouts{
			Read:  time.Second,
			Write: time.Second,
		},
	})
}

func stubSigningKey() *storage.SigningKey {
	return &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-unit-test"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("ciphertext"),
		IsCurrent:           true,
		ActivatesAt:         time.Now().UTC(),
		CreatedAt:           time.Now().UTC(),
	}
}

func assertStorageErrorKind(t *testing.T, err error, operation string, kind storage.ErrorKind) {
	t.Helper()

	var se *storage.StorageError
	require.ErrorAs(t, err, &se)
	assert.Equal(t, operation, se.Operation)
	assert.Equal(t, kind, se.Kind)
}

func TestSigningKeyRepo_ErrorClassification(t *testing.T) {
	tests := []struct {
		name      string
		cfg       signingKeyRepoTestConfig
		operation string
		kind      storage.ErrorKind
		call      func(repo *SigningKeyRepo) error
	}{
		{
			name:      "Create maps deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{execErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.Create",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				return repo.Create(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "Create maps generic exec error to connection",
			cfg:       signingKeyRepoTestConfig{execErr: errors.New("database unavailable")},
			operation: "SigningKeyRepo.Create",
			kind:      storage.ErrorKindConnection,
			call: func(repo *SigningKeyRepo) error {
				return repo.Create(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "Create maps unique violation to conflict",
			cfg:       signingKeyRepoTestConfig{execErr: &pgconn.PgError{Code: "23505"}},
			operation: "SigningKeyRepo.Create",
			kind:      storage.ErrorKindConflict,
			call: func(repo *SigningKeyRepo) error {
				return repo.Create(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "CreateAndSetCurrent maps begin deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{beginErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.CreateAndSetCurrent",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				return repo.CreateAndSetCurrent(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "CreateAndSetCurrent maps exec deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{execErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.CreateAndSetCurrent",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				return repo.CreateAndSetCurrent(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "CreateAndSetCurrent maps unique violation on insert to conflict",
			cfg:       signingKeyRepoTestConfig{execErrs: []error{nil, &pgconn.PgError{Code: "23505"}}},
			operation: "SigningKeyRepo.CreateAndSetCurrent",
			kind:      storage.ErrorKindConflict,
			call: func(repo *SigningKeyRepo) error {
				return repo.CreateAndSetCurrent(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "CreateAndSetCurrent maps commit generic error to connection",
			cfg:       signingKeyRepoTestConfig{commitErr: errors.New("commit failed")},
			operation: "SigningKeyRepo.CreateAndSetCurrent",
			kind:      storage.ErrorKindConnection,
			call: func(repo *SigningKeyRepo) error {
				return repo.CreateAndSetCurrent(context.Background(), stubSigningKey())
			},
		},
		{
			name:      "ListActive maps deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{queryErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.ListActive",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				_, err := repo.ListActive(context.Background())
				return err
			},
		},
		{
			name:      "ListActive maps generic query error to connection",
			cfg:       signingKeyRepoTestConfig{queryErr: errors.New("query failed")},
			operation: "SigningKeyRepo.ListActive",
			kind:      storage.ErrorKindConnection,
			call: func(repo *SigningKeyRepo) error {
				_, err := repo.ListActive(context.Background())
				return err
			},
		},
		{
			name:      "SetCurrent maps begin deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{beginErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.SetCurrent",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				return repo.SetCurrent(context.Background(), id.NewKeyID("kid-set-current"))
			},
		},
		{
			name:      "SetCurrent maps exec deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{execErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.SetCurrent",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				return repo.SetCurrent(context.Background(), id.NewKeyID("kid-set-current"))
			},
		},
		{
			name:      "SetCurrent maps commit generic error to connection",
			cfg:       signingKeyRepoTestConfig{commitErr: errors.New("commit failed")},
			operation: "SigningKeyRepo.SetCurrent",
			kind:      storage.ErrorKindConnection,
			call: func(repo *SigningKeyRepo) error {
				return repo.SetCurrent(context.Background(), id.NewKeyID("kid-set-current"))
			},
		},
		{
			name:      "Delete maps deadline exceeded to timeout",
			cfg:       signingKeyRepoTestConfig{execErr: context.DeadlineExceeded},
			operation: "SigningKeyRepo.Delete",
			kind:      storage.ErrorKindTimeout,
			call: func(repo *SigningKeyRepo) error {
				return repo.Delete(context.Background(), id.NewKeyID("kid-delete"))
			},
		},
		{
			name:      "Delete maps generic exec error to connection",
			cfg:       signingKeyRepoTestConfig{execErr: errors.New("delete failed")},
			operation: "SigningKeyRepo.Delete",
			kind:      storage.ErrorKindConnection,
			call: func(repo *SigningKeyRepo) error {
				return repo.Delete(context.Background(), id.NewKeyID("kid-delete"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newUnitTestSigningKeyRepo(t, tt.cfg)

			err := tt.call(repo)
			require.Error(t, err)
			assertStorageErrorKind(t, err, tt.operation, tt.kind)
		})
	}
}
