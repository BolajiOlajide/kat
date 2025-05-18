package migration

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/types"
)

func TestComputeDefinition(t *testing.T) {
	tests := []struct {
		name         string
		files        fstest.MapFS
		migrationDir string
		expectedDef  types.Definition
		expectError  bool
	}{
		{
			name: "valid migration",
			files: fstest.MapFS{
				"1651234567_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				"1651234567_create_users/down.sql":      {Data: []byte("DROP TABLE users;\n")},
				"1651234567_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
			},
			migrationDir: "1651234567_create_users",
			expectedDef: types.Definition{
				MigrationMetadata: types.MigrationMetadata{
					Name:      "create_users",
					Timestamp: 1651234567,
					Parents:   []int64{},
				},
				// We won't compare the actual query content since sqlf.Query doesn't easily support equality testing
			},
		},
		{
			name: "migration with parents",
			files: fstest.MapFS{
				"1651234568_create_posts/up.sql":        {Data: []byte("CREATE TABLE posts (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES users(id));\n")},
				"1651234568_create_posts/down.sql":      {Data: []byte("DROP TABLE posts;\n")},
				"1651234568_create_posts/metadata.yaml": {Data: []byte("name: create_posts\ntimestamp: 1651234568\nparents: [1651234567]\n")},
			},
			migrationDir: "1651234568_create_posts",
			expectedDef: types.Definition{
				MigrationMetadata: types.MigrationMetadata{
					Name:      "create_posts",
					Timestamp: 1651234568,
					Parents:   []int64{1651234567},
				},
				// We won't compare the actual query content since sqlf.Query doesn't easily support equality testing
			},
		},
		{
			name: "missing up.sql file",
			files: fstest.MapFS{
				"1651234567_create_users/down.sql":      {Data: []byte("DROP TABLE users;\n")},
				"1651234567_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
			},
			migrationDir: "1651234567_create_users",
			expectError:  true,
		},
		{
			name: "missing down.sql file",
			files: fstest.MapFS{
				"1651234567_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				"1651234567_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
			},
			migrationDir: "1651234567_create_users",
			expectError:  true,
		},
		{
			name: "missing metadata.yaml file",
			files: fstest.MapFS{
				"1651234567_create_users/up.sql":   {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				"1651234567_create_users/down.sql": {Data: []byte("DROP TABLE users;\n")},
			},
			migrationDir: "1651234567_create_users",
			expectError:  true,
		},
		{
			name: "invalid metadata.yaml file",
			files: fstest.MapFS{
				"1651234567_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
				"1651234567_create_users/down.sql":      {Data: []byte("DROP TABLE users;\n")},
				"1651234567_create_users/metadata.yaml": {Data: []byte("invalid: yaml: format")},
			},
			migrationDir: "1651234567_create_users",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := computeDefinition(tt.files, tt.migrationDir)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Check metadata fields
			assert.Equal(t, tt.expectedDef.Name, result.Name)
			assert.Equal(t, tt.expectedDef.Timestamp, result.Timestamp)
			assert.ElementsMatch(t, tt.expectedDef.Parents, result.Parents)

			// Check that queries are not nil
			assert.NotNil(t, result.UpQuery)
			assert.NotNil(t, result.DownQuery)

			// Verify query content by converting to string and comparing
			upSQL := result.UpQuery.Query(sqlf.PostgresBindVar)
			downSQL := result.DownQuery.Query(sqlf.PostgresBindVar)

			// Get expected SQL from the map and trim newlines for comparison
			expectedUpSQL := strings.TrimSpace(string(tt.files[tt.migrationDir+"/up.sql"].Data))
			expectedDownSQL := strings.TrimSpace(string(tt.files[tt.migrationDir+"/down.sql"].Data))

			// Compare SQL strings
			assert.Equal(t, expectedUpSQL, upSQL)
			assert.Equal(t, expectedDownSQL, downSQL)
		})
	}
}
