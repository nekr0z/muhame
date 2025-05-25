package main

import (
	"os"
)

func main() {
	foo()
	os.Exit(1) // want "os.Exit called directly in main func of the main package"
}

func foo() {
	os.Exit(2) // no diagnostics
}
