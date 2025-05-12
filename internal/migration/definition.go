package migration

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sort"

	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"gopkg.in/yaml.v3"
)

func ComputeDefinitions(fs fs.FS) ([]types.Definition, error) {
	// Make sure the root directory exists. All migrations must be in a subdirectory.
	// Also using `http.FS` here because it's API is more robust than `fs.FS`.
	root, err := http.FS(fs).Open("/")
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	migrations, err := root.Readdir(0)
	if err != nil {
		return nil, err
	}

	definitions := make([]types.Definition, 0, len(migrations))
	for _, file := range migrations {
		if !file.IsDir() {
			// if this is not a directory, skip it
			continue
		}

		def, err := computeDefinition(fs, file.Name())
		if err != nil {
			return nil, errors.Wrap(err, "malformed migration definition")
		}

		definitions = append(definitions, def)
	}

	// We sort the definitions by their ID so that they are executed in the correct order.
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Timestamp < definitions[j].Timestamp })
	return definitions, nil
}

func computeDefinition(fs fs.FS, filename string) (types.Definition, error) {
	// Read all migration files in a single function call
	upQuery, downQuery, metadata, err := readMigrationFiles(fs, filename)
	if err != nil {
		return types.Definition{}, err
	}

	return populateDefinition(upQuery, downQuery, metadata)
}

func readMigrationFiles(fs fs.FS, dirname string) (*sqlf.Query, *sqlf.Query, []byte, error) {
	upFilename := fmt.Sprintf("%s/up.sql", dirname)
	downFilename := fmt.Sprintf("%s/down.sql", dirname)
	metadataFilename := fmt.Sprintf("%s/metadata.yaml", dirname)

	// Read up.sql file
	upFile, err := fs.Open(upFilename)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to open up.sql for migration %s", dirname)
	}
	defer upFile.Close()

	// Read down.sql file
	downFile, err := fs.Open(downFilename)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to open down.sql for migration %s", dirname)
	}
	defer downFile.Close()

	// Read metadata.yaml file
	metadataFile, err := fs.Open(metadataFilename)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to open metadata.yaml for migration %s", dirname)
	}
	defer metadataFile.Close()

	// Use buffered readers for all files
	upReader := bufio.NewReader(upFile)
	downReader := bufio.NewReader(downFile)
	metadataReader := bufio.NewReader(metadataFile)

	// Read file contents
	upContent, err := io.ReadAll(upReader)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read up.sql for migration %s", dirname)
	}

	downContent, err := io.ReadAll(downReader)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read down.sql for migration %s", dirname)
	}

	metadata, err := io.ReadAll(metadataReader)
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
