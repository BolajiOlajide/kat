package update

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/BolajiOlajide/kat/internal/version"

	"github.com/stretchr/testify/assert"
)

func TestCheckForUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
			"tag_name": "v1.0.1",
			"assets": [
				{
					"name": "kat_1.0.1_linux_amd64.tar.gz",
					"browser_download_url": "https://example.com/download/linux"
				},
				{
					"name": "kat_1.0.1_darwin_amd64.tar.gz",
					"browser_download_url": "https://example.com/download/darwin"
				},
        {
					"name": "kat_1.0.1_darwin_arm64.tar.gz",
					"browser_download_url": "https://example.com/download/darwin"
				},
				{
					"name": "kat_1.0.1_windows_amd64.zip",
					"browser_download_url": "https://example.com/download/windows"
				}
			]
		}`)
	}))
	defer server.Close()

	// Save the original constant value and restore it after the test
	originalReleaseURL := gitHubReleaseURL
	gitHubReleaseURL = server.URL

	// Mock current version
	version.MockVersion = func() string { return "v1.0.0" }

	t.Cleanup(func() {
		version.MockVersion = nil
		gitHubReleaseURL = originalReleaseURL
	})

	// Test when update is available
	hasUpdate, latestVersion, downloadURL, err := CheckForUpdates()
	assert.NoError(t, err)
	assert.True(t, hasUpdate)
	assert.Equal(t, "1.0.1", latestVersion)

	// URL should match the platform
	expectedURL := ""
	switch runtime.GOOS {
	case "linux":
		expectedURL = "https://example.com/download/linux"
	case "darwin":
		expectedURL = "https://example.com/download/darwin"
	case "windows":
		expectedURL = "https://example.com/download/windows"
	}
	assert.Equal(t, expectedURL, downloadURL)

	// Test when already up to date
	version.MockVersion = func() string { return "v1.0.1" }
	hasUpdate, _, _, err = CheckForUpdates()
	assert.NoError(t, err)
	assert.False(t, hasUpdate)
}

func TestDownloadAndReplace(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "kat-update-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a fake binary for testing
	origPath := filepath.Join(tempDir, "kat")
	if runtime.GOOS == "windows" {
		origPath += ".exe"
	}

	err = os.WriteFile(origPath, []byte("original"), 0755)
	assert.NoError(t, err)

	// Create a test server that returns a binary
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		io.WriteString(w, "updated")
	}))
	defer server.Close()

	// Test download and replacement
	progressWriter := io.Discard // Discard progress output in tests
	err = DownloadAndReplace(server.URL, origPath, progressWriter)
	assert.NoError(t, err)

	// Verify content was replaced
	content, err := os.ReadFile(origPath)
	assert.NoError(t, err)
	assert.Equal(t, "updated", string(content))
}
