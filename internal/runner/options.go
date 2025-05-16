package runner

import (
	"github.com/dominikbraun/graph"

	"github.com/BolajiOlajide/kat/internal/types"
)

// Options represents the options for the runner.
type Options struct {
	Operation     types.MigrationOperationType
	Definitions   graph.Graph[int64, types.Definition]
	MigrationInfo types.MigrationInfo
	DryRun        bool
	Verbose       bool
}
