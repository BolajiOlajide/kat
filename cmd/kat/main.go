package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/constants"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
)

func main() {
	if err := kat.RunContext(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var verbose bool

var descriptionFlag = &cli.StringFlag{
	Name:    "description",
	Usage:   "-d <description of the migration>",
	Aliases: []string{"d"},
}

var configFlag = &cli.PathFlag{
	Name:    "config",
	Usage:   "the configuration file for kat",
	Aliases: []string{"c"},
	EnvVars: []string{"KAT_CONFIG_FILE"},
	Value:   constants.KatConfigurationFileName,
}

var countFlag = &cli.IntFlag{
	Name:    "count",
	Aliases: []string{"n"},
	Usage:   "number of migrations to roll back (default: 1)",
	Value:   1,
}

var dryRunFlag = &cli.BoolFlag{
	Name:    "dry-run",
	Usage:   "validate migrations without applying them",
	Aliases: []string{"d"},
	EnvVars: []string{"KAT_DRY_RUN"},
	Value:   false,
}

var retryCountFlag = &cli.IntFlag{
	Name:    "retry-count",
	Usage:   "number of times to retry operations on transient errors (min: 0, max: 7)",
	Aliases: []string{"r"},
	EnvVars: []string{"KAT_RETRY_COUNT"},
	Value:   3,
}

var retryDelayFlag = &cli.IntFlag{
	Name:    "retry-delay",
	Usage:   "initial delay between retries in milliseconds (min: 100, max: 3000)",
	Aliases: []string{"rd"},
	EnvVars: []string{"KAT_RETRY_DELAY"},
	Value:   500,
}

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
			Name:        "export",
			Usage:       "Export migration graph",
			Description: "Generate a directed acyclic graph (DAG) visualization of migrations",
			Action:      exportExec,
			Before:      config.ParseConfig,
			Flags: []cli.Flag{
				configFlag,
				&cli.StringFlag{
					Name:    "file",
					Usage:   "filename to write the directed acyclic graph to (defaults to stdout)",
					Aliases: []string{"f"},
				},
			},
		},
		{
			Name:        "update",
			Usage:       "Update kat to the latest version",
			Description: "Check for and install the latest version of kat",
			Action:      update,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "yes",
					Usage:   "skip confirmation prompt",
					Aliases: []string{"y"},
					Value:   false,
				},
			},
		},
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
				&cli.StringFlag{
					Name:    "dbUser",
					Usage:   "database username",
					EnvVars: []string{"KAT_DB_USER"},
					Value:   "postgres",
				},
				&cli.StringFlag{
					Name:    "dbPassword",
					Usage:   "database password",
					EnvVars: []string{"KAT_DB_PASSWORD"},
					Value:   "postgres",
				},
				&cli.StringFlag{
					Name:    "dbName",
					Usage:   "database name",
					EnvVars: []string{"KAT_DB_NAME"},
					Value:   "myapp",
				},
				&cli.StringFlag{
					Name:    "dbPort",
					Usage:   "database port",
					EnvVars: []string{"KAT_DB_PORT"},
					Value:   "5432",
				},
				&cli.StringFlag{
					Name:    "dbHost",
					Usage:   "database host",
					EnvVars: []string{"KAT_DB_HOST"},
					Value:   "localhost",
				},
				&cli.StringFlag{
					Name:    "dbSSLMode",
					Usage:   "database SSL mode (disable, allow, prefer, require, verify-ca, verify-full)",
					EnvVars: []string{"KAT_DB_SSL_MODE"},
					Value:   "disable",
				},
			},
		},
		{
			Name:        "version",
			Usage:       "Show version",
			Description: "This command returns the version of kat",
			Action:      getVersion,
		},
		{
			Name:        "add",
			ArgsUsage:   "<n>",
			Usage:       "Create migration",
			Description: "Creates a new migration file in the migrations directory",
			Action:      addExec,
			Before:      config.ParseConfig,
			Flags:       []cli.Flag{configFlag, descriptionFlag},
		},
		{
			Name:        "up",
			Usage:       "Run migrations",
			Description: "Apply migrations",
			Action:      upExec,
			Before:      config.ParseConfig,
			Flags:       []cli.Flag{countFlag, configFlag, dryRunFlag},
		},
		{
			Name:        "down",
			Usage:       "Rollback migrations",
			Description: "Rollback the most recent migration or specify a count with --count flag",
			Action:      downExec,
			Before:      config.ParseConfig,
			Flags:       []cli.Flag{countFlag, configFlag, dryRunFlag},
		},
		{
			Name:        "ping",
			Usage:       "Test database connection",
			Description: "Verifies the database connection with optional retry capabilities",
			Action:      ping,
			Before:      config.ParseConfig,
			Flags:       []cli.Flag{configFlag, retryCountFlag, retryDelayFlag},
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
			fmt.Fprintln(os.Stderr, fmt.Sprintf("%s%s%s", output.StyleFailure, errMsg, output.StyleReset))
		}

		// Determine exit code
		var exitErr cli.ExitCoder
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	},
}
