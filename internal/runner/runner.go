package runner

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dominikbraun/graph"
	"github.com/keegancsmith/sqlf"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Runner is the interface that every runner must implement.
type Runner interface {
	Run(context.Context, Options) error
}

type runner struct {
	db database.DB
}

// successfulMigration tracks information about a successfully executed migration
type successfulMigration struct {
	Name      string
	Operation string
	Duration  time.Duration
}

var _ Runner = (*runner)(nil)

// NewRunner returns a new instance of the runner.
func NewRunner(ctx context.Context, db database.DB) (Runner, error) {
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}
	return &runner{db: db}, nil
}

func (r *runner) Run(ctx context.Context, options Options) error {
	createMigrationLogQuery, err := computeCreateMigrationLogQuery(options.MigrationInfo.TableName)
	if err != nil {
		return errors.Wrap(err, "compute migration log query")
	}

	// create migration log table if it doesn't exist. This action is idempotent.
	// No retry for migrations
	if err := r.db.Exec(ctx, sqlf.Sprintf(createMigrationLogQuery)); err != nil {
		return errors.Wrap(err, "initializing migration table")
	}

	mcols := computeMigrationLogColumns(options.MigrationInfo.TableName)
	selectLogQuery, err := computeSelectMigrationLogQuery(options.MigrationInfo.TableName)
	if err != nil {
		return errors.Wrap(err, "compute select log query")
	}

	insertLogQuery, err := computeInsertMigrationLogQuery(options.MigrationInfo.TableName)
	if err != nil {
		return errors.Wrap(err, "compute insert log query")
	}

	// Get existing migration logs
	query := sqlf.Sprintf(
		selectLogQuery,
		sqlf.Join(mcols, ", "),
	)
	rows, err := r.db.Query(ctx, query)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, "scanning log")
	}
	defer rows.Close()

	var logsMap = map[string]*types.MigrationLog{}
	for rows.Next() {
		log, err := scanMigrationLog(rows)
		if err != nil {
			return err
		}
		logsMap[log.Name] = log
	}

	var noOfMigrations int
	var successfulMigrations []successfulMigration

	// we use a topological sort to determine the correct sequence of execution
	sortedDefs, err := graph.TopologicalSort(options.Definitions)
	if err != nil {
		return errors.Wrap(err, "sorting definitions")
	}

	for _, hash := range sortedDefs {
		definition, err := options.Definitions.Vertex(hash)
		if err != nil {
			return err
		}

		// Use retry functionality for transaction if configured
		var txFunc = func(tx database.Tx) (err error) {
			if options.Operation == types.UpMigrationOperation {
				// Skip migrations that have already been executed
				if logsMap[definition.Name] != nil {
					return nil
				}
			} else if options.Operation == types.DownMigrationOperation {
				// Skip migrations that haven't been executed
				if logsMap[definition.Name] == nil {
					return nil
				}
			}

			noOfMigrations++
			fmt.Printf("%s%s%s ", output.StyleInfo, definition.Name, output.StyleReset)

			var q *sqlf.Query
			var migrationKind string
			if options.Operation == types.UpMigrationOperation {
				q = definition.UpQuery
				migrationKind = "up"
			} else {
				q = definition.DownQuery
				migrationKind = "down"
			}

			// In dry-run mode, don't execute the SQL
			if options.DryRun {
				fmt.Printf("%s[DRY RUN] Would execute %s migration for %s%s\n",
					output.StyleInfo, migrationKind, definition.Name, output.StyleReset)

				// Add to successful migrations list for summary
				successfulMigrations = append(successfulMigrations, successfulMigration{
					Name:      definition.Name,
					Operation: migrationKind,
				})

				return nil
			}

			start := time.Now()
			if err := tx.Exec(ctx, q); err != nil {
				return errors.Wrapf(err, "executing %s query", migrationKind)
			}
			duration := time.Since(start)

			// For UP operations, insert a log entry
			// For DOWN operations, remove the log entry
			if options.Operation == types.UpMigrationOperation {
				migrationTime := start.Format("2006-01-02 15:04:05.999-07")
				insertQuery := sqlf.Sprintf(
					insertLogQuery,
					sqlf.Join(migrationLogInsertColumns, ", "),
					sqlf.Join(
						[]*sqlf.Query{
							sqlf.Sprintf("%s", definition.Name),
							sqlf.Sprintf("%s", migrationTime),
							sqlf.Sprintf("%d * interval '1 microsecond'", duration),
						},
						", ",
					),
				)
				err = tx.Exec(ctx, insertQuery)
				if err != nil {
					return errors.Wrap(err, "inserting log entry")
				}
			} else {
				// Delete the migration log entry for DOWN operations
				deleteQuery := sqlf.Sprintf(
					"DELETE FROM %s WHERE name = %s",
					sqlf.Sprintf(options.MigrationInfo.TableName),
					definition.Name,
				)
				err = tx.Exec(ctx, deleteQuery)
				if err != nil {
					return errors.Wrap(err, "deleting log entry")
				}
			}

			// Add to successful migrations list for summary
			successfulMigrations = append(successfulMigrations, successfulMigration{
				Name:      definition.Name,
				Operation: migrationKind,
				Duration:  duration,
			})

			// add a new line incase there's an error
			fmt.Print("\n")
			return nil
		}

		// Execute transaction without retry
		if err := r.db.WithTransact(ctx, txFunc); err != nil {
			// Print detailed error information
			fmt.Printf("\n%sMigration failed: %s%s\n", output.StyleFailure, definition.Name, output.StyleReset)
			fmt.Printf("%sError details: %s%s\n", output.StyleFailure, err.Error(), output.StyleReset)
			fmt.Printf("%sMigration process stopped to preserve database integrity%s\n", output.StyleInfo, output.StyleReset)

			return errors.Wrapf(err, "executing %s", definition.Name)
		}
	}

	if noOfMigrations > 0 {
		// Print basic summary line
		if options.DryRun {
			if options.Operation == types.UpMigrationOperation {
				fmt.Printf("%sDRY RUN: Validated %d migration(s) without applying them%s\n", output.StyleInfo, noOfMigrations, output.StyleReset)
			} else {
				fmt.Printf("%sDRY RUN: Validated %d migration(s) without rolling them back%s\n", output.StyleInfo, noOfMigrations, output.StyleReset)
			}
		} else {
			if options.Operation == types.UpMigrationOperation {
				fmt.Printf("%sSuccessfully applied %d migration(s)%s\n", output.StyleInfo, noOfMigrations, output.StyleReset)
			} else {
				fmt.Printf("%sSuccessfully rolled back %d migration(s)%s\n", output.StyleInfo, noOfMigrations, output.StyleReset)
			}
		}

		// Print detailed migration summary if verbose mode is enabled
		if options.Verbose {
			printMigrationSummary(successfulMigrations, options.Operation, options.DryRun)
		}
	} else {
		if options.Operation == types.UpMigrationOperation {
			fmt.Printf("%sNo new migration(s) to apply%s\n", output.StyleInfo, output.StyleReset)
		} else {
			fmt.Printf("%sNo migration(s) to roll back%s\n", output.StyleInfo, output.StyleReset)
		}
	}
	return nil
}

