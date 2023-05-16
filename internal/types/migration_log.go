package types

import "time"

type MigrationLog struct {
	ID            int
	Name          string
	MigrationTime time.Time
	Duration      time.Duration
}
