package migration

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"

	"github.com/BolajiOlajide/kat/internal/config"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
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

	m := types.TemporaryMigrationInfo{
		Up:        filepath.Join(cfg.Migration.Directory, migrationDirName, "up.sql"),
		Down:      filepath.Join(cfg.Migration.Directory, migrationDirName, "down.sql"),
		Metadata:  filepath.Join(cfg.Migration.Directory, migrationDirName, "metadata.yaml"),
		Timestamp: timestamp,
	}

	f, err := getMigrationsFS(cfg.Migration.Directory)
	if err != nil {
		return err
	}

	defs, err := ComputeDefinitions(f)
	if err != nil {
		return err
	}

	leaves, err := defs.Leaves()
	if err != nil {
		return err
	}

	md := types.MigrationMetadata{
		Name:        sanitizedName,
		Timestamp:   timestamp,
		Description: c.String("description"),
		Parents:     leaves,
	}

	if err := saveMigration(m, md); err != nil {
		return err
	}

	output.Default.Success("Migration created successfully!")
	if cfg.Verbose {
		output.Default.Infof("Up query file: %s", m.Up)
		output.Default.Infof("Down query file: %s", m.Down)
		output.Default.Infof("Metadata file: %s", m.Metadata)
	}

	return nil
}
