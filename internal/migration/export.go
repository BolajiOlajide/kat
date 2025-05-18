package migration

import (
	"io"

	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/types"
)

// ExportGraph exports a visualization of the migration graph.
// It supports exporting in DOT format (for Graphviz) or JSON format.
// The output is written to the provided writer (typically stdout).
func ExportGraph(w io.Writer, cfg types.Config) error {
	// get filesystem for the migrations directory
	dirFS, err := getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return err
	}

	// Compute the migration graph
	g, err := ComputeDefinitions(dirFS)
	if err != nil {
		return errors.Wrap(err, "failed to compute migration definitions")
	}

	// Check if there are any migrations
	order, err := g.Order()
	if err != nil {
		return errors.Wrap(err, "failed to compute graph order")
	}

	if order == 0 {
		return errors.New("no migrations found")
	}

	return g.Draw(w)
}
