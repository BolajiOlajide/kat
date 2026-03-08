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
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keegancsmith/sqlf"
	_ "modernc.org/sqlite"

	dbdriver "github.com/BolajiOlajide/kat/internal/database/driver"
	"github.com/BolajiOlajide/kat/internal/loggr"
)

var _ DB = &database{}

// DBConfig holds database connection configuration options
type DBConfig struct {
	// Connection settings
	ConnectTimeout   time.Duration // Timeout for establishing connections
	StatementTimeout time.Duration // Timeout for individual SQL statements

	// Connection pool settings
	MaxOpenConns    int           // Maximum number of open connections
	MaxIdleConns    int           // Maximum number of idle connections
	ConnMaxLifetime time.Duration // Maximum connection lifetime

	// Default timeouts for operations
	DefaultTimeout time.Duration // Default timeout for operations without explicit context deadline
}

// DefaultDBConfig returns sensible default configuration for Kat migrations
// These defaults prioritize backward compatibility while providing basic protection
func DefaultDBConfig(drv dbdriver.DatabaseDriver) DBConfig {
	maxOpenConns := 2
	maxIdleConns := 2
	connMaxLifetime := 5 * time.Minute

	if drv.IsSQLite() {
		maxIdleConns = 1
		maxOpenConns = 1
		connMaxLifetime = 2 * time.Minute
	}

	return DBConfig{
		ConnectTimeout:   10 * time.Second, // Reasonable connection timeout
		StatementTimeout: 0,                // Disabled by default for compatibility
		MaxOpenConns:     maxOpenConns,     // Reasonable limit for migration tool
		MaxIdleConns:     maxIdleConns,     // Conservative idle connection count
		ConnMaxLifetime:  connMaxLifetime,  // Reasonable connection lifetime
		DefaultTimeout:   0,                // Disabled by default for compatibility
	}
}

type database struct {
	db     *sql.DB
	logger loggr.Logger
	config DBConfig
	driver dbdriver.DatabaseDriver
}

// withDefaultTimeout wraps a context with a default timeout if none is set and timeout is > 0
func (d *database) withDefaultTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {} // Return no-op cancel function
	}
	if timeout <= 0 {
		return ctx, func() {} // No timeout configured
	}
	return context.WithTimeout(ctx, timeout)
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

	// Apply default timeout if not set
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()

	// If retryCount is 0, just do a simple ping
	if retryCount == 0 {
		return d.db.PingContext(ctx)
	}

	// Otherwise, use retry logic
	return withRetry(ctx, d.logger, retryCount, retryDelay, func(retryCtx context.Context) error {
		return d.db.PingContext(retryCtx)
	})
}

func (d *database) Driver() dbdriver.DatabaseDriver {
	return d.driver
}

// Ping checks if the database connection is alive
func (d *database) Ping(ctx context.Context) error {
	// Regular ping with no retries
	return d.PingWithRetry(ctx, 0, 0)
}

func (d *database) Exec(ctx context.Context, query *sqlf.Query) error {
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()

	_, err := d.db.ExecContext(ctx, query.Query(d.driver.BindVar()), query.Args()...)
	return err
}

func (d *database) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	// Don't apply default timeout here — the context must remain valid
	// until the caller calls Row.Scan()
	return d.db.QueryRowContext(ctx, query.Query(d.driver.BindVar()), query.Args()...)
}

func (d *database) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	// For Query operations, we don't apply default timeout as the context
	// needs to remain valid for the lifetime of the Rows
	// Users should set their own timeout if needed
	return d.db.QueryContext(ctx, query.Query(d.driver.BindVar()), query.Args()...)
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) WithTransact(ctx context.Context, f func(Tx) error) error {
	if f == nil {
		return errors.New("WithTransact: nil callback")
	}

	// Apply default timeout if not set
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err = f(&databaseTx{tx: tx, driver: d.driver, config: d.config}); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Wrapf(err, "transaction failed; rollback also failed: %v", rbErr)
		}
		return errors.Wrap(err, "transaction failed")
	}

	if commitErr := tx.Commit(); commitErr != nil {
		_ = tx.Rollback() // ensure connection cleanup
		return errors.Wrap(commitErr, "failed to commit transaction")
	}

	return nil
}

// isTransientError determines if an error is likely transient and can be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Try to get the pgconn error, handling wrapped errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
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
func withRetry(ctx context.Context, l loggr.Logger, retryCount int, initialDelay int, f func(context.Context) error) error {
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
		// Check context before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// First attempt or subsequent retries
		err = f(ctx)

		// If no error or non-transient error, return immediately
		if err == nil || !isTransientError(err) {
			return err
		}

		// Don't sleep on the last attempt
		if attempt < retryCount {
			l.Error(fmt.Sprintf("Transient error detected: %s. Retrying in %v (attempt %d/%d)...", err.Error(), delay, attempt+1, retryCount))

			// Add jitter to prevent thundering herd
			jitter := time.Duration(rand.Int64N(int64(delay / 4)))
			actualDelay := delay + jitter

			// Sleep with context cancellation awareness
			select {
			case <-time.After(actualDelay):
				// Continue to next attempt
			case <-ctx.Done():
				return ctx.Err()
			}

			// Exponential backoff: double the delay for the next attempt
			delay *= 2
		}
	}

	// If we reached here, all retries failed
	return errors.Wrapf(err, "failed after %d retries", retryCount)
}

