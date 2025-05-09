package kat

import (
	"context"
	"io/fs"

	"github.com/BolajiOlajide/kat/v0/internal/database"
	"github.com/BolajiOlajide/kat/v0/internal/migration"
	"github.com/BolajiOlajide/kat/v0/internal/types"
	"github.com/BolajiOlajide/kat/v0/internal/version"
)

// Version returns the current version of the application.
// This function delegates to version.Version() to get the version string.
func Version() string {
	return version.Version()
}

type Migration struct {
	db                 database.DB
	definitions        []types.Definition
	migrationTableName string
}

func New(connStr string, f fs.FS, migrationTableName string) (*Migration, error) {
	db, err := database.New(connStr)
	if err != nil {
		return nil, err
	}

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
func (m *Migration) Up(ctx context.Context) error {
	return migration.UpWithFS(ctx, m.db, m.definitions, types.Config{
		Migration: types.MigrationInfo{
			TableName: m.migrationTableName,
		},
	}, false)
}

// Down rolls back migrations in the database.
// It takes a context, database connection string, and a slice of migration definitions.
// The migrations are rolled back in reverse order and removed from the migration table.
func (m *Migration) Down(ctx context.Context, count int) error {
	return migration.DownWithFS(ctx, m.db, m.definitions, types.Config{
		Migration: types.MigrationInfo{
			TableName: m.migrationTableName,
		},
	}, count, false)
}
