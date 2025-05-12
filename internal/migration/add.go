package migration

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
)

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
func Add(c *cli.Context, name string) error {
	cfg, err := config.GetKatConfigFromCtx(c)
	if err != nil {
		return err
	}

	_, err = getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return errors.Wrap(err, "getting migrations")
	}

	timestamp := time.Now().UTC().Unix()
	sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
		strings.ReplaceAll(strings.ToLower(name), " ", "_"), "",
	)
	migrationDirName := fmt.Sprintf("%d_%s", timestamp, sanitizedName)

	m := types.Migration{
		Up:        filepath.Join(cfg.Migration.Directory, migrationDirName, "up.sql"),
		Down:      filepath.Join(cfg.Migration.Directory, migrationDirName, "down.sql"),
		Metadata:  filepath.Join(cfg.Migration.Directory, migrationDirName, "metadata.yaml"),
		Timestamp: timestamp,
	}

	md := types.MigrationMetadata{
		Name:        sanitizedName,
		Timestamp:   timestamp,
		Description: c.String("description"),
		Parents:     nil,
	}

	if err := saveMigration(m, md); err != nil {
		return err
	}

	fmt.Printf("%sMigration created successfully!%s\n", output.StyleSuccess, output.StyleReset)
	if cfg.Verbose {
		fmt.Printf("%sUp query file: %s%s\n", output.StyleInfo, m.Up, output.StyleReset)
		fmt.Printf("%sDown query file: %s%s\n", output.StyleInfo, m.Down, output.StyleReset)
		fmt.Printf("%sMetadata file: %s%s\n", output.StyleInfo, m.Metadata, output.StyleReset)
	}

	return nil
}
