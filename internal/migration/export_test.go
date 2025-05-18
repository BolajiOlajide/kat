package migration

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/BolajiOlajide/kat/internal/types"
)

func TestExportGraph(t *testing.T) {
	testCases := []struct {
		name               string
		migrationsDir      string
		expectError        bool
		createMigrationDir bool
		definitions        []types.Definition
	}{
		{
			name:               "success - valid migrations directory",
			migrationsDir:      "testdata/valid",
			expectError:        false,
			createMigrationDir: true,
			definitions: []types.Definition{
				{
					UpQuery:   queryFromString("CREATE TABLE test (id INT);"),
					DownQuery: queryFromString("DROP TABLE test;"),
					MigrationMetadata: types.MigrationMetadata{
						Name:        "init",
						Timestamp:   1747578808,
						Description: "Initial migration",
					},
				},
				{
					UpQuery:   queryFromString("CREATE TABLE test_2 (id INT);"),
					DownQuery: queryFromString("DROP TABLE test_2;"),
					MigrationMetadata: types.MigrationMetadata{
						Name:        "another_migration",
						Timestamp:   1747578819,
						Description: "Another migration",
						Parents:     []int64{1747578808},
					},
				},
			},
		},
		{
			name:               "error - empty migrations directory",
			migrationsDir:      "testdata/empty",
			expectError:        false,
			createMigrationDir: true,
		},
		{
			name:          "error - non-existent directory",
			migrationsDir: "testdata/nonexistent",
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary dir for test migrations
			testDir := t.TempDir()

			// Set up test migrations dir
			migrationsPath := filepath.Join(testDir, tc.migrationsDir)
			if tc.createMigrationDir {
				require.NoError(t, os.MkdirAll(migrationsPath, 0755), "failed to create migrations dir")
			}

			for _, def := range tc.definitions {
				migrationDir := filepath.Join(migrationsPath, fmt.Sprintf("%d_%s", def.Timestamp, def.Name))
				require.NoError(t, os.MkdirAll(migrationDir, 0755), "failed to create migrations dir")

				metadataContent, err := yaml.Marshal(def.MigrationMetadata)
				require.NoError(t, err)

				require.NoError(t, os.WriteFile(filepath.Join(migrationDir, "up.sql"), []byte(def.UpQuery.Query(sqlf.PostgresBindVar)), 0644), "failed to write up migration")
				require.NoError(t, os.WriteFile(filepath.Join(migrationDir, "down.sql"), []byte(def.UpQuery.Query(sqlf.PostgresBindVar)), 0644), "failed to write down migration")
				require.NoError(t, os.WriteFile(filepath.Join(migrationDir, "metadata.yaml"), metadataContent, 0644), "failed to write migration metadata")
			}

			// Configure Kat
			cfg := types.Config{
				Migration: types.MigrationInfo{
					Directory: migrationsPath,
				},
			}

			// Call ExportGraph
			var buf bytes.Buffer
			err := ExportGraph(&buf, cfg)

			if tc.expectError {
				require.Error(t, err, "expected an error but got none")
			} else {
				require.NoError(t, err, "unexpected error")
				require.NotEmpty(t, buf.String(), "expected non-empty output")
			}
		})
	}
}
