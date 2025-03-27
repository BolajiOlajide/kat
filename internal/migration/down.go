package migration

import (
	"database/sql"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Down is the command that rolls back migrations.
// It rolls back the most recent migration by default,
// or a specific number of migrations if specified.
func Down(c *cli.Context, cfg types.Config, dryRun bool, skipValidation bool) error {
	fs, err := getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return err
	}

	definitions, err := computeDefinitions(fs)
	if err != nil {
		return err
	}

	dbConn, err := cfg.Database.ConnString()
	if err != nil {
		return err
	}

	db, err := database.NewDB(dbConn, sqlf.PostgresBindVar)
	if err != nil {
		return err
	}
	defer db.Close()

	count := c.Int("count")
	if count < 1 {
		return errors.New("count must be a positive number")
	}

	// Get applied migrations to determine which ones to roll back
	// No retry logic for migrations
	migrationsToRollback, err := getAppliedMigrationsToRollback(c, db, cfg.Migration.TableName, count)

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
	r, err := runner.NewRunner(c.Context, db)
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}

	return r.Run(c.Context, runner.Options{
		Operation:      types.DownMigrationOperation,
		Definitions:    filteredDefinitions,
		MigrationInfo:  cfg.Migration,
		DryRun:         dryRun,
		SkipValidation: skipValidation,
	})
}

// getAppliedMigrationsToRollback returns the names of migrations that should be rolled back
func getAppliedMigrationsToRollback(c *cli.Context, db database.DB, tableName string, count int) ([]string, error) {
	// Check if the migrations table exists
	var exists bool
	err := db.QueryRow(c.Context, sqlf.Sprintf(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = %s)",
		tableName,
	)).Scan(&exists)

	if err != nil {
		return nil, errors.Wrap(err, "checking if migrations table exists")
	}

	if !exists {
		return nil, nil // No migrations table means no migrations to roll back
	}

	// Get the most recent migrations in descending order
	query := sqlf.Sprintf(
		"SELECT name FROM %s ORDER BY migration_time DESC LIMIT %d",
		sqlf.Sprintf("%s", tableName),
		count,
	)

	rows, err := db.Query(c.Context, query)
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
