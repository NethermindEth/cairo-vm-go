package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cairo-vm <path_to_file>")
		os.Exit(1)
	}

	path := os.Args[1]

	runCairoFile(path)
}

func runCairoFile(path string) {
	fmt.Printf("Running Cairo file at path: %s\n", path)
}
