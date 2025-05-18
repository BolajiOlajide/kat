package migration

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"gopkg.in/yaml.v3"

	"github.com/BolajiOlajide/kat/internal/types"
)

// extractMigrationFiles reads the root directory of the provided filesystem and returns
// a list of FileInfo objects for all entries. It serves as the first step in migration discovery,
// returning all potential migration directories that will be later filtered and processed.
//
// The function uses http.FS for more robust filesystem operations and ensures the root
// directory exists before attempting to read its contents.
func extractMigrationFiles(f fs.FS) ([]fs.FileInfo, error) {
	// Make sure the root directory exists. All migrations must be in a subdirectory.
	// Also using `http.FS` here because it's API is more robust than `fs.FS`.
	root, err := http.FS(f).Open("/")
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	fis, err := root.Readdir(0)
	if err != nil {
		return nil, err
	}

	// Sort files alphabetically by name to ensure consistent, deterministic ordering
	// across different operating systems and filesystems. This is necessary because
	// the order returned by Readdir is filesystem-dependent and not guaranteed to be consistent.
	sort.Slice(fis, func(i, j int) bool {
		return fis[i].Name() < fis[j].Name()
	})
	return fis, nil
}

func computeDefinition(fs fs.FS, filename string) (types.Definition, error) {
	// Read all migration files in a single function call
	upQuery, downQuery, metadata, err := readMigrationFiles(fs, filename)
	if err != nil {
		return types.Definition{}, err
	}

	return populateDefinition(upQuery, downQuery, metadata)
}

func readMigrationFiles(f fs.FS, dirname string) (*sqlf.Query, *sqlf.Query, []byte, error) {
	upFilename := fmt.Sprintf("%s/up.sql", dirname)
	downFilename := fmt.Sprintf("%s/down.sql", dirname)
	metadataFilename := fmt.Sprintf("%s/metadata.yaml", dirname)

	// Read up.sql file
	upFile, err := f.Open(upFilename)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to open up.sql for migration %s", dirname)
	}
	defer upFile.Close()

	// Read down.sql file
	downFile, err := f.Open(downFilename)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to open down.sql for migration %s", dirname)
	}
	defer downFile.Close()

	// Read metadata.yaml file
	metadataFile, err := f.Open(metadataFilename)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to open metadata.yaml for migration %s", dirname)
	}
	defer metadataFile.Close()

	// Read file contents
	upContent, err := readFile(upFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read up.sql for migration %s", dirname)
	}

	downContent, err := readFile(downFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read down.sql for migration %s", dirname)
	}

	metadata, err := readFile(metadataFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read metadata.yaml for migration %s", dirname)
	}

	// Convert SQL content to queries
	upQuery := queryFromString(string(upContent))
	downQuery := queryFromString(string(downContent))

	return upQuery, downQuery, metadata, nil
}

func populateDefinition(upQuery, downQuery *sqlf.Query, metadata []byte) (types.Definition, error) {
	var payload types.MigrationMetadata
	if err := yaml.Unmarshal(metadata, &payload); err != nil {
		return types.Definition{}, err
	}

	return types.Definition{
		UpQuery:           upQuery,
		DownQuery:         downQuery,
		MigrationMetadata: payload,
	}, nil
}

func readFile(fi fs.File) ([]byte, error) {
	stat, err := fi.Stat()
	if err != nil {
		return nil, err
	}

	size := stat.Size()
	buf := make([]byte, size)
	if _, err := io.ReadFull(fi, buf); err != nil {
		return nil, err
	}

	return buf, nil
}
