// Package kat provides a lightweight, powerful CLI tool for PostgreSQL database migrations.
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
//	// Create a new migration manager
//	m, err := kat.New("postgres://user:pass@localhost:5432/db", fsys, "migrations")
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
	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Migration manages database schema migrations using a graph-based approach.
// It tracks applied migrations in a database table and ensures dependencies
// are respected when applying or rolling back migrations.
type Migration struct {
	db                 database.DB
	definitions        *graph.Graph
	migrationTableName string
}

// New creates a new Migration instance with a database connection string.
// It establishes a connection to the PostgreSQL database and loads migration
// definitions from the provided filesystem.
//
// Parameters:
//   - connStr: PostgreSQL connection string (e.g., "postgres://user:pass@host:port/db")
//   - f: Filesystem containing migration directories
//   - migrationTableName: Name of the table to track applied migrations
//
// Returns a Migration instance or an error if connection fails or migration
// definitions cannot be loaded.
func New(connStr string, f fs.FS, migrationTableName string) (*Migration, error) {
	db, err := database.New(connStr)
	if err != nil {
		return nil, err
	}

	return newMigration(db, f, migrationTableName)
}

// NewWithDB creates a new Migration instance using an existing database connection.
// This is useful when you already have a *sql.DB instance and want to reuse it
// for migrations.
//
// Parameters:
//   - db: Existing *sql.DB connection to PostgreSQL
//   - f: Filesystem containing migration directories
//   - migrationTableName: Name of the table to track applied migrations
//
// Returns a Migration instance or an error if the database wrapper cannot be
// created or migration definitions cannot be loaded.
func NewWithDB(db *sql.DB, f fs.FS, migrationTableName string) (*Migration, error) {
	d, err := database.NewWithDB(db)
	if err != nil {
		return nil, err
	}
	return newMigration(d, f, migrationTableName)
}

func newMigration(db database.DB, f fs.FS, migrationTableName string) (*Migration, error) {
	definitions, err := migration.ComputeDefinitions(f)
	if err != nil {
		return nil, err
	}

	return &Migration{
		db:                 db,
		definitions:        definitions,
		migrationTableName: migrationTableName,
	}, nil
}

// Up applies pending migrations to the database.
// Migrations are executed in dependency order as determined by the migration graph.
// Each migration runs within a transaction for safety.
//
// Parameters:
//   - ctx: Context for the operation (supports cancellation)
//   - count: Number of migrations to apply (0 means apply all pending)
//
// Returns an error if count is negative, if any migration fails, or if
// database operations fail. Applied migrations are tracked in the migration table.
func (m *Migration) Up(ctx context.Context, count int) error {
	if count < 0 {
		return errors.New("count cannot be a negative number")
	}

	cfg := types.Config{Migration: types.MigrationInfo{TableName: m.migrationTableName}}
	return migration.Execute(ctx, m.db, m.definitions, cfg, count, types.UpMigrationOperation, false)
}

// Down rolls back applied migrations from the database.
// Migrations are rolled back in reverse dependency order. Each rollback
// runs within a transaction and removes the migration record from the tracking table.
//
// Parameters:
//   - ctx: Context for the operation (supports cancellation)
//   - count: Number of migrations to roll back (must be positive)
//
// Returns an error if count is not a positive number, if any rollback fails,
// or if database operations fail. Successfully rolled back migrations are
// removed from the migration table.
func (m *Migration) Down(ctx context.Context, count int) error {
	if count < 1 {
		return errors.New("count must be a non-zero positive number")
	}

	cfg := types.Config{Migration: types.MigrationInfo{TableName: m.migrationTableName}}
	return migration.Execute(ctx, m.db, m.definitions, cfg, count, types.DownMigrationOperation, false)
}
