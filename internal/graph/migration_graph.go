package graph

import (
	"sort"

	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/cockroachdb/errors"
)

// MigrationGraph represents a directed acyclic graph of migrations
type MigrationGraph struct {
	nodes map[string]*graphNode
}

// graphNode represents a node in the migration dependency graph
type graphNode struct {
	name      string
	parents   []*graphNode
	children  []*graphNode
	definition types.Definition
	visited    bool
	temporary  bool // Used for cycle detection
}

// NewMigrationGraph creates a new migration graph from a list of definitions
func NewMigrationGraph(definitions []types.Definition) (*MigrationGraph, error) {
	graph := &MigrationGraph{
		nodes: make(map[string]*graphNode),
	}

	// First pass: create nodes for all migrations
	for _, def := range definitions {
		graph.nodes[def.Name] = &graphNode{
			name:       def.Name,
			definition: def,
			parents:    []*graphNode{},
			children:   []*graphNode{},
		}
	}

	// Second pass: connect nodes based on parent relationships
	for _, def := range definitions {
		node := graph.nodes[def.Name]
		for _, parentTimestamp := range def.MigrationMetadata.Parents {
			// Find parent node by timestamp
			var parentNode *graphNode
			var found bool

			for _, n := range graph.nodes {
				if n.definition.Timestamp == parentTimestamp {
					parentNode = n
					found = true
					break
				}
			}

			if !found {
				// This happens when a migration references a parent that doesn't exist
				return nil, errors.Newf("migration %s references non-existent parent timestamp %d", def.Name, parentTimestamp)
			}

			// Add the parent to this node's parents list
			node.parents = append(node.parents, parentNode)
			
			// Add this node to the parent's children list
			parentNode.children = append(parentNode.children, node)
		}
	}

	// Check for cycles
	if err := graph.checkForCycles(); err != nil {
		return nil, err
	}

	return graph, nil
}

// checkForCycles detects cycles in the graph using depth-first search
func (g *MigrationGraph) checkForCycles() error {
	// Reset all visited flags
	for _, node := range g.nodes {
		node.visited = false
		node.temporary = false
	}

	// Check for cycles starting from each node
	for _, node := range g.nodes {
		if !node.visited {
			if err := g.visitNode(node); err != nil {
				return err
			}
		}
	}

	return nil
}

// visitNode performs depth-first search from a node to detect cycles
func (g *MigrationGraph) visitNode(node *graphNode) error {
	// If we've already processed this node, we're good
	if node.visited {
		return nil
	}

	// If we see a temporary mark, we have a cycle
	if node.temporary {
		return errors.Newf("cycle detected in migration graph involving %s", node.name)
	}

	// Mark this node as temporarily visited (in process)
	node.temporary = true

	// Visit all children
	for _, child := range node.children {
		if err := g.visitNode(child); err != nil {
			return err
		}
	}

	// Mark as permanently visited
	node.temporary = false
	node.visited = true

	return nil
}

// TopologicalSort returns migrations in order they should be executed
func (g *MigrationGraph) TopologicalSort() []types.Definition {
	// Reset all visited flags
	for _, node := range g.nodes {
		node.visited = false
	}

	// Create a sorted list
	var sorted []types.Definition

	// Get the nodes without parents (roots)
	roots := g.getRootNodes()

	// Process nodes in timestamp order within each level
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].definition.Timestamp < roots[j].definition.Timestamp
	})

	// Process each root node
	for _, node := range roots {
		g.visit(node, &sorted)
	}

	return sorted
}

// getRootNodes returns all nodes that have no parents
func (g *MigrationGraph) getRootNodes() []*graphNode {
	var roots []*graphNode
	for _, node := range g.nodes {
		if len(node.parents) == 0 {
			roots = append(roots, node)
		}
	}
	return roots
}

// visit traverses the graph in topological order
func (g *MigrationGraph) visit(node *graphNode, sorted *[]types.Definition) {
	if node.visited {
		return
	}

	node.visited = true

	// Sort children by timestamp so siblings are processed in timestamp order
	sort.Slice(node.children, func(i, j int) bool {
		return node.children[i].definition.Timestamp < node.children[j].definition.Timestamp
	})

	// Process all children
	for _, child := range node.children {
		g.visit(child, sorted)
	}

	// Add this node to the sorted list
	*sorted = append(*sorted, node.definition)
}

// GetMissingMigrations returns migrations that need to be applied
func (g *MigrationGraph) GetMissingMigrations(appliedMigrations map[string]*types.MigrationLog) []types.Definition {
	// Get a topologically sorted order of all migrations
	allSorted := g.TopologicalSort()

	// Filter out already applied migrations
	var missing []types.Definition
	for _, def := range allSorted {
		if _, ok := appliedMigrations[def.Name]; !ok {
			missing = append(missing, def)
		}
	}

	return missing
}