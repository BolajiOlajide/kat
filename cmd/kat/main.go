package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
	"github.com/urfave/cli/v2"
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
)

var kat = &cli.App{
	Usage:       "Database migration tool",
	Description: "Database migration tool based on Sourcegraph's internal tooling.",
	Version:     version.Version(),
	Compiled:    time.Now(),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "toggle verbose mode",
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
			Usage:       "Initializes kat",
			Description: "Creates a new configuration file for Kat and a migration directory.",
			Action:      initialize,
		},
		{
			Name:        "add",
			ArgsUsage:   "<name>",
			Usage:       "Add a new migration file",
			Description: "Creates a new migration file in the migrations directory",
			Action:      add,
		},
		{
			Name:        "up",
			Usage:       "Apply all migrations",
			Description: "Apply migrations",
			Action:      up,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "url",
					Usage:       "database url",
					Aliases:     []string{"u"},
					EnvVars:     []string{"KAT_DATABASE_URL"},
					Destination: &databaseURL,
					Required:    true,
				},
				// &cli.StringFlag{}
			},
		},
	},

	UseShortOptionHandling: true,

	HideVersion:     true,
	HideHelpCommand: true,
	ExitErrHandler: func(cmd *cli.Context, err error) {
		if err == nil {
			return
		}

		// Show help text only
		if errors.Is(err, flag.ErrHelp) {
			cli.ShowSubcommandHelpAndExit(cmd, 1)
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
