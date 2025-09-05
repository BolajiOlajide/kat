package kat

import (
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/loggr"
)

// MigrationOption is a function that configures a Migration instance.
// Options are applied during migration creation and can return errors
// if the configuration is invalid.
type MigrationOption func(*Migration) error

// WithLogger configures the migration to use a custom logger implementation.
// The logger must implement the Logger interface with Debug, Info, Warn, and Error methods.
//
// Example:
//
//	m, err := kat.New(connStr, fsys, "migrations",
//		kat.WithLogger(customLogger),
//	)
func WithLogger(logger loggr.Logger) MigrationOption {
	return func(m *Migration) error {
		m.logger = logger
		return nil
	}
}

// WithSqlDB configures the migration to use an existing *sql.DB connection
// instead of creating a new one from the connection string.
// When this option is used, the connection string parameter in New() is ignored.
// By default, this assumes a PostgreSQL database (uses sqlf.PostgresBindVar).
// For SQLite databases, use WithSqlDBAndDriver instead.
//
// Example:
//
//	db, err := sql.Open("pgx", "postgres://...")
//	if err != nil {
//		return err
//	}
//	m, err := kat.New("", fsys, "migrations",
//		kat.WithSqlDB(db),
//	)
func WithSqlDB(db *sql.DB) MigrationOption {
	return WithSqlDBAndDriver(db, "postgres")
}

// WithSqlDBAndDriver configures the migration to use an existing *sql.DB connection
// with the specified driver type. This allows proper handling of different database
// bind variable formats (PostgreSQL uses $1,$2 while SQLite uses ?).
//
// Example for SQLite:
//
//	db, err := sql.Open("sqlite", "database.db")
//	if err != nil {
//		return err
//	}
//	m, err := kat.New("", fsys, "migrations",
//		kat.WithSqlDBAndDriver(db, "sqlite3"),
//	)
func WithSqlDBAndDriver(db *sql.DB, driver string) MigrationOption {
	return func(m *Migration) error {
		var bindVar sqlf.BindVar
		switch driver {
		case "postgres":
			bindVar = sqlf.PostgresBindVar
		case "sqlite3", "sqlite":
			bindVar = sqlf.SimpleBindVar
			// SQLite allows only one writer at a time, so limit connections to avoid "database is locked" errors
			db.SetMaxOpenConns(1)
		default:
			bindVar = sqlf.PostgresBindVar // default to postgres for backward compatibility
		}
		
		d, err := database.NewWithDB(db, bindVar, m.logger)
		if err != nil {
			return err
		}
		m.db = d
		return nil
	}
}
