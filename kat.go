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

type Migration struct {
	db                 database.DB
	definitions        *graph.Graph
	migrationTableName string
}

func New(connStr string, f fs.FS, migrationTableName string) (*Migration, error) {
	db, err := database.New(connStr)
	if err != nil {
		return nil, err
	}

	return newMigration(db, f, migrationTableName)
}

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

// Up runs all pending migrations in the database.
// It takes a context, database connection string, and a slice of migration definitions.
// The migrations are executed in order and tracked in the specified migration table.
func (m *Migration) Up(ctx context.Context, count int) error {
	if count < 0 {
		return errors.New("count cannot be a negative number")
	}

	cfg := types.Config{Migration: types.MigrationInfo{TableName: m.migrationTableName}}
	return migration.Execute(ctx, m.db, m.definitions, cfg, count, types.UpMigrationOperation, false)
}

// Down rolls back migrations in the database.
// It takes a context, database connection string, and a slice of migration definitions.
// The migrations are rolled back in reverse order and removed from the migration table.
func (m *Migration) Down(ctx context.Context, count int) error {
	if count < 1 {
		return errors.New("count must be a non-zero positive number")
	}

	cfg := types.Config{Migration: types.MigrationInfo{TableName: m.migrationTableName}}
	return migration.Execute(ctx, m.db, m.definitions, cfg, count, types.DownMigrationOperation, false)
}
