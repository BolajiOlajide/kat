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

	db, err := database.New(connStr)
	require.NoError(t, err, "creating database service")

	r, err := NewRunner(ctx, db)
	require.NoError(t, err, "initializing runner")

	createMigrationLogQuery, err := computeCreateMigrationLogQuery(migrationTableName)
	require.NoError(t, err, "create migration log query")

	tests := []struct {
		name           string
		options        Options
		expectedSchema []dbSchema
		pre            string
	}{
		{
			name: "up migration",
			options: Options{
				Operation: types.UpMigrationOperation,
				Definitions: createMigrationDef(t, types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name: "create_users",
					},
					UpQuery:   sqlf.Sprintf("CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
					DownQuery: sqlf.Sprintf("DROP TABLE users;"),
				}),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema),
		},
		{
			name: "DRYRUN: up migration",
			options: Options{
				Operation: types.UpMigrationOperation,
				Definitions: createMigrationDef(t, types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name: "create_users",
					},
					UpQuery:   sqlf.Sprintf("CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
					DownQuery: sqlf.Sprintf("DROP TABLE users;"),
				}),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
				DryRun: true,
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "up migration (all migrations applied)",
			pre: createMigrationLogQuery + `
        INSERT INTO "migration_logs"("name","migration_time","duration")
          VALUES
        ('create_users','2025-04-14 19:41:23.39-04','00:00:13.147291');

        CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation: types.UpMigrationOperation,
				Definitions: createMigrationDef(t, types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name: "create_users",
					},
					UpQuery:   sqlf.Sprintf("CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
					DownQuery: sqlf.Sprintf("DROP TABLE users;"),
				}),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema),
		},
		{
			name: "down migration",
			pre: createMigrationLogQuery + `
		INSERT INTO "migration_logs"("name","migration_time","duration")
		 VALUES
		('create_users','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		('create_transactions','2025-04-14 19:41:23.39-04','00:00:13.147291');

		CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation: types.DownMigrationOperation,
				Definitions: createMigrationDef(t, types.Definition{
					MigrationMetadata: types.MigrationMetadata{
						Name:      "create_users",
						Timestamp: 1747525262,
					},
					UpQuery:   sqlf.Sprintf("CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
					DownQuery: sqlf.Sprintf("DROP TABLE users;"),
				},
					types.Definition{
						MigrationMetadata: types.MigrationMetadata{
							Name:      "create_transactions",
							Timestamp: 1747525318,
						},
						UpQuery:   sqlf.Sprintf("CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
						DownQuery: sqlf.Sprintf("DROP TABLE transactions;"),
					}),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: migrationLogsSchema,
		},
		{
			name: "DRYRUN: down migration",
			pre: createMigrationLogQuery + `
		    INSERT INTO "migration_logs"("name","migration_time","duration")
		      VALUES
		    ('create_users','2025-04-14 19:41:23.39-04','00:00:13.147291'),
		    ('create_transactions','2025-04-14 19:41:23.39-04','00:00:13.147291');

		    CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
		    CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation: types.DownMigrationOperation,
				DryRun:    true,
				Definitions: createMigrationDef(t, []types.Definition{
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
						},
						UpQuery:   sqlf.Sprintf("CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
						DownQuery: sqlf.Sprintf("DROP TABLE transactions;"),
					},
				}...),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema, transactionsSchema),
		},
		{
			name: "down migration (partial)",
			pre: createMigrationLogQuery + `
        INSERT INTO "migration_logs"("name","migration_time","duration")
          VALUES
        ('create_users','2025-04-14 19:41:23.39-04','00:00:13.147291'),
        ('create_transactions','2025-04-14 19:41:23.39-04','00:00:13.147291');

        CREATE TABLE users (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);
        CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);`,
			options: Options{
				Operation: types.DownMigrationOperation,
				Definitions: createMigrationDef(t, []types.Definition{
					{
						MigrationMetadata: types.MigrationMetadata{
							Name:      "create_transactions",
							Timestamp: 1747525318,
						},
						UpQuery:   sqlf.Sprintf("CREATE TABLE transactions (id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY);"),
						DownQuery: sqlf.Sprintf("DROP TABLE transactions;"),
					},
				}...),
				MigrationInfo: types.MigrationInfo{
					TableName: migrationTableName,
				},
			},
			expectedSchema: append(migrationLogsSchema, usersSchema),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Exec(ctx, sqlf.Sprintf(dropRecreateQuery))
			require.NoError(t, err, "dropping and recreating public schema")

			if tt.pre != "" {
				err := db.Exec(ctx, sqlf.Sprintf(tt.pre))
				require.NoError(t, err, "executing pre query")
			}

			err = r.Run(ctx, tt.options)

			// if tt.wantErr {
			// 	require.ErrorContains(t, err, tt.errMsg, "expected error from Run() method")
			// } else {
			require.NoError(t, err, "expected error to be nil from Run() method")
			// }

			rows, err := db.Query(ctx, sqlf.Sprintf(dumpSchemaQuery))
			require.NoError(t, err, "fetching schema from database")

			var schemas []dbSchema
			for rows.Next() {
				schema, err := scanDBSchema(rows)
				require.NoError(t, err)
				schemas = append(schemas, *schema)
			}

			require.ElementsMatch(t, schemas, tt.expectedSchema)
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
	err := sc.Scan(
		&schema.TableSchema,
		&schema.TableName,
		&schema.ColumnName,
		&schema.DataType,
		&schema.IsNullable,
		&schema.ColumnDefault,
	)
	return &schema, err
}
