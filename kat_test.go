package kat

import (
	"context"
	"database/sql"
	"testing"
	"testing/fstest"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var migrations = fstest.MapFS{
	"1651234567/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
	"1651234567/down.sql":      {Data: []byte("DROP TABLE users;\n")},
	"1651234567/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
	"1651234568/up.sql":        {Data: []byte("CREATE TABLE posts (id SERIAL PRIMARY KEY, user_id INT REFERENCES users(id));\n")},
	"1651234568/down.sql":      {Data: []byte("DROP TABLE posts;\n")},
	"1651234568/metadata.yaml": {Data: []byte("name: create_posts\ntimestamp: 1651234568\nparents: [1651234567]\n")},
}

func TestValidateTableName(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		wantErr   bool
	}{
		{name: "simple lowercase", tableName: "migrations", wantErr: false},
		{name: "with underscore", tableName: "migration_logs", wantErr: false},
		{name: "leading underscore", tableName: "_private", wantErr: false},
		{name: "mixed case", tableName: "MyTable", wantErr: false},
		{name: "letter and digit", tableName: "t1", wantErr: false},
		{name: "starts with number", tableName: "1starts_with_number", wantErr: true},
		{name: "has dashes", tableName: "has-dashes", wantErr: true},
		{name: "has spaces", tableName: "has spaces", wantErr: true},
		{name: "has dots", tableName: "has.dots", wantErr: true},
		{name: "quoted", tableName: `"quoted"`, wantErr: true},
		{name: "semicolon", tableName: "semi;colon", wantErr: true},
		{name: "sql injection attempt", tableName: "drop;--", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTableName(tt.tableName)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid migration table name")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScanMigrationLog(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	kmo, err := NewWithDB(PostgresDriver, db, migrations, "migration_logs")
	require.NoError(t, err)

	require.NoError(t, kmo.Up(ctx, 0))

	// Add more migrations - build a new MapFS that includes everything
	expandedMigrations := fstest.MapFS{
		// original ones
		"1651234567/up.sql":        {Data: []byte("CREATE TABLE users (id SERIAL PRIMARY KEY);\n")},
		"1651234567/down.sql":      {Data: []byte("DROP TABLE users;\n")},
		"1651234567/metadata.yaml": {Data: []byte("name: create_users\ntimestamp: 1651234567\nparents: []\n")},
		"1651234568/up.sql":        {Data: []byte("CREATE TABLE posts (id SERIAL PRIMARY KEY, user_id INT REFERENCES users(id));\n")},
		"1651234568/down.sql":      {Data: []byte("DROP TABLE posts;\n")},
		"1651234568/metadata.yaml": {Data: []byte("name: create_posts\ntimestamp: 1651234568\nparents: [1651234567]\n")},

		// new one
		"1651234569/up.sql":        {Data: []byte("CREATE TABLE comments (id SERIAL PRIMARY KEY);\n")},
		"1651234569/down.sql":      {Data: []byte("DROP TABLE comments;\n")},
		"1651234569/metadata.yaml": {Data: []byte("name: create_comments\ntimestamp: 1651234569\nparents: [1651234568]\n")},
	}

	kmo, err = NewWithDB(PostgresDriver, db, expandedMigrations, "migration_logs")
	require.NoError(t, err)
	require.NoError(t, kmo.Up(ctx, 0)) // only applies the new pending one
}
