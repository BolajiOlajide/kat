package update

import (
	"archive/zip"
	"bytes"
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
	tempDir := t.TempDir()

	// Determine binary name and archive format based on platform
	binaryName := "kat"
	archiveExt := ".tar.gz"
	if runtime.GOOS == "windows" {
		binaryName = "kat.exe"
		archiveExt = ".zip"
	}

	// Create a fake binary for testing
	origPath := filepath.Join(tempDir, binaryName)
	require.NoError(t, os.WriteFile(origPath, []byte("original"), 0755))

	// Create another temp directory to prepare the archive
	archiveDir := t.TempDir()

	// Create the mock binary that will be inside the archive
	mockBinaryPath := filepath.Join(archiveDir, binaryName)
	require.NoError(t, os.WriteFile(mockBinaryPath, []byte("updated"), 0755))

	// Create the appropriate archive based on platform
	archivePath := filepath.Join(archiveDir, "kat"+archiveExt)
	var archiveContent []byte

	var err error
	if runtime.GOOS == "windows" {
		// Create ZIP archive
		archiveContent, err = createZipArchive(mockBinaryPath, binaryName)
		require.NoError(t, err)
	} else {
		// Create tar.gz archive using external tar command
		cmd := exec.Command("tar", "czf", archivePath, "-C", archiveDir, binaryName)
		require.NoError(t, cmd.Run())
		archiveContent, err = os.ReadFile(archivePath)
		require.NoError(t, err)
	}

	// Create a test server that returns the archive
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(archiveContent)
	}))
	defer server.Close()

	// Test download and replacement
	require.NoError(t, DownloadAndReplace(server.URL, origPath, io.Discard))

	// Verify content was replaced
	content, err := os.ReadFile(origPath)
	require.NoError(t, err)
	require.Equal(t, "updated", string(content))
}

// createZipArchive creates a ZIP archive containing the specified file
func createZipArchive(filePath, fileName string) ([]byte, error) {
	// Create a buffer to hold the ZIP archive
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add the file to the ZIP
	fileWriter, err := zipWriter.Create(fileName)
	if err != nil {
		return nil, err
	}

	// Read the source file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Write the content to the ZIP
	_, err = fileWriter.Write(fileContent)
	if err != nil {
		return nil, err
	}

	// Close the ZIP writer
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
