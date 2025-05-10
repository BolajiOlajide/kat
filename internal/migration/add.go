package migration

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
func Add(name string, cfg types.Config) error {
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

	// Get the current parent migrations from the database
	parents, err := getCurrentParentMigrations(cfg)
	if err != nil {
		return errors.Wrap(err, "getting current parent migrations")
	}

	err = saveMigration(m, migrationName, parents)
	if err != nil {
		return err
	}

	fmt.Printf("%sMigration created successfully!%s\n", output.StyleSuccess, output.StyleReset)
	if cfg.Verbose {
		fmt.Printf("%sUp query file: %s%s\n", output.StyleInfo, m.Up, output.StyleReset)
		fmt.Printf("%sDown query file: %s%s\n", output.StyleInfo, m.Down, output.StyleReset)
		fmt.Printf("%sMetadata file: %s%s\n", output.StyleInfo, m.Metadata, output.StyleReset)
		
		if len(parents) > 0 {
			fmt.Printf("%sParent migrations:%s\n", output.StyleInfo, output.StyleReset)
			for _, parent := range parents {
				fmt.Printf("%s  - %s%s\n", output.StyleInfo, parent, output.StyleReset)
			}
		} else {
			fmt.Printf("%sNo parent migrations (first migration)%s\n", output.StyleInfo, output.StyleReset)
		}
	}

	return nil
}

// getCurrentParentMigrations retrieves the list of already applied migrations to use as parents
func getCurrentParentMigrations(cfg types.Config) ([]int64, error) {
	// Connect to the database
	dbConn, err := cfg.Database.ConnString()
	if err != nil {
		return nil, errors.Wrap(err, "getting database connection string")
	}

	db, err := database.New(dbConn)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to database")
	}
	defer db.Close()

	ctx := context.Background()
	
	// Check if migration table exists
	query := sqlf.Sprintf(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = %s)",
		cfg.Migration.TableName,
	)
	
	var tableExists bool
	row := db.QueryRow(ctx, query)
	if err := row.Scan(&tableExists); err != nil {
		// If we can't determine if the table exists, assume no parents
		return []int64{}, nil
	}
	
	if !tableExists {
		// If migration table doesn't exist yet, this is the first migration
		return []int64{}, nil
	}

	// Query the existing migrations - extract timestamps from names
	selectQuery := sqlf.Sprintf(
		"SELECT name FROM %s ORDER BY migration_time DESC",
		sqlf.Sprintf(cfg.Migration.TableName),
	)

	rows, err := db.Query(ctx, selectQuery)
	if err != nil {
		return nil, errors.Wrap(err, "querying migrations")
	}
	defer rows.Close()

	// Collect the timestamps from migration names
	// Migration name format is "TIMESTAMP_name"
	var parents []int64
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, errors.Wrap(err, "scanning migration name")
		}

		// Extract timestamp from the migration name
		timestampStr := strings.Split(name, "_")[0] 
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parsing timestamp from migration name")
		}
		parents = append(parents, timestamp)
	}

	return parents, nil
}
