package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/stmansour/psim/util"
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
		fmt.Printf("You must supply the filename with the -f option\n")
		os.Exit(1)
	}
	// Open the CSV file
	file, err := os.Open(app.filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV file:", err)
		return
	}

	//---------------------------------------------------------------
	// look at the column headers, make sure we have what we need...
	//---------------------------------------------------------------
	var d1, d2 time.Time
	cols := records[0]
	if cols[0] != "Date" {
		fmt.Printf("*ERROR* the first column of the csv file must be Date\n")
		os.Exit(1)
	}
	d1, err = util.StringToDate(records[1][0])
	if err != nil {
		fmt.Println("Error parsing date:", err)
		os.Exit(1)
	}
	d2, err = util.StringToDate(records[len(records)-1][0])
	if err != nil {
		fmt.Println("Error parsing date:", err)
		os.Exit(1)
	}
	d2 = d2.AddDate(0, 0, 1)

	// fmt.Printf("Start date: %s\n", d1.Format("01/02/2006"))
	// fmt.Printf("Stop date: %s\n", d3.Format("01/02/2006"))

	//---------------------------------------------------------------
	// Start by printing out the column headers...
	//---------------------------------------------------------------
	for i := 0; i < len(cols); i++ {
		s := cols[i]
		if i < len(cols)-1 {
			s += ","
		}
		fmt.Printf("%s", s)
	}
	fmt.Printf("\n")

	// We want the CSV file to contain values by day. Some statistics
	// are published monthly, quarterly, yearly.  When this occurs
	// just repeat values to fill in the dates.  For example, if
	// we have a statistic that is published once a month, and the
	// first value we have is for March 1, then we use the March 1
	// value for every day of March.
	//---------------------------------------------------------------
	printFromThisRecord := 1
	for dtLoop := d1; dtLoop.Before(d2); dtLoop = dtLoop.Add(24 * time.Hour) {
		fmt.Printf("%s", dtLoop.Format("1/02/2006")) // This is the date we're on now

		// what's the next date in the records...
		dtNextRecord, err := util.StringToDate(records[printFromThisRecord+1][0])
		if err != nil {
			fmt.Println("Error parsing date:", err)
			os.Exit(1)
		}

		// has our loop date (dtLoop) reached the date of the next record (dtNextRecord)?
		if dtLoop.Equal(dtNextRecord) || dtLoop.After(dtNextRecord) { // has the current loop date reached or passed that of the record?
			printFromThisRecord++ // we're now going to be printing THIS record until the loop date reaches or passes this row's date
		}

		// print all columns after the date...
		for j := 1; j < len(records[printFromThisRecord]); j++ {
			fmt.Printf(",%s", records[printFromThisRecord][j])
		}
		fmt.Printf("\n")
	}

}
