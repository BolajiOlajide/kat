package migration

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/BolajiOlajide/kat/internal/constants"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
)

// GenerateConfigFile creates a configuration file from the init.tmpl template
// using the provided parameters
func GenerateConfigFile(tableName, directory string, driver constants.Driver) ([]byte, error) {
	// Load the embedded template
	tmpl, err := template.ParseFS(templatesFS, "templates/init.tmpl")
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		TableName string
		Directory string
		Driver    constants.Driver
	}{
		TableName: tableName,
		Directory: directory,
		Driver:    driver,
	}); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	return buf.Bytes(), nil
}

// Init initializes a project for use with kat.
func Init(c *cli.Context) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "getting working directory")
	}

	configFilePath := fmt.Sprintf("%s/%s", wd, constants.KatConfigurationFileName)

	_, err = os.Stat(configFilePath)
	if !os.IsNotExist(err) {
		return errors.New("kat is already initialized")
	}

	// Get parameters from CLI context
	tableName := c.String("tableName")
	if tableName == "" {
		tableName = "migrations"
	}

	directory := c.String("directory")
	if directory == "" {
		directory = fmt.Sprintf("%s/%s", wd, "migrations")
	}

	unformattedDriver := c.String("driver")
	if unformattedDriver == "" {
		unformattedDriver = "postgres"
	}

	drv := constants.Driver(unformattedDriver)
	if !drv.Valid() {
		return errors.New("invalid driver: (postgres/sqlite)")
	}

	// Generate config file from template
	configContent, err := GenerateConfigFile(tableName, directory, drv)
	if err != nil {
		return errors.Wrap(err, "generating config file")
	}

	// Save the generated config to file
	err = os.WriteFile(constants.KatConfigurationFileName, configContent, os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "writing configuration file")
	}

	fmt.Printf("%sKat initialized successfully!%s\n", output.StyleSuccess, output.StyleReset)
	fmt.Printf("%sConfig file: %s%s\n", output.StyleInfo, configFilePath, output.StyleReset)
	return nil
}
