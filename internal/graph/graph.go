package graph

import (
	"io"
	"sort"

	"github.com/cockroachdb/errors"
	graphlib "github.com/dominikbraun/graph"
	graphlibdraw "github.com/dominikbraun/graph/draw"

	"github.com/BolajiOlajide/kat/internal/types"
)

var definitionHash = func(d types.Definition) int64 {
	return d.Timestamp
}

// New creates a DAG (Directed Acyclic graph) to represent a migration. This makes it easy
// to compute execution order for migrations and also to export the graph as a digraph
// diagram.
//
// You can visualize using an online tool like:
// https://dreampuf.github.io/GraphvizOnline
func New() *Graph {
	g := graphlib.New(definitionHash, graphlib.Acyclic(), graphlib.Directed())
	return &Graph{graph: g}
}

type Graph struct {
	graph graphlib.Graph[int64, types.Definition]
}

func (g *Graph) AddDefinition(def types.Definition) error {
	if err := g.graph.AddVertex(def, graphlib.VertexAttribute("name", def.FileName())); err != nil {
		return errors.Wrap(err, "error adding vertex")
	}

	// Then we define the relationship with its parent by adding its edges.
	for _, parent := range def.Parents {
		if err := g.graph.AddEdge(parent, def.Timestamp); err != nil {
			return errors.Wrapf(err, "error adding edge for parent: %d", parent)
		}
	}

	return nil
}

func (g *Graph) AddDefinitions(defs ...types.Definition) error {
	for _, def := range defs {
		if err := g.AddDefinition(def); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) Leaves() ([]int64, error) {
	// build the adjacency map: for each vertex, map of outgoing edges
	adj, err := g.graph.AdjacencyMap()
	if err != nil {
		return nil, errors.Wrap(err, "getting adjacency map")
	}

	var leaves []int64
	for v, outs := range adj {
		// no outgoing neighbors → it's a leaf
		if len(outs) == 0 {
			leaves = append(leaves, v)
		}
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
	return leaves, nil
}

func (g *Graph) AdjacencyMap() (map[int64]map[int64]graphlib.Edge[int64], error) {
	return g.graph.AdjacencyMap()
}

func (g *Graph) GetDefinition(timestamp int64) (types.Definition, error) {
	return g.graph.Vertex(timestamp)
}

// TopologicalSort returns a valid topological ordering of all the vertices in the graph.
// It uses StableTopologicalSort from the graph library to ensure that elements with
// valid topological ordering are consistently returned in order of their timestamps (i < j),
// making the results deterministic and predictable.
func (g *Graph) TopologicalSort() ([]int64, error) {
	return graphlib.StableTopologicalSort(g.graph, func(i, j int64) bool {
		return i < j
	})
}

func (g *Graph) Order() (int, error) {
	return g.graph.Order()
}

func (g *Graph) Draw(w io.Writer) error {
	return graphlibdraw.DOT(g.graph, w)
}
