package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/keegancsmith/sqlf"
)

var _ Tx = &databaseTx{}

type databaseTx struct {
	tx      *sql.Tx
	bindVar sqlf.BindVar
}

func (d *databaseTx) Ping(ctx context.Context) error {
	return errors.New("ping not supported in transaction")
}

func (d *databaseTx) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := d.tx.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

func (d *databaseTx) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return d.tx.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *databaseTx) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return d.tx.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *databaseTx) Close() error {
	return errors.New("close method not supported in transaction")
}

func (d *databaseTx) WithTransact(ctx context.Context, f func(Tx) error) error {
	return errors.New("nested transactions are not supported")
}

func (d *databaseTx) Commit() error {
	return d.tx.Commit()
}

func (d *databaseTx) Rollback() error {
	return d.tx.Rollback()
}
