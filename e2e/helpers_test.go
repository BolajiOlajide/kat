//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "modernc.org/sqlite"
	"gopkg.in/yaml.v3"

	"github.com/BolajiOlajide/kat"
)

var (
	katBinary      string
	sharedPGConnStr string
)

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "kat-e2e-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	katBinary = filepath.Join(tmp, "kat")
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get project root: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", katBinary, "./cmd/kat")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build kat binary: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("kat_e2e"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	sharedPGConnStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get connection string: %v\n", err)
		pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	code := m.Run()
	pgContainer.Terminate(ctx)
	os.RemoveAll(tmp)
	os.Exit(code)
}

// dbProvider abstracts database setup for dual-DB testing.
type dbProvider struct {
	name   string
	driver kat.Driver
	setup  func(t *testing.T) (connStr string, cleanup func())
	// driverName is the database/sql driver name for opening connections.
	driverName string
	// configSnippet returns the YAML database config section for CLI tests.
	configSnippet func(connStr string) string
}

var postgresProvider = dbProvider{
	name:       "postgres",
	driver:     kat.PostgresDriver,
	driverName: "pgx",
	setup: func(t *testing.T) (string, func()) {
		t.Helper()

		db, err := sql.Open("pgx", sharedPGConnStr)
		require.NoError(t, err, "opening shared postgres connection")
		resetPostgresDB(t, db)
		db.Close()

		return sharedPGConnStr, func() {}
	},
	configSnippet: func(connStr string) string {
		u, err := url.Parse(connStr)
		if err != nil {
			panic(fmt.Sprintf("invalid connStr: %v", err))
		}
		password, _ := u.User.Password()
		cfg := map[string]interface{}{
			"migration": map[string]interface{}{
				"tablename": "migration_logs",
				"directory": "migrations",
			},
			"database": map[string]interface{}{
				"driver":   "postgres",
				"host":     u.Hostname(),
				"port":     u.Port(),
				"user":     u.User.Username(),
				"password": password,
				"name":     strings.TrimPrefix(u.Path, "/"),
				"sslmode":  "disable",
			},
		}
		out, err := yaml.Marshal(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal config: %v", err))
		}
		return string(out)
	},
}

var sqliteProvider = dbProvider{
	name:       "sqlite",
	driver:     kat.SQLiteDriver,
	driverName: "sqlite",
	setup: func(t *testing.T) (string, func()) {
		t.Helper()
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "kat_e2e.db")
		return dbPath, func() {}
	},
	configSnippet: func(connStr string) string {
		cfg := map[string]interface{}{
			"migration": map[string]interface{}{
				"tablename": "migration_logs",
				"directory": "migrations",
			},
			"database": map[string]interface{}{
				"driver": "sqlite",
				"path":   connStr,
			},
		}
		out, err := yaml.Marshal(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal config: %v", err))
		}
		return string(out)
	},
}

var allProviders = []dbProvider{postgresProvider, sqliteProvider}

