package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
)

type DB interface {
	Close() error
	Ping(context.Context) error
	PingWithRetry(context.Context, int, int) error // Only used by ping command
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
	// For transactions, we only implement the basic DB interface methods,
	// not the retry versions since retries happen at the DB level
	Close() error
	Ping(context.Context) error
	Exec(context.Context, *sqlf.Query) error
	QueryRow(context.Context, *sqlf.Query) *sql.Row
	Query(context.Context, *sqlf.Query) (*sql.Rows, error)
	ValidateQuery(context.Context, *sqlf.Query) error
	WithTransact(ctx context.Context, f func(Tx) error) error

	Commit() error
	Rollback() error
}
