package runner

import "text/template"

// createMigrationTableTmpl is a template used to create the migrations tracking table for PostgreSQL.
// you cannot use SQL parameters to specify table names, column names, or other structural elements of a SQL query.
// We construct this statement using a template so as to prevent SQL injection
var createMigrationTableTmpl = template.Must(template.New("createMigrationsLogSQL").Parse(`CREATE TABLE IF NOT EXISTS "{{ .TableName }}" (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    migration_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    duration INTERVAL NOT NULL
);`))

// createMigrationTableSQLiteTmpl is the SQLite-compatible variant of the migration log table.
var createMigrationTableSQLiteTmpl = template.Must(template.New("createMigrationsLogSQLite").Parse(`CREATE TABLE IF NOT EXISTS "{{ .TableName }}" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    migration_time TEXT NOT NULL DEFAULT (datetime('now')),
    duration TEXT NOT NULL
);`))

var selectMigrationsTmpl = template.Must(template.New("selectMigrationLogTemplate").Parse(`SELECT %s FROM "{{ .TableName }}"`))

var insertMigrationTmpl = template.Must(
	template.
		New("insertMigrationLogTemplate").
		Parse(`INSERT INTO "{{ .TableName }}" (%s)
VALUES (%s)`),
)

var deleteMigrationTmpl = template.Must(
	template.
		New("deleteMigrationLogTemplate").
		Parse(`DELETE FROM "{{ .TableName }}" WHERE name = %s`),
)
