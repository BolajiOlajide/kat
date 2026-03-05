package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDBTimeouts(t *testing.T) {
	tests := []struct {
		name     string
		info     DatabaseInfo
		expected *DBTimeouts
		wantErr  string
	}{
		{
			name:     "all fields empty returns nil",
			info:     DatabaseInfo{},
			expected: nil,
		},
		{
			name: "all duration fields set",
			info: DatabaseInfo{
				ConnectTimeout:   "5s",
				StatementTimeout: "30s",
				MaxOpenConns:     20,
				MaxIdleConns:     10,
				ConnMaxLifetime:  "1h",
				DefaultTimeout:   "60s",
			},
			expected: &DBTimeouts{
				ConnectTimeout:   5 * time.Second,
				StatementTimeout: 30 * time.Second,
				MaxOpenConns:     20,
				MaxIdleConns:     10,
				ConnMaxLifetime:  1 * time.Hour,
				DefaultTimeout:   60 * time.Second,
			},
		},
		{
			name: "only pool settings triggers parsing",
			info: DatabaseInfo{
				MaxOpenConns: 25,
			},
			expected: &DBTimeouts{
				MaxOpenConns: 25,
			},
		},
		{
			name: "only connect_timeout set",
			info: DatabaseInfo{
				ConnectTimeout: "10s",
			},
			expected: &DBTimeouts{
				ConnectTimeout: 10 * time.Second,
			},
		},
		{
			name: "only max_idle_conns set",
			info: DatabaseInfo{
				MaxIdleConns: 3,
			},
			expected: &DBTimeouts{
				MaxIdleConns: 3,
			},
		},
		{
			name:    "invalid connect_timeout",
			info:    DatabaseInfo{ConnectTimeout: "notaduration"},
			wantErr: "invalid connect_timeout",
		},
		{
			name:    "invalid statement_timeout",
			info:    DatabaseInfo{StatementTimeout: "bad"},
			wantErr: "invalid statement_timeout",
		},
		{
			name:    "invalid conn_max_lifetime",
			info:    DatabaseInfo{ConnMaxLifetime: "xyz"},
			wantErr: "invalid conn_max_lifetime",
		},
		{
			name:    "invalid default_timeout",
			info:    DatabaseInfo{DefaultTimeout: "???"},
			wantErr: "invalid default_timeout",
		},
		{
			name: "millisecond durations",
			info: DatabaseInfo{
				ConnectTimeout: "500ms",
				DefaultTimeout: "200ms",
			},
			expected: &DBTimeouts{
				ConnectTimeout: 500 * time.Millisecond,
				DefaultTimeout: 200 * time.Millisecond,
			},
		},
		{
			name: "minute durations",
			info: DatabaseInfo{
				StatementTimeout: "5m",
				ConnMaxLifetime:  "30m",
			},
			expected: &DBTimeouts{
				StatementTimeout: 5 * time.Minute,
				ConnMaxLifetime:  30 * time.Minute,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.info.ParseDBTimeouts()

			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)

			if tc.expected == nil {
				require.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			require.Equal(t, tc.expected.ConnectTimeout, result.ConnectTimeout)
			require.Equal(t, tc.expected.StatementTimeout, result.StatementTimeout)
			require.Equal(t, tc.expected.MaxOpenConns, result.MaxOpenConns)
			require.Equal(t, tc.expected.MaxIdleConns, result.MaxIdleConns)
			require.Equal(t, tc.expected.ConnMaxLifetime, result.ConnMaxLifetime)
			require.Equal(t, tc.expected.DefaultTimeout, result.DefaultTimeout)
		})
	}
}
