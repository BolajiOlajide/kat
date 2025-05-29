package migration

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetMigrationsFS(t *testing.T) {
	testCases := []struct {
		name        string
		setupDir    bool
		expectError bool
		errorType   error
	}{
		{
			name:        "success - existing directory",
			setupDir:    true,
			expectError: false,
		},
		{
			name:        "error - non-existent directory",
			setupDir:    false,
			expectError: true,
			errorType:   ErrMigrationsDirNotExist,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testDir := t.TempDir()
			migrationsPath := filepath.Join(testDir, "migrations")

			if tc.setupDir {
				require.NoError(t, os.MkdirAll(migrationsPath, 0755), "failed to create migrations dir")
			}

			result, err := getMigrationsFS(migrationsPath)

			if tc.expectError {
				require.Error(t, err, "expected an error but got none")
				require.ErrorIs(t, err, tc.errorType, "error type mismatch")
				require.Nil(t, result, "expected nil filesystem")
			} else {
				require.NoError(t, err, "unexpected error")
				require.NotNil(t, result, "expected non-nil filesystem")
				require.Implements(t, (*fs.FS)(nil), result, "result should implement fs.FS")
			}
		})
	}
}