package types

import "time"

type MigrationLog struct {
	ID        int
	Name      string
	Timestamp int64
	CreatedAt time.Time
}
