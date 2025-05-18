package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/types"
)

// ExportGraph exports a visualization of the migration graph.
// It supports exporting in DOT format (for Graphviz) or JSON format.
// The output is written to the provided writer (typically stdout).
func ExportGraph(ctx context.Context, w io.Writer, cfg types.Config, format string) error {
	// Validate format first, before any processing
	switch format {
	case "dot", "json":
		// Valid formats, continue
	default:
		return errors.Newf("unsupported format: %s", format)
	}

	// Create filesystem for the migrations directory
	dirFS := os.DirFS(cfg.Migration.Directory)

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

	// Export in the validated format
	switch format {
	case "dot":
		return exportDOT(w, g, dirFS)
	case "json":
		return exportJSON(w, g, dirFS)
	default:
		// This should never happen due to validation above, but maintain as a safeguard
		return errors.Newf("unsupported format: %s", format)
	}
}

// exportDOT exports the migration graph in DOT format for visualization with Graphviz.
func exportDOT(w io.Writer, g *graph.Graph, fs fs.FS) error {
	// Get the internal graph representation
	adjMap, err := g.AdjacencyMap()
	if err != nil {
		return errors.Wrap(err, "failed to get adjacency map")
	}

	// Create a mapping of vertex IDs to names for a more readable graph
	vertexNames := make(map[int64]string)
	for vertex := range adjMap {
		def, err := g.GetDefinition(vertex)
		if err != nil {
			return errors.Wrap(err, "failed to get vertex definition")
		}
		vertexNames[vertex] = fmt.Sprintf("%s\n(%d)", def.Name, def.Timestamp)
	}

	// Write the DOT representation to the output writer
	fmt.Fprintf(w, "digraph Migrations {\n")
	fmt.Fprintf(w, "  node [shape=box];\n")

	// Add all vertices with custom labels
	for vertex, name := range vertexNames {
		fmt.Fprintf(w, "  \"%d\" [label=\"%s\"];\n", vertex, name)
	}

	// Add all edges
	for vertex, edges := range adjMap {
		for target := range edges {
			fmt.Fprintf(w, "  \"%d\" -> \"%d\";\n", vertex, target)
		}
	}

	fmt.Fprintf(w, "}\n")
	return nil
}

// exportJSON exports the migration graph in JSON format for programmatic use.
func exportJSON(w io.Writer, g *graph.Graph, fs fs.FS) error {
	// Prepare a data structure for the JSON output
	type MigrationNode struct {
		Timestamp int64    `json:"timestamp"`
		Name      string   `json:"name"`
		Parents   []int64  `json:"parents"`
		Children  []int64  `json:"children"`
	}

	adjMap, err := g.AdjacencyMap()
	if err != nil {
		return errors.Wrap(err, "failed to get adjacency map")
	}

	// Create a reverse adjacency map for getting parents
	reverseAdj := make(map[int64][]int64)
	for parent, children := range adjMap {
		for child := range children {
			reverseAdj[child] = append(reverseAdj[child], parent)
		}
	}

	// Convert to nodes for JSON export
	nodes := make(map[string]MigrationNode)
	for vertex := range adjMap {
		def, err := g.GetDefinition(vertex)
		if err != nil {
			return errors.Wrap(err, "failed to get vertex definition")
		}

		// Get children
		var children []int64
		for child := range adjMap[vertex] {
			children = append(children, child)
		}

		// Create node
		nodes[fmt.Sprintf("%d", vertex)] = MigrationNode{
			Timestamp: vertex,
			Name:      def.Name,
			Parents:   reverseAdj[vertex],
			Children:  children,
		}
	}

	// Write JSON to the output writer
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(nodes)
}