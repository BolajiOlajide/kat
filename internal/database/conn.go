package database

import (
	"database/sql"

	// Import the postgres driver
	_ "github.com/lib/pq"
)

// NewDB returns a new instance of the database
func NewDB(conn string) (Database, error) {
	dbconn, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	defer dbconn.Close()
	return &db{dbconn}, nil
}
