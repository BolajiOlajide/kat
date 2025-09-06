package constants

type contextKey int

const KatConfigurationFileName string = "kat.conf.yaml"

const (
	KatConfigKey contextKey = iota
)

type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite"
)

func (d Driver) Valid() bool {
	switch d {
	case DriverPostgres, DriverSQLite:
		return true
	default:
		return false
	}
}
