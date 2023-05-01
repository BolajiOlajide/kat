package migration

import (
	"strings"

	"github.com/keegancsmith/sqlf"
)

// queryFromString creates a sqlf Query object from the conetents of a file or serialized
// string literal. The resulting query is canonicalized. SQL placeholder values are also
// escaped, so when sqlf.Query renders it the placeholders will be valid and not replaced
// by a "missing" parameterized value.
func queryFromString(query string) *sqlf.Query {
	return sqlf.Sprintf(strings.ReplaceAll(canonicalizeQuery(query), "%", "%%"))
}

func canonicalizeQuery(query string) string {
	return strings.TrimSpace(
		strings.TrimSuffix(
			strings.TrimPrefix(
				strings.TrimSpace(query),
				"BEGIN;",
			),
			"COMMIT;",
		),
	)
}
