// Package kat provides a lightweight, powerful CLI tool for PostgreSQL and SQLite database migrations.
//
// Kat allows you to manage your database schema using SQL files with a simple,
// intuitive workflow. It features:
//
//   - Simple SQL Migrations: Write raw SQL for both up and down migrations
//   - Graph-Based Migration System: Manages parent-child relationships between migrations
//   - Explicit Dependencies: Migrations can declare parent dependencies
//   - Transaction Support: Migrations run within transactions for safety
//   - Migration Tracking: Applied migrations are recorded in a database table
//   - Rollback Support: Easily revert migrations
//
// Basic usage:
//
//	// Create a new migration manager with a connection string
//	m, err := kat.New(kat.PostgresDriver, "postgres://user:pass@localhost:5432/db", fsys, "migrations")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Or use an existing *sql.DB connection
//	m, err = kat.NewWithDB(kat.PostgresDriver, db, fsys, "migrations")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Apply all pending migrations
//	err = m.Up(context.Background(), 0)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Roll back the most recent migration
//	err = m.Down(context.Background(), 1)
//	if err != nil {
//		log.Fatal(err)
//	}
package kat

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/loggr"
	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/types"
)

// migrationConfig holds configuration gathered from options before construction.
type migrationConfig struct {
	logger   Logger
	dbConfig *DBConfig
}

func defaultConfig() migrationConfig {
	return migrationConfig{
		logger: loggr.NewDefault(),
	}
}

func applyOptions(opts []MigrationOption, cfg *migrationConfig) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

// Migration manages database schema migrations using a graph-based approach.
// It tracks applied migrations in a database table and ensures dependencies
// are respected when applying or rolling back migrations.
type Migration struct {
	db                 database.DB
	definitions        *graph.Graph
	migrationTableName string
	logger             Logger
}

// New creates a new Migration instance that opens a database connection from a connection string.
//
// Parameters:
//   - drv: Database driver (kat.PostgresDriver or kat.SQLiteDriver)
//   - connStr: Database connection string (e.g., "postgres://user:pass@host:port/db")
//   - f: Filesystem containing migration directories
//   - migrationTableName: Name of the table to track applied migrations
//   - options: Optional configuration (WithLogger, WithDBConfig, WithConnectTimeout, WithPoolLimits)
func New(drv Driver, connStr string, f fs.FS, migrationTableName string, options ...MigrationOption) (*Migration, error) {
	if !drv.Valid() {
		return nil, errors.New("driver must be one of `sqlite` or `postgres`")
	}

	definitions, err := migration.ComputeDefinitions(f)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig()
	if err := applyOptions(options, &cfg); err != nil {
		return nil, err
	}

	dbConfig := DefaultDBConfig()
	if cfg.dbConfig != nil {
		dbConfig = *cfg.dbConfig
	}

	db, err := database.NewWithConfig(drv, connStr, cfg.logger, dbConfig)
	if err != nil {
		return nil, err
	}

	return &Migration{
		db:                 db,
		definitions:        definitions,
		migrationTableName: migrationTableName,
		logger:             cfg.logger,
	}, nil
}

// NewWithDB creates a new Migration instance using an existing *sql.DB connection.
// The caller is responsible for managing the connection's lifecycle and pool settings.
// For SQLite connections, kat will set MaxOpenConns to 1 to avoid "database is locked" errors.
//
// Parameters:
//   - drv: Database driver (kat.PostgresDriver or kat.SQLiteDriver)
//   - sqlDB: An existing *sql.DB connection
//   - f: Filesystem containing migration directories
//   - migrationTableName: Name of the table to track applied migrations
//   - options: Optional configuration (WithLogger)
func NewWithDB(drv Driver, sqlDB *sql.DB, f fs.FS, migrationTableName string, options ...MigrationOption) (*Migration, error) {
	if !drv.Valid() {
		return nil, errors.New("driver must be one of `sqlite` or `postgres`")
	}

	definitions, err := migration.ComputeDefinitions(f)
	if err != nil {
		return nil, err
	}

	cfg := defaultConfig()
	if err := applyOptions(options, &cfg); err != nil {
		return nil, err
	}

	if drv.IsSQLite() {
		sqlDB.SetMaxOpenConns(1)
	}

	db, err := database.NewWithDB(sqlDB, drv.BindVar(), cfg.logger)
	if err != nil {
		return nil, err
	}

	return &Migration{
		db:                 db,
		definitions:        definitions,
		migrationTableName: migrationTableName,
		logger:             cfg.logger,
	}, nil
}

// Up applies pending migrations to the database.
// Migrations are executed in dependency order as determined by the migration graph.
// Each migration runs within a transaction for safety.
//
// Parameters:
//   - ctx: Context for the operation (supports cancellation)
//   - count: Number of migrations to apply (0 means apply all pending)
func (m *Migration) Up(ctx context.Context, count int) error {
	if count < 0 {
		return errors.New("count cannot be a negative number")
	}

	cfg := types.Config{Migration: types.MigrationInfo{TableName: m.migrationTableName}}
	return migration.Execute(ctx, m.db, m.logger, m.definitions, cfg, count, types.UpMigrationOperation, false)
}

// Down rolls back applied migrations from the database.
// Migrations are rolled back in reverse dependency order. Each rollback
// runs within a transaction and removes the migration record from the tracking table.
//
// Parameters:
//   - ctx: Context for the operation (supports cancellation)
//   - count: Number of migrations to roll back (must be positive)
func (m *Migration) Down(ctx context.Context, count int) error {
	if count < 1 {
		return errors.New("count must be a non-zero positive number")
	}

	cfg := types.Config{Migration: types.MigrationInfo{TableName: m.migrationTableName}}
	return migration.Execute(ctx, m.db, m.logger, m.definitions, cfg, count, types.DownMigrationOperation, false)
}
