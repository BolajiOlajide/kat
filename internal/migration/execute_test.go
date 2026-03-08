package migration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/database"
	dbdriver "github.com/BolajiOlajide/kat/internal/database/driver"
	"github.com/BolajiOlajide/kat/internal/types"
)

func TestDBConfigFromCfg(t *testing.T) {
	drivers := []struct {
		name   string
		driver dbdriver.DatabaseDriver
	}{
		{"postgres", dbdriver.PostgresDriver},
		{"sqlite", dbdriver.SqliteDriver},
	}

	for _, drv := range drivers {
		t.Run(drv.name, func(t *testing.T) {
			defaults := database.DefaultDBConfig(drv.driver)

			tests := []struct {
				name     string
				cfg      types.Config
				expected database.DBConfig
				wantErr  string
			}{
				{
					name: "no timeouts configured returns defaults",
					cfg: types.Config{
						Database: types.DatabaseInfo{Driver: drv.driver},
					},
					expected: defaults,
				},
				{
					name: "all timeouts configured",
					cfg: types.Config{
						Database: types.DatabaseInfo{
							Driver:           drv.driver,
							ConnectTimeout:   "5s",
							StatementTimeout: "30s",
							MaxOpenConns:     20,
							MaxIdleConns:     10,
							ConnMaxLifetime:  "1h",
							DefaultTimeout:   "60s",
						},
					},
					expected: database.DBConfig{
						ConnectTimeout:   5 * time.Second,
						StatementTimeout: 30 * time.Second,
						MaxOpenConns:     20,
						MaxIdleConns:     10,
						ConnMaxLifetime:  1 * time.Hour,
						DefaultTimeout:   60 * time.Second,
					},
				},
				{
					name: "partial config overrides only specified fields",
					cfg: types.Config{
						Database: types.DatabaseInfo{
							Driver:         drv.driver,
							ConnectTimeout: "3s",
							MaxOpenConns:   50,
						},
					},
					expected: database.DBConfig{
						ConnectTimeout:   3 * time.Second,
						StatementTimeout: defaults.StatementTimeout,
						MaxOpenConns:     50,
						MaxIdleConns:     defaults.MaxIdleConns,
						ConnMaxLifetime:  defaults.ConnMaxLifetime,
						DefaultTimeout:   defaults.DefaultTimeout,
					},
				},
				{
					name: "invalid duration returns error",
					cfg: types.Config{
						Database: types.DatabaseInfo{
							Driver:         drv.driver,
							ConnectTimeout: "notaduration",
						},
					},
					wantErr: "invalid connect_timeout",
				},
				{
					name: "zero-value pool settings keep defaults",
					cfg: types.Config{
						Database: types.DatabaseInfo{
							Driver:         drv.driver,
							ConnectTimeout: "5s",
							MaxOpenConns:   0,
							MaxIdleConns:   0,
						},
					},
					expected: database.DBConfig{
						ConnectTimeout:   5 * time.Second,
						StatementTimeout: defaults.StatementTimeout,
						MaxOpenConns:     defaults.MaxOpenConns,
						MaxIdleConns:     defaults.MaxIdleConns,
						ConnMaxLifetime:  defaults.ConnMaxLifetime,
						DefaultTimeout:   defaults.DefaultTimeout,
					},
				},
			}

			for _, tc := range tests {
				t.Run(tc.name, func(t *testing.T) {
					result, err := DBConfigFromCfg(tc.cfg)

					if tc.wantErr != "" {
						require.Error(t, err)
						require.Contains(t, err.Error(), tc.wantErr)
						return
					}

					require.NoError(t, err)
					require.Equal(t, tc.expected, result)
				})
			}
		})
	}
}
