package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
		var candidates []string
		
		// Collect all matching assets
		for _, asset := range release.Assets {
			if strings.Contains(asset.Name, platformSuffix) {
				candidates = append(candidates, asset.BrowserDownloadURL)
			}
		}
		
		// Select the best format for platform
		for _, candidate := range candidates {
			if runtime.GOOS == "windows" && strings.HasSuffix(candidate, ".zip") {
				downloadURL = candidate
				break
			} else if runtime.GOOS != "windows" && strings.HasSuffix(candidate, ".tar.gz") {
				downloadURL = candidate
				break
			}
		}
		
		// Fallback to any matching asset if preferred format not found
		if downloadURL == "" && len(candidates) > 0 {
			downloadURL = candidates[0]
		}

		// If we couldn't find any matching asset
		if downloadURL == "" {
			return false, "", "", errors.Newf("no matching binary found for your platform: %s", platformSuffix)
		}
	}

	return hasUpdate, latestVersion, downloadURL, nil
}

// DownloadAndReplace downloads a new binary and replaces the current one
func DownloadAndReplace(downloadURL, execPath string, progressWriter io.Writer) error {
	// Create a temporary directory to work in
	tempDir, err := os.MkdirTemp("", "kat-update-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary directory")
	}
	defer os.RemoveAll(tempDir) // Clean up temp directory on exit

	// Determine archive format and file extension based on platform
	var archiveExt, binaryName string
	if runtime.GOOS == "windows" {
		archiveExt = ".zip"
		binaryName = "kat.exe"
	} else {
		archiveExt = ".tar.gz"
		binaryName = "kat"
	}

	// Create a temporary file for the archive
	tempArchivePath := filepath.Join(tempDir, "kat"+archiveExt)
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
	tempBinaryPath := filepath.Join(tempDir, binaryName)

	// Extract based on platform
	if runtime.GOOS == "windows" {
		err = extractZip(tempArchivePath, tempDir, binaryName)
	} else {
		err = extractTarGz(tempArchivePath, tempDir, binaryName)
	}
	if err != nil {
		return errors.Wrap(err, "failed to extract archive")
	}

	// Make the extracted binary executable (Unix only)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempBinaryPath, 0755); err != nil {
			return errors.Wrap(err, "failed to make binary executable")
		}
	}

	// Replace the binary with platform-specific handling
	execDir := filepath.Dir(execPath)
	execName := filepath.Base(execPath)

	if runtime.GOOS == "windows" {
		// On Windows, copy instead of rename to avoid cross-volume issues
		// and handle locked executable files
		return replaceWindowsBinary(tempBinaryPath, execPath, progressWriter)
	}

	// Unix-like systems: atomic rename
	newExecPath := filepath.Join(execDir, execName+".new")
	if err := os.Rename(tempBinaryPath, newExecPath); err != nil {
		return errors.Wrap(err, "failed to move new executable to destination directory")
	}

	// Replace the current executable with the new one
	if err := os.Rename(newExecPath, execPath); err != nil {
		return errors.Wrap(err, "failed to replace current executable")
	}

	fmt.Fprintf(progressWriter, "%sUpdate successfully installed%s\n", output.StyleSuccess, output.StyleReset)
	return nil
}

// extractZip extracts a specific file from a ZIP archive
func extractZip(archivePath, destDir, fileName string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return errors.Wrap(err, "failed to open ZIP archive")
	}
	defer reader.Close()

	// Find the target file in the archive
	for _, file := range reader.File {
		if file.Name == fileName || filepath.Base(file.Name) == fileName {
			// Open the file in the archive
			rc, err := file.Open()
			if err != nil {
				return errors.Wrap(err, "failed to open file in ZIP archive")
			}

			// Create destination file
			destPath := filepath.Join(destDir, fileName)
			destFile, err := os.Create(destPath)
			if err != nil {
				rc.Close()
				return errors.Wrap(err, "failed to create destination file")
			}

			// Copy the content
			_, err = io.Copy(destFile, rc)
			destFile.Close()
			rc.Close()

			if err != nil {
				return errors.Wrap(err, "failed to extract file from ZIP")
			}

			return nil
		}
	}

	return errors.Newf("file %s not found in ZIP archive", fileName)
}

// extractTarGz extracts a specific file from a tar.gz archive
func extractTarGz(archivePath, destDir, fileName string) error {
	// Open the tar.gz file
	file, err := os.Open(archivePath)
	if err != nil {
		return errors.Wrap(err, "failed to open tar.gz archive")
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return errors.Wrap(err, "failed to create gzip reader")
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Iterate through files in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to read tar archive")
		}

		// Check if this is the file we want
		if header.Name == fileName || filepath.Base(header.Name) == fileName {
			// Create destination file
			destPath := filepath.Join(destDir, fileName)
			destFile, err := os.Create(destPath)
			if err != nil {
				return errors.Wrap(err, "failed to create destination file")
			}

			// Copy the content
			_, err = io.Copy(destFile, tarReader)
			destFile.Close()

			if err != nil {
				return errors.Wrap(err, "failed to extract file from tar.gz")
			}

			return nil
		}
	}

	return errors.Newf("file %s not found in tar.gz archive", fileName)
}

// replaceWindowsBinary handles Windows-specific binary replacement
func replaceWindowsBinary(newBinaryPath, currentBinaryPath string, progressWriter io.Writer) error {
	// Copy the new binary to a .new file first using stream copy
	newPath := currentBinaryPath + ".new"
	
	// Stream copy instead of loading entire binary in memory
	src, err := os.Open(newBinaryPath)
	if err != nil {
		return errors.Wrap(err, "failed to open new binary")
	}
	defer src.Close()
	
	dst, err := os.Create(newPath)
	if err != nil {
		return errors.Wrap(err, "failed to create new binary file")
	}
	defer dst.Close()
	
	if _, err := io.Copy(dst, src); err != nil {
		return errors.Wrap(err, "failed to copy new binary")
	}
	dst.Close() // Close explicitly before attempting rename
	
	// Try atomic replacement with backup for safety
	backupPath := currentBinaryPath + ".bak"
	
	// First, try to backup the current binary
	if err := os.Rename(currentBinaryPath, backupPath); err == nil {
		// Backup successful, now try to install new version
		if err := os.Rename(newPath, currentBinaryPath); err == nil {
			// Success! Clean up backup
			os.Remove(backupPath)
			fmt.Fprintf(progressWriter, "%sUpdate successfully installed%s\n", output.StyleSuccess, output.StyleReset)
			return nil
		} else {
			// Installation failed, restore backup
			if restoreErr := os.Rename(backupPath, currentBinaryPath); restoreErr != nil {
				return errors.Wrapf(err, "update failed and could not restore backup (backup at %s): %v", backupPath, restoreErr)
			}
			return errors.Wrap(err, "update failed, original binary restored")
		}
	}
	
	// Fallback: leave the .new file and inform user (binary is likely running)
	fmt.Fprintf(progressWriter, "%sWarning: Could not replace running executable%s\n", output.StyleWarning, output.StyleReset)
	fmt.Fprintf(progressWriter, "%sNew version saved as: %s%s\n", output.StyleInfo, newPath, output.StyleReset)
	fmt.Fprintf(progressWriter, "%sRestart your terminal and run the command again to complete the update%s\n", output.StyleInfo, output.StyleReset)
	return nil
}
