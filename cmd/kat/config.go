package main

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/constants"
)

func checkConfigPath(c *cli.Context) error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "getting working directory")
	}

	path := c.String("config")
	configFilePath := fmt.Sprintf("%s/%s", wd, constants.KatConfigurationFileName)
	_, err = os.Stat(configFilePath)
	if path == "" && os.IsNotExist(err) {
		return errors.Newf("config file not provided. Provide one via the `-c` flag or have the %s in the current working directory.", constants.KatConfigurationFileName)
	}

	return nil
}
