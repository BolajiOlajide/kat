package migration

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
