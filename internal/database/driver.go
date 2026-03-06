package database

import (
	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

type driver string

const (
	postgresDriver driver = "postgres"
	sqliteDriver   driver = "sqlite"
)

func newDriver(d string) (driver, error) {
	switch d {
	// we want to always default to postgres when the driver is not specified
	// this is to enforce backward compatibility
	case "postgres", "":
		return postgresDriver, nil
	case "sqlite3", "sqlite":
		return sqliteDriver, nil
	default:
		return "", errors.Newf("unsupported database driver: %s", d)
	}
}

func (d driver) IsPostgres() bool {
	return d == postgresDriver
}

func (d driver) IsSQLite() bool {
	return d == sqliteDriver
}

func (d driver) Valid() bool {
	switch d {
	case postgresDriver:
		return true
	case sqliteDriver:
		return true
	default:
		return false
	}
}

func (d driver) BindVar() sqlf.BindVar {
	switch d {
	case postgresDriver:
		return sqlf.PostgresBindVar
	case sqliteDriver:
		return sqlf.SimpleBindVar
	default:
		return nil
	}
}

func (d driver) DriverName() string {
	switch d {
	case postgresDriver:
		return "pgx"
	case sqliteDriver:
		return "sqlite3"
	default:
		return ""
	}
}
