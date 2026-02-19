package runner

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"
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

func (n *noTransactTx) Exec(ctx context.Context, q *sqlf.Query) error         { return n.db.Exec(ctx, q) }
func (n *noTransactTx) QueryRow(ctx context.Context, q *sqlf.Query) *sql.Row   { return n.db.QueryRow(ctx, q) }
func (n *noTransactTx) Query(ctx context.Context, q *sqlf.Query) (*sql.Rows, error) { return n.db.Query(ctx, q) }
func (n *noTransactTx) Ping(ctx context.Context) error                         { return n.db.Ping(ctx) }
func (n *noTransactTx) PingWithRetry(ctx context.Context, c, d int) error      { return n.db.PingWithRetry(ctx, c, d) }
func (n *noTransactTx) WithTransact(_ context.Context, _ func(database.Tx) error) error { return errors.New("transactions not supported for no-transaction execution") }

func (n *noTransactTx) Close() error    { return errors.New("close not supported for no-transaction execution") }
func (n *noTransactTx) Commit() error   { return errors.New("commit not supported for no-transaction execution") }
func (n *noTransactTx) Rollback() error { return errors.New("rollback not supported for no-transaction execution") }
