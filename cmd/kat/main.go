package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
)

func main() {
	if err := kat.RunContext(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var (
	// Global verbose mode
	verbose bool

	// database connection string
	databaseURL string

	// the path to Kat configuration
	configPath string
)

var kat = &cli.App{
	Usage:       "Database migration tool",
	Description: "Database migration tool based on Sourcegraph's internal tooling.",
	Version:     version.Version(),
	Compiled:    time.Now(),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "toggle verbose mode (default: false)",
			Aliases:     []string{"v"},
			EnvVars:     []string{"KAT_VERBOSE"},
			Value:       false,
			Destination: &verbose,
		},
	},
	Commands: []*cli.Command{
		{
			Name:        "init",
			ArgsUsage:   "<directory>",
			Usage:       "Initializes kat by creating a configuration file",
			Description: "Creates a new configuration file for Kat and a migration directory.",
			Action:      initialize,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "tableName",
					Usage:   "the name of the database table for tracking migrations",
					Aliases: []string{"t"},
					EnvVars: []string{"KAT_MIGRATION_TABLE_NAME"},
					Value:   "migrations",
				},
				&cli.StringFlag{
					Name:    "databaseURL",
					Usage:   "the URL of the database to run the migrations against (default: '')",
					Aliases: []string{"u"},
					EnvVars: []string{"KAT_MIGRATION_DATABASE_URL"},
				},
				&cli.StringFlag{
					Name:    "directory",
					Usage:   "the name of the directory where migrations will be stored",
					Aliases: []string{"d"},
					EnvVars: []string{"KAT_MIGRATION_DIRECTORY"},
					Value:   "migrations",
				},
			},
		},
		{
			Name:        "version",
			Usage:       "returns the current version of kat",
			Description: "This command returns the version of kat",
			Action:      getVersion,
		},
		{
			Name:        "add",
			ArgsUsage:   "<name>",
			Usage:       "Adds a new migration",
			Description: "Creates a new migration file in the migrations directory",
			Action:      add,
			Before:      checkConfigPath,
			Flags: []cli.Flag{
				&cli.PathFlag{
					Name:    "config",
					Usage:   "the configuration file for kat",
					Aliases: []string{"c"},
					EnvVars: []string{"KAT_CONFIG_FILE"},
					Value:   "kat.config.yaml",
				},
			},
		},
		{
			Name:        "up",
			Usage:       "Apply all migrations",
			Description: "Apply migrations",
			Action:      up,
			Before:      checkConfigPath,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "url",
					Usage:       "database url",
					Aliases:     []string{"u"},
					EnvVars:     []string{"KAT_DATABASE_URL"},
					Destination: &databaseURL,
					Required:    true,
				},
			},
		},
	},
	UseShortOptionHandling: true,
	HideVersion:            true,
	HideHelpCommand:        true,
	ExitErrHandler: func(c *cli.Context, err error) {
		if err == nil {
			return
		}

		// Show help text only
		if errors.Is(err, flag.ErrHelp) {
			cli.ShowSubcommandHelpAndExit(c, 1)
		}

		errMsg := err.Error()
		if errMsg != "" {
			f := fmt.Sprintf("%s%s%s", output.StyleFailure, errMsg, output.StyleReset)
			fmt.Fprintln(os.Stderr, f)
		}

		// Determine exit code
		if exitErr, ok := err.(cli.ExitCoder); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	},
}
