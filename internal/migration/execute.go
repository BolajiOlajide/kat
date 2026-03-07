package migration

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/loggr"
	"github.com/BolajiOlajide/kat/internal/runner"
	"github.com/BolajiOlajide/kat/internal/types"
)

// DBConfigFromCfg builds a database.DBConfig from the config file's timeout settings.
// Falls back to database.DefaultDBConfig() for any unset fields.
func DBConfigFromCfg(cfg types.Config) (database.DBConfig, error) {
	isSQLite := cfg.Database.Driver.IsSQLite()

	dbConfig := database.DefaultDBConfig()

	if isSQLite {
		dbConfig.MaxOpenConns = 1
		dbConfig.MaxIdleConns = 1
		dbConfig.ConnMaxLifetime = 2 * time.Minute
	}

	timeouts, err := cfg.Database.ParseDBTimeouts()
	if err != nil {
		return dbConfig, err
	}
	if timeouts == nil {
		return dbConfig, nil
	}

	if timeouts.ConnectTimeout > 0 {
		dbConfig.ConnectTimeout = timeouts.ConnectTimeout
	}
	if timeouts.StatementTimeout > 0 {
		dbConfig.StatementTimeout = timeouts.StatementTimeout
	}
	if timeouts.MaxOpenConns > 0 {
		dbConfig.MaxOpenConns = timeouts.MaxOpenConns
	}
	if timeouts.MaxIdleConns > 0 {
		dbConfig.MaxIdleConns = timeouts.MaxIdleConns
	}
	if timeouts.ConnMaxLifetime > 0 {
		dbConfig.ConnMaxLifetime = timeouts.ConnMaxLifetime
	}
	if timeouts.DefaultTimeout > 0 {
		dbConfig.DefaultTimeout = timeouts.DefaultTimeout
	}

	return dbConfig, nil
}

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

	dbConfig, err := DBConfigFromCfg(cfg)
	if err != nil {
		return err
	}

	logger := loggr.NewDefault()

	db, err := database.NewWithConfig(cfg.Database.Driver, dbConn, logger, dbConfig)
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

	dbConfig, err := DBConfigFromCfg(cfg)
	if err != nil {
		return err
	}

	logger := loggr.NewDefault()

	db, err := database.NewWithConfig(cfg.Database.Driver, dbConn, logger, dbConfig)
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
