package migration

import (
	"context"
	"database/sql"
	"fmt"
	"iter"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

func Down(c *cli.Context, cfg types.Config, dryRun bool) error {
	f, err := getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return err
	}

	g, err := ComputeDefinitions(f)
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
	defer db.Close()

	count := c.Int("count")
	if count < 1 {
		return errors.New("count must be a non-zero positive number")
	}

	return RollbackMigrations(c.Context, db, g, cfg, count, dryRun)
}

// RollbackMigrations is the command that rolls back migrations.
// It rolls back the most recent migration by default,
// or a specific number of migrations if specified.
func RollbackMigrations(ctx context.Context, db database.DB, definitions *graph.Graph, cfg types.Config, count int, dryRun bool) error {
	// Get applied migrations to determine which ones to roll back
	// No retry logic for migrations
	migrationsToRollback, err := getAppliedMigrationsToRollback(ctx, db, cfg.Migration.TableName, count)
	if err != nil {
		return err
	}

	// Filter definitions to include only those that need to be rolled back
	filteredDefinitions, err := filterDefinitionsForRollback(definitions, migrationsToRollback)
	if err != nil {
		return err
	}

	noOfVertices, err := filteredDefinitions.Order()
	if err != nil {
		return errors.Wrap(err, "ordering vertices")
	}
	if noOfVertices == 0 {
		fmt.Printf("%sNo migrations to roll back%s\n", output.StyleInfo, output.StyleReset)
		return nil
	}

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
func getAppliedMigrationsToRollback(ctx context.Context, db database.DB, tableName string, count int) (iter.Seq2[int64, error], error) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "querying migrations to roll back")
	}
	defer rows.Close()

	var migrations = make([]string, 0, count)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, errors.Wrap(err, "scanning migration name")
		}

		migrations = append(migrations, name)
	}

	return func(yield func(int64, error) bool) {
		for _, name := range migrations {
			var ts int64
			_, err := fmt.Sscanf(name, "%d_", &ts)
			if !yield(ts, err) {
				return
			}
		}
	}, nil
}

// filterDefinitionsForRollback filters definitions to match migrations that should be rolled back.
// It creates a new graph containing only the migrations that need to be rolled back while
// preserving their dependency relationships.
func filterDefinitionsForRollback(definitions *graph.Graph, appliedMigrations iter.Seq2[int64, error]) (*graph.Graph, error) {
	// Create a new graph with the same properties as the original
	filteredGraph := graph.New()

	// Add all vertices that need to be rolled back to the filtered graph
	for tsToRollback, err := range appliedMigrations {
		if err != nil {
			return nil, errors.Wrap(err, "invalid name for migration")
		}

		def, err := definitions.GetDefinition(tsToRollback)
		if err != nil {
			return nil, err
		}

		// Successfully found the vertex, add it to our filtered graph
		if err := filteredGraph.AddDefinition(def); err != nil {
			return nil, err
		}
	}

	return filteredGraph, nil
}
