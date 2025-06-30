// Package migration provides functionality for creating, managing, and executing
// database migrations. It handles the lifecycle of migrations including creation,
// dependency resolution, and execution in both up and down directions.
//
// The package supports:
//   - Creating new migration files with proper structure
//   - Managing migration dependencies through parent-child relationships
//   - Executing migrations in dependency order
//   - Rolling back migrations safely
//   - Tracking applied migrations in the database
package migration

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"

	"github.com/BolajiOlajide/kat/internal/types"
)

// FilePerm is the standard permission for migration files (readable by all, writable by owner)
const FilePerm = 0644

func saveMigration(m types.TemporaryMigrationInfo, metadata types.MigrationMetadata) (err error) {
	defer func() {
		if err != nil {
			// undo any changes to the fs on error. we don't care about the errors here.
			_ = os.Remove(m.Up)
			_ = os.Remove(m.Down)
			_ = os.Remove(m.Metadata)
		}
	}()

	// Create directory once for all files
	dir := filepath.Dir(m.Up)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create migration directory")
	}

	// Prepare all file contents
	upContent := []byte(upMigrationFileTemplate)
	downContent := []byte(downMigrationFileTemplate)
	metadataContent, err := yaml.Marshal(&metadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshal metadata")
	}

	// Write all files
	files := map[string][]byte{
		m.Up:       upContent,
		m.Down:     downContent,
		m.Metadata: metadataContent,
	}

	for path, content := range files {
		if err := os.WriteFile(path, content, os.FileMode(FilePerm)); err != nil {
			return errors.Wrapf(err, "failed to write %s", filepath.Base(path))
		}
	}

	return nil
}

var nonAlphaNumericOrUnderscore = regexp.MustCompile("[^a-z0-9_]+")
