package database

import (
	"database/sql"

	// Import the postgres driver
	_ "github.com/lib/pq"
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
