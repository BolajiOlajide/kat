//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat"
)

func TestLib_New(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, singleMigration, "migration_logs")
			require.NoError(t, err)
			require.NoError(t, m.Close())
		})
	}
}

func TestLib_NewWithDB(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			db := openDB(t, p, connStr)
			if p.driver == kat.SQLiteDriver {
				db.SetMaxOpenConns(1)
			}

			m, err := kat.NewWithDB(p.driver, db, singleMigration, "migration_logs")
			require.NoError(t, err)
			require.NoError(t, m.Close())

			// Verify the caller's db is still usable after Close
			err = db.Ping()
			require.NoError(t, err, "caller's db should still be usable after Migration.Close()")
		})
	}
}

func TestLib_InvalidDriver(t *testing.T) {
	_, err := kat.New("mysql", "fake://conn", singleMigration, "migration_logs")
	require.Error(t, err)
	require.Contains(t, err.Error(), "driver must be one of")
}

func TestLib_NilDB(t *testing.T) {
	_, err := kat.NewWithDB(kat.PostgresDriver, nil, singleMigration, "migration_logs")
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil database connection")
}

func TestLib_EmptyTableName(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			_, err := kat.New(p.driver, connStr, singleMigration, "")
			require.Error(t, err)
			require.Contains(t, err.Error(), "migrationTableName cannot be empty")
		})
	}
}

func TestLib_UpAll(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
			assertTableExists(t, db, p, "migration_logs")
		})
	}
}

func TestLib_UpCount(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 1))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")

			// Apply the remaining one
			require.NoError(t, m.Up(ctx, 1))
			assertTableExists(t, db, p, "posts")
		})
	}
}

func TestLib_UpNegativeCount(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			err = m.Up(context.Background(), -1)
			require.Error(t, err)
			require.Contains(t, err.Error(), "count cannot be a negative number")
		})
	}
}

func TestLib_UpIdempotent(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))
			require.NoError(t, m.Up(ctx, 0))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
			// Migration log should have exactly 2 entries
			require.Equal(t, 2, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_DownCount(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))
			require.NoError(t, m.Down(ctx, 1))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
			require.Equal(t, 1, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_DownZero(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			err = m.Down(context.Background(), 0)
			require.Error(t, err)
			require.Contains(t, err.Error(), "count must be a non-zero positive number")
		})
	}
}

func TestLib_DownAll(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))
			require.NoError(t, m.Down(ctx, 2))

			db := openDB(t, p, connStr)
			assertTableNotExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
			require.Equal(t, 0, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_FullRoundTrip(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()

			// Up all
			require.NoError(t, m.Up(ctx, 0))
			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
			require.Equal(t, 2, countRows(t, db, "migration_logs"))

			// Down all
			require.NoError(t, m.Down(ctx, 2))
			assertTableNotExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")
			require.Equal(t, 0, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_DAGOrdering(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, dagMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")
			assertTableExists(t, db, p, "comments")
			require.Equal(t, 4, countRows(t, db, "migration_logs"))

			// Verify the email column was added
			if p.driver == kat.PostgresDriver {
				var exists bool
				err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='email')").Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists, "email column should exist on users table")
			} else {
				rows, err := db.Query("PRAGMA table_info(users)")
				require.NoError(t, err)
				defer rows.Close()
				foundEmail := false
				for rows.Next() {
					var cid int
					var name, ctype string
					var notnull int
					var dflt *string
					var pk int
					require.NoError(t, rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk))
					if name == "email" {
						foundEmail = true
					}
				}
				require.True(t, foundEmail, "email column should exist on users table")
			}
		})
	}
}

func TestLib_IncrementalMigrations(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			// Apply initial set
			m, err := kat.New(p.driver, connStr, singleMigration, "migration_logs")
			require.NoError(t, err)

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))
			require.NoError(t, m.Close())

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			require.Equal(t, 1, countRows(t, db, "migration_logs"))

			// Now create a new Migration with expanded migrations
			m2, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m2.Close()

			require.NoError(t, m2.Up(ctx, 0))
			assertTableExists(t, db, p, "posts")
			require.Equal(t, 2, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_ContextCancellation(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err = m.Up(ctx, 0)
			require.Error(t, err)
		})
	}
}

