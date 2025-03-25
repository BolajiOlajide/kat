package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
)

type DB interface {
	Close() error
	Ping(context.Context) error
	Exec(context.Context, *sqlf.Query) error
	QueryRow(context.Context, *sqlf.Query) *sql.Row
	Query(context.Context, *sqlf.Query) (*sql.Rows, error)
	ValidateQuery(context.Context, *sqlf.Query) error

	WithTransact(ctx context.Context, f func(Tx) error) error
}

type Scanner interface {
	Scan(dest ...any) error
}

type Tx interface {
	DB

	Commit() error
	Rollback() error
}
