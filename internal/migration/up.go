package migration

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Up is the command that runs the up migration operation.
func Up(c *cli.Context, cfg types.Config, dryRun bool) error {
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
	defer db.Close()

	if err := ApplyMigrations(c.Context, db, definitions, cfg, dryRun); err != nil {
		return err
	}
	return nil
}

func ApplyMigrations(ctx context.Context, db database.DB, definitions *graph.Graph, cfg types.Config, dryRun bool) error {
	r, err := runner.NewRunner(ctx, db)
	if err != nil {
		return errors.Wrap(err, "initializing runner")
	}

	return r.Run(ctx, runner.Options{
		Operation:     types.UpMigrationOperation,
		Definitions:   definitions,
		MigrationInfo: cfg.Migration,
		DryRun:        dryRun,
		Verbose:       cfg.Verbose,
	})
}
