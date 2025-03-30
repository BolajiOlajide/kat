package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
	"github.com/keegancsmith/sqlf"
)

func add(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) == 0 {
		return cli.Exit("no migration name specified", 1)
	}
	if len(args) != 1 {
		return cli.Exit("too many arguments", 1)
	}

	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}
	return migration.Add(args[0], cfg)
}

func up(c *cli.Context) error {
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	// Get command flags
	dryRun := c.Bool("dry-run")

	if dryRun {
		fmt.Fprintf(os.Stdout, "%sDRY RUN: Migrations will not be applied%s\n", output.StyleInfo, output.StyleReset)
	}

	// Note: Retry is not used for migrations, only for the ping command
	return migration.Up(c, cfg, dryRun)
}

func down(c *cli.Context) error {
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	// Get command flags
	dryRun := c.Bool("dry-run")

	if dryRun {
		fmt.Fprintf(os.Stdout, "%sDRY RUN: Migrations will not be rolled back%s\n", output.StyleInfo, output.StyleReset)
	}

	// Note: Retry is not used for migrations, only for the ping command
	return migration.Down(c, cfg, dryRun)
}

func initialize(c *cli.Context) error {
	return migration.Init(c)
}

func getVersion(c *cli.Context) error {
	fmt.Fprintf(os.Stdout, "%sVersion: %s%s\n", output.StyleInfo, version.Version(), output.StyleReset)
	return nil
}

func ping(c *cli.Context) error {
	// Get database configuration
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	// Create database connection string
	dbConn, err := cfg.Database.ConnString()
	if err != nil {
		return err
	}

	// Get retry parameters
	retryCount := c.Int("retry-count")
	retryDelay := c.Int("retry-delay")

	// Create DB connection
	db, err := database.NewDB(dbConn, sqlf.PostgresBindVar)
	if err != nil {
		return err
	}
	defer db.Close()

	// Use PingWithRetry with the provided parameters
	fmt.Fprintf(os.Stdout, "%sAttempting to ping database%s\n", output.StyleInfo, output.StyleReset)
	if retryCount > 0 {
		fmt.Fprintf(os.Stdout, "%sUsing retry count: %d, initial delay: %dms%s\n",
			output.StyleInfo, retryCount, retryDelay, output.StyleReset)
	}

	err = db.PingWithRetry(c.Context, retryCount, retryDelay)

	if err != nil {
		fmt.Fprintf(os.Stdout, "%sFailed to connect to database: %s%s\n",
			output.StyleFailure, err.Error(), output.StyleReset)
		return err
	}

	fmt.Fprintf(os.Stdout, "%sSuccessfully connected to database!%s\n",
		output.StyleSuccess, output.StyleReset)
	return nil
}
