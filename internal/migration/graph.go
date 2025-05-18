package migration

import (
	"io/fs"

	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/graph"
)

// ComputeDefinitions builds a directed acyclic graph (DAG) from migration definitions.
// It scans the provided filesystem for migration directories, creates vertices for each migration,
// and establishes parent-child relationships between them. The resulting graph enables proper
// migration ordering and dependency resolution.
//
// The function expects directories to be timestamp-prefixed following Kat's convention.
// It returns the constructed graph and any errors encountered during graph construction.
func ComputeDefinitions(f fs.FS) (*graph.Graph, error) {
	g := graph.New()

	mf, err := extractMigrationFiles(f)
	if err != nil {
		return nil, err
	}

	// Because Kat prepends it's directories with a timestamp, we expect the migrations
	// to be automatically sorted, so we can pretty much assume the loop is going to be
	// in ascending order from the oldest migrations to newest.
	//
	// If this isn't the case, adding an edge will result in an error, because the parent
	// edge won't be found in the graph.
	for _, file := range mf {
		if !file.IsDir() {
			// if this is not a directory, skip it.

			// Kat expects migrations to live in directories, so this is one of the most
			// important validations we can have
			continue
		}

		definition, err := computeDefinition(f, file.Name())
		if err != nil {
			return g, errors.Wrap(err, "malformed migration definition")
		}

		if err := g.AddDefinition(definition); err != nil {
			return nil, errors.Wrap(err, "failed to add definition to graph")
		}
	}

	return g, nil
}
