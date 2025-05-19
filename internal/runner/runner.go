package runner

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/cockroachdb/errors"
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

func (r *runner) executeMigrationLogQuery(ctx context.Context, tblName string) error {
	createMigrationLogQuery, err := computeCreateMigrationLogQuery(tblName)
	if err != nil {
		return errors.Wrap(err, "compute migration log query")
	}

	// create migration log table if it doesn't exist. This action is idempotent.
	// No retry for migrations
	if err := r.db.Exec(ctx, sqlf.Sprintf(createMigrationLogQuery)); err != nil {
		return errors.Wrap(err, "initializing migration table")
	}

	return nil
}

func (r *runner) getAppliedMigrations(ctx context.Context, tblName string) (map[string]*types.MigrationLog, error) {
	migrationLogColumns := computeMigrationLogColumns(tblName)
	selectLogQuery, err := computeSelectMigrationLogQuery(tblName)
	if err != nil {
		return nil, errors.Wrap(err, "compute select log query")
	}

	// Get existing migration logs
	query := sqlf.Sprintf(
		selectLogQuery,
		sqlf.Join(migrationLogColumns, ", "),
	)
	rows, err := r.db.Query(ctx, query)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()

	var logsMap = make(map[string]*types.MigrationLog)
	for rows.Next() {
		l, err := scanMigrationLog(rows)
		if err != nil {
			return nil, err
		}
		logsMap[l.Name] = l
	}

	return logsMap, rows.Err()
}

func (r *runner) computePostExecutionQuery(fileName, tblName string, duration time.Duration, migrationStart time.Time, operation types.MigrationOperationType) (*sqlf.Query, error) {
	// For UP operations, insert a log entry
	// For DOWN operations, remove the log entry
	if operation.IsUpMigration() {
		insertLogQuery, err := computeInsertMigrationLogQuery(tblName)
		if err != nil {
			return nil, errors.Wrap(err, "compute insert log query")
		}

		migrationTime := migrationStart.Format("2006-01-02 15:04:05.999-07")
		return sqlf.Sprintf(
			insertLogQuery,
			sqlf.Join(migrationLogInsertColumns, ", "),
			sqlf.Join(
				[]*sqlf.Query{
					sqlf.Sprintf("%s", fileName),
					sqlf.Sprintf("%s", migrationTime),
					sqlf.Sprintf("%d * interval '1 microsecond'", duration),
				},
				", ",
			),
		), nil
	}

	// Delete the migration log entry for DOWN operations
	deleteLogQuery := sqlf.Sprintf(
		"DELETE FROM %s WHERE name = %s",
		sqlf.Sprintf(tblName),
		fileName,
	)
	return deleteLogQuery, nil
}

func (r *runner) Run(ctx context.Context, options Options) error {
	if err := r.executeMigrationLogQuery(ctx, options.MigrationInfo.TableName); err != nil {
		return err
	}

	logsMap, err := r.getAppliedMigrations(ctx, options.MigrationInfo.TableName)
	if err != nil {
		return err
	}

	var successfulMigrations []successfulMigration

	// we use a topological sort to determine the correct sequence of execution
	// The sort is stable but depending on whether it's a up or down migration, we need to reverse both
	// the definitions and the sorting of elements with the same order.
	sortedDefs, err := options.Definitions.TopologicalSort(options.Operation.IsUpMigration())
	if err != nil {
		return err
	}

	if options.Operation.IsDownMigration() {
		slices.Reverse(sortedDefs)
	}

	for _, hash := range sortedDefs {
		// We want to respect the count flag when it's provided, so we don't exceed the number
		// of migrations the user expects to be processed.
		if options.Count > 0 && options.Count >= len(successfulMigrations) {
			break
		}

		definition, err := options.Definitions.GetDefinition(hash)
		if err != nil {
			return err
		}

		// Use retry functionality for transaction if configured
		var txFunc = func(tx database.Tx) (err error) {
			var q *sqlf.Query
			if options.Operation.IsUpMigration() {
				// Skip migrations that have already been executed
				if _, exists := logsMap[definition.FileName()]; exists {
					return nil
				}
				q = definition.UpQuery
			} else {
				// Skip migrations that haven't been executed
				if _, exists := logsMap[definition.FileName()]; !exists {
					return nil
				}
				q = definition.DownQuery
			}

			// In dry-run mode, don't execute the SQL
			if options.DryRun {
				fmt.Printf("%s[DRY RUN] Would execute %s migration for %q%s\n",
					output.StyleInfo, options.Operation, definition.FileName(), output.StyleReset)

				// Add to successful migrations list for summary
				successfulMigrations = append(successfulMigrations, successfulMigration{
					Name:      definition.FileName(),
					Operation: options.Operation.String(),
				})

				return nil
			}

			start := time.Now()
			if err := tx.Exec(ctx, q); err != nil {
				return errors.Wrapf(err, "executing %s query", options.Operation)
			}
			duration := time.Since(start)

			query, err := r.computePostExecutionQuery(definition.FileName(), options.MigrationInfo.TableName, duration, start, options.Operation)
			if err := tx.Exec(ctx, query); err != nil {
				return err
			}

			// Add to successful migrations list for summary
			successfulMigrations = append(successfulMigrations, successfulMigration{
				Name:      definition.FileName(),
				Operation: options.Operation.String(),
				Duration:  duration,
			})
			return nil
		}

		// Execute transaction without retry
		if err := r.db.WithTransact(ctx, txFunc); err != nil {
			// Print detailed error information
			fmt.Printf("\n%sMigration failed: %s%s\n", output.StyleFailure, definition.Name, output.StyleReset)
			fmt.Printf("%sError details: %s%s\n", output.StyleFailure, err.Error(), output.StyleReset)
			fmt.Printf("%sMigration process stopped to preserve database integrity%s\n", output.StyleInfo, output.StyleReset)

			return errors.Wrapf(err, "executing %s", definition.FileName())
		}
	}

	printMigrationSummary(successfulMigrations, options.Operation, options.DryRun, options.Verbose)
	return nil
}

// printMigrationSummary prints a summary of successful migrations
func printMigrationSummary(migrations []successfulMigration, operation types.MigrationOperationType, dryRun, verbose bool) {
	var executionVerb = "apply"
	if operation.IsDownMigration() {
		executionVerb = "roll back"
	}

	if len(migrations) == 0 {
		fmt.Printf("%sNo migration(s) to %s.%s\n", output.StyleInfo, executionVerb, output.StyleReset)
		return
	}

	if verbose {
		// Print summary header
		fmt.Printf("\n%sMigration Summary%s\n\n", output.StyleHeading, output.StyleReset)

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
	var migrationLog types.MigrationLog
	return &migrationLog, sc.Scan(
		&migrationLog.ID,
		&migrationLog.Name,
		&migrationLog.MigrationTime,
		&migrationLog.Duration,
	)
}
