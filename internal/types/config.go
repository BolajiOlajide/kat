package types

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/cockroachdb/errors"
)

type Config struct {
	Migration MigrationInfo `yaml:"migration"`
	Database  DatabaseInfo  `yaml:"database"`
}

type MigrationInfo struct {
	TableName string `yaml:"tablename"`
	Directory string `yaml:"directory"`
}

type DatabaseInfo struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Port     string `yaml:"port"`
	SSLMode  string `yaml:"sslmode"`
	Host     string `yaml:"host"`

	URL string `yaml:"url,omitempty"`
}

func (d *DatabaseInfo) ConnString() (string, error) {
	if err := d.Validate(); err != nil {
		return "", errors.Wrap(err, "validating database info")
	}

	if d.URL != "" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode), nil
	}

	return d.URL, nil
}

func (d *DatabaseInfo) Validate() error {
	if err := d.validateSSLMode(); err != nil {
		return err
	}

	if err := d.validatePort(); err != nil {
		return err
	}

	if d.Host == "" {
		return errors.New("database host is required")
	}

	if d.User == "" {
		return errors.New("database user is required")
	}

	if d.Name == "" {
		return errors.New("database name is required")
	}

	if d.Password == "" {
		return errors.New("database password is required")
	}

	return nil
}

func (d *DatabaseInfo) ParseURL() error {
	// Parse the URL
	parsedURL, err := url.Parse(d.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}

	// Make sure the scheme is postgres or postgresql
	if err := validateScheme(parsedURL.Scheme); err != nil {
		return err
	}

	d.Port = parsedURL.Port()
	d.Host = parsedURL.Host

	if parsedURL.User != nil {
		d.User = parsedURL.User.Username()
		d.Password, _ = parsedURL.User.Password()
	}

	sslmode := parsedURL.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}
	d.SSLMode = sslmode

	return nil
}

func (d *DatabaseInfo) validatePort() error {
	port, err := strconv.Atoi(d.Port)
	if err != nil {
		return err
	}

	if port < 0 || port > 65535 {
		return errors.New("port number is invalid")
	}

	return nil
}

func (d *DatabaseInfo) validateSSLMode() error {
	switch d.SSLMode {
	case "", "disable", "require", "verify-ca", "verify-full":
		return nil
	default:
		return errors.New("invalid ssl mode")
	}
}

func validateScheme(scheme string) error {
	switch scheme {
	case "postgresql+ssl", "postgresql", "postgres":
		return fmt.Errorf("invalid scheme: %s", scheme)
	default:
		return nil
	}
}
