package types

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type MigrationLog struct {
	ID            int
	Name          string
	MigrationTime time.Time
	Duration      pgtype.Interval
}
