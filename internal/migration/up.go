package migration

import (
    "context"
    "github.com/dominikbraun/graph"

    "github.com/cockroachdb/errors"
    "github.com/urfave/cli/v2"

    "github.com/BolajiOlajide/kat/internal/database"
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

    return UpWithFS(c.Context, db, definitions, cfg, dryRun)
}

func UpWithFS(ctx context.Context, db database.DB, definitions graph.Graph[int64, types.Definition], cfg types.Config, dryRun bool) error {
    defer db.Close()

    // No retry for migrations, just basic connection
    r, err := runner.NewRunner(ctx, db)
    if err != nil {
        return errors.Wrap(err, "connecting to database")
    }

    return r.Run(ctx, runner.Options{
        Operation:     types.UpMigrationOperation,
        Definitions:   definitions,
        MigrationInfo: cfg.Migration,
        DryRun:        dryRun,
        Verbose:       cfg.Verbose,
    })
}
