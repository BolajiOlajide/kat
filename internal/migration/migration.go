package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BolajiOlajide/kat/internal/types"
)

// FilePerm is the standard permission for migration files (readable by all, writable by owner)
const FilePerm = 0644

func saveMigration(m types.Migration, name string, parents []int64) (err error) {
	defer func() {
		if err != nil {
			// undo any changes to the fs on error. we don't care about the errors here.
			_ = os.Remove(m.Up)
			_ = os.Remove(m.Down)
			_ = os.Remove(m.Metadata)
		}
	}()

	// write the up.sql file
	if err := os.MkdirAll(filepath.Dir(m.Up), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(m.Up, []byte(upMigrationFileTemplate), os.FileMode(FilePerm)); err != nil {
		return err
	}

	// write the down.sql file
	if err := os.MkdirAll(filepath.Dir(m.Down), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(m.Down, []byte(downMigrationFileTemplate), os.FileMode(FilePerm)); err != nil {
		return err
	}

	// Format parents as YAML list
	parentsYAML := formatParentsAsYAML(parents)

	// write the metadata.yaml file
	if err := os.MkdirAll(filepath.Dir(m.Metadata), os.ModePerm); err != nil {
		return err
	}
	metadata := fmt.Sprintf(metadataFileTemplate, name, m.Timestamp, parentsYAML)
	if err := os.WriteFile(m.Metadata, []byte(metadata), os.FileMode(FilePerm)); err != nil {
		return err
	}

	return nil
}

// formatParentsAsYAML converts a slice of parent migration timestamps to YAML inline array format
func formatParentsAsYAML(parents []int64) string {
	if len(parents) == 0 {
		return "[]"
	}

	var result strings.Builder
	result.WriteString("[")
	for i, parent := range parents {
		if i > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("%d", parent))
	}
	result.WriteString("]")
	return result.String()
}

var nonAlphaNumericOrUnderscore = regexp.MustCompile("[^a-z0-9_]+")
