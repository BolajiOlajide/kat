package kat

import (
	"time"

	"github.com/cockroachdb/errors"
)

// MigrationOption is a function that configures migration settings.
// Options are applied during construction and only affect configuration,
// not connection source (use New vs NewWithDB for that).
type MigrationOption func(*migrationConfig) error

// WithLogger configures the migration to use a custom logger implementation.
// The logger must implement the Logger interface with Debug, Info, Warn, and Error methods.
//
// Example:
//
//	m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
//		kat.WithLogger(customLogger),
//	)
func WithLogger(logger Logger) MigrationOption {
	return func(cfg *migrationConfig) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		cfg.logger = logger
		return nil
	}
}

// WithDBConfig configures the migration to use custom database settings.
// This allows fine-tuning of connection timeouts, pool limits, and statement timeouts
// for production deployments. Only applicable when using New (not NewWithDB).
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
//	m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
//		kat.WithDBConfig(config),
//	)
func WithDBConfig(config DBConfig) MigrationOption {
	return func(cfg *migrationConfig) error {
		cfg.dbConfig = &config
		return nil
	}
}

// WithConnectTimeout configures just the connection establishment timeout.
// This is a convenience function for the most common configuration need.
// Only applicable when using New (not NewWithDB).
//
// Example:
//
//	m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
//		kat.WithConnectTimeout(5 * time.Second),
//	)
func WithConnectTimeout(timeout time.Duration) MigrationOption {
	return func(cfg *migrationConfig) error {
		if cfg.dbConfig == nil {
			config := DefaultDBConfig()
			cfg.dbConfig = &config
		}
		cfg.dbConfig.ConnectTimeout = timeout
		return nil
	}
}

// WithPoolLimits configures the database connection pool limits.
// This is useful for controlling resource usage in production environments.
// Only applicable when using New (not NewWithDB).
//
// Example:
//
//	m, err := kat.New(kat.PostgresDriver, connStr, fsys, "migrations",
//		kat.WithPoolLimits(20, 10, 1*time.Hour),
//	)
func WithPoolLimits(maxOpen, maxIdle int, connMaxLifetime time.Duration) MigrationOption {
	return func(cfg *migrationConfig) error {
		if cfg.dbConfig == nil {
			config := DefaultDBConfig()
			cfg.dbConfig = &config
		}
		cfg.dbConfig.MaxOpenConns = maxOpen
		cfg.dbConfig.MaxIdleConns = maxIdle
		cfg.dbConfig.ConnMaxLifetime = connMaxLifetime
		return nil
	}
}