func TestLib_WithLogger(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			logger := &testLogger{}
			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs",
				kat.WithLogger(logger),
			)
			require.NoError(t, err)
			defer m.Close()

			require.NoError(t, m.Up(context.Background(), 0))
			require.True(t, len(logger.messages) > 0, "logger should have received messages")
		})
	}
}

type testLogger struct {
	messages []string
}

func (l *testLogger) Debug(msg string) { l.messages = append(l.messages, "DEBUG: "+msg) }
func (l *testLogger) Info(msg string)  { l.messages = append(l.messages, "INFO: "+msg) }
func (l *testLogger) Warn(msg string)  { l.messages = append(l.messages, "WARN: "+msg) }
func (l *testLogger) Error(msg string) { l.messages = append(l.messages, "ERROR: "+msg) }

func TestLib_WithDBConfig(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			config := kat.DefaultDBConfig(p.driver)
			config.ConnectTimeout = 5 * time.Second
			config.MaxOpenConns = 20

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs",
				kat.WithDBConfig(config),
			)
			require.NoError(t, err)
			defer m.Close()

			require.NoError(t, m.Up(context.Background(), 0))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
		})
	}
}

func TestLib_WithConnectTimeout(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs",
				kat.WithConnectTimeout(10*time.Second),
			)
			require.NoError(t, err)
			defer m.Close()

			require.NoError(t, m.Up(context.Background(), 0))
		})
	}
}

func TestLib_WithPoolLimits(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs",
				kat.WithPoolLimits(15, 5, 30*time.Minute),
			)
			require.NoError(t, err)
			defer m.Close()

			require.NoError(t, m.Up(context.Background(), 0))
		})
	}
}

func TestLib_DBConfigWithNewWithDB(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			db := openDB(t, p, connStr)
			if p.driver == kat.SQLiteDriver {
				db.SetMaxOpenConns(1)
			}

			_, err := kat.NewWithDB(p.driver, db, basicMigrations, "migration_logs",
				kat.WithDBConfig(kat.DefaultDBConfig(p.driver)),
			)
			require.Error(t, err)
			require.Contains(t, err.Error(), "not supported with NewWithDB")
		})
	}
}

func TestLib_Close(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, singleMigration, "migration_logs")
			require.NoError(t, err)

			require.NoError(t, m.Close())
			// Double close should be safe
			require.NoError(t, m.Close())
		})
	}
}

func TestLib_CloseWithNewWithDB(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			db := openDB(t, p, connStr)
			if p.driver == kat.SQLiteDriver {
				db.SetMaxOpenConns(1)
			}

			m, err := kat.NewWithDB(p.driver, db, singleMigration, "migration_logs")
			require.NoError(t, err)
			require.NoError(t, m.Close())

			// The caller's DB should still be usable
			require.NoError(t, db.Ping())
		})
	}
}

func TestLib_TransactionRollback(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			// First apply the valid migration
			m1, err := kat.New(p.driver, connStr, singleMigration, "migration_logs")
			require.NoError(t, err)
			require.NoError(t, m1.Up(context.Background(), 0))
			require.NoError(t, m1.Close())

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")

			// Now try to apply a broken migration set that includes the existing + a bad one
			brokenMigrations := fstest.MapFS{
				"1000000001_create_users/up.sql":        {Data: []byte("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n")},
				"1000000001_create_users/down.sql":      {Data: []byte("DROP TABLE IF EXISTS users;\n")},
				"1000000001_create_users/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1000000001\nparents: []\n")},

				"1000000002_broken/up.sql":        {Data: []byte("CREAT TABL broken (id INTEGER);\n")},
				"1000000002_broken/down.sql":      {Data: []byte("DROP TABLE IF EXISTS broken;\n")},
				"1000000002_broken/metadata.yaml": {Data: []byte("name: broken\ntimestamp: 1000000002\nparents:\n  - 1000000001\n")},
			}

			m2, err := kat.New(p.driver, connStr, brokenMigrations, "migration_logs")
			require.NoError(t, err)
			defer m2.Close()

			err = m2.Up(context.Background(), 0)
			require.Error(t, err)

			// The users table from the first migration should still exist
			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "broken")
			require.Equal(t, 1, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_NoTransactionMigration(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, noTransactionMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))

			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "orders")
			require.Equal(t, 2, countRows(t, db, "migration_logs"))

			// Verify index was created
			if p.driver == kat.PostgresDriver {
				var exists bool
				err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_orders_status')").Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists, "index idx_orders_status should exist")
			} else {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_orders_status'").Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 1, count, "index idx_orders_status should exist")
			}
		})
	}
}

