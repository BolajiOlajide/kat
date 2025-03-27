package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

// Ensure databaseTx implements the Tx interface
var _ Tx = &databaseTx{}

type databaseTx struct {
	tx      *sql.Tx
	bindVar sqlf.BindVar
}

// For transaction objects, we don't implement retry methods.
// Retry functionality is limited to the ping command only.

func (d *databaseTx) Ping(ctx context.Context) error {
	return errors.New("ping not supported in transaction")
}

func (d *databaseTx) PingWithRetry(ctx context.Context, retryCount int, retryDelay int) error {
	return errors.New("ping with retry not supported in transaction")
}

func (d *databaseTx) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := d.tx.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

// ValidateQuery checks if a SQL query is valid without executing it
// It uses Postgres's EXPLAIN to validate the query syntax
func (d *databaseTx) ValidateQuery(ctx context.Context, query *sqlf.Query) error {
	// Extract the SQL query
	sqlQuery := query.Query(d.bindVar)
	
	// Skip empty queries
	if strings.TrimSpace(sqlQuery) == "" {
		return errors.New("empty SQL query")
	}
	
	// For non-SELECT queries, we need to wrap them in EXPLAIN
	// This validates the query without executing it
	var explainQuery string
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlQuery)), "SELECT") {
		explainQuery = fmt.Sprintf("EXPLAIN %s", sqlQuery)
	} else {
		explainQuery = fmt.Sprintf("EXPLAIN (ANALYZE FALSE) %s", sqlQuery)
	}
	
	// Execute the EXPLAIN query to verify syntax
	_, err := d.tx.ExecContext(ctx, explainQuery, query.Args()...)
	if err != nil {
		return errors.Wrap(err, "SQL validation failed")
	}
	
	return nil
}

func (d *databaseTx) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return d.tx.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *databaseTx) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
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
