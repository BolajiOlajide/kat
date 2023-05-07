package runner

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/BolajiOlajide/kat/internal/database"
	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
)

// Runner is the interface that every runner must implement.
type Runner interface {
	Run(context.Context, Options) error
}

type runner struct {
	db database.DB
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
	if err = r.db.Exec(ctx, sqlf.Sprintf(createMigrationLogQuery)); err != nil {
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
	for _, definition := range options.Definitions {
		err := r.db.WithTransact(ctx, func(tx database.Tx) (err error) {
			// this means this migration has already been executed
			if logsMap[definition.Name] != nil {
				return nil
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

			err = r.db.Exec(ctx, q)
			if err != nil {
				return errors.Wrapf(err, "executing %s query", migrationKind)
			}

			t := time.Now()
			migrationTime := t.Format("2006-01-02 15:04:05.999-07")
			insertQuery := sqlf.Sprintf(
				insertLogQuery,
				sqlf.Join(migrationLogInsertColumns, ", "),
				sqlf.Join(
					[]*sqlf.Query{
						sqlf.Sprintf("%s", definition.Name),
						sqlf.Sprintf("%s", migrationTime),
					},
					", ",
				),
			)
			err = tx.Exec(ctx, insertQuery)
			if err != nil {
				return errors.Wrap(err, "inserting log entry")
			}

			// add a new line incase there's an error
			fmt.Print("\n")
			return nil
		})
		if err != nil {
			return errors.Wrapf(err, "executing %s", definition.Name)
		}
	}

	if noOfMigrations > 0 {
		fmt.Printf("%sSuccessfully executed %d migrations%s\n", output.StyleInfo, noOfMigrations, output.StyleReset)
	} else {
		fmt.Printf("%sNo new migrations%s\n", output.StyleInfo, output.StyleReset)
	}
	return nil
}

func scanMigrationLog(sc database.Scanner) (*types.MigrationLog, error) {
	var mlog types.MigrationLog
	if err := sc.Scan(
		&mlog.ID,
		&mlog.Name,
		&mlog.MigrationTime,
	); err != nil {
		return nil, err
	}

	return &mlog, nil
}
