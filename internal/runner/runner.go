package runner

import (
	"context"
	"database/sql"
)

// Runner is the interface that every runner must implement.
type Runner interface {
	Run(context.Context, Options) error
}

type runner struct {
	db *sql.DB
}

var _ Runner = (*runner)(nil)

// NewRunner returns a new instance of the runner.
func NewRunner(db *sql.DB) Runner {
	return &runner{db: db}
}

func (r *runner) Run(ctx context.Context, options Options) error {
	return nil
}
