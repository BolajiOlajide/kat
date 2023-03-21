package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/keegancsmith/sqlf"
)

type DB interface {
	Exec(ctx context.Context, query sqlf.Query) error
	QueryRow(ctx context.Context, query sqlf.Query) *sql.Row
	Query(ctx context.Context, query sqlf.Query) (*sql.Rows, error)
	Ping(ctx context.Context) error
	Transact(context.Context, func(*sql.Tx) error) error
}

type db struct {
	*sql.DB
}

func (db *db) Exec(ctx context.Context, query sqlf.Query) error {
	_, err := db.ExecContext(ctx, query.Query(&sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		return fmt.Errorf("exec query %q failed: %w", query, err)
	}
	return nil
}

func (db *db) QueryRow(ctx context.Context, query sqlf.Query) *sql.Row {
	return db.QueryRowContext(ctx, query.Query(&sqlf.PostgresBindVar), query.Args()...)
}

func (db *db) Query(ctx context.Context, query sqlf.Query) (*sql.Rows, error) {
	return db.QueryContext(ctx, query.Query(&sqlf.PostgresBindVar), query.Args()...)
}

func (db *db) Transact(ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	if err := f(tx); err != nil {
		if txErr := tx.Rollback(); txErr != nil {
			return fmt.Errorf("transaction rollback failed: %v (original error: %v)", txErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (db *db) Ping(ctx context.Context) error {
	return db.PingContext(ctx)
}