// isKeyValueDSN detects whether a connection string is in key=value format
// (e.g. "host=localhost port=5432") as opposed to URL format.
func isKeyValueDSN(dsn string) bool {
	return !strings.Contains(dsn, "://")
}

// ensureTimeoutsInDSN adds timeout parameters to the connection string if not present.
// Only applies to PostgreSQL connection strings. Supports both URL and key=value formats.
func ensureTimeoutsInDSN(connStr string, connectTimeout, statementTimeout time.Duration) (string, error) {
	if isKeyValueDSN(connStr) {
		return ensureTimeoutsInKeyValueDSN(connStr, connectTimeout, statementTimeout), nil
	}

	u, err := url.Parse(connStr)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse connection URL")
	}

	query := u.Query()

	// Only add connect_timeout if not already present and timeout > 0
	if connectTimeout > 0 && query.Get("connect_timeout") == "" {
		timeoutSeconds := max(int(connectTimeout.Seconds()), 1)
		query.Set("connect_timeout", strconv.Itoa(timeoutSeconds))
	}

	// Only add statement_timeout if not already present and timeout > 0
	if statementTimeout > 0 && query.Get("statement_timeout") == "" {
		timeoutMs := int(statementTimeout.Milliseconds())
		query.Set("statement_timeout", strconv.Itoa(timeoutMs))
	}

	u.RawQuery = query.Encode()
	return u.String(), nil
}

// ensureTimeoutsInKeyValueDSN adds timeout parameters to a key=value DSN string.
func ensureTimeoutsInKeyValueDSN(dsn string, connectTimeout, statementTimeout time.Duration) string {
	if connectTimeout > 0 && !strings.Contains(dsn, "connect_timeout=") {
		timeoutSeconds := max(int(connectTimeout.Seconds()), 1)
		dsn += fmt.Sprintf(" connect_timeout=%d", timeoutSeconds)
	}
	if statementTimeout > 0 && !strings.Contains(dsn, "statement_timeout=") {
		timeoutMs := int(statementTimeout.Milliseconds())
		dsn += fmt.Sprintf(" statement_timeout=%d", timeoutMs)
	}
	return dsn
}

// NewWithConfig returns a new database instance with custom configuration
func NewWithConfig(dd dbdriver.DatabaseDriver, url string, logger loggr.Logger, config DBConfig) (DB, error) {
	var err error
	finalURL := url
	// Only apply DSN timeout enrichment for PostgreSQL
	if dd.IsPostgres() {
		finalURL, err = ensureTimeoutsInDSN(url, config.ConnectTimeout, config.StatementTimeout)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open(dd.DriverName(), finalURL)
	if err != nil {
		return nil, err
	}

	if dd.IsSQLite() {
		// SQLite allows only one writer at a time
		db.SetMaxOpenConns(1)
		maxIdle := min(config.MaxIdleConns, 1)
		db.SetMaxIdleConns(maxIdle)
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		db.SetMaxOpenConns(config.MaxOpenConns)
		db.SetMaxIdleConns(config.MaxIdleConns)
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// Test the connection with timeout if configured
	var ctx context.Context
	var cancel context.CancelFunc = func() {}
	if config.ConnectTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), config.ConnectTimeout)
	} else {
		ctx = context.Background()
	}
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		timeoutMsg := "no timeout"
		if config.ConnectTimeout > 0 {
			timeoutMsg = fmt.Sprintf("within %v", config.ConnectTimeout)
		}
		return nil, errors.Wrapf(err, "failed to establish database connection %s", timeoutMsg)
	}

	logger.Info(fmt.Sprintf("Database connection established - Pool: max_open=%d, max_idle=%d, max_lifetime=%v",
		config.MaxOpenConns, config.MaxIdleConns, config.ConnMaxLifetime))

	d := &database{
		db:     db,
		logger: logger,
		config: config,
		driver: dd,
	}

	if config.StatementTimeout > 0 {
		logger.Info(fmt.Sprintf("Session statement timeout configured for %v", config.StatementTimeout))
	}

	return d, nil
}

// New returns a new instance of the database with default configuration
func New(drv dbdriver.DatabaseDriver, url string, logger loggr.Logger) (DB, error) {
	return NewWithConfig(drv, url, logger, DefaultDBConfig(drv))
}

func NewWithDB(db *sql.DB, driver dbdriver.DatabaseDriver, logger loggr.Logger) (DB, error) {
	// SQLite allows only one writer at a time. Enforce this on externally
	// supplied connections to prevent "database is locked" errors.
	if driver.IsSQLite() {
		db.SetMaxOpenConns(1)
	}

	return &database{
		db:     db,
		driver: driver,
		logger: logger,
		config: DefaultDBConfig(driver),
	}, nil
}
