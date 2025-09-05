// Package database provides database connectivity and operations for PostgreSQL and SQLite.
// It wraps the standard database/sql package with additional functionality
// specific to migration management, including retry logic, transaction handling,
// and migration table management.
//
// The package supports both pgx driver for PostgreSQL and modernc.org/sqlite for SQLite,
// providing abstractions for database operations that are safe for concurrent use.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keegancsmith/sqlf"
	_ "modernc.org/sqlite"

	"github.com/BolajiOlajide/kat/internal/loggr"
)

var _ DB = &database{}

type database struct {
	db      *sql.DB
	bindVar sqlf.BindVar
	logger  loggr.Logger
}

// PingWithRetry pings the database with configurable retries and exponential backoff
func (d *database) PingWithRetry(ctx context.Context, retryCount int, retryDelay int) error {
	// Validate retry parameters
	if retryCount < 0 {
		retryCount = 0 // No retries
	} else if retryCount > 7 {
		retryCount = 7 // Maximum 7 retries
	}

	if retryDelay < 100 {
		retryDelay = 100 // Minimum 100ms delay
	} else if retryDelay > 3000 {
		retryDelay = 3000 // Maximum 3000ms delay
	}

	// If retryCount is 0, just do a simple ping
	if retryCount == 0 {
		return d.db.PingContext(ctx)
	}

	// Otherwise, use retry logic
	return withRetry(d.logger, retryCount, retryDelay, func() error {
		return d.db.PingContext(ctx)
	})
}

// Ping checks if the database connection is alive
func (d *database) Ping(ctx context.Context) error {
	// Regular ping with no retries
	return d.PingWithRetry(ctx, 0, 0)
}

func (d *database) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := d.db.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

func (d *database) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return d.db.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *database) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) WithTransact(ctx context.Context, f func(Tx) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	err = f(&databaseTx{tx: tx, bindVar: d.bindVar})
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// isTransientError determines if an error is likely transient and can be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Try to get the pgconn error for PostgreSQL
	pgErr, ok := err.(*pgconn.PgError)
	if ok {
		// Check PostgreSQL error codes for transient errors
		switch pgErr.Code {
		case
			"08003", // connection_exception
			"08006", // connection_failure
			"08001", // sqlclient_unable_to_establish_sqlconnection
			"08004", // sqlserver_rejected_establishment_of_sqlconnection
			"08007", // connection_failure_during_transaction
			"57P01", // admin_shutdown
			"57P02", // crash_shutdown
			"57P03", // cannot_connect_now
			"53300", // too_many_connections
			"53301": // too_many_connections_for_role
			return true
		}
		return false
	}

	// For SQLite and other databases, check error messages for common transient conditions
	errMsg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errMsg, "database is locked"):
		return true
	case strings.Contains(errMsg, "database is busy"):
		return true
	case strings.Contains(errMsg, "locked:"):
		return true
	case strings.Contains(errMsg, "resource deadlock avoided"):
		return true
	case strings.Contains(errMsg, "database schema has changed"):
		return true
	case strings.Contains(errMsg, "connection refused"):
		return true
	case strings.Contains(errMsg, "connection reset"):
		return true
	case strings.Contains(errMsg, "broken pipe"):
		return true
	default:
		return false
	}
}

// withRetry executes a function with retries for transient errors
func withRetry(l loggr.Logger, retryCount int, initialDelay int, f func() error) error {
	// Validate retry parameters
	if retryCount < 1 {
		retryCount = 1 // Minimum 1 retry
	} else if retryCount > 7 {
		retryCount = 7 // Maximum 7 retries
	}

	if initialDelay < 100 {
		initialDelay = 100 // Minimum 100ms delay
	} else if initialDelay > 3000 {
		initialDelay = 3000 // Maximum 3000ms delay
	}

	var err error
	delay := time.Duration(initialDelay) * time.Millisecond

	for attempt := 0; attempt <= retryCount; attempt++ {
		// First attempt or subsequent retries
		err = f()

		// If no error or non-transient error, return immediately
		if err == nil || !isTransientError(err) {
			return err
		}

		// Don't sleep on the last attempt
		if attempt < retryCount {
			l.Error(fmt.Sprintf("Transient error detected: %s. Retrying in %v (attempt %d/%d)...", err.Error(), delay, attempt+1, retryCount))
			time.Sleep(delay)

			// Exponential backoff: double the delay for the next attempt
			delay *= 2
		}
	}

	// If we reached here, all retries failed
	return errors.Wrapf(err, "failed after %d retries", retryCount)
}

// New returns a new instance of the database
func New(driver, url string, logger loggr.Logger) (DB, error) {
	var sqlDriverName string
	var bindVar sqlf.BindVar

	switch driver {
	case "postgres":
		sqlDriverName = "pgx"
		bindVar = sqlf.PostgresBindVar
	case "sqlite3", "sqlite":
		sqlDriverName = "sqlite"
		bindVar = sqlf.SimpleBindVar
	default:
		return nil, errors.Newf("unsupported database driver: %s", driver)
	}

	db, err := sql.Open(sqlDriverName, url)
	if err != nil {
		return nil, err
	}

	// SQLite allows only one writer at a time, so limit connections to avoid "database is locked" errors
	if driver == "sqlite3" {
		db.SetMaxOpenConns(1)
	}

	return NewWithDB(db, bindVar, logger)
}

func NewWithDB(db *sql.DB, bindVar sqlf.BindVar, logger loggr.Logger) (DB, error) {
	return &database{db: db, bindVar: bindVar, logger: logger}, nil
}
