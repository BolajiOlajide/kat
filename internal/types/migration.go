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
	Name      string `yaml:"name"`
	Timestamp int64  `yaml:"timestamp"`
}

// MigrationOperationType represents the type of migration operation.
type MigrationOperationType int

const (
	// MigrationOperationTypeUpgrade represents an upgrade operation.
	MigrationOperationTypeUpgrade MigrationOperationType = iota

	// MigrationOperationTypeDowngrade represents a downgrade operation.
	MigrationOperationTypeDowngrade
)
