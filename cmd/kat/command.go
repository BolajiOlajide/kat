package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
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
	
	// Pass the dry-run flag to the migration
	dryRun := c.Bool("dry-run")
	if dryRun {
		fmt.Fprintf(os.Stdout, "%sDRY RUN: Migrations will be validated but not applied%s\n", output.StyleInfo, output.StyleReset)
	}
	
	return migration.Up(c, cfg, dryRun)
}

func down(c *cli.Context) error {
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}
	
	// Pass the dry-run flag to the migration
	dryRun := c.Bool("dry-run")
	if dryRun {
		fmt.Fprintf(os.Stdout, "%sDRY RUN: Migrations will be validated but not rolled back%s\n", output.StyleInfo, output.StyleReset)
	}
	
	return migration.Down(c, cfg, dryRun)
}

func initialize(c *cli.Context) error {
	return migration.Init(c)
}

func getVersion(c *cli.Context) error {
	fmt.Fprintf(os.Stdout, "%sVersion: %s%s\n", output.StyleInfo, version.Version(), output.StyleReset)
	return nil
}
