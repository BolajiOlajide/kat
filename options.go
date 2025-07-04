package kat

import (
	"database/sql"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/loggr"
)

type MigrationOption func(*Migration) error

func WithLogger(logger loggr.Logger) MigrationOption {
	return func(m *Migration) error {
		m.logger = logger
		return nil
	}
}

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
