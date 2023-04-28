package runner

import "text/template"

// createMigrationTableTmpl is a template used to create the migrations tracking table.
// you cannot use SQL parameters to specify table names, column names, or other structural elements of a SQL query.
// We construct this statement using a template so as to prevent SQL injection
var createMigrationTableTmpl = template.Must(template.New("createMigrationsLogSQL").Parse(`CREATE TABLE IF NOT EXISTS {{ .TableName }} (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    migration_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);`))

var dropMigrationTableStmt = template.Must(template.New("dropMigrationsLogSQL").Parse(`DROP TABLE IF EXISTS {{ .TableName }}`))

var selectMigrationTmpl = template.Must(template.New("selectMigrationLogTemplate").Parse(`SELECT %s FROM {{ .TableName }} WHERE %s`))

var insertMigrationTmpl = template.Must(
	template.
		New("insertMigrationLogTemplate").
		Parse(`INSERT INTO {{ .TableName }} (%s)
VALUES (%s)`),
)