func TestLib_SQLite_FullRoundTrip(t *testing.T) {
	p := sqliteProvider
	connStr, cleanup := p.setup(t)
	defer cleanup()

	m, err := kat.New(p.driver, connStr, basicMigrations, "migration_logs")
	require.NoError(t, err)
	defer m.Close()

	ctx := context.Background()
	require.NoError(t, m.Up(ctx, 0))

	db := openDB(t, p, connStr)
	assertTableExists(t, db, p, "users")
	assertTableExists(t, db, p, "posts")

	require.NoError(t, m.Down(ctx, 2))
	assertTableNotExists(t, db, p, "users")
	assertTableNotExists(t, db, p, "posts")
}

func TestLib_SQLite_NewWithDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	defer db.Close()

	m, err := kat.NewWithDB(kat.SQLiteDriver, db, basicMigrations, "migration_logs")
	require.NoError(t, err)
	defer m.Close()

	ctx := context.Background()
	require.NoError(t, m.Up(ctx, 0))

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='users'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Verify db is still usable after Close
	require.NoError(t, m.Close())
	require.NoError(t, db.Ping())
}

func TestLib_NilLogger(t *testing.T) {
	_, err := kat.New(kat.SQLiteDriver, "/tmp/test.db", singleMigration, "migration_logs",
		kat.WithLogger(nil),
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "logger cannot be nil")
}

func TestLib_DAG_DownReverseOrder(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, dagMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()
			require.NoError(t, m.Up(ctx, 0))

			db := openDB(t, p, connStr)
			require.Equal(t, 4, countRows(t, db, "migration_logs"))

			// Roll back 1 - should remove the leaf (comments)
			require.NoError(t, m.Down(ctx, 1))
			assertTableNotExists(t, db, p, "comments")
			assertTableExists(t, db, p, "users")
			assertTableExists(t, db, p, "posts")

			// Roll back 2 more - should remove posts and email
			require.NoError(t, m.Down(ctx, 2))
			assertTableExists(t, db, p, "users")
			assertTableNotExists(t, db, p, "posts")

			// Roll back the last one
			require.NoError(t, m.Down(ctx, 1))
			assertTableNotExists(t, db, p, "users")
			require.Equal(t, 0, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_UpCount_DAG(t *testing.T) {
	for _, p := range allProviders {
		t.Run(p.name, func(t *testing.T) {
			connStr, cleanup := p.setup(t)
			defer cleanup()

			m, err := kat.New(p.driver, connStr, dagMigrations, "migration_logs")
			require.NoError(t, err)
			defer m.Close()

			ctx := context.Background()

			// Apply 1 - should apply the root (create_users)
			require.NoError(t, m.Up(ctx, 1))
			db := openDB(t, p, connStr)
			assertTableExists(t, db, p, "users")
			require.Equal(t, 1, countRows(t, db, "migration_logs"))

			// Apply 2 more - should apply add_email and create_posts
			require.NoError(t, m.Up(ctx, 2))
			assertTableExists(t, db, p, "posts")
			require.Equal(t, 3, countRows(t, db, "migration_logs"))

			// Apply remaining - should apply create_comments
			require.NoError(t, m.Up(ctx, 0))
			assertTableExists(t, db, p, "comments")
			require.Equal(t, 4, countRows(t, db, "migration_logs"))
		})
	}
}

func TestLib_ParseDriver(t *testing.T) {
	tests := []struct {
		input       string
		expected    kat.Driver
		expectError bool
	}{
		{"postgres", kat.PostgresDriver, false},
		{"postgresql", kat.PostgresDriver, false},
		{"", kat.PostgresDriver, false},
		{"sqlite", kat.SQLiteDriver, false},
		{"sqlite3", kat.SQLiteDriver, false},
		{"mysql", "", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			drv, err := kat.ParseDriver(tt.input)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, drv)
			}
		})
	}
}
