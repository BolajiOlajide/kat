package kat

import (
	"context"

	"github.com/BolajiOlajide/kat/v0/internal/database"
	"github.com/BolajiOlajide/kat/v0/internal/runner"
	"github.com/BolajiOlajide/kat/v0/internal/types"
	"github.com/BolajiOlajide/kat/v0/internal/version"
	"github.com/keegancsmith/sqlf"
)

// Version returns the current version of the application.
// This function delegates to version.Version() to get the version string.
func Version() string {
	return version.Version()
}

// Up runs all pending migrations in the database.
// It takes a context, database connection string, and a slice of migration definitions.
// The migrations are executed in order and tracked in the specified migration table.
func Up(ctx context.Context, connStr string, definitions []types.Definition, tableName string) error {
	db, err := database.NewDB(connStr, sqlf.PostgresBindVar)
	if err != nil {
		return err
	}
	defer db.Close()

	r, err := runner.NewRunner(ctx, db)
	if err != nil {
		return err
	}

	options := runner.Options{
		Operation: types.UpMigrationOperation,
		Definitions: definitions,
		MigrationInfo: types.MigrationInfo{
			TableName: tableName,
		},
	}

	return r.Run(ctx, options)
}

// Down rolls back migrations in the database.
// It takes a context, database connection string, and a slice of migration definitions.
// The migrations are rolled back in reverse order and removed from the migration table.
func Down(ctx context.Context, connStr string, definitions []types.Definition, tableName string) error {
	db, err := database.NewDB(connStr, sqlf.PostgresBindVar)
	if err != nil {
		return err
	}
	defer db.Close()

	r, err := runner.NewRunner(ctx, db)
	if err != nil {
		return err
	}

	options := runner.Options{
		Operation: types.DownMigrationOperation,
		Definitions: definitions,
		MigrationInfo: types.MigrationInfo{
			TableName: tableName,
		},
	}

	return r.Run(ctx, options)
}
