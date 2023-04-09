package migration

import (
	"fmt"
	"os"

	"github.com/BolajiOlajide/kat/internal/constants"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var defaultConfig = types.Config{
	Migration: types.MigrationInfo{
		TableName: "migrations",
		Directory: "migrations",
	},
	Database: types.DatabaseInfo{},
}

// Init initliazes a project for use with kat.
func Init(c *cli.Context) (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("%sAn error occurred while initializing kat!%s\n", output.StyleFailure, output.StyleReset)
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "getting working directory")
	}

	configFilePath := fmt.Sprintf("%s/%s", wd, constants.KatConfigurationFileName)

	_, err = os.Stat(configFilePath)
	if !os.IsNotExist(err) {
		return errors.New("kat is already initialized")
	}

	tableName := c.String("tableName")
	if tableName == "" {
		tableName = "migrations"
	}

	databaseURL := c.String("databaseURL")
	if databaseURL == "" {
		databaseURL = ""
	}

	directory := c.String("directory")
	if directory == "" {
		directory = fmt.Sprintf("%s/%s", wd, "migrations")
	}

	config := types.Config{
		Database: types.DatabaseInfo{
			URL: databaseURL,
		},
		Migration: types.MigrationInfo{
			Directory: directory,
			TableName: tableName,
		},
	}
	mc, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "marshalling config during initialization")
	}

	err = os.WriteFile(constants.KatConfigurationFileName, mc, os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "writing configuration file")
	}

	fmt.Printf("%sKat initialized successfully!%s\n", output.StyleSuccess, output.StyleReset)
	fmt.Printf("%sConfig file: %s%s\n", output.StyleInfo, configFilePath, output.StyleReset)
	return nil
}
