package runner

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/graph"
	"github.com/BolajiOlajide/kat/internal/loggr"
	"github.com/BolajiOlajide/kat/internal/types"
)

const dropRecreateQuery = `-- as a superuser or the database owner:
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;

-- (re-grant privileges if needed)
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;`

const migrationTableName = "migration_logs"

func createMigrationDef(t *testing.T, defs ...types.Definition) *graph.Graph {
	t.Helper()
	g := graph.New()
	require.NoError(t, g.AddDefinitions(defs...))
	return g
}

var noTransactionDefinitions = []types.Definition{
	{
		MigrationMetadata: types.MigrationMetadata{
			Name:      "create_orders",
			Timestamp: 1747525100,
		},
		UpQuery:   sqlf.Sprintf("CREATE TABLE orders (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, status TEXT NOT NULL DEFAULT 'pending');"),
		DownQuery: sqlf.Sprintf("DROP TABLE orders;"),
	},
	{
		MigrationMetadata: types.MigrationMetadata{
			Name:          "add_orders_status_index",
			Timestamp:     1747525200,
			Parents:       []int64{1747525100},
			NoTransaction: true,
		},
		UpQuery:   sqlf.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status ON orders (status);"),
		DownQuery: sqlf.Sprintf("DROP INDEX CONCURRENTLY IF EXISTS idx_orders_status;"),
	},
}

