package kat

import (
	"database/sql"

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
	return func(m *Migration) error {
		d, err := database.NewWithDB(db, m.logger)
		if err != nil {
			return err
		}
		m.db = d
		return nil
	}
}
