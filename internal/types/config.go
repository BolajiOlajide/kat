package types

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"
)

type Config struct {
	Migration MigrationInfo `yaml:"migration"`
	Database  DatabaseInfo  `yaml:"database"`
}

type MigrationInfo struct {
	TableName string `yaml:"tablename"`
	Directory string `yaml:"directory"`
	DryRun    bool   `yaml:"dryRun,omitempty"`
}

func (c *Config) SetDefault() {
	if c.Migration.Directory == "" {
		c.Migration.Directory = "migrations"
	}

	if c.Migration.TableName == "" {
		c.Migration.TableName = "migrations"
	}

	// We assume when the URL isn't provided, the user has specified database credentials manually
	// so we set SSL mode to `disable` if the user doesn't have it defined.
	if c.Database.URL == "" && c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
}

type DatabaseInfo struct {
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	Name     string `yaml:"name,omitempty"`
	Port     string `yaml:"port,omitempty"`
	SSLMode  string `yaml:"sslmode,omitempty"`
	Host     string `yaml:"host,omitempty"`

	URL string `yaml:"url,omitempty"`
}

func (d *DatabaseInfo) ConnString() (string, error) {
	if d.URL != "" {
		err := d.parseURL()
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode), nil
}

func (d *DatabaseInfo) parseURL() error {
	// Parse the URL
	parsedURL, err := url.Parse(d.URL)
	if err != nil {
		return errors.Newf("failed to parse URL: %v", err)
	}

	// Make sure the scheme is postgres or postgresql
	if err := validateScheme(parsedURL.Scheme); err != nil {
		return err
	}

	port := parsedURL.Port()
	if port == "" {
		port = "5432" // default postgres port
	}
	d.Port = port

	d.Host = parsedURL.Hostname()

	if parsedURL.User != nil {
		d.User = parsedURL.User.Username()
		d.Password, _ = parsedURL.User.Password()
	}

	sslmode := parsedURL.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}
	d.SSLMode = sslmode

	d.Name = strings.ReplaceAll(parsedURL.Path, "/", "")
	return nil
}

func validateScheme(scheme string) error {
	switch scheme {
	case "postgresql+ssl", "postgresql", "postgres":
		return nil
	default:
		return fmt.Errorf("invalid scheme: %s", scheme)
	}
}
