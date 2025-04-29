package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "development version",
			version:  "dev",
			expected: "dev",
		},
		{
			name:     "specific version",
			version:  "v1.0.0",
			expected: "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original version
			originalVersion := version
			// Set test version
			version = tt.version
			// Restore original version after test
			defer func() { version = originalVersion }()

			got := Version()
			if got != tt.expected {
				t.Errorf("Version() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "development version",
			version:  "dev",
			expected: true,
		},
		{
			name:     "specific version",
			version:  "v1.0.0",
			expected: false,
		},
		{
			name:     "empty version",
			version:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original version
			originalVersion := version
			// Set test version
			version = tt.version
			// Restore original version after test
			defer func() { version = originalVersion }()

			got := IsDev()
			if got != tt.expected {
				t.Errorf("IsDev() = %v, want %v", got, tt.expected)
			}
		})
	}
}
