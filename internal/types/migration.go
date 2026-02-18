// Package types defines the core data structures and types used throughout
// the kat migration system. It provides type definitions for migrations,
// configuration, and metadata structures.
//
// The package includes:
//   - Migration definitions and metadata structures
//   - Configuration types for database and migration settings
//   - Operation types for different migration actions
//   - Validation logic for migration data
package types

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
)

// Definition represents the definition of a single migration.
// It contains what gets executed by the migration operation.
type Definition struct {
	MigrationMetadata

	UpQuery   *sqlf.Query
	DownQuery *sqlf.Query
}

func (d Definition) FileName() string {
	return fmt.Sprintf("%d_%s", d.Timestamp, d.Name)
}

// TemporaryMigrationInfo represents a temporary migration file definition for creation.
type TemporaryMigrationInfo struct {
	Up        string
	Down      string
	Metadata  string
	Timestamp int64
}

// MigrationMetadata represents the metadata of a migration file.
type MigrationMetadata struct {
	Name        string `yaml:"name"`
	Timestamp   int64  `yaml:"timestamp"`
	Description string `yaml:"description,omitempty"`

	// Parents represents the timestamps of parent migrations in the dependency graph.
	// This field is used to define dependencies between migrations.
	Parents []int64 `yaml:"parents,omitempty,flow"`

	// NoTransaction indicates that this migration should not be wrapped in a transaction.
	// This is required for operations like CREATE INDEX CONCURRENTLY which cannot run
	// inside a transaction block.
	NoTransaction bool `yaml:"no_transaction,omitempty"`
}

// MigrationOperationType represents the type of migration operation.
type MigrationOperationType int

func (m MigrationOperationType) IsUpMigration() bool {
	return m == UpMigrationOperation
}

func (m MigrationOperationType) IsDownMigration() bool {
	return m == DownMigrationOperation
}

func (m MigrationOperationType) String() string {
	switch m {
	case UpMigrationOperation:
		return "up"
	case DownMigrationOperation:
		return "down"
	default:
		return ""
	}
}

const (
	// UpMigrationOperation represents an upgrade operation.
	UpMigrationOperation MigrationOperationType = iota

	// DownMigrationOperation represents a downgrade operation.
	DownMigrationOperation
)
