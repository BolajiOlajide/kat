package migration

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

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

	timestamp := time.Now().UTC().Unix()
	sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
		strings.ReplaceAll(strings.ToLower(name), " ", "_"), "",
	)
	migrationName := fmt.Sprintf("%d_%s", timestamp, sanitizedName)

	m := types.Migration{
		Up:        filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/up.sql", migrationName)),
		Down:      filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/down.sql", migrationName)),
		Metadata:  filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/metadata.yaml", migrationName)),
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
