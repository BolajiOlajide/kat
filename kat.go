package kat

import "fmt"

func Up() {
	fmt.Println("performing an up migration")
}

func Down() {
	fmt.Println("performing a down migration")
}

func Undo() {
	fmt.Println("performing an undo migration")
}
