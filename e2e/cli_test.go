//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BolajiOlajide/kat"
	"github.com/stretchr/testify/require"
)

func TestCLI_Version(t *testing.T) {
	stdout, _, exitCode := runKat(t, t.TempDir(), []string{"version"}, nil)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout, "Version:")
}

func TestCLI_Init(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, exitCode := runKat(t, tmpDir, []string{"init"}, nil)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout, "Kat initialized successfully")

	// Config file should exist
	_, err := os.Stat(filepath.Join(tmpDir, "kat.conf.yaml"))
	require.NoError(t, err, "kat.conf.yaml should exist")

	// Re-running should fail
	_, stderr, exitCode := runKat(t, tmpDir, []string{"init"}, nil)
	require.NotEqual(t, 0, exitCode)
	require.Contains(t, stderr, "already initialized")
}

func TestCLI_InitCustomFlags(t *testing.T) {
	tmpDir := t.TempDir()

	args := []string{
		"init",
		"--tableName", "custom_migrations",
		"--directory", "db/migrations",
		"--dbUser", "testuser",
		"--dbPassword", "testpass",
		"--dbName", "testdb",
		"--dbPort", "5433",
		"--dbHost", "customhost",
	}

	stdout, _, exitCode := runKat(t, tmpDir, args, nil)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout, "Kat initialized successfully")

	configData, err := os.ReadFile(filepath.Join(tmpDir, "kat.conf.yaml"))
	require.NoError(t, err)
	config := string(configData)

	require.Contains(t, config, "custom_migrations")
	require.Contains(t, config, "db/migrations")
	require.Contains(t, config, "testuser")
	require.Contains(t, config, "testdb")
	require.Contains(t, config, "5433")
	require.Contains(t, config, "customhost")
}

func TestCLI_Add(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize first
	_, _, exitCode := runKat(t, tmpDir, []string{"init"}, nil)
	require.Equal(t, 0, exitCode)

	// Create the migrations directory
	migDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(migDir, 0755))

	// Add a migration
	stdout, _, exitCode := runKat(t, tmpDir, []string{"add", "create_users"}, nil)
	require.Equal(t, 0, exitCode)
	require.Contains(t, stdout, "Migration created successfully")

	// Verify the migration directory was created with the expected files
	entries, err := os.ReadDir(migDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	migDirName := entries[0].Name()
	require.Contains(t, migDirName, "create_users")

	// Check files exist
	_, err = os.Stat(filepath.Join(migDir, migDirName, "up.sql"))
	require.NoError(t, err, "up.sql should exist")
	_, err = os.Stat(filepath.Join(migDir, migDirName, "down.sql"))
	require.NoError(t, err, "down.sql should exist")
	_, err = os.Stat(filepath.Join(migDir, migDirName, "metadata.yaml"))
	require.NoError(t, err, "metadata.yaml should exist")
}

func TestCLI_Ping(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			stdout, _, exitCode := runKat(t, projDir, []string{"ping"}, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout, "Successfully connected to database")
		})
	}
}

func TestCLI_PingRetry(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			stdout, _, exitCode := runKat(t, projDir, []string{"ping", "--retry-count", "2", "--retry-delay", "200"}, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout, "Successfully connected to database")
			require.Contains(t, stdout, "retry count: 2")
		})
	}
}

func TestCLI_Up(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
		})
	}
}

func TestCLI_UpCount(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up", "--count", "1"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
		})
	}
}

func TestCLI_UpIdempotent(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			// Run again - should be a no-op
			_, _, exitCode = runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			require.Equal(t, 2, countRows(t, db, "migration_logs"))
		})
	}
}

func TestCLI_UpDryRun(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			stdout, _, exitCode := runKat(t, projDir, []string{"up", "--dry-run"}, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, strings.ToUpper(stdout), "DRY RUN")

			db := openDB(t, p, connStr)
			assertTableNotExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
		})
	}
}

func TestCLI_UpDAG(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "dag"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
			assertTableExists(t, db, p, "comments")
			require.Equal(t, 4, countRows(t, db, "migration_logs"))
		})
	}
}

func TestCLI_Down(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			_, _, exitCode = runKat(t, projDir, []string{"down"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
		})
	}
}

func TestCLI_DownCount(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			_, _, exitCode = runKat(t, projDir, []string{"down", "--count", "2"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableNotExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
		})
	}
}

func TestCLI_DownDryRun(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			stdout, _, exitCode := runKat(t, projDir, []string{"down", "--dry-run"}, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, strings.ToUpper(stdout), "DRY RUN")

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
			require.Equal(t, 2, countRows(t, db, "migration_logs"))
		})
	}
}

func TestCLI_DownAll(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			_, _, exitCode = runKat(t, projDir, []string{"down", "--count", "2"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableNotExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
			assertTableExists(t, db, p, "migration_logs")
			require.Equal(t, 0, countRows(t, db, "migration_logs"))
		})
	}
}

