package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Migration represents a migration definition.
type Migration struct {
	Up        string
	Down      string
	Metadata  string
	Timestamp int64
}

func saveMigration(m Migration, name string) (err error) {
	defer func() {
		if err != nil {
			// undo any changes to the fs on error
			_ = os.Remove(m.Up)
			_ = os.Remove(m.Down)
			_ = os.Remove(m.Metadata)
		}
	}()

	// write the up.sql file
	if err := os.MkdirAll(filepath.Dir(m.Up), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(m.Up, []byte(upMigrationFileTemplate), os.FileMode(0644)); err != nil {
		return err
	}

	// write the down.sql file
	if err := os.MkdirAll(filepath.Dir(m.Down), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(m.Down, []byte(downMigrationFileTemplate), os.FileMode(0644)); err != nil {
		return err
	}

	// write the metadata.yaml file
	if err := os.MkdirAll(filepath.Dir(m.Metadata), 0755); err != nil {
		return err
	}
	metadata := fmt.Sprintf(metadataFileTemplate, name, m.Timestamp)
	if err := os.WriteFile(m.Metadata, []byte(metadata), os.FileMode(0644)); err != nil {
		return err
	}

	return nil
}

var nonAlphaNumericOrUnderscore = regexp.MustCompile("[^a-z0-9_]+")
