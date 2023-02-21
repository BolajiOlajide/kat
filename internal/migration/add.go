package migration

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BolajiOlajide/kat/internal/output"
)

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
func Add(name string) error {
	migrationsBaseDir, err := getMigrationsPath()
	if err != nil {
		return err
	}

	timestamp := time.Now().UTC().Unix()
	sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
		strings.ReplaceAll(strings.ToLower(name), " ", "_"), "",
	)
	migrationName := fmt.Sprintf("%d_%s", timestamp, sanitizedName)

	m := Migration{
		Up:        filepath.Join(migrationsBaseDir, fmt.Sprintf("%s/up.sql", migrationName)),
		Down:      filepath.Join(migrationsBaseDir, fmt.Sprintf("%s/down.sql", migrationName)),
		Metadata:  filepath.Join(migrationsBaseDir, fmt.Sprintf("%s/metadata.yaml", migrationName)),
		Timestamp: timestamp,
	}
	err = saveMigration(m, name)
	if err != nil {
		return err
	}

	fmt.Printf("%sMigration created successfully!%s\n", output.StyleSuccess, output.StyleReset)
	fmt.Printf("%sUp query file: %s%s\n", output.StyleInfo, m.Up, output.StyleReset)
	fmt.Printf("%sDown query file: %s%s\n", output.StyleInfo, m.Down, output.StyleReset)
	fmt.Printf("%sMetadata file: %s%s\n", output.StyleInfo, m.Metadata, output.StyleReset)
	return nil
}
