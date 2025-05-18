package migration

import (
	"testing"
	"testing/fstest"

	"github.com/dominikbraun/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/types"
)

func TestComputeDefinitions(t *testing.T) {
	tests := []struct {
		name             string
		files            fstest.MapFS
		expectError      bool
		expectedVertices int
		expectedEdges    map[int64][]int64
	}{
		{
			name:             "empty filesystem",
			files:            fstest.MapFS{},
			expectedVertices: 0,
			expectedEdges:    map[int64][]int64{},
		},
		{
			name: "single migration",
			files: fstest.MapFS{
				"1651234567/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				"1651234567/down.sql":      {Data: []byte("DROP TABLE users;\n")},
				"1651234567/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
			},
			expectedVertices: 1,
			expectedEdges:    map[int64][]int64{1651234567: {}},
		},
		{
			name: "multiple migrations with parent relationship",
			files: fstest.MapFS{
				"1651234567/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				"1651234567/down.sql":      {Data: []byte("DROP TABLE users;\n")},
				"1651234567/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
				"1651234568/up.sql":        {Data: []byte("CREATE TABLE posts (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id));\n")},
				"1651234568/down.sql":      {Data: []byte("DROP TABLE posts;\n")},
				"1651234568/metadata.yaml": {Data: []byte("name: create_posts\ntimestamp: 1651234568\nparents: [1651234567]\n")},
			},
			expectedVertices: 2,
			expectedEdges: map[int64][]int64{
				1651234567: {1651234568},
				1651234568: {},
			},
		},
		{
			name: "missing required files",
			files: fstest.MapFS{
				"1651234567/up.sql": {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				// Missing down.sql and metadata.yaml
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ComputeDefinitions(tt.files)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Get the adjacency map from the result
			actualAdj, err := result.AdjacencyMap()
			require.NoError(t, err)

			// Check the number of vertices
			assert.Equal(t, tt.expectedVertices, len(actualAdj), "Graph should have the expected number of vertices")

			// Check that the edges match what we expect
			for vertex, expectedTargets := range tt.expectedEdges {
				actualTargets, exists := actualAdj[vertex]

				if tt.expectedVertices > 0 {
					// If we expect this vertex to exist
					assert.True(t, exists, "Vertex %d should exist in graph", vertex)

					if exists {
						// Compare the outgoing edges
						var actualEdges []int64
						for target := range actualTargets {
							actualEdges = append(actualEdges, target)
						}

						// Check that the expected targets match the actual ones
						assert.ElementsMatch(t, expectedTargets, actualEdges,
							"Outgoing edges from vertex %d should match expected", vertex)
					}
				}
			}
		})
	}
}

func TestComputeLeaves(t *testing.T) {
	tests := []struct {
		name           string
		graphSetup     func(t *testing.T) graph.Graph[int64, types.Definition]
		expectedLeaves []int64
		description    string // Description of what the test is checking
	}{
		{
			name:        "empty graph",
			description: "An empty graph should return an empty slice of leaves",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				return graph.New(definitionHash, graph.Acyclic(), graph.Directed())
			},
			expectedLeaves: []int64{},
		},
		{
			name:        "single node graph",
			description: "A graph with a single vertex should return that vertex as a leaf",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())
				def := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}
				require.NoError(t, g.AddVertex(def))
				return g
			},
			expectedLeaves: []int64{1651234567},
		},
		{
			name:        "linear graph",
			description: "A linear chain of migrations should return only the last vertex as a leaf",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())
				def1 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}
				def2 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_posts",
						Timestamp: 1651234568,
						Parents:   []int64{1651234567},
					},
				}
				def3 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_comments",
						Timestamp: 1651234569,
						Parents:   []int64{1651234568},
					},
				}
				require.NoError(t, g.AddVertex(def1))
				require.NoError(t, g.AddVertex(def2))
				require.NoError(t, g.AddVertex(def3))
				require.NoError(t, g.AddEdge(1651234567, 1651234568))
				require.NoError(t, g.AddEdge(1651234568, 1651234569))
				return g
			},
			expectedLeaves: []int64{1651234569},
		},
		{
			name:        "tree graph with multiple leaves",
			description: "A tree-like graph with multiple branches should return all terminal vertices as leaves",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())
				def1 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}
				def2 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_posts",
						Timestamp: 1651234568,
						Parents:   []int64{1651234567},
					},
				}
				def3 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_comments",
						Timestamp: 1651234569,
						Parents:   []int64{1651234567},
					},
				}
				require.NoError(t, g.AddVertex(def1))
				require.NoError(t, g.AddVertex(def2))
				require.NoError(t, g.AddVertex(def3))
				require.NoError(t, g.AddEdge(1651234567, 1651234568))
				require.NoError(t, g.AddEdge(1651234567, 1651234569))
				return g
			},
			expectedLeaves: []int64{1651234568, 1651234569},
		},
		{
			name:        "complex DAG with multiple leaves",
			description: "A complex directed acyclic graph should properly identify all terminal vertices as leaves",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())
				def1 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}
				def2 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_posts",
						Timestamp: 1651234568,
						Parents:   []int64{1651234567},
					},
				}
				def3 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_comments",
						Timestamp: 1651234569,
						Parents:   []int64{1651234568},
					},
				}
				def4 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_likes",
						Timestamp: 1651234570,
						Parents:   []int64{1651234567},
					},
				}
				def5 := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_views",
						Timestamp: 1651234571,
						Parents:   []int64{1651234570},
					},
				}
				require.NoError(t, g.AddVertex(def1))
				require.NoError(t, g.AddVertex(def2))
				require.NoError(t, g.AddVertex(def3))
				require.NoError(t, g.AddVertex(def4))
				require.NoError(t, g.AddVertex(def5))
				require.NoError(t, g.AddEdge(1651234567, 1651234568))
				require.NoError(t, g.AddEdge(1651234568, 1651234569))
				require.NoError(t, g.AddEdge(1651234567, 1651234570))
				require.NoError(t, g.AddEdge(1651234570, 1651234571))
				return g
			},
			expectedLeaves: []int64{1651234569, 1651234571},
		},
		{
			name:        "diamond shaped DAG",
			description: "A diamond-shaped DAG should identify only the bottom vertex as a leaf",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())
				// Create a diamond pattern:
				//    A
				//   / \
				//  B   C
				//   \ /
				//    D
				defA := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_schema",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}
				defB := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234568,
						Parents:   []int64{1651234567},
					},
				}
				defC := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_posts",
						Timestamp: 1651234569,
						Parents:   []int64{1651234567},
					},
				}
				defD := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_user_posts",
						Timestamp: 1651234570,
						Parents:   []int64{1651234568, 1651234569},
					},
				}
				require.NoError(t, g.AddVertex(defA))
				require.NoError(t, g.AddVertex(defB))
				require.NoError(t, g.AddVertex(defC))
				require.NoError(t, g.AddVertex(defD))
				require.NoError(t, g.AddEdge(1651234567, 1651234568)) // A -> B
				require.NoError(t, g.AddEdge(1651234567, 1651234569)) // A -> C
				require.NoError(t, g.AddEdge(1651234568, 1651234570)) // B -> D
				require.NoError(t, g.AddEdge(1651234569, 1651234570)) // C -> D
				return g
			},
			expectedLeaves: []int64{1651234570}, // Only D is a leaf
		},
		{
			name:        "multi-root DAG",
			description: "A graph with multiple root nodes should identify terminal vertices as leaves",
			graphSetup: func(t *testing.T) graph.Graph[int64, types.Definition] {
				t.Helper()
				g := graph.New(definitionHash, graph.Acyclic(), graph.Directed())
				// Two separate roots with their own paths
				defA := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users_schema",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}
				defB := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234568,
						Parents:   []int64{1651234567},
					},
				}
				defX := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_products_schema",
						Timestamp: 1651234600,
						Parents:   []int64{},
					},
				}
				defY := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_products",
						Timestamp: 1651234601,
						Parents:   []int64{1651234600},
					},
				}
				require.NoError(t, g.AddVertex(defA))
				require.NoError(t, g.AddVertex(defB))
				require.NoError(t, g.AddVertex(defX))
				require.NoError(t, g.AddVertex(defY))
				require.NoError(t, g.AddEdge(1651234567, 1651234568)) // A -> B
				require.NoError(t, g.AddEdge(1651234600, 1651234601)) // X -> Y
				return g
			},
			expectedLeaves: []int64{1651234568, 1651234601}, // B and Y are leaves
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the test graph according to the test case
			g := tt.graphSetup(t)

			// Execute the function being tested
			leaves, err := ComputeLeaves(g)

			// Verify the result
			require.NoError(t, err, "Got unexpected error: %v", err)

			// Check that the number of leaves matches expected
			assert.Len(t, leaves, len(tt.expectedLeaves),
				"Expected %d leaves but got %d", len(tt.expectedLeaves), len(leaves))

			// Check that the leaves match expected values
			assert.ElementsMatch(t, tt.expectedLeaves, leaves,
				"Leaves don't match expected values")

			// Verify each leaf is actually a leaf by checking adjacency map
			adjMap, err := g.AdjacencyMap()
			require.NoError(t, err, "Failed to get adjacency map")

			for _, leaf := range leaves {
				outNeighbors, exists := adjMap[leaf]
				assert.True(t, exists, "Leaf %d not found in graph", leaf)
				assert.Empty(t, outNeighbors, "Vertex %d has outgoing edges, not a leaf", leaf)
			}
		})
	}
}
