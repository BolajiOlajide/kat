package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/types"
)

func TestGraph_Leaves(t *testing.T) {
	tests := []struct {
		name           string
		graphSetup     func(t *testing.T) *Graph
		expectedLeaves []int64
		description    string // Description of what the test is checking
	}{
		{
			name:        "empty graph",
			description: "An empty graph should return an empty slice of leaves",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				return New()
			},
			expectedLeaves: []int64{},
		},
		{
			name:        "single node graph",
			description: "A graph with a single vertex should return that vertex as a leaf",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				g := New()
				def := types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1651234567,
						Parents:   []int64{},
					},
				}

				require.NoError(t, g.AddDefinition(def))
				return g
			},
			expectedLeaves: []int64{1651234567},
		},
		{
			name:        "linear graph",
			description: "A linear chain of migrations should return only the last vertex as a leaf",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				g := New()
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
				require.NoError(t, g.AddDefinitions(def1, def2, def3))
				return g
			},
			expectedLeaves: []int64{1651234569},
		},
		{
			name:        "tree graph with multiple leaves",
			description: "A tree-like graph with multiple branches should return all terminal vertices as leaves",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				g := New()
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

				require.NoError(t, g.AddDefinitions(def1, def2, def3))
				return g
			},
			expectedLeaves: []int64{1651234568, 1651234569},
		},
		{
			name:        "complex DAG with multiple leaves",
			description: "A complex directed acyclic graph should properly identify all terminal vertices as leaves",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				g := New()
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
				require.NoError(t, g.AddDefinitions(def1, def2, def3, def4, def5))
				return g
			},
			expectedLeaves: []int64{1651234569, 1651234571},
		},
		{
			name:        "diamond shaped DAG",
			description: "A diamond-shaped DAG should identify only the bottom vertex as a leaf",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				g := New()
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
				require.NoError(t, g.AddDefinitions(defA, defB, defC, defD))
				return g
			},
			expectedLeaves: []int64{1651234570}, // Only D is a leaf
		},
		{
			name:        "multi-root DAG",
			description: "A graph with multiple root nodes should identify terminal vertices as leaves",
			graphSetup: func(t *testing.T) *Graph {
				t.Helper()
				g := New()
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
				require.NoError(t, g.AddDefinitions(defA, defB, defX, defY))
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
			leaves, err := g.Leaves()

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
