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
func GenerateConfigFile(tableName, directory, databaseURL, dbUser, dbPassword, dbName, dbPort, dbHost, dbSSLMode string) ([]byte, error) {
	// Load the embedded template
	tmpl, err := template.ParseFS(templatesFS, "templates/init.tmpl")
	if err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}

	// Prepare template data
	data := struct {
		TableName     string
		Directory     string
		DatabaseURL   string
		DBUser        string
		DBPassword    string
		DBName        string
		DBPort        string
		DBHost        string
		DBSSLMode     string
		UseConnString bool
	}{
		TableName:     tableName,
		Directory:     directory,
		DatabaseURL:   databaseURL,
		DBUser:        dbUser,
		DBPassword:    dbPassword,
		DBName:        dbName,
		DBPort:        dbPort,
		DBHost:        dbHost,
		DBSSLMode:     dbSSLMode,
		UseConnString: databaseURL != "",
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
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

	databaseURL := c.String("databaseURL")

	directory := c.String("directory")
	if directory == "" {
		directory = fmt.Sprintf("%s/%s", wd, "migrations")
	}

	// Get DB connection parameters
	dbUser := c.String("dbUser")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := c.String("dbPassword")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dbName := c.String("dbName")
	if dbName == "" {
		dbName = "myapp"
	}

	dbPort := c.String("dbPort")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbHost := c.String("dbHost")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbSSLMode := c.String("dbSSLMode")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	// Generate config file from template
	configContent, err := GenerateConfigFile(
		tableName, directory, databaseURL,
		dbUser, dbPassword, dbName, dbPort, dbHost, dbSSLMode,
	)
	if err != nil {
		return errors.Wrap(err, "generating config file")
	}

	// Note: For backward compatibility, we would normally create a config struct here,
	// but since we're using the template directly, we don't need it anymore.
	// The template-based approach provides more flexibility and better output formatting.

	// Save the generated config to file
	err = os.WriteFile(constants.KatConfigurationFileName, configContent, os.FileMode(0755))
	if err != nil {
		return errors.Wrap(err, "writing configuration file")
	}

	fmt.Printf("%sKat initialized successfully!%s\n", output.StyleSuccess, output.StyleReset)
	fmt.Printf("%sConfig file: %s%s\n", output.StyleInfo, configFilePath, output.StyleReset)
	return nil
}
