package runner

import (
	"context"

	"github.com/BolajiOlajide/kat/internal/database"
)

// Runner is the interface that every runner must implement.
type Runner interface {
	Run(context.Context, Options) error
}

type runner struct {
	db database.DB
}

var _ Runner = (*runner)(nil)

// NewRunner returns a new instance of the runner.
func NewRunner(db database.DB) Runner {
	return &runner{db: db}
}

func (r *runner) Run(ctx context.Context, options Options) error {
	return nil
}
