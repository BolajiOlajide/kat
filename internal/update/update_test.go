package update

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/BolajiOlajide/kat/internal/version"
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
	require.NoError(t, err)
	require.True(t, hasUpdate)
	require.Equal(t, "1.0.1", latestVersion)

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
	require.Equal(t, expectedURL, downloadURL)

	// Test when already up to date
	version.MockVersion = func() string { return "v1.0.1" }
	hasUpdate, _, _, err = CheckForUpdates()
	require.NoError(t, err)
	require.False(t, hasUpdate)
}

func TestDownloadAndReplace(t *testing.T) {
	// Skip if not on a system where we can create tar archives
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}

	tempDir := t.TempDir()

	// Create a fake binary for testing
	origPath := filepath.Join(tempDir, "kat")
	if runtime.GOOS == "windows" {
		origPath += ".exe"
	}

	require.NoError(t, os.WriteFile(origPath, []byte("original"), 0755))

	// Create another temp directory to prepare a tar.gz file
	tarDir := t.TempDir()

	// Create the mock binary that will be inside the tar
	mockBinaryPath := filepath.Join(tarDir, "kat")
	require.NoError(t, os.WriteFile(mockBinaryPath, []byte("updated"), 0755))

	// Create a tar.gz file
	tarPath := filepath.Join(tarDir, "kat.tar.gz")
	cmd := exec.Command("tar", "czf", tarPath, "-C", tarDir, "kat")
	require.NoError(t, cmd.Run())

	// Read the tar.gz content
	tarContent, err := os.ReadFile(tarPath)
	require.NoError(t, err)

	// Create a test server that returns the tar.gz
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(tarContent)
	}))
	defer server.Close()

	// Test download and replacement
	require.NoError(t, DownloadAndReplace(server.URL, origPath, io.Discard))

	// Verify content was replaced
	content, err := os.ReadFile(origPath)
	require.NoError(t, err)
	require.Equal(t, "updated", string(content))
}
