package database

import (
	"context"
	"database/sql"

	// Import the postgres driver

	"github.com/keegancsmith/sqlf"
	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DB) Exec(ctx context.Context, query *sqlf.Query, args ...any) error {
	_, err := d.db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), args...)
	return err
}

// NewDB returns a new instance of the database
func NewDB(url string) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}
