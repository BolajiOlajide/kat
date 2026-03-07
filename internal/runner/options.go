package runner

import (
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Options represents the options for the runner.
type Options struct {
	Operation     types.MigrationOperationType
	Definitions   *graph.Graph
	MigrationInfo types.MigrationInfo
	DryRun        bool
	Verbose       bool
	Count         int
}