// runKat executes the kat binary and returns stdout, stderr, and exit code.
func runKat(t *testing.T, dir string, args []string, env []string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(katBinary, args...)
	cmd.Dir = dir

	cmd.Env = append(os.Environ(), env...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run kat: %v", err)
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// createTempProject sets up a temp directory with kat.conf.yaml and a copy of fixture migrations.
func createTempProject(t *testing.T, p dbProvider, connStr string, fixtureDir string) string {
	t.Helper()

	tmpDir := t.TempDir()

	migDir := filepath.Join(tmpDir, "migrations")
	err := copyDir(fixtureDir, migDir)
	require.NoError(t, err, "copying fixture migrations")

	configContent := p.configSnippet(connStr)
	err = os.WriteFile(filepath.Join(tmpDir, "kat.conf.yaml"), []byte(configContent), 0644)
	require.NoError(t, err, "writing kat.conf.yaml")

	return tmpDir
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

// openDB opens a *sql.DB for assertion queries.
func openDB(t *testing.T, p dbProvider, connStr string) *sql.DB {
	t.Helper()
	db, err := sql.Open(p.driverName, connStr)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// assertTableExists checks that a table exists in the database.
func assertTableExists(t *testing.T, db *sql.DB, p dbProvider, tableName string) {
	t.Helper()
	var exists bool
	var query string
	if p.driver == kat.PostgresDriver {
		query = "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)"
	} else {
		query = "SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)"
	}
	err := db.QueryRow(query, tableName).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "expected table %q to exist", tableName)
}

// assertTableNotExists checks that a table does NOT exist in the database.
func assertTableNotExists(t *testing.T, db *sql.DB, p dbProvider, tableName string) {
	t.Helper()
	var exists bool
	var query string
	if p.driver == kat.PostgresDriver {
		query = "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)"
	} else {
		query = "SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)"
	}
	err := db.QueryRow(query, tableName).Scan(&exists)
	require.NoError(t, err)
	require.False(t, exists, "expected table %q to NOT exist", tableName)
}

// countRows returns the number of rows in a table.
func countRows(t *testing.T, db *sql.DB, tableName string) int {
	t.Helper()
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	require.NoError(t, err)
	return count
}

// resetPostgresDB drops and recreates the public schema for test isolation.
func resetPostgresDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		DROP SCHEMA public CASCADE;
		CREATE SCHEMA public;
		GRANT ALL ON SCHEMA public TO postgres;
		GRANT ALL ON SCHEMA public TO public;
	`)
	require.NoError(t, err, "resetting postgres database")
}

// fixturesPath returns the absolute path to a fixture migrations directory.
func fixturesPath(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("migrations", name))
	require.NoError(t, err)
	return p
}

// In-memory migration fixtures for library tests (fstest.MapFS).

var basicMigrations = fstest.MapFS{
	"1000000001_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n")},
	"1000000001_create_users/down.sql":      {Data: []byte("DROP TABLE IF EXISTS users;\n")},
	"1000000001_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1000000001\nparents: []\n")},

	"1000000002_create_posts/up.sql":        {Data: []byte("CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, title TEXT NOT NULL);\n")},
	"1000000002_create_posts/down.sql":      {Data: []byte("DROP TABLE IF EXISTS posts;\n")},
	"1000000002_create_posts/metadata.yaml": {Data: []byte("name: create_posts\ntimestamp: 1000000002\nparents:\n  - 1000000001\n")},
}

var dagMigrations = fstest.MapFS{
	"1000000001_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n")},
	"1000000001_create_users/down.sql":      {Data: []byte("DROP TABLE IF EXISTS users;\n")},
	"1000000001_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1000000001\nparents: []\n")},

	"1000000002_add_email/up.sql":        {Data: []byte("ALTER TABLE users ADD COLUMN email TEXT;\n")},
	"1000000002_add_email/down.sql":      {Data: []byte("ALTER TABLE users DROP COLUMN email;\n")},
	"1000000002_add_email/metadata.yaml": {Data: []byte("name: add_email\ntimestamp: 1000000002\nparents:\n  - 1000000001\n")},

	"1000000003_create_posts/up.sql":        {Data: []byte("CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL, title TEXT NOT NULL);\n")},
	"1000000003_create_posts/down.sql":      {Data: []byte("DROP TABLE IF EXISTS posts;\n")},
	"1000000003_create_posts/metadata.yaml": {Data: []byte("name: create_posts\ntimestamp: 1000000003\nparents:\n  - 1000000001\n")},

	"1000000004_create_comments/up.sql":        {Data: []byte("CREATE TABLE comments (id INTEGER PRIMARY KEY, post_id INTEGER NOT NULL, body TEXT NOT NULL);\n")},
	"1000000004_create_comments/down.sql":      {Data: []byte("DROP TABLE IF EXISTS comments;\n")},
	"1000000004_create_comments/metadata.yaml": {Data: []byte("name: create_comments\ntimestamp: 1000000004\nparents:\n  - 1000000002\n  - 1000000003\n")},
}

var singleMigration = fstest.MapFS{
	"1000000001_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n")},
	"1000000001_create_users/down.sql":      {Data: []byte("DROP TABLE IF EXISTS users;\n")},
	"1000000001_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1000000001\nparents: []\n")},
}

var invalidSQLMigrations = fstest.MapFS{
	"1000000001_bad_migration/up.sql":        {Data: []byte("CREAT TABL broken_syntax (id INTEGER);\n")},
	"1000000001_bad_migration/down.sql":      {Data: []byte("DROP TABLE IF EXISTS broken_syntax;\n")},
	"1000000001_bad_migration/metadata.yaml": {Data: []byte("name: bad_migration\ntimestamp: 1000000001\nparents: []\n")},
}

var noTransactionMigrations = fstest.MapFS{
	"1000000001_create_orders/up.sql":        {Data: []byte("CREATE TABLE orders (id INTEGER PRIMARY KEY, status TEXT NOT NULL DEFAULT 'pending');\n")},
	"1000000001_create_orders/down.sql":      {Data: []byte("DROP TABLE IF EXISTS orders;\n")},
	"1000000001_create_orders/metadata.yaml": {Data: []byte("name: create_orders\ntimestamp: 1000000001\nparents: []\n")},

	"1000000002_add_status_index/up.sql":        {Data: []byte("CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);\n")},
	"1000000002_add_status_index/down.sql":      {Data: []byte("DROP INDEX IF EXISTS idx_orders_status;\n")},
	"1000000002_add_status_index/metadata.yaml": {Data: []byte("name: add_status_index\ntimestamp: 1000000002\nno_transaction: true\nparents:\n  - 1000000001\n")},
}
