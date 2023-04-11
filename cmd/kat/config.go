package main

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"

	"github.com/BolajiOlajide/kat/internal/constants"
	"github.com/BolajiOlajide/kat/internal/types"
)

var errConfigNotFound = errors.Newf("config file not provided. Provide one via the `-c` flag or have the %s in the current working directory.", constants.KatConfigurationFileName)

func checkConfigPath(c *cli.Context) error {
	var (
		f   []byte
		err error
	)

	path := c.String("config")
	if path != "" {
		f, err = os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return errConfigNotFound
			}
			return err
		}
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "getting working directory")
		}
		configFilePath := fmt.Sprintf("%s/%s", wd, constants.KatConfigurationFileName)
		f, err = os.ReadFile(configFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return errConfigNotFound
			}
			return err
		}
	}

	var config types.Config
	err = yaml.Unmarshal(f, config)
	if err != nil {
		return err
	}
	fmt.Println(config, "<====")
	return nil
}
