package types

import (
	"fmt"
	"net/url"
	"time"

	dbdriver "github.com/BolajiOlajide/kat/internal/database/driver"
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

	if !c.Database.Driver.Valid() {
		c.Database.Driver = dbdriver.PostgresDriver
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

	Driver dbdriver.DatabaseDriver `yaml:"driver,omitempty"`

	Path string `yaml:"path,omitempty"`
	URL  string `yaml:"url,omitempty"`

	ConnectTimeout   string `yaml:"connect_timeout,omitempty"`
	StatementTimeout string `yaml:"statement_timeout,omitempty"`
	MaxOpenConns     int    `yaml:"max_open_conns,omitempty"`
	MaxIdleConns     int    `yaml:"max_idle_conns,omitempty"`
	ConnMaxLifetime  string `yaml:"conn_max_lifetime,omitempty"`
	DefaultTimeout   string `yaml:"default_timeout,omitempty"`
}

// DBTimeouts holds parsed database timeout and pool configuration.
type DBTimeouts struct {
	ConnectTimeout   time.Duration
	StatementTimeout time.Duration
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration
	DefaultTimeout   time.Duration
}

// ParseDBTimeouts parses the timeout string fields into a DBTimeouts struct.
// Returns nil if no timeout fields are configured.
func (d *DatabaseInfo) ParseDBTimeouts() (*DBTimeouts, error) {
	if d.ConnectTimeout == "" && d.StatementTimeout == "" && d.ConnMaxLifetime == "" && d.DefaultTimeout == "" && d.MaxOpenConns == 0 && d.MaxIdleConns == 0 {
		return nil, nil
	}

	t := &DBTimeouts{
		MaxOpenConns: d.MaxOpenConns,
		MaxIdleConns: d.MaxIdleConns,
	}

	var err error
	if d.ConnectTimeout != "" {
		t.ConnectTimeout, err = time.ParseDuration(d.ConnectTimeout)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid connect_timeout %q", d.ConnectTimeout)
		}
	}
	if d.StatementTimeout != "" {
		t.StatementTimeout, err = time.ParseDuration(d.StatementTimeout)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid statement_timeout %q", d.StatementTimeout)
		}
	}
	if d.ConnMaxLifetime != "" {
		t.ConnMaxLifetime, err = time.ParseDuration(d.ConnMaxLifetime)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid conn_max_lifetime %q", d.ConnMaxLifetime)
		}
	}
	if d.DefaultTimeout != "" {
		t.DefaultTimeout, err = time.ParseDuration(d.DefaultTimeout)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid default_timeout %q", d.DefaultTimeout)
		}
	}

	return t, nil
}

func (d *DatabaseInfo) ConnString() (string, error) {
	// For SQLite, return the database file path directly
	if d.Driver.IsSQLite() {
		if d.Path == "" {
			return "", errors.New("database path is required for SQLite driver; set the 'path' field in your database config")
		}
		return d.Path, nil
	}

	// at this point, we can assume the driver is postgres
	if d.URL != "" {
		// Validate the scheme but return the original URL unchanged to preserve
		// query params, special characters in passwords, and connection options.
		if err := d.validateURL(); err != nil {
			return "", err
		}
		return d.URL, nil
	}

	// if a url isn't provided, use the traditional connection string format
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode), nil
}

func (d *DatabaseInfo) validateURL() error {
	parsedURL, err := url.Parse(d.URL)
	if err != nil {
		return errors.Newf("failed to parse URL: %v", err)
	}

	return validateScheme(parsedURL.Scheme)
}

func validateScheme(scheme string) error {
	switch scheme {
	case "postgresql+ssl", "postgresql", "postgres":
		return nil
	default:
		return errors.Newf("invalid scheme: %s", scheme)
	}
}
