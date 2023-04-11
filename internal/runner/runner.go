package runner

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/keegancsmith/sqlf"
)

// Runner is the interface that every runner must implement.
type Runner interface {
	Run(context.Context, Options) error
}

type runner struct {
	db *database.DB
}

var _ Runner = (*runner)(nil)

// NewRunner returns a new instance of the runner.
func NewRunner(ctx context.Context, db *database.DB) (Runner, error) {
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}
	return &runner{db: db}, nil
}

func (r *runner) Run(ctx context.Context, options Options) error {
	for _, definition := range options.Definitions {
		fmt.Printf("%s%s%s ", output.StyleInfo, definition.Name, output.StyleReset)

		var q *sqlf.Query
		var migrationKind string
		if options.Operation == types.UpMigrationOperation {
			q = definition.UpQuery
			migrationKind = "up"
		} else {
			q = definition.DownQuery
			migrationKind = "down"
		}

		err := r.db.Exec(ctx, q, q.Args()...)
		if err != nil {
			return errors.Wrapf(err, "executing %s query", migrationKind)
		}

		// add a new line incase there's an error
		fmt.Print("\n")
	}

	return nil
}
