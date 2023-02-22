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

// CanonicalizeQuery removes old cruft from historic definitions to make them conform to
// the new standards. This includes YAML metadata frontmatter as well as explicit tranaction
// blocks around golang-migrate-era migration definitions.
func canonicalizeQuery(query string) string {
	// Strip out embedded yaml frontmatter (existed temporarily)
	parts := strings.SplitN(query, "-- +++\n", 3)
	if len(parts) == 3 {
		query = parts[2]
	}

	// Strip outermost transactions
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
