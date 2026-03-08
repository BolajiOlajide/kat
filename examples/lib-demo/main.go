package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/BolajiOlajide/kat"
)

type matrix struct {
	driver    kat.Driver
	pathOrUrl string
}

func main() {
	ctx := context.Background()

	// Define the migration directory path
	migrationDir := "./migrations"
	// Check if the migration directory exists
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		log.Fatalf("Migration directory does not exist: %s", migrationDir)
	}
	// Create a fs.FS value representing the migration directory
	migrationFS := os.DirFS(migrationDir)

	matrixesssssss := []matrix{
		{
			driver:    kat.PostgresDriver,
			pathOrUrl: "postgres://bolaji:andela@localhost:5432/kat3",
		},
		{
			driver:    kat.SQLiteDriver,
			pathOrUrl: "./kat3.db",
		},
	}

	for _, m := range matrixesssssss {
		fmt.Printf("executing demo for %s\n", m.driver)
		db, err := sql.Open(m.driver.DriverName(), m.pathOrUrl)
		if err != nil {
			panic(err)
		}
		if m.driver == kat.SQLiteDriver {
			db.SetMaxOpenConns(1)
		}
		km, err := kat.NewWithDB(m.driver, db, migrationFS, "logs")
		if err != nil {
			panic(err)
		}

		err = km.Up(ctx, 0)
		if err != nil {
			panic(err)
		}

		km.Close()
		fmt.Printf("hello %v\n\n", m.driver)
	}

}
