package migration

import (
	"io/fs"

	"github.com/cockroachdb/errors"
	"github.com/dominikbraun/graph"

	"github.com/BolajiOlajide/kat/internal/types"
)

var definitionHash = func(d types.Definition) int64 {
	return d.Timestamp
}

// ComputeDefinitions builds a directed acyclic graph (DAG) from migration definitions.
// It scans the provided filesystem for migration directories, creates vertices for each migration,
// and establishes parent-child relationships between them. The resulting graph enables proper
// migration ordering and dependency resolution.
//
// The function expects directories to be timestamp-prefixed following Kat's convention.
// It returns the constructed graph and any errors encountered during graph construction.
func ComputeDefinitions(f fs.FS) (graph.Graph[int64, types.Definition], error) {
	// create a DAG (Directed Acyclic graph) to represent a migration. This makes it easy
	// to compute execution order for migrations and also to export the graph as a digraph
	// diagram.
	// You can visualize using an online tool like:
	// https://dreampuf.github.io/GraphvizOnline
	g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())

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
			return nil, errors.Wrap(err, "malformed migration definition")
		}

		// We add each migration as a vertex to the graph
		if err := g.AddVertex(definition); err != nil {
			return nil, errors.Wrap(err, "error adding vertex")
		}

		// Then we define the relationship with its parent by adding its edges.
		for _, parent := range definition.Parents {
			if err := g.AddEdge(parent, definition.Timestamp); err != nil {
				return nil, errors.Wrapf(err, "error adding edge for parent: %d", parent)
			}
		}
	}

	return g, nil
}

func ComputeLeaves(g graph.Graph[int64, types.Definition]) ([]int64, error) {
	// build the adjacency map: for each vertex, map of outgoing edges
	adj, err := g.AdjacencyMap()
	if err != nil {
		return nil, errors.Wrap(err, "adjacency map")
	}

	var leaves []int64
	for v, outs := range adj {
		// no outgoing neighbors → it's a leaf
		if len(outs) == 0 {
			leaves = append(leaves, v)
		}
	}
	return leaves, nil
}
