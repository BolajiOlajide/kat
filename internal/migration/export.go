package migration

import (
	"io"

	"github.com/BolajiOlajide/kat/internal/types"
)

func ExportGraph(w io.Writer, cfg types.Config) error {
	// get filesystem for the migrations directory
	dirFS, err := getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return err
	}

	// Compute the migration graph
	g, err := ComputeDefinitions(dirFS)
	if err != nil {
		return err
	}

	return g.Draw(w)
}
