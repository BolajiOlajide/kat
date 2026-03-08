package database

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbdriver "github.com/BolajiOlajide/kat/internal/database/driver"
)

func TestDefaultDBConfig(t *testing.T) {
	t.Run("postgres", func(t *testing.T) {
		cfg := DefaultDBConfig(dbdriver.PostgresDriver)

		require.Equal(t, 10*time.Second, cfg.ConnectTimeout)
		require.Equal(t, time.Duration(0), cfg.StatementTimeout)
		require.Equal(t, 2, cfg.MaxOpenConns)
		require.Equal(t, 2, cfg.MaxIdleConns)
		require.Equal(t, 5*time.Minute, cfg.ConnMaxLifetime)
		require.Equal(t, time.Duration(0), cfg.DefaultTimeout)
	})

	t.Run("sqlite", func(t *testing.T) {
		cfg := DefaultDBConfig(dbdriver.SqliteDriver)

		require.Equal(t, 10*time.Second, cfg.ConnectTimeout)
		require.Equal(t, time.Duration(0), cfg.StatementTimeout)
		require.Equal(t, 1, cfg.MaxOpenConns)
		require.Equal(t, 1, cfg.MaxIdleConns)
		require.Equal(t, 2*time.Minute, cfg.ConnMaxLifetime)
		require.Equal(t, time.Duration(0), cfg.DefaultTimeout)
	})
}

func TestEnsureTimeoutsInDSN(t *testing.T) {
	tests := []struct {
		name             string
		connURL          string
		connectTimeout   time.Duration
		statementTimeout time.Duration
		wantErr          bool
		check            func(t *testing.T, result string)
	}{
		{
			name:             "adds connect_timeout to URL",
			connURL:          "postgres://user:pass@localhost:5432/db?sslmode=disable",
			connectTimeout:   10 * time.Second,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "10", u.Query().Get("connect_timeout"))
			},
		},
		{
			name:             "adds statement_timeout to URL",
			connURL:          "postgres://user:pass@localhost:5432/db?sslmode=disable",
			connectTimeout:   0,
			statementTimeout: 30 * time.Second,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "30000", u.Query().Get("statement_timeout"))
			},
		},
		{
			name:             "adds both timeouts",
			connURL:          "postgres://user:pass@localhost:5432/db",
			connectTimeout:   5 * time.Second,
			statementTimeout: 1 * time.Minute,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "5", u.Query().Get("connect_timeout"))
				require.Equal(t, "60000", u.Query().Get("statement_timeout"))
			},
		},
		{
			name:             "does not override existing connect_timeout",
			connURL:          "postgres://user:pass@localhost:5432/db?connect_timeout=30",
			connectTimeout:   5 * time.Second,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "30", u.Query().Get("connect_timeout"))
			},
		},
		{
			name:             "does not override existing statement_timeout",
			connURL:          "postgres://user:pass@localhost:5432/db?statement_timeout=5000",
			connectTimeout:   0,
			statementTimeout: 30 * time.Second,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "5000", u.Query().Get("statement_timeout"))
			},
		},
		{
			name:             "zero timeouts leave URL unchanged",
			connURL:          "postgres://user:pass@localhost:5432/db?sslmode=disable",
			connectTimeout:   0,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Empty(t, u.Query().Get("connect_timeout"))
				require.Empty(t, u.Query().Get("statement_timeout"))
			},
		},
		{
			name:             "sub-second connect_timeout rounds up to 1",
			connURL:          "postgres://user:pass@localhost:5432/db",
			connectTimeout:   500 * time.Millisecond,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "1", u.Query().Get("connect_timeout"))
			},
		},
		{
			name:             "preserves other query params",
			connURL:          "postgres://user:pass@localhost:5432/db?sslmode=require&application_name=kat",
			connectTimeout:   10 * time.Second,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				u, err := url.Parse(result)
				require.NoError(t, err)
				require.Equal(t, "require", u.Query().Get("sslmode"))
				require.Equal(t, "kat", u.Query().Get("application_name"))
				require.Equal(t, "10", u.Query().Get("connect_timeout"))
			},
		},
		{
			name:             "delegates to key=value path for key=value DSN",
			connURL:          "host=localhost port=5432 dbname=testdb",
			connectTimeout:   10 * time.Second,
			statementTimeout: 30 * time.Second,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "connect_timeout=10")
				require.Contains(t, result, "statement_timeout=30000")
				require.Contains(t, result, "host=localhost")
				require.Contains(t, result, "port=5432")
				require.Contains(t, result, "dbname=testdb")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ensureTimeoutsInDSN(tc.connURL, tc.connectTimeout, tc.statementTimeout)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			tc.check(t, result)
		})
	}
}

func TestEnsureTimeoutsInKeyValueDSN(t *testing.T) {
	tests := []struct {
		name             string
		dsn              string
		connectTimeout   time.Duration
		statementTimeout time.Duration
		check            func(t *testing.T, result string)
	}{
		{
			name:             "adds connect_timeout",
			dsn:              "host=localhost port=5432",
			connectTimeout:   10 * time.Second,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "connect_timeout=10")
			},
		},
		{
			name:             "adds statement_timeout",
			dsn:              "host=localhost port=5432",
			connectTimeout:   0,
			statementTimeout: 30 * time.Second,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "statement_timeout=30000")
			},
		},
		{
			name:             "adds both timeouts",
			dsn:              "host=localhost port=5432",
			connectTimeout:   10 * time.Second,
			statementTimeout: 30 * time.Second,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "connect_timeout=10")
				require.Contains(t, result, "statement_timeout=30000")
			},
		},
		{
			name:             "preserves existing params",
			dsn:              "host=localhost port=5432 sslmode=disable",
			connectTimeout:   10 * time.Second,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "sslmode=disable")
				require.Contains(t, result, "connect_timeout=10")
			},
		},
		{
			name:             "does not override existing connect_timeout",
			dsn:              "host=localhost port=5432 connect_timeout=30",
			connectTimeout:   5 * time.Second,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "connect_timeout=30")
				require.NotContains(t, result, "connect_timeout=5")
			},
		},
		{
			name:             "does not override existing statement_timeout",
			dsn:              "host=localhost port=5432 statement_timeout=5000",
			connectTimeout:   0,
			statementTimeout: 30 * time.Second,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "statement_timeout=5000")
				require.NotContains(t, result, "statement_timeout=30000")
			},
		},
		{
			name:             "zero timeouts leave DSN unchanged",
			dsn:              "host=localhost port=5432",
			connectTimeout:   0,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				require.Equal(t, "host=localhost port=5432", result)
			},
		},
		{
			name:             "sub-second connect_timeout rounds up to 1",
			dsn:              "host=localhost port=5432",
			connectTimeout:   500 * time.Millisecond,
			statementTimeout: 0,
			check: func(t *testing.T, result string) {
				require.Contains(t, result, "connect_timeout=1")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ensureTimeoutsInKeyValueDSN(tc.dsn, tc.connectTimeout, tc.statementTimeout)
			tc.check(t, result)
		})
	}
}
