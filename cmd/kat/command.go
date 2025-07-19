package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/loggr"
	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/output"
	updatepkg "github.com/BolajiOlajide/kat/internal/update"
	"github.com/BolajiOlajide/kat/internal/version"
)

func addExec(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) == 0 {
		return cli.Exit("no migration name specified", 1)
	}
	if len(args) != 1 {
		return cli.Exit("too many arguments", 1)
	}

	return migration.Add(c, args[0])
}

func upExec(c *cli.Context) error {
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	// Get command flags
	dryRun := c.Bool("dry-run")

	if dryRun {
		fmt.Printf("%sDRY RUN: Migrations will not be applied%s\n", output.StyleInfo, output.StyleReset)
	}

	// Note: Retry is not used for migrations, only for the ping command
	return migration.Up(c, cfg, dryRun)
}

func downExec(c *cli.Context) error {
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	// Get command flags
	dryRun := c.Bool("dry-run")

	if dryRun {
		fmt.Printf("%sDRY RUN: Migrations will not be rolled back%s\n", output.StyleInfo, output.StyleReset)
	}

	// Note: Retry is not used for migrations, only for the ping command
	return migration.Down(c, cfg, dryRun)
}

func initialize(c *cli.Context) error {
	return migration.Init(c)
}

func getVersion(c *cli.Context) error {
	fmt.Printf("%sVersion: %s%s\n", output.StyleInfo, version.Version(), output.StyleReset)
	return nil
}

func updateExec(c *cli.Context) error {
	if version.IsDev() {
		fmt.Printf("%sYou are running kat in dev mode. The update command is not available in dev mode.%s\n", output.StyleInfo, output.StyleReset)
		return nil
	}

	// Check if a newer version is available
	hasUpdate, latestVersion, downloadURL, err := updatepkg.CheckForUpdates()
	if err != nil {
		return errors.Wrap(err, "failed to check for updates")
	}

	// No update available
	if !hasUpdate {
		fmt.Printf("%sKat is already at the latest version.%s\n", output.StyleSuccess, output.StyleReset)
		return nil
	}

	// Update available - notify the user
	fmt.Printf("%sA new version of Kat is available: %s%s\n",
		output.StyleInfo, latestVersion, output.StyleReset)

	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "failed to get executable path")
	}

	// In case the executable is a symlink, get the real path
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return errors.Wrap(err, "failed to resolve executable path")
	}

	// Confirm the update with the user, unless -y flag is provided
	if !c.Bool("yes") {
		fmt.Printf("\nDo you want to update to version %s? [y/N]: ", latestVersion)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return err
		}
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Printf("%sUpdate cancelled.%s\n", output.StyleInfo, output.StyleReset)
			return nil
		}
	}

	// Download and install the update
	err = updatepkg.DownloadAndReplace(downloadURL, execPath, os.Stdout)
	if err != nil {
		return errors.Wrap(err, "failed to update")
	}

	fmt.Printf("%sKat has been updated to version %s%s\n",
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

	logger := loggr.NewDefault()

	// Create DB connection
	db, err := database.New(dbConn, logger)
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

func exportExec(c *cli.Context) error {
	// Get configuration
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	var wrt io.Writer

	// Get format parameter
	file := c.String("file")
	if file == "" {
		wrt = os.Stdout
	} else {
		f, err := os.Create(file)
		if err != nil {
			return err
		}
		defer f.Close()
		wrt = f
	}

	// Export the graph
	return migration.ExportGraph(wrt, cfg)
}
