package runner

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/keegancsmith/sqlf"
)

var migrationLogColumns = []string{
	"id",
	"name",
	"migration_time",
}

var migrationLogInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("migration_time"),
}

func computeMigrationLogColumns(tableName string) []*sqlf.Query {
	var cols = make([]*sqlf.Query, len(migrationLogColumns))
	for index, column := range migrationLogColumns {
		cols[index] = sqlf.Sprintf(fmt.Sprintf("%s.%s", tableName, column))
	}
	return cols
}

func computeCreateMigrationLogQuery(tableName string) (string, error) {
	return computeSQLQueryFromTemplate(tableName, createMigrationTableTmpl)
}

func computeSelectMigrationLogQuery(tableName string) (string, error) {
	return computeSQLQueryFromTemplate(tableName, selectMigrationTmpl)
}

func computeInsertMigrationLogQuery(tableName string) (string, error) {
	return computeSQLQueryFromTemplate(tableName, insertMigrationTmpl)
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
