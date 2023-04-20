package database

import (
	"context"
	"database/sql"

	// Import the postgres driver

	"github.com/keegancsmith/sqlf"
	_ "github.com/lib/pq"
)

type DB struct {
	db      *sql.DB
	bindVar sqlf.BindVar
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *DB) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := d.db.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

func (d *DB) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return d.db.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *DB) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
}

// NewDB returns a new instance of the database
func NewDB(url string) (*DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	return &DB{db: db, bindVar: sqlf.PostgresBindVar}, nil
}
