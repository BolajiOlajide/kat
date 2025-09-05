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
	Verbose   bool          `yaml:"verbose"`
}

type MigrationInfo struct {
	TableName string `yaml:"tablename"`
	Directory string `yaml:"directory"`
}

func (c *Config) SetDefault() {
	if c.Migration.Directory == "" {
		c.Migration.Directory = "migrations"
	}

	if c.Migration.TableName == "" {
		c.Migration.TableName = "migrations"
	}

	if c.Database.Driver == "" {
		c.Database.Driver = "postgres"
	}

	// We assume when the URL isn't provided, the user has specified database credentials manually
	// so we set SSL mode to `disable` if the user doesn't have it defined.
	if c.Database.URL == "" && c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
}

type DatabaseInfo struct {
	Driver   string `yaml:"driver,omitempty"`
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

	// For SQLite, return the database file path directly
	if d.Driver == "sqlite3" {
		if d.Name == "" {
			return "", errors.New("database name/path is required for SQLite")
		}
		return d.Name, nil
	}

	// For PostgreSQL, use the traditional connection string format
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode), nil
}

func (d *DatabaseInfo) parseURL() error {
	// Parse the URL
	parsedURL, err := url.Parse(d.URL)
	if err != nil {
		return errors.Newf("failed to parse URL: %v", err)
	}

	// Make sure the scheme is valid
	if err := validateScheme(parsedURL.Scheme); err != nil {
		return err
	}

	// Handle SQLite URLs differently
	if parsedURL.Scheme == "sqlite" || parsedURL.Scheme == "file" {
		// For SQLite, the URL is just the database file path
		// Auto-detect driver from URL scheme if not explicitly set
		if d.Driver == "" || d.Driver == "postgres" {
			d.Driver = "sqlite3"
		}
		// SQLite doesn't need host/port/user/password/sslmode
		d.Name = parsedURL.Path
		if d.Name == "" && parsedURL.Opaque != "" {
			d.Name = parsedURL.Opaque // Handle file:database.db format
		}
		// Handle sqlite://database.db format where database name becomes the host
		if d.Name == "" && parsedURL.Host != "" {
			d.Name = parsedURL.Host
		}
		// Preserve query parameters for SQLite (e.g., pragmas, cache settings)
		if parsedURL.RawQuery != "" {
			d.Name += "?" + parsedURL.RawQuery
		}
		return nil
	}

	// PostgreSQL URL parsing
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
	case "postgresql+ssl", "postgresql", "postgres", "sqlite", "file":
		return nil
	default:
		return errors.Newf("invalid scheme: %s", scheme)
	}
}
