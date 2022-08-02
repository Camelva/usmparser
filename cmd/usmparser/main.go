package main

import (
	parser "USMparser"
	"fmt"
	"os"
)

func main() {
	args := os.Args

	if len(args) < 3 {
		displayHelp()
	}

	parser.DumpFile(args[1], args[2])
}

func displayHelp() {
	fmt.Println("Usage: \n" +
		"\tusmparse input output\n" +
		"\n\n" +
		"If output doesn't exist - it will be created")
	os.Exit(0)
}
