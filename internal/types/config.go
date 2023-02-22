package types

// Config is the configuration structure Kat uses for it's operations.
// It contains the database connection information and the path to the
// migrations directory.
type Config struct {
	Migration MigrationConfig    `yaml:"migration"`
	Database  DatabaseConnection `yaml:"database"`
}

// DatabaseConnection is the configuration structure for the database connection.
// This is required to run migration operations such as (up, down, reset).
// One of URL or (Host, Port, User, Password, Name) must be set.
type DatabaseConnection struct {
	URL string `yaml:"url"`

	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
}

// Valid checks if the database connection is set..
func (d DatabaseConnection) Valid() bool {
	return d.URL != "" || (d.Host != "" && d.Port != 0 && d.Password != "" && d.Name != "" && d.User != "")
}

// MigrationConfig is the configuration structure for the migration directory.
type MigrationConfig struct {
	Directory string `yaml:"directory"`
	TableName string `yaml:"table_name"`
}
