package types

// Migration represents a migration file definition.
type Migration struct {
	Up        string
	Down      string
	Metadata  string
	Timestamp int64
}

// MigrationMetadata represents the metadata of a migration file.
type MigrationMetadata struct {
	Name        string  `yaml:"name"`
	Timestamp   int64   `yaml:"timestamp"`
	Description string  `yaml:"description,omitempty"`
	
	// Parents represents the timestamps of parent migrations in the dependency graph.
	// This field is used to define dependencies between migrations.
	Parents     []int64 `yaml:"parents,omitempty,flow"`
}

// MigrationOperationType represents the type of migration operation.
type MigrationOperationType int

const (
	// UpMigrationOperation represents an upgrade operation.
	UpMigrationOperation MigrationOperationType = iota

	// DownMigrationOperation represents a downgrade operation.
	DownMigrationOperation
)
