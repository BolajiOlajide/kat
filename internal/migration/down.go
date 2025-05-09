package migration

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/v0/internal/database"
	"github.com/BolajiOlajide/kat/v0/internal/output"
	"github.com/BolajiOlajide/kat/v0/internal/runner"
	"github.com/BolajiOlajide/kat/v0/internal/types"
)

func Down(c *cli.Context, cfg types.Config, dryRun bool) error {
	f, err := getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return err
	}

	definitions, err := ComputeDefinitions(f)
	if err != nil {
		return err
	}

	dbConn, err := cfg.Database.ConnString()
	if err != nil {
		return err
	}

	db, err := database.New(dbConn)
	if err != nil {
		return err
	}

	count := c.Int("count")
	if count < 1 {
		return errors.New("count must be a positive number")
	}

	return DownWithFS(c.Context, db, definitions, cfg, count, dryRun)
}

// Down is the command that rolls back migrations.
// It rolls back the most recent migration by default,
// or a specific number of migrations if specified.
func DownWithFS(ctx context.Context, db database.DB, definitions []types.Definition, cfg types.Config, count int, dryRun bool) error {
	defer db.Close()

	// Get applied migrations to determine which ones to roll back
	// No retry logic for migrations
	migrationsToRollback, err := getAppliedMigrationsToRollback(ctx, db, cfg.Migration.TableName, count)
	if err != nil {
		return err
	}

	if len(migrationsToRollback) == 0 {
		fmt.Printf("%sNo migrations to roll back%s\n", output.StyleInfo, output.StyleReset)
		return nil
	}

	// Filter definitions to include only those that need to be rolled back
	filteredDefinitions := filterDefinitionsForRollback(definitions, migrationsToRollback)

	// No retry for migrations
	r, err := runner.NewRunner(ctx, db)
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}

	return r.Run(ctx, runner.Options{
		Operation:     types.DownMigrationOperation,
		Definitions:   filteredDefinitions,
		MigrationInfo: cfg.Migration,
		DryRun:        dryRun,
		Verbose:       cfg.Verbose,
	})
}

// getAppliedMigrationsToRollback returns the names of migrations that should be rolled back
func getAppliedMigrationsToRollback(ctx context.Context, db database.DB, tableName string, count int) ([]string, error) {
	// Check if the migrations table exists
	var exists bool
	if err := db.QueryRow(ctx, sqlf.Sprintf(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = %s)",
		tableName,
	)).Scan(&exists); err != nil {
		return nil, errors.Wrap(err, "checking if migrations table exists")
	}

	if !exists {
		return nil, nil // No migrations table means no migrations to roll back
	}

	// Get the most recent migrations in descending order
	query := sqlf.Sprintf(
		"SELECT name FROM %s ORDER BY migration_time DESC LIMIT %d",
		sqlf.Sprintf(tableName),
		count,
	)

	rows, err := db.Query(ctx, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "querying migrations to roll back")
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, errors.Wrap(err, "scanning migration name")
		}
		migrations = append(migrations, name)
	}

	return migrations, nil
}

// filterDefinitionsForRollback filters and sorts definitions to match migrations that should be rolled back
func filterDefinitionsForRollback(definitions []types.Definition, migrationsToRollback []string) []types.Definition {
	// Create a map for quick lookups
	migrationMap := make(map[string]bool)
	for _, name := range migrationsToRollback {
		migrationMap[name] = true
	}

	// Filter definitions to only include those that need to be rolled back
	var filteredDefinitions []types.Definition
	for _, def := range definitions {
		if migrationMap[def.Name] {
			filteredDefinitions = append(filteredDefinitions, def)
		}
	}

	// Sort in reverse order of creation (newest first) to roll back in the correct order
	for i, j := 0, len(filteredDefinitions)-1; i < j; i, j = i+1, j-1 {
		filteredDefinitions[i], filteredDefinitions[j] = filteredDefinitions[j], filteredDefinitions[i]
	}

	return filteredDefinitions
}
