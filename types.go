package kat

import (
	"github.com/BolajiOlajide/kat/internal/database"
	dbdriver "github.com/BolajiOlajide/kat/internal/database/driver"
	"github.com/BolajiOlajide/kat/internal/loggr"
)

// Driver represents a supported database driver type.
type Driver = dbdriver.DatabaseDriver

const (
	// PostgresDriver is the driver for PostgreSQL databases.
	PostgresDriver = dbdriver.PostgresDriver
	// SQLiteDriver is the driver for SQLite databases.
	SQLiteDriver = dbdriver.SqliteDriver
)

// ParseDriver converts a driver name string into a Driver.
// Accepted values: "postgres", "postgresql", "" (defaults to postgres),
// "sqlite", "sqlite3".
func ParseDriver(name string) (Driver, error) {
	return dbdriver.ParseDBDriver(name)
}

// Logger is the interface used for logging within kat.
// Implement this interface to provide custom logging behavior.
type Logger = loggr.Logger

// DBConfig holds database connection configuration options.
type DBConfig = database.DBConfig

// DefaultDBConfig returns sensible default configuration for Kat migrations.
func DefaultDBConfig(drv Driver) DBConfig {
	return database.DefaultDBConfig(drv)
}
