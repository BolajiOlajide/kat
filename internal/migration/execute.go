package migration

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/loggr"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Up is the command that runs the up migration operation.
func Up(c *cli.Context, cfg types.Config, dryRun bool) error {
	count := c.Int("count")
	if count < 0 {
		return errors.New("count cannot be a negative number")
	}

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

	logger := loggr.NewDefault()

	db, err := database.New(dbConn, logger)
	if err != nil {
		return err
	}
	defer db.Close()

	return Execute(c.Context, db, logger, definitions, cfg, count, types.UpMigrationOperation, dryRun)
}

func Down(c *cli.Context, cfg types.Config, dryRun bool) error {
	count := c.Int("count")
	if count < 1 {
		return errors.New("count must be a non-zero positive number")
	}

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

	logger := loggr.NewDefault()

	db, err := database.New(dbConn, logger)
	if err != nil {
		return err
	}
	defer db.Close()

	return Execute(c.Context, db, logger, g, cfg, count, types.DownMigrationOperation, dryRun)
}

func Execute(ctx context.Context, db database.DB, logger loggr.Logger, definitions *graph.Graph, cfg types.Config, count int, op types.MigrationOperationType, dryRun bool) error {
	r, err := runner.NewRunner(ctx, db, logger)
	if err != nil {
		return errors.Wrap(err, "initializing runner")
	}

	return r.Run(ctx, runner.Options{
		Operation:     op,
		Definitions:   definitions,
		MigrationInfo: cfg.Migration,
		DryRun:        dryRun,
		Verbose:       cfg.Verbose,
		Count:         count,
	})
}
