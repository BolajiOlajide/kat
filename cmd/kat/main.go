package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := kat.RunContext(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const version = "dev"

var (
	// Global verbose mode
	verbose bool
)

var kat = &cli.App{
	Usage:       "Database Migration Tool",
	Description: "Database Migration Tool based on Sourcegraph's internal tooling.",
	Version:     version,
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
		addCommand,
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

var (
	addCommand = &cli.Command{
		Name:        "add",
		ArgsUsage:   "<name>",
		Usage:       "Add a new migration file",
		Description: "Creates a new migration file in the migrations directory",
		Action:      addMigration,
	}
)

func addMigration(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return cli.Exit("no migration name specified", 1)
	}
	if len(args) != 1 {
		return cli.Exit("too many arguments", 1)
	}
	return migration.Add(args[0])
}
