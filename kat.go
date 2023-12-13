package kat

import "fmt"

// Up executes all migrations in the order they were created. It is safe to call this
// before your server starts when running Kat in an API.
func Up() {
	fmt.Println("performing an up migration")
}

func Down() {
	fmt.Println("performing a down migration")
}

func Undo() {
	fmt.Println("performing an undo migration")
}
