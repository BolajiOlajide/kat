package main

import (
	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/migration"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

func addExec(c *cli.Context) error {
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

	name := args[0]
	description := c.String("description")
	prog := tea.NewProgram(migration.NewAddModel(c.Context, cfg, name, description), tea.WithContext(c.Context))
	_, err = prog.Run()
	return err
}
