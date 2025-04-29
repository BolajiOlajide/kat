package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/errors"
)

// GitHubReleaseURL is the URL for GitHub's latest release API
var GitHubReleaseURL = "https://api.github.com/repos/BolajiOlajide/kat/releases/latest"

// ReleaseInfo represents the GitHub release API response
type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckForUpdates checks GitHub for a newer version of Kat
// Returns: hasUpdate, latestVersion, downloadURL, error
func CheckForUpdates() (bool, string, string, error) {
	// Get current version without "v" prefix
	currentVersion := strings.TrimPrefix(version.Version(), "v")

	cv, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, "", "", errors.Newf("failed to parse current version %q: %w", currentVersion, err)
	}

	// Create a request to the GitHub API
	req, err := http.NewRequest("GET", GitHubReleaseURL, nil)
	if err != nil {
		return false, "", "", errors.Newf("error creating request: %w", err)
	}

	// Set User-Agent to avoid GitHub API rate limiting
	req.Header.Set("User-Agent", "kat-updater/"+currentVersion)

	// Execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "", "", errors.Newf("error checking for updates: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return false, "", "", errors.Newf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	// Decode the response
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return false, "", "", errors.Newf("error parsing release info: %w", err)
	}

	// Get the version number without the "v" prefix
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	lv, err := semver.NewVersion(latestVersion)
	if err != nil {
		return false, "", "", errors.Newf("failed to parse latest version %q: %w", latestVersion, err)
	}

	// Compare versions (simple string comparison for now)
	// This works for semantic versioning as long as the format is consistent
	hasUpdate := lv.GreaterThan(cv)

	// Find the appropriate download URL for current platform
	downloadURL := ""
	if hasUpdate {
		platformSuffix := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, platformSuffix) {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		// If we couldn't find a matching asset
		if downloadURL == "" {
			return false, "", "", errors.Newf("no matching binary found for your platform: %s", platformSuffix)
		}
	}

	return hasUpdate, latestVersion, downloadURL, nil
}

// DownloadAndReplace downloads a new binary and replaces the current one
func DownloadAndReplace(downloadURL, execPath string, progressWriter io.Writer) error {
	// Create a temporary file to download to
	tempFile, err := os.CreateTemp("", "kat-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath) // Clean up temp file on exit

	// Download the file
	fmt.Fprintf(progressWriter, "%sDownloading update...%s\n", output.StyleInfo, output.StyleReset)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	// Check if download was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download update: HTTP %d", resp.StatusCode)
	}

	// Make the temporary file executable (important for Unix-like systems)
	if err := tempFile.Chmod(0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Copy the downloaded content to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write downloaded file: %w", err)
	}

	// Close the file before replacing
	tempFile.Close()

	// On Unix-like systems, we can replace the binary directly
	execDir := filepath.Dir(execPath)
	execName := filepath.Base(execPath)

	// Move the new executable to the same directory as the current one
	newExecPath := filepath.Join(execDir, execName+".new")
	if err := os.Rename(tempFilePath, newExecPath); err != nil {
		return fmt.Errorf("failed to move new executable to destination directory: %w", err)
	}

	// Replace the current executable with the new one
	if err := os.Rename(newExecPath, execPath); err != nil {
		return fmt.Errorf("failed to replace current executable: %w", err)
	}

	fmt.Fprintf(progressWriter, "%sUpdate successfully installed%s\n", output.StyleSuccess, output.StyleReset)
	return nil
}
