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
	database string
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
		&cli.StringFlag{
			Name:        "database",
			Usage:       "database connection string",
			Aliases:     []string{"d"},
			EnvVars:     []string{"KAT_DATABASE_URL"},
			Destination: &database,
		},
	},
	Commands: []*cli.Command{
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
		},
	},

	Before: func(ctx *cli.Context) error {
		if verbose {
			fmt.Fprintln(os.Stderr, "Verbose mode enabled")
		}
		return nil
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
			f := fmt.Sprintf("%s%s%s\n", output.StyleFailure, errMsg, output.StyleReset)
			fmt.Fprintln(os.Stderr, f)
		}

		// Determine exit code
		if exitErr, ok := err.(cli.ExitCoder); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	},
}
