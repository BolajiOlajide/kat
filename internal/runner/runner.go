package runner

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/loggr"
	"github.com/BolajiOlajide/kat/internal/types"
)

// Runner is the interface that every runner must implement.
type Runner interface {
	Run(context.Context, Options) error
}

type runner struct {
	db     database.DB
	logger loggr.Logger
}

// executionDetails tracks information of a successful execution.
type executionDetails struct {
	Name      string
	Operation string
	Duration  time.Duration
}

var _ Runner = (*runner)(nil)

// NewRunner returns a new instance of the runner.
func NewRunner(ctx context.Context, db database.DB, logger loggr.Logger) (Runner, error) {
	if err := db.Ping(ctx); err != nil {
		return nil, err
	}
	return &runner{db: db, logger: logger}, nil
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
	defer rows.Close()

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
					sqlf.Sprintf("%d * interval '1 millisecond'", duration.Milliseconds()),
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

	// we use a topological sort to determine the correct sequence of execution
	// The sort is stable but depending on whether it's a up or down migration, we need to reverse both
	// the definitions and the sorting of elements with the same order.
	sortedDefs, err := options.Definitions.TopologicalSort()
	if err != nil {
		return err
	}

	if options.Operation.IsDownMigration() {
		slices.Reverse(sortedDefs)
	}

	var execs []executionDetails

	for _, hash := range sortedDefs {
		// We want to respect the count flag when it's provided, so we don't exceed the number
		// of migrations the user expects to be processed.
		if options.Count > 0 && len(execs) >= options.Count {
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
				r.logger.Info(fmt.Sprintf("[DRY RUN] Would execute %s migration for %q", options.Operation, definition.FileName()))

				// Add to successful migrations list for summary
				execs = append(execs, executionDetails{
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
			execs = append(execs, executionDetails{
				Name:      definition.FileName(),
				Operation: options.Operation.String(),
				Duration:  duration,
			})
			return nil
		}

		var execErr error
		if definition.NoTransaction {
			execErr = r.runNoTransaction(ctx, definition, options, logsMap, &execs)
		} else {
			// Execute within a transaction
			execErr = r.db.WithTransact(ctx, txFunc)
		}

		if execErr != nil {
			// Print detailed error information
			r.logger.Error(fmt.Sprintf("Migration failed: %s", definition.Name))
			r.logger.Error(fmt.Sprintf("Error: %s", execErr.Error()))
			r.logger.Error("Migration process stopped to preserve database integrity")

			return errors.Wrapf(execErr, "executing %s", definition.FileName())
		}
	}

	r.printMigrationSummary(execs, options.Operation, options.DryRun, options.Verbose)
	return nil
}

// runNoTransaction executes a migration without wrapping it in a transaction.
// The migration SQL runs in autocommit mode (required for operations like CREATE INDEX
// CONCURRENTLY), while the bookkeeping log update is wrapped in its own transaction
// to reduce the chance of "applied but not recorded" drift.
func (r *runner) runNoTransaction(ctx context.Context, definition types.Definition, options Options, logsMap map[string]*types.MigrationLog, execs *[]executionDetails) error {
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
		r.logger.Info(fmt.Sprintf("[DRY RUN] Would execute %s migration for %q (no transaction)", options.Operation, definition.FileName()))
		*execs = append(*execs, executionDetails{
			Name:      definition.FileName(),
			Operation: options.Operation.String(),
		})
		return nil
	}

	r.logger.Warn(fmt.Sprintf("Executing %q without a transaction; partial application is possible on failure", definition.FileName()))

	// Warn if the migration contains multiple statements
	if hasMultipleStatements(q) {
		r.logger.Warn(fmt.Sprintf("Migration %q contains multiple SQL statements; each will commit independently outside a transaction", definition.FileName()))
	}

	// Execute the migration SQL directly (autocommit mode)
	start := time.Now()
	if err := r.db.Exec(ctx, q); err != nil {
		return errors.Wrapf(err, "executing %s query", options.Operation)
	}
	duration := time.Since(start)

	// Record the migration log in a transaction for bookkeeping integrity
	if err := r.db.WithTransact(ctx, func(tx database.Tx) error {
		query, err := r.computePostExecutionQuery(definition.FileName(), options.MigrationInfo.TableName, duration, start, options.Operation)
		if err != nil {
			return err
		}
		return tx.Exec(ctx, query)
	}); err != nil {
		r.logger.Error(fmt.Sprintf("Migration SQL for %q executed successfully but failed to update migration log; you may need to update the record manually", definition.FileName()))
		return errors.Wrap(err, "updating migration log")
	}

	*execs = append(*execs, executionDetails{
		Name:      definition.FileName(),
		Operation: options.Operation.String(),
		Duration:  duration,
	})
	return nil
}

// hasMultipleStatements checks if a SQL query contains multiple semicolon-terminated statements.
func hasMultipleStatements(q *sqlf.Query) bool {
	rendered := strings.TrimSpace(q.Query(sqlf.PostgresBindVar))
	// Remove trailing semicolon so a single statement ending with ";" doesn't count
	rendered = strings.TrimRight(rendered, "; \t\n")
	return strings.Contains(rendered, ";")
}

// printMigrationSummary prints a summary of successful migrations
func (r *runner) printMigrationSummary(details []executionDetails, operation types.MigrationOperationType, dryRun, verbose bool) {
	var executionVerb = "apply"
	if operation.IsDownMigration() {
		executionVerb = "roll back"
	}

	if len(details) == 0 {
		r.logger.Info(fmt.Sprintf("No migration(s) to %s.", executionVerb))
		return
	}

	if verbose {
		// Print summary header
		r.logger.Info("Migration Summary")

		// Print list of successful migrations with duration
		if dryRun {
			r.logger.Info("Validated migrations: ")
		} else {
			r.logger.Info("Successful migrations: ")
		}

		for _, migration := range details {
			if dryRun {
				r.logger.Info(fmt.Sprintf("  ✓ %s (%s)", migration.Name, migration.Operation))
			} else {
				r.logger.Info(fmt.Sprintf("  ✓ %s (%s) - %s", migration.Name, migration.Operation, migration.Duration))
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

	r.logger.Info(fmt.Sprintf("Total: %d migration(s) %s.", len(details), operationName))
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