var allDefinitions = []types.Definition{
	{
		MigrationMetadata: types.MigrationMetadata{
			Name:      "create_users",
			Timestamp: 1747525262,
		},
		UpQuery:   sqlf.Sprintf("CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
		DownQuery: sqlf.Sprintf("DROP TABLE users;"),
	},
	{
		MigrationMetadata: types.MigrationMetadata{
			Name:      "create_transactions",
			Timestamp: 1747525318,
			Parents:   []int64{},
		},
		UpQuery:   sqlf.Sprintf("CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
		DownQuery: sqlf.Sprintf("DROP TABLE transactions;"),
	},
	{
		MigrationMetadata: types.MigrationMetadata{
			Name:      "create_products",
			Timestamp: 1747527900,
			Parents:   []int64{1747525262, 1747525318},
		},
		UpQuery:   sqlf.Sprintf("CREATE TABLE products (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
		DownQuery: sqlf.Sprintf("DROP TABLE products;"),
	},
	{
		MigrationMetadata: types.MigrationMetadata{
			Name:      "create_roles",
			Timestamp: 1749554911,
			Parents:   []int64{1747527900},
		},
		UpQuery:   sqlf.Sprintf("CREATE TABLE roles (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
		DownQuery: sqlf.Sprintf("DROP TABLE roles;"),
	},
}

func TestRun(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15.3-alpine",
		// postgres.WithInitScripts(filepath.Join("..", "testdata", "init-db.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err, "starting up postgres container")

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "fetching connection string")

	logger := loggr.NewDefault()

	db, err := database.New(connStr, logger)
	require.NoError(t, err, "creating database service")

	r, err := NewRunner(ctx, db, logger)
	require.NoError(t, err, "initializing runner")

	createMigrationLogQuery, err := computeCreateMigrationLogQuery(migrationTableName)
	require.NoError(t, err, "create migration log query")

	tests := []struct {
		name            string
		options         Options
		expectedSchema  []dbSchema
		expectedIndexes []string // index names that should exist after migration
		unexpectedIndexes []string // index names that should NOT exist after migration
		pre             string
	}{
		{
			name: "up migration",
			options: Options{
				Operation:     types.UpMigrationOperation,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema, transactionsSchema, productsSchema, rolesSchema),
		},
		{
			name: "up migration (with count=1)",
			options: Options{
				Operation:     types.UpMigrationOperation,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
				Count:         1,
			},
			expectedSchema: append(migrationLogsSchema, usersSchema),
		},
		{
			name: "up migration (with count=2)",
			options: Options{
				Operation:     types.UpMigrationOperation,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
				Count:         2,
			},
			expectedSchema: append(migrationLogsSchema, usersSchema, transactionsSchema),
		},
		{
			name: "DRYRUN: up migration",
			options: Options{
				Operation:   types.UpMigrationOperation,
				Definitions: createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
				DryRun: true,
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "up migration (all migrations applied including existing ones)",
			pre: createMigrationLogQuery + `
		INSERT INTO "migration_logs"("name","migration_time","duration")
		 VALUES
		('1747525262_create_users','2025-04-14 19:41:23.39-04','00:00:13.147291');

		CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation:   types.UpMigrationOperation,
				Definitions: createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema, transactionsSchema, productsSchema, rolesSchema),
		},
		{
			name: "down migration (all migrations applied including existing ones)",
			pre: createMigrationLogQuery + `
		INSERT INTO "migration_logs"("name","migration_time","duration")
		VALUES
		('1747525262_create_users','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747525318_create_transactions','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747527900_create_products','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1749554911_create_roles','2025-04-14 19:41:23.39-04','00:00:13.147291');

		CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE products (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE roles (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation:     types.DownMigrationOperation,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "down migration (count=2)",
			pre: createMigrationLogQuery + `
		INSERT INTO "migration_logs"("name","migration_time","duration")
		VALUES
		('1747525262_create_users','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747525318_create_transactions','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747527900_create_products','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1749554911_create_roles','2025-04-14 19:41:23.39-04','00:00:13.147291');

		CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE products (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE roles (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation:     types.DownMigrationOperation,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
				Count:         2,
			},
			expectedSchema: append(migrationLogsSchema, usersSchema, transactionsSchema),
		},
		{
			name: "DRYRUN: down migration",
			pre: createMigrationLogQuery + `
		INSERT INTO "migration_logs"("name","migration_time","duration")
		VALUES
		('1747525262_create_users','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747525318_create_transactions','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747527900_create_products','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1749554911_create_roles','2025-04-14 19:41:23.39-04','00:00:13.147291');

		CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE products (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE roles (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation:     types.DownMigrationOperation,
				DryRun:        true,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema, transactionsSchema, productsSchema, rolesSchema),
		},
		{
			name: "down migration (no existing migrations applied)",
			pre:  createMigrationLogQuery,
			options: Options{
				Operation:     types.DownMigrationOperation,
				Definitions:   createMigrationDef(t, allDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "no migrations to apply (up)",
			options: Options{
				Operation:   types.UpMigrationOperation,
				Definitions: graph.New(),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "no migrations to apply (down)",
			options: Options{
				Operation:   types.DownMigrationOperation,
				Definitions: graph.New(),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "DRYRUN: no migrations to apply (up)",
			options: Options{
				Operation:   types.UpMigrationOperation,
				Definitions: graph.New(),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
				DryRun: true,
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "DRYRUN: no migrations to apply (down)",
			options: Options{
				Operation:   types.DownMigrationOperation,
				Definitions: graph.New(),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
				DryRun: true,
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "up migration with no_transaction (CREATE INDEX CONCURRENTLY)",
			options: Options{
				Operation:     types.UpMigrationOperation,
				Definitions:   createMigrationDef(t, noTransactionDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
			},
			expectedSchema:  append(migrationLogsSchema, ordersSchema...),
			expectedIndexes: []string{"idx_orders_status"},
		},
		{
			name: "down migration with no_transaction (DROP INDEX CONCURRENTLY)",
			pre: createMigrationLogQuery + `
		INSERT INTO "migration_logs"("name","migration_time","duration")
		VALUES
		('1747525100_create_orders','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('1747525200_add_orders_status_index','2025-04-14 19:41:23.39-04','00:00:13.147291');

		CREATE TABLE orders (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, status TEXT NOT NULL DEFAULT 'pending');
		CREATE INDEX idx_orders_status ON orders (status);`,
			options: Options{
				Operation:     types.DownMigrationOperation,
				Definitions:   createMigrationDef(t, noTransactionDefinitions...),
				MigrationInfo: types.MigrationInfo{TableName: migrationTableName},
			},
			expectedSchema:    migrationLogsSchema,
			unexpectedIndexes: []string{"idx_orders_status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Exec(ctx, sqlf.Sprintf(dropRecreateQuery))
			require.NoError(t, err, "dropping and recreating public schema")

			if tt.pre != "" {
				require.NoError(
					t, db.Exec(ctx, sqlf.Sprintf(tt.pre)),
					"executing pre query",
				)
			}

			require.NoError(t, r.Run(ctx, tt.options), "expected error to be nil from Run() method")

			rows, err := db.Query(ctx, sqlf.Sprintf(dumpSchemaQuery))
			require.NoError(t, err, "fetching schema from database")

			var schemas []dbSchema
			for rows.Next() {
				schema, err := scanDBSchema(rows)
				require.NoError(t, err)
				schemas = append(schemas, *schema)
			}

			require.ElementsMatch(t, schemas, tt.expectedSchema)

			// Assert expected indexes exist
			for _, idxName := range tt.expectedIndexes {
				var exists bool
				row := db.QueryRow(ctx, sqlf.Sprintf("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = %s)", idxName))
				require.NoError(t, row.Scan(&exists), "checking index %s existence", idxName)
				require.True(t, exists, "expected index %s to exist", idxName)
			}

			// Assert unexpected indexes do not exist
			for _, idxName := range tt.unexpectedIndexes {
				var exists bool
				row := db.QueryRow(ctx, sqlf.Sprintf("SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = %s)", idxName))
				require.NoError(t, row.Scan(&exists), "checking index %s absence", idxName)
				require.False(t, exists, "expected index %s to not exist", idxName)
			}
		})
	}
}

var dumpSchemaQuery = `SELECT
  table_schema AS "tableSchema",
  table_name AS "tableName",
  column_name AS "columnName",
  data_type AS "dataType",
  is_nullable AS "isNullable",
  column_default AS "columnDefault"
FROM information_schema.columns
WHERE table_schema NOT IN ('information_schema','pg_catalog')
ORDER BY table_schema, table_name, ordinal_position;`

type dbSchema struct {
	TableSchema   string
	TableName     string
	ColumnName    string
	DataType      string
	IsNullable    string
	ColumnDefault *string
}

func scanDBSchema(sc database.Scanner) (*dbSchema, error) {
	var schema dbSchema
	return &schema, sc.Scan(
		&schema.TableSchema,
		&schema.TableName,
		&schema.ColumnName,
		&schema.DataType,
		&schema.IsNullable,
		&schema.ColumnDefault,
	)
}
