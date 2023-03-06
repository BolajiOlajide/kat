package database

import (
	"database/sql"

	// Import the postgres driver
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// NewDB returns a new instance of the database
func NewDB(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db, nil
}

// NewDBWithPing returns a new instance of the database with a ping.
func NewDBWithPing(conn string) (*sql.DB, error) {
	db, err := NewDB(conn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	return db, nil
}
