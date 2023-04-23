package database

import (
	"context"
	"database/sql"

	// Import the postgres driver

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keegancsmith/sqlf"
)

type database struct {
	db      *sql.DB
	bindVar sqlf.BindVar
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *database) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := d.db.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

func (d *database) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return d.db.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *database) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
}

// NewDB returns a new instance of the database
func NewDB(url string) (DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	return &database{db: db, bindVar: sqlf.PostgresBindVar}, nil
}
