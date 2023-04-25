package types

import "github.com/keegancsmith/sqlf"

// Definition represents the definition of a single migration.
// It contains what gets executed by the migration operation.
type Definition struct {
	Name          string
	Timestamp     int64
	UpQuery       *sqlf.Query
	DownQuery     *sqlf.Query
	IndexMetadata *IndexMetadata
}

// IndexMetadata represents the metadata of an index on a table.
type IndexMetadata struct {
	TableName string
	IndexName string
}
