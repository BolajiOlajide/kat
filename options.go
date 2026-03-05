package kat

import (
	"database/sql"
	"time"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/loggr"
)

// MigrationOption is a function that configures a Migration instance.
// Options are applied during migration creation and can return errors
// if the configuration is invalid.
type MigrationOption func(*Migration) error

// WithLogger configures the migration to use a custom logger implementation.
// The logger must implement the Logger interface with Debug, Info, Warn, and Error methods.
//
// Example:
//
//	m, err := kat.New(connStr, fsys, "migrations",
//		kat.WithLogger(customLogger),
//	)
func WithLogger(logger loggr.Logger) MigrationOption {
	return func(m *Migration) error {
		m.logger = logger
		return nil
	}
}

// WithSqlDB configures the migration to use an existing *sql.DB connection
// instead of creating a new one from the connection string.
// When this option is used, the connection string parameter in New() is ignored.
//
// Example:
//
//	db, err := sql.Open("pgx", "postgres://...")
//	if err != nil {
//		return err
//	}
//	m, err := kat.New("", fsys, "migrations",
//		kat.WithSqlDB(db),
//	)
func WithSqlDB(db *sql.DB) MigrationOption {
	return func(m *Migration) error {
		d, err := database.NewWithDB(db, m.logger)
		if err != nil {
			return err
		}
		m.db = d
		return nil
	}
}

// DBConfig holds database connection configuration options  
type DBConfig = database.DBConfig

// DefaultDBConfig returns sensible default configuration for Kat migrations
func DefaultDBConfig() DBConfig {
	return database.DefaultDBConfig()
}

// WithDBConfig configures the migration to use custom database settings.
// This allows fine-tuning of connection timeouts, pool limits, and statement timeouts
// for production deployments.
//
// Example:
//
//	config := kat.DBConfig{
//		ConnectTimeout:   5 * time.Second,
//		StatementTimeout: 5 * time.Minute,
//		MaxOpenConns:     20,
//		MaxIdleConns:     10,
//		ConnMaxLifetime:  1 * time.Hour,
//		DefaultTimeout:   60 * time.Second,
//	}
//	m, err := kat.New(connStr, fsys, "migrations",
//		kat.WithDBConfig(config),
//	)
func WithDBConfig(config DBConfig) MigrationOption {
	return func(m *Migration) error {
		m.dbConfig = &config
		return nil
	}
}

// WithConnectTimeout configures just the connection establishment timeout.
// This is a convenience function for the most common configuration need.
//
// Example:
//
//	m, err := kat.New(connStr, fsys, "migrations",
//		kat.WithConnectTimeout(5 * time.Second),
//	)
func WithConnectTimeout(timeout time.Duration) MigrationOption {
	return func(m *Migration) error {
		if m.dbConfig == nil {
			config := DefaultDBConfig()
			m.dbConfig = &config
		}
		m.dbConfig.ConnectTimeout = timeout
		return nil
	}
}

// WithPoolLimits configures the database connection pool limits.
// This is useful for controlling resource usage in production environments.
//
// Example:
//
//	m, err := kat.New(connStr, fsys, "migrations",
//		kat.WithPoolLimits(20, 10, 1*time.Hour),
//	)
func WithPoolLimits(maxOpen, maxIdle int, connMaxLifetime time.Duration) MigrationOption {
	return func(m *Migration) error {
		if m.dbConfig == nil {
			config := DefaultDBConfig()
			m.dbConfig = &config
		}
		m.dbConfig.MaxOpenConns = maxOpen
		m.dbConfig.MaxIdleConns = maxIdle
		m.dbConfig.ConnMaxLifetime = connMaxLifetime
		return nil
	}
}
