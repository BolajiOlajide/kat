package migration

import (
	"github.com/urfave/cli/v2"
)

// Up is the command that runs the up migration operation.
func Up(c *cli.Context) error {
	// ctx := c.Context
	// path, err := getMigrationsPath()
	// if err != nil {
	// 	return err
	// }

	// fs, err := getMigrationsFS(path)
	// if err != nil {
	// 	return err
	// }

	// definitions, err := computeDefinitions(fs)
	// if err != nil {
	// 	return err
	// }

	// dbConn, err := types.Database.ConnString()
	// if err != nil {
	// 	return err
	// }

	// db, err := database.NewDB(dbConn)
	// if err != nil {
	// 	return err
	// }

	// r := runner.NewRunner(db)
	// err = r.Run(ctx, runner.Options{
	// 	Operation:   types.MigrationOperationTypeUpgrade,
	// 	Definitions: definitions,
	// })
	// if err != nil {
	// 	return errors.Wrap(err, "running up migration command")
	// }

	return nil
}
