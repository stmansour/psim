package main

import (
	"flag"
	"fmt"
	"os"
)

var app struct {
	filename string
}

func readCommandLineArgs() {
	fPtr := flag.String("f", "", "csv file name to open and process")
	flag.Parse()
	app.filename = *fPtr
}

func main() {
	readCommandLineArgs()
	if len(app.filename) == 0 {
		fmt.Printf("Please specify the csv filename.  Example:  -f xyz.csv")
		os.Exit(1)
	}
	DoIt(app.filename)
}
