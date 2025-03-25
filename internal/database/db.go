package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	// Import the postgres driver
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/keegancsmith/sqlf"
)

var _ DB = &database{}

type database struct {
	db      *sql.DB
	bindVar sqlf.BindVar
}

func (d *database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *database) Exec(ctx context.Context, query *sqlf.Query) error {
	_, err := d.db.ExecContext(ctx, query.Query(d.bindVar), query.Args()...)
	return err
}

// ValidateQuery checks if a SQL query is valid without executing it
// It uses Postgres's EXPLAIN to validate the query syntax
func (d *database) ValidateQuery(ctx context.Context, query *sqlf.Query) error {
	// Extract the SQL query
	sqlQuery := query.Query(d.bindVar)
	
	// Skip empty queries
	if strings.TrimSpace(sqlQuery) == "" {
		return errors.New("empty SQL query")
	}
	
	// For non-SELECT queries, we need to wrap them in EXPLAIN
	// This validates the query without executing it
	var explainQuery string
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sqlQuery)), "SELECT") {
		explainQuery = fmt.Sprintf("EXPLAIN %s", sqlQuery)
	} else {
		explainQuery = fmt.Sprintf("EXPLAIN (ANALYZE FALSE) %s", sqlQuery)
	}
	
	// Execute the EXPLAIN query to verify syntax
	_, err := d.db.ExecContext(ctx, explainQuery, query.Args()...)
	if err != nil {
		return errors.Wrap(err, "SQL validation failed")
	}
	
	return nil
}

func (d *database) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	return d.db.QueryRowContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *database) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query.Query(d.bindVar), query.Args()...)
}

func (d *database) Close() error {
	return d.db.Close()
}

func (d *database) WithTransact(ctx context.Context, f func(Tx) error) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	return f(&databaseTx{tx: tx, bindVar: d.bindVar})
}

// NewDB returns a new instance of the database
func NewDB(url string, bindvar sqlf.BindVar) (DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	return &database{db: db, bindVar: bindvar}, nil
}
