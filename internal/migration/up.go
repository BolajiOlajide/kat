package migration

import (
	"github.com/keegancsmith/sqlf"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Up is the command that runs the up migration operation.
func Up(c *cli.Context, cfg types.Config) error {
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

	r, err := runner.NewRunner(c.Context, db)
	if err != nil {
		return err
	}

	return r.Run(c.Context, runner.Options{
		Operation:     types.UpMigrationOperation,
		Definitions:   definitions,
		MigrationInfo: cfg.Migration,
	})
}
