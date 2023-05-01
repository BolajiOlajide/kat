package types

import "github.com/keegancsmith/sqlf"

// Definition represents the definition of a single migration.
// It contains what gets executed by the migration operation.
type Definition struct {
	MigrationMetadata

	UpQuery   *sqlf.Query
	DownQuery *sqlf.Query
}