func TestCLI_UpThenDown(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			// Up
			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")

			// Down all
			_, _, exitCode = runKat(t, projDir, []string{"down", "--count", "2"}, nil)
			require.Equal(t, 0, exitCode)

			assertTableNotExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
		})
	}
}

func TestCLI_Export(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "dag"))
			dotFile := filepath.Join(projDir, "graph.dot")

			_, _, exitCode := runKat(t, projDir, []string{"export", "--file", dotFile}, nil)
			require.Equal(t, 0, exitCode)

			data, err := os.ReadFile(dotFile)
			require.NoError(t, err)
			content := string(data)
			require.Contains(t, content, "digraph")
		})
	}
}

func TestCLI_ExportStdout(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "dag"))

			stdout, _, exitCode := runKat(t, projDir, []string{"export"}, nil)
			require.Equal(t, 0, exitCode)
			require.Contains(t, stdout, "digraph")
		})
	}
}

func TestCLI_InvalidSQL(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "invalid_sql"))

			_, stderr, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.NotEqual(t, 0, exitCode)
			// Should contain some error indication
			require.True(t, len(stderr) > 0, "stderr should contain error output")
		})
	}
}

func TestCLI_Verbose(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			stdout, _, exitCode := runKat(t, projDir, []string{"--verbose", "up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			// Verbose mode should produce more output
			require.True(t, len(stdout) > 0, "verbose mode should produce output")
		})
	}
}

func TestCLI_ConfigFlag(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "basic"))

			// Rename the config to a custom name
			oldPath := filepath.Join(projDir, "kat.conf.yaml")
			newPath := filepath.Join(projDir, "custom.yaml")
			require.NoError(t, os.Rename(oldPath, newPath))

			// Without -c flag it should fail
			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.NotEqual(t, 0, exitCode)

			// With -c flag it should work
			_, _, exitCode = runKat(t, projDir, []string{"up", "-c", "custom.yaml"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
		})
	}
}

func TestCLI_EnvVarSubstitution(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			tmpDir := t.TempDir()

			// Copy migrations
			err := copyDir(fixturesPath(t, "basic"), filepath.Join(tmpDir, "migrations"))
			require.NoError(t, err)

			// Write config with env var reference
			var configContent string
			if p.driver == kat.PostgresDriver {
				configContent = "migration:\n  tablename: migration_logs\n  directory: migrations\n\ndatabase:\n  driver: postgres\n  url: ${E2E_DATABASE_URL}\n"
			} else {
				configContent = "migration:\n  tablename: migration_logs\n  directory: migrations\n\ndatabase:\n  driver: sqlite\n  path: ${E2E_DATABASE_URL}\n"
			}
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "kat.conf.yaml"), []byte(configContent), 0644))

			env := []string{fmt.Sprintf("E2E_DATABASE_URL=%s", connStr)}

			_, _, exitCode := runKat(t, tmpDir, []string{"up"}, env)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
		})
	}
}

func TestCLI_NoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	_, stderr, exitCode := runKat(t, tmpDir, []string{"up"}, nil)
	require.NotEqual(t, 0, exitCode)
	require.Contains(t, stderr, "config file not provided")
}

func TestCLI_NoTransactionMigration(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "no_transaction"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "orders")
			require.Equal(t, 2, countRows(t, db, "migration_logs"))

			// Roll back
			_, _, exitCode = runKat(t, projDir, []string{"down", "--count", "2"}, nil)
			require.Equal(t, 0, exitCode)

			assertTableNotExists(t, db, p, "orders")
		})
	}
}

func TestCLI_DAG_UpThenDownPartial(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "dag"))

			// Apply all
			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			require.Equal(t, 4, countRows(t, db, "migration_logs"))

			// Roll back 2
			_, _, exitCode = runKat(t, projDir, []string{"down", "--count", "2"}, nil)
			require.Equal(t, 0, exitCode)

			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "comments")
			require.Equal(t, 2, countRows(t, db, "migration_logs"))

			// Re-apply all
			_, _, exitCode = runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			assertTableExists(t, db, p, "comments")
			require.Equal(t, 4, countRows(t, db, "migration_logs"))
		})
	}
}

func TestCLI_SingleMigration(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			projDir := createTempProject(t, p, connStr, fixturesPath(t, "single"))

			_, _, exitCode := runKat(t, projDir, []string{"up"}, nil)
			require.Equal(t, 0, exitCode)

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			require.Equal(t, 1, countRows(t, db, "migration_logs"))

			_, _, exitCode = runKat(t, projDir, []string{"down"}, nil)
			require.Equal(t, 0, exitCode)

			assertTableNotExists(t, db, p, "users")
			require.Equal(t, 0, countRows(t, db, "migration_logs"))
		})
	}
}
