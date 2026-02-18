package runner

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/BolajiOlajide/kat/internal/database"
)

// noTransactTx implements database.Tx by delegating directly to the underlying
// database.DB without wrapping operations in a transaction. This is needed for
// migrations that use operations like CREATE INDEX CONCURRENTLY which cannot
// run inside a transaction block.
type noTransactTx struct {
	db database.DB
}

var _ database.Tx = &noTransactTx{}

func (n *noTransactTx) Close() error                                        { return n.db.Close() }
func (n *noTransactTx) Ping(ctx context.Context) error                      { return n.db.Ping(ctx) }
func (n *noTransactTx) PingWithRetry(ctx context.Context, c, d int) error   { return n.db.PingWithRetry(ctx, c, d) }
func (n *noTransactTx) Exec(ctx context.Context, q *sqlf.Query) error       { return n.db.Exec(ctx, q) }
func (n *noTransactTx) QueryRow(ctx context.Context, q *sqlf.Query) *sql.Row { return n.db.QueryRow(ctx, q) }
func (n *noTransactTx) Query(ctx context.Context, q *sqlf.Query) (*sql.Rows, error) { return n.db.Query(ctx, q) }
func (n *noTransactTx) WithTransact(ctx context.Context, f func(database.Tx) error) error { return n.db.WithTransact(ctx, f) }
func (n *noTransactTx) Commit() error                                       { return nil }
func (n *noTransactTx) Rollback() error                                     { return nil }
