package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/output"
	updatepkg "github.com/BolajiOlajide/kat/internal/update"
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

func update(c *cli.Context) error {
	if version.IsDev() {
		fmt.Fprintf(os.Stdout, "%sYou are running kat in dev mode. The update command is not available in dev mode.%s\n", output.StyleInfo, output.StyleReset)
		return nil
	}

	fmt.Fprintf(os.Stdout, "%sChecking for updates...%s\n", output.StyleInfo, output.StyleReset)

	// Check if a newer version is available
	hasUpdate, latestVersion, downloadURL, err := updatepkg.CheckForUpdates()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	// No update available
	if !hasUpdate {
		fmt.Fprintf(os.Stdout, "%sKat is already at the latest version.%s\n", output.StyleSuccess, output.StyleReset)
		return nil
	}

	// Update available - notify the user
	fmt.Fprintf(os.Stdout, "%sA new version of Kat is available: %s%s\n",
		output.StyleInfo, latestVersion, output.StyleReset)

	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// In case the executable is a symlink, get the real path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Confirm the update with the user, unless -y flag is provided
	if !c.Bool("yes") {
		fmt.Fprintf(os.Stdout, "\nDo you want to update to version %s? [y/N]: ", latestVersion)
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Fprintf(os.Stdout, "%sUpdate cancelled.%s\n", output.StyleInfo, output.StyleReset)
			return nil
		}
	}

	// Download and install the update
	err = updatepkg.DownloadAndReplace(downloadURL, execPath, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Fprintf(os.Stdout, "%sKat has been updated to version %s%s\n",
		output.StyleSuccess, latestVersion, output.StyleReset)
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
