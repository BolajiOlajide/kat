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

type DatabaseInfo struct {
	Driver   Driver `yaml:"driver"`
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
		return errors.Newf("invalid scheme: %s", scheme)
	}
}
