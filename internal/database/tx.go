package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

// Ensure databaseTx implements the Tx interface
var _ Tx = &databaseTx{}



type databaseTx struct {
	tx      *sql.Tx
	bindVar sqlf.BindVar
	config  DBConfig
}

// For transaction objects, we don't implement retry methods.
// Retry functionality is limited to the ping command only.

func (d *databaseTx) Ping(ctx context.Context) error {
	return errors.New("ping not supported in transaction")
}

func (d *databaseTx) PingWithRetry(ctx context.Context, retryCount int, retryDelay int) error {
	return errors.New("ping with retry not supported in transaction")
}

// withDefaultTimeout wraps a context with a default timeout if none is set and timeout is > 0
func (d *databaseTx) withDefaultTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {} // Return no-op cancel function
	}
	if timeout <= 0 {
		return ctx, func() {} // No timeout configured
	}
	return context.WithTimeout(ctx, timeout)
}

func (d *databaseTx) Exec(ctx context.Context, query *sqlf.Query) error {
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()
	
	_, err := d.tx.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

func (d *databaseTx) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()
	
	return d.tx.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *databaseTx) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	// For Query operations, we don't apply default timeout as the context
	// needs to remain valid for the lifetime of the Rows
	// Users should set their own timeout if needed
	return d.tx.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *databaseTx) Close() error {
	return errors.New("close method not supported in transaction")
}

func (d *databaseTx) WithTransact(ctx context.Context, f func(Tx) error) error {
	return errors.New("nested transactions are not supported")
}

func (d *databaseTx) Commit() error {
	return d.tx.Commit()
}

func (d *databaseTx) Rollback() error {
	return d.tx.Rollback()
}
