package driver

import (
	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

type DatabaseDriver string

const (
	PostgresDriver DatabaseDriver = "postgres"
	SqliteDriver   DatabaseDriver = "sqlite"
)

func ParseDBDriver(d string) (DatabaseDriver, error) {
	switch d {
	// we want to always default to postgres when the driver is not specified
	// this is to enforce backward compatibility
	case "postgres", "postgresql", "":
		return PostgresDriver, nil
	case "sqlite3", "sqlite":
		return SqliteDriver, nil
	default:
		return "", errors.Newf("unsupported database driver: %s", d)
	}
}

func (d DatabaseDriver) String() string {
	return string(d)
}

func (d DatabaseDriver) IsPostgres() bool {
	return d == PostgresDriver
}

func (d DatabaseDriver) IsSQLite() bool {
	return d == SqliteDriver
}

func (d DatabaseDriver) Valid() bool {
	switch d {
	case PostgresDriver:
		return true
	case SqliteDriver:
		return true
	default:
		return false
	}
}

func (d DatabaseDriver) BindVar() sqlf.BindVar {
	switch d {
	case PostgresDriver:
		return sqlf.PostgresBindVar
	case SqliteDriver:
		return sqlf.SimpleBindVar
	default:
		return nil
	}
}

func (d DatabaseDriver) DriverName() string {
	switch d {
	case PostgresDriver:
		return "pgx"
	case SqliteDriver:
		return "sqlite3"
	default:
		return ""
	}
}
