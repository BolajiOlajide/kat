package runner

import "github.com/BolajiOlajide/kat/internal/types"

// Options represents the options for the runner.
type Options struct {
	Operation     types.MigrationOperationType
	Definitions   []types.Definition
	MigrationInfo types.MigrationInfo
}
