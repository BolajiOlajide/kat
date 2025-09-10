// Package database provides database connectivity and operations for PostgreSQL.
// It wraps the standard database/sql package with additional functionality
// specific to migration management, including retry logic, transaction handling,
// and migration table management.
//
// The package uses pgx driver for PostgreSQL connectivity and provides
// abstractions for database operations that are safe for concurrent use.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keegancsmith/sqlf"

	"github.com/BolajiOlajide/kat/internal/loggr"
)

func init() {
	// Seed random number generator for jitter in retry logic
	rand.Seed(time.Now().UnixNano())
}

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
func DefaultDBConfig() DBConfig {
	return DBConfig{
		ConnectTimeout:   10 * time.Second, // Reasonable connection timeout
		StatementTimeout: 0,                // Disabled by default for compatibility
		MaxOpenConns:     10,               // Reasonable limit for migration tool
		MaxIdleConns:     5,                // Conservative idle connection count
		ConnMaxLifetime:  30 * time.Minute, // Reasonable connection lifetime
		DefaultTimeout:   0,                // Disabled by default for compatibility
	}
}

type database struct {
	db      *sql.DB
	bindVar sqlf.BindVar
	logger  loggr.Logger
	config  DBConfig
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

// Ping checks if the database connection is alive
func (d *database) Ping(ctx context.Context) error {
	// Regular ping with no retries
	return d.PingWithRetry(ctx, 0, 0)
}

func (d *database) Exec(ctx context.Context, query *sqlf.Query) error {
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()
	
	_, err := d.db.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

func (d *database) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	ctx, cancel := d.withDefaultTimeout(ctx, d.config.DefaultTimeout)
	defer cancel()
	
	return d.db.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *database) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	// For Query operations, we don't apply default timeout as the context
	// needs to remain valid for the lifetime of the Rows
	// Users should set their own timeout if needed
	return d.db.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
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



	if err = f(&databaseTx{tx: tx, bindVar: d.bindVar, config: d.config}); err != nil {
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
	if !errors.As(err, &pgErr) {
		return false
	}

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
			jitter := time.Duration(rand.Int63n(int64(delay / 4)))
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

// ensureTimeoutsInDSN adds timeout parameters to the connection URL if not present
func ensureTimeoutsInDSN(connURL string, connectTimeout, statementTimeout time.Duration) (string, error) {
	u, err := url.Parse(connURL)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse connection URL")
	}

	query := u.Query()
	
	// Only add connect_timeout if not already present and timeout > 0
	if connectTimeout > 0 && query.Get("connect_timeout") == "" {
		timeoutSeconds := int(connectTimeout.Seconds())
		if timeoutSeconds < 1 {
			timeoutSeconds = 1 // Minimum 1 second
		}
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

// NewWithConfig returns a new database instance with custom configuration
func NewWithConfig(url string, logger loggr.Logger, config DBConfig) (DB, error) {
	// Ensure timeouts are set in the DSN
	finalURL, err := ensureTimeoutsInDSN(url, config.ConnectTimeout, config.StatementTimeout)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", finalURL)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

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
		db:      db, 
		bindVar: sqlf.PostgresBindVar, 
		logger:  logger,
		config:  config,
	}

	// Set session-level statement timeout if configured
	// We'll add it to the DSN instead of running SET, for reliability
	if config.StatementTimeout > 0 {
		logger.Info(fmt.Sprintf("Session statement timeout configured for %v", config.StatementTimeout))
	}

	return d, nil
}

// New returns a new instance of the database with default configuration
func New(url string, logger loggr.Logger) (DB, error) {
	return NewWithConfig(url, logger, DefaultDBConfig())
}

func NewWithDB(db *sql.DB, logger loggr.Logger) (DB, error) {
	// Apply default pool configuration to provided db
	config := DefaultDBConfig()
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	
	return &database{
		db:      db, 
		bindVar: sqlf.PostgresBindVar, 
		logger:  logger,
		config:  config,
	}, nil
}
