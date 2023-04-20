package migration

import (
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Up is the command that runs the up migration operation.
func Up(c *cli.Context, cfg types.Config) error {
	ctx := c.Context

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

	db, err := database.NewDB(dbConn)
	if err != nil {
		return err
	}
	defer db.Close()

	r, err := runner.NewRunner(ctx, db)
	if err != nil {
		return err
	}

	return r.Run(ctx, runner.Options{
		Operation:     types.UpMigrationOperation,
		Definitions:   definitions,
		MigrationInfo: cfg.Migration,
	})
}
