package migration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/types"
)

func TestExportGraph(t *testing.T) {
	// Create a test directory
	tmpDir := t.TempDir()

	// Create test migration directories (using timestamp pattern)
	dir1 := filepath.Join(tmpDir, "1620000000000_initial_schema")
	dir2 := filepath.Join(tmpDir, "1620100000000_add_users_table")
	
	// Create directories and files
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.MkdirAll(dir2, 0755))
	
	// Create migration files
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "up.sql"), []byte("CREATE TABLE users;"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "down.sql"), []byte("DROP TABLE users;"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "up.sql"), []byte("CREATE TABLE posts;"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "down.sql"), []byte("DROP TABLE posts;"), 0644))
	
	// Create metadata files
	metadata1 := filepath.Join(dir1, "metadata.yaml")
	metadata2 := filepath.Join(dir2, "metadata.yaml")
	
	// Write test metadata (migration2 depends on migration1)
	require.NoError(t, os.WriteFile(metadata1, []byte("name: initial_schema\ntimestamp: 1620000000000\nparents: []\n"), 0644))
	require.NoError(t, os.WriteFile(metadata2, []byte("name: add_users_table\ntimestamp: 1620100000000\nparents: [1620000000000]\n"), 0644))
	
	// Create test configuration
	cfg := types.Config{
		Migration: types.MigrationInfo{
			Directory: tmpDir,
		},
	}
	
	// Test DOT format export
	t.Run("DOT format export", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := ExportGraph(context.Background(), buf, cfg, "dot")
		require.NoError(t, err)
		
		// Check output contains expected elements
		dot := buf.String()
		require.Contains(t, dot, "digraph Migrations")
		require.Contains(t, dot, "\"1620000000000\" [label=\"initial_schema")
		require.Contains(t, dot, "\"1620100000000\" [label=\"add_users_table")
		require.Contains(t, dot, "\"1620000000000\" -> \"1620100000000\"")
	})
	
	// Test JSON format export
	t.Run("JSON format export", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := ExportGraph(context.Background(), buf, cfg, "json")
		require.NoError(t, err)
		
		// Check output contains expected elements
		json := buf.String()
		require.Contains(t, json, "\"timestamp\": 1620000000000")
		require.Contains(t, json, "\"name\": \"initial_schema\"")
		require.Contains(t, json, "\"timestamp\": 1620100000000")
		require.Contains(t, json, "\"name\": \"add_users_table\"")
		require.True(t, strings.Contains(json, "\"parents\": [1620000000000]") || 
			strings.Contains(json, "\"parents\": [\n"))
	})
	
	// Test error case - invalid format
	t.Run("Invalid format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := ExportGraph(context.Background(), buf, cfg, "invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported format: invalid")
	})
}