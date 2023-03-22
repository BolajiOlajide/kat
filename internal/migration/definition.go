package migration

import (
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

func computeDefinitions(fs fs.FS) ([]types.Definition, error) {
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

		definition, err := computeDefinition(fs, file.Name())
		if err != nil {
			return nil, errors.Wrap(err, "malformed migration definition")
		}

		definitions = append(definitions, definition)
	}

	// We sort the definitions by their ID so that they are executed in the correct order.
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })
	return definitions, nil
}

func computeDefinition(fs fs.FS, filename string) (types.Definition, error) {
	upFilename := fmt.Sprintf("%s/up.sql", filename)
	downFilename := fmt.Sprintf("%s/down.sql", filename)
	metadataFilename := fmt.Sprintf("%s/metadata.yaml", filename)

	upQuery, err := readQueryFromFile(fs, upFilename)
	if err != nil {
		return types.Definition{}, err
	}

	downQuery, err := readQueryFromFile(fs, downFilename)
	if err != nil {
		return types.Definition{}, err
	}

	metadata, err := readFile(fs, metadataFilename)
	if err != nil {
		return types.Definition{}, err
	}

	return populateDefinition(upQuery, downQuery, metadata)
}

func populateDefinition(upQuery, downQuery *sqlf.Query, metadata []byte) (types.Definition, error) {
	var payload struct {
		Name      string `yaml:"name"`
		Timestamp int    `yaml:"timestamp"`
	}
	if err := yaml.Unmarshal(metadata, &payload); err != nil {
		return types.Definition{}, err
	}

	var definition types.Definition

	definition.UpQuery = upQuery
	definition.DownQuery = downQuery
	definition.Name = payload.Name
	definition.ID = payload.Timestamp

	return definition, nil
}

func readFile(fs fs.FS, filepath string) ([]byte, error) {
	file, err := fs.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func readQueryFromFile(fs fs.FS, filepath string) (*sqlf.Query, error) {
	file, err := readFile(fs, filepath)
	if err != nil {
		return nil, err
	}

	return queryFromString(string(file)), nil
}
