package types

type Driver string

const (
	DriverPostgres Driver = "postgres"
	DriverSQLite   Driver = "sqlite"
)

func (d Driver) IsEmpty() bool {
	return d == ""
}

func (d Driver) Valid() bool {
	switch d {
	case DriverPostgres, DriverSQLite:
		return true
	default:
		return false
	}
}
