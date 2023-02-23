package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Config is a struct that holds the configuration information for Kat.
type Config struct {
	host     string
	port     string
	password string
	name     string
	sslmode  string
	user     string
}

// ConnString returns a PostgreSQL connection string.
func (c Config) ConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.host, c.port, c.user, c.password, c.name, c.sslmode)
}

// ParsePgConnString parses a PostgreSQL connection string and returns a Config struct.
func ParsePgConnString(connStr string) (Config, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return Config{}, err
	}

	if !isSchemeValid(u.Scheme) {
		return Config{}, fmt.Errorf("invalid PostgreSQL connection scheme: %s", u.Scheme)
	}

	passwd, _ := u.User.Password()

	return Config{
		host:     u.Hostname(),
		port:     u.Port(),
		password: passwd,
		name:     strings.TrimPrefix(u.Path, "/"),
		sslmode:  u.Query().Get("sslmode"),
		user:     u.User.Username(),
	}, nil
}

func isSchemeValid(scheme string) bool {
	switch scheme {
	case "postgres", "postgresql", "postgresql+unix", "postgresql+ssl":
		return true
	default:
		return false
	}
}
