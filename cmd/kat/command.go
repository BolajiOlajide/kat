package main

import (
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/migration"
	"github.com/BolajiOlajide/kat/internal/types"
)

func add(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return cli.Exit("no migration name specified", 1)
	}
	if len(args) != 1 {
		return cli.Exit("too many arguments", 1)
	}
	return migration.Add(args[0])
}

func up(ctx *cli.Context) error {
	// databaseURL := ctx.String("url")
	// config, err := conf.Init(databaseURL)
	// if err != nil {
	// 	return err
	// }
	return migration.Up(ctx, types.Config{})
}

func initialize(ctx *cli.Context) error {
	return migration.Init(ctx)
}