// printMigrationSummary prints a summary of successful migrations
func printMigrationSummary(migrations []successfulMigration, operation types.MigrationOperationType, dryRun bool) {
	if len(migrations) == 0 {
		return
	}

	// Print summary header
	fmt.Printf("\n%sMigration Summary%s\n", output.StyleHeading, output.StyleReset)

	// Print list of successful migrations with duration
	if dryRun {
		fmt.Printf("%sValidated migrations:%s\n", output.StyleSuccess, output.StyleReset)
	} else {
		fmt.Printf("%sSuccessful migrations:%s\n", output.StyleSuccess, output.StyleReset)
	}

	for _, migration := range migrations {
		if dryRun {
			fmt.Printf("  %s✓ %s (%s)%s\n",
				output.StyleSuccess, migration.Name, migration.Operation, output.StyleReset)
		} else {
			fmt.Printf("  %s✓ %s (%s) - %s%s\n",
				output.StyleSuccess, migration.Name, migration.Operation, migration.Duration, output.StyleReset)
		}
	}

	// Print total count
	operationName := "applied"
	if operation == types.DownMigrationOperation {
		operationName = "rolled back"
	}
	if dryRun {
		operationName = "validated"
	}

	fmt.Printf("\n%sTotal: %d migration(s) %s%s\n",
		output.StyleInfo, len(migrations), operationName, output.StyleReset)
}

func scanMigrationLog(sc database.Scanner) (*types.MigrationLog, error) {
	var mlog types.MigrationLog
	if err := sc.Scan(
		&mlog.ID,
		&mlog.Name,
		&mlog.MigrationTime,
		&mlog.Duration,
	); err != nil {
		return nil, err
	}

	return &mlog, nil
}
