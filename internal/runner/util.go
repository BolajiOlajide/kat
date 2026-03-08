package runner

import (
	"bytes"
	"text/template"

	"github.com/keegancsmith/sqlf"
)

var migrationLogColumns = []string{
	"id",
	"name",
	"migration_time",
	"duration",
}

var migrationLogInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("migration_time"),
	sqlf.Sprintf("duration"),
}

func computeMigrationLogColumns() []*sqlf.Query {
	var cols = make([]*sqlf.Query, len(migrationLogColumns))
	for index, column := range migrationLogColumns {
		cols[index] = sqlf.Sprintf(column)
	}
	return cols
}

func computeCreateMigrationLogQuery(tableName string, isSQLite bool) (string, error) {
	tmpl := createMigrationTableTmpl
	if isSQLite {
		tmpl = createMigrationTableSQLiteTmpl
	}
	return computeSQLQueryFromTemplate(tableName, tmpl)
}

func computeSelectMigrationLogQuery(tableName string) (string, error) {
	return computeSQLQueryFromTemplate(tableName, selectMigrationsTmpl)
}

func computeInsertMigrationLogQuery(tableName string) (string, error) {
	return computeSQLQueryFromTemplate(tableName, insertMigrationTmpl)
}

func computeDeleteMigrationLogQuery(tableName string) (string, error) {
	return computeSQLQueryFromTemplate(tableName, deleteMigrationTmpl)
}

func computeSQLQueryFromTemplate(tableName string, tmpl *template.Template) (string, error) {
	query := new(bytes.Buffer)
	if err := tmpl.Execute(query, struct {
		TableName string
	}{
		TableName: tableName,
	}); err != nil {
		return "", err
	}
	return query.String(), nil
}
