package runner

import (
	"testing"
	"text/template"

	"github.com/keegancsmith/sqlf"
)

func TestComputeSQLQueryFromTemplate(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		tmpl      *template.Template
		want      string
		wantErr   bool
	}{
		{
			name:      "valid template with table name",
			tableName: "users",
			tmpl:      template.Must(template.New("test").Parse("SELECT * FROM {{.TableName}}")),
			want:      "SELECT * FROM users",
			wantErr:   false,
		},
		{
			name:      "valid template with empty table name",
			tableName: "",
			tmpl:      template.Must(template.New("test").Parse("SELECT * FROM {{.TableName}}")),
			want:      "SELECT * FROM ",
			wantErr:   false,
		},
		{
			name:      "invalid template",
			tableName: "users",
			tmpl:      template.Must(template.New("test").Parse("SELECT * FROM {{.InvalidField}}")),
			want:      "",
			wantErr:   true,
		},
		{
			name:      "template with multiple table references",
			tableName: "users",
			tmpl:      template.Must(template.New("test").Parse("SELECT * FROM {{.TableName}} WHERE id IN (SELECT user_id FROM {{.TableName}}_history)")),
			want:      "SELECT * FROM users WHERE id IN (SELECT user_id FROM users_history)",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeSQLQueryFromTemplate(tt.tableName, tt.tmpl)
			if (err != nil) != tt.wantErr {
				t.Errorf("computeSQLQueryFromTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("computeSQLQueryFromTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasMultipleStatements(t *testing.T) {
	tests := []struct {
		name     string
		query    *sqlf.Query
		expected bool
	}{
		{
			name:     "single statement without trailing semicolon",
			query:    sqlf.Sprintf("CREATE INDEX CONCURRENTLY idx_foo ON bar (baz)"),
			expected: false,
		},
		{
			name:     "single statement with trailing semicolon",
			query:    sqlf.Sprintf("CREATE INDEX CONCURRENTLY idx_foo ON bar (baz);"),
			expected: false,
		},
		{
			name:     "multiple statements",
			query:    sqlf.Sprintf("CREATE INDEX CONCURRENTLY idx_foo ON bar (baz); CREATE INDEX CONCURRENTLY idx_qux ON bar (qux);"),
			expected: true,
		},
		{
			name:     "multiple statements without trailing semicolon",
			query:    sqlf.Sprintf("CREATE INDEX idx_foo ON bar (baz); CREATE INDEX idx_qux ON bar (qux)"),
			expected: true,
		},
		{
			name:     "empty query",
			query:    sqlf.Sprintf(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasMultipleStatements(tt.query)
			if got != tt.expected {
				t.Errorf("hasMultipleStatements() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestComputeMigrationLogColumns(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		want      []*sqlf.Query
	}{
		{
			name:      "valid table name",
			tableName: "migrations",
			want: []*sqlf.Query{
				sqlf.Sprintf("migrations.id"),
				sqlf.Sprintf("migrations.name"),
				sqlf.Sprintf("migrations.migration_time"),
				sqlf.Sprintf("migrations.duration"),
			},
		},
		{
			name:      "empty table name",
			tableName: "",
			want: []*sqlf.Query{
				sqlf.Sprintf(".id"),
				sqlf.Sprintf(".name"),
				sqlf.Sprintf(".migration_time"),
				sqlf.Sprintf(".duration"),
			},
		},
		{
			name:      "table name with schema",
			tableName: "public.migrations",
			want: []*sqlf.Query{
				sqlf.Sprintf("public.migrations.id"),
				sqlf.Sprintf("public.migrations.name"),
				sqlf.Sprintf("public.migrations.migration_time"),
				sqlf.Sprintf("public.migrations.duration"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeMigrationLogColumns(tt.tableName)
			if len(got) != len(tt.want) {
				t.Errorf("computeMigrationLogColumns() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Query(sqlf.PostgresBindVar) != tt.want[i].Query(sqlf.PostgresBindVar) {
					t.Errorf("computeMigrationLogColumns()[%d] = %v, want %v", i, got[i].Query(sqlf.PostgresBindVar), tt.want[i].Query(sqlf.PostgresBindVar))
				}
			}
		})
	}
}
