package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/cockroachdb/errors"

	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/version"
)

// GitHubReleaseURL is the URL for GitHub's latest release API
var gitHubReleaseURL = "https://api.github.com/repos/BolajiOlajide/kat/releases/latest"

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
	req, err := http.NewRequest("GET", gitHubReleaseURL, nil)
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
	platformSuffix := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := ""

	if hasUpdate {
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
	execDir := filepath.Dir(execPath)
	execName := filepath.Base(execPath)

	// Preflight: verify we can write to the destination directory before downloading.
	probe, err := os.CreateTemp(execDir, execName+".probe-*")
	if err != nil {
		if os.IsPermission(err) {
			return errors.Newf("permission denied: cannot write to %s — try running with sudo", execDir)
		}
		return errors.Wrap(err, "cannot write to destination directory")
	}
	_ = probe.Close()
	_ = os.Remove(probe.Name())

	// Create a temporary directory to work in
	tempDir, err := os.MkdirTemp("", "kat-update-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory on exit

	// Create a temporary file for the archive
	tempArchivePath := filepath.Join(tempDir, "kat.tar.gz")
	tempArchive, err := os.Create(tempArchivePath)
	if err != nil {
		return errors.Wrap(err, "failed to create temporary file")
	}

	// Download the file
	fmt.Fprintf(progressWriter, "%sDownloading update...%s\n", output.StyleInfo, output.StyleReset)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return errors.Wrap(err, "failed to download update")
	}
	defer resp.Body.Close()

	// Check if download was successful
	if resp.StatusCode != http.StatusOK {
		return errors.Newf("failed to download update: HTTP %d", resp.StatusCode)
	}

	// Copy the downloaded content to the temporary archive file
	_, err = io.Copy(tempArchive, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write downloaded file")
	}

	// Close the archive file
	if err := tempArchive.Close(); err != nil {
		return errors.Wrap(err, "failed to close the temporary archive file")
	}
	// Extract the binary from the archive
	fmt.Fprintf(progressWriter, "%sExtracting binary...%s\n", output.StyleInfo, output.StyleReset)

	// Create path for the extracted binary
	tempBinaryPath := filepath.Join(tempDir, "kat")

	// Use tar to extract the file
	cmd := exec.Command("tar", "xzf", tempArchivePath, "-C", tempDir)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to extract archive")
	}

	// Make the extracted binary executable
	if err := os.Chmod(tempBinaryPath, 0755); err != nil {
		return errors.Wrap(err, "failed to make binary executable")
	}

	// Copy the new binary into the destination directory first.
	// os.Rename cannot cross filesystem boundaries (e.g. from /var/folders to
	// /usr/local/bin on macOS), so we copy into the same directory and then
	// do an atomic same-directory rename.
	newExecPath := filepath.Join(execDir, execName+".new")
	if err := copyFile(tempBinaryPath, newExecPath, 0755); err != nil {
		if os.IsPermission(err) {
			return errors.Newf("permission denied: cannot write to %s — try running with sudo", execDir)
		}
		return errors.Wrap(err, "failed to copy new executable to destination directory")
	}

	// Replace the current executable with the new one (atomic within same directory)
	if err := os.Rename(newExecPath, execPath); err != nil {
		_ = os.Remove(newExecPath)
		if os.IsPermission(err) {
			return errors.Newf("permission denied: cannot replace %s — try running with sudo", execPath)
		}
		return errors.Wrap(err, "failed to replace current executable")
	}

	fmt.Fprintf(progressWriter, "%sUpdate successfully installed%s\n", output.StyleSuccess, output.StyleReset)
	return nil
}

// copyFile copies src to dst with the given permission bits.
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
