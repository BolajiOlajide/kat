package database

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/loggr"
)

func TestSQLiteDriver(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	logger := loggr.NewDefault()

	// Test creating a new SQLite database connection
	db, err := New("sqlite3", dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Test ping
	err = db.Ping(context.Background())
	require.NoError(t, err)

	// Test executing a simple query
	createTableQuery := sqlf.Sprintf("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	err = db.Exec(context.Background(), createTableQuery)
	require.NoError(t, err)

	// Test insert with SQLite bind variables (the sqlf library handles the ? conversion)
	insertQuery := sqlf.Sprintf("INSERT INTO test_table (name) VALUES (%s)", "test_name")
	t.Logf("Generated SQL: %s", insertQuery.Query(sqlf.SimpleBindVar))
	t.Logf("Generated Args: %v", insertQuery.Args())
	err = db.Exec(context.Background(), insertQuery)
	require.NoError(t, err)

	// Test querying
	selectQuery := sqlf.Sprintf("SELECT COUNT(*) FROM test_table")
	row := db.QueryRow(context.Background(), selectQuery)
	
	var count int
	err = row.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSQLiteTransaction(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_tx.db")

	logger := loggr.NewDefault()
	db, err := New("sqlite3", dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Create test table
	createTableQuery := sqlf.Sprintf("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	err = db.Exec(context.Background(), createTableQuery)
	require.NoError(t, err)

	// Test transaction
	err = db.WithTransact(context.Background(), func(tx Tx) error {
		insertQuery := sqlf.Sprintf("INSERT INTO test_table (name) VALUES (%s)", "tx_test")
		return tx.Exec(context.Background(), insertQuery)
	})
	require.NoError(t, err)

	// Verify the data was inserted
	selectQuery := sqlf.Sprintf("SELECT name FROM test_table")
	row := db.QueryRow(context.Background(), selectQuery)
	
	var name string
	err = row.Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "tx_test", name)
}

func TestSQLiteErrorHandling(t *testing.T) {
	logger := loggr.NewDefault()
	
	// Test with invalid database path (should fail when pinging)
	db, err := New("sqlite3", "/dev/null/test.db", logger) // /dev/null is not a directory
	require.NoError(t, err) // Connection creation may succeed
	defer db.Close()
	
	// But pinging should fail
	err = db.Ping(context.Background())
	require.Error(t, err)
}

func TestSQLiteConcurrency(t *testing.T) {
	// Test that SQLite connections are properly limited to avoid "database is locked" errors
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "concurrent_test.db")

	logger := loggr.NewDefault()
	db, err := New("sqlite3", dbPath, logger)
	require.NoError(t, err)
	defer db.Close()

	// Create test table
	createTableQuery := sqlf.Sprintf("CREATE TABLE test_concurrent (id INTEGER PRIMARY KEY, data TEXT)")
	err = db.Exec(context.Background(), createTableQuery)
	require.NoError(t, err)

	// Test concurrent writes don't fail with "database is locked"
	done := make(chan bool, 2)
	for i := 0; i < 2; i++ {
		go func(id int) {
			defer func() { done <- true }()
			insertQuery := sqlf.Sprintf("INSERT INTO test_concurrent (data) VALUES (%s)", fmt.Sprintf("data_%d", id))
			err := db.Exec(context.Background(), insertQuery)
			assert.NoError(t, err, "Concurrent write should not fail")
		}(i)
	}

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify data was inserted
	selectQuery := sqlf.Sprintf("SELECT COUNT(*) FROM test_concurrent")
	row := db.QueryRow(context.Background(), selectQuery)
	
	var count int
	err = row.Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestBothSQLiteDriverNames(t *testing.T) {
	tempDir := t.TempDir()
	logger := loggr.NewDefault()

	// Test that both "sqlite3" and "sqlite" work
	for _, driver := range []string{"sqlite3", "sqlite"} {
		dbPath := filepath.Join(tempDir, fmt.Sprintf("test_%s.db", driver))
		db, err := New(driver, dbPath, logger)
		require.NoError(t, err, "Should accept driver: %s", driver)
		db.Close()
	}
}

func TestUnsupportedDriver(t *testing.T) {
	logger := loggr.NewDefault()
	
	// Test with unsupported driver
	_, err := New("mysql", "test.db", logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported database driver: mysql")
}
