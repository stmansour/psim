package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"
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
	d1, err = time.Parse("1/02/2006", records[1][0])
	if err != nil {
		fmt.Println("Error parsing date:", err)
		os.Exit(1)
	}
	d2, err = time.Parse("1/02/2006", records[len(records)-1][0])
	if err != nil {
		fmt.Println("Error parsing date:", err)
		os.Exit(1)
	}
	d2 = d2.AddDate(0, 0, 1)

	fmt.Printf("Start date: %s\n", d1.Format("01/02/2006"))
	fmt.Printf("Stop date: %s\n", d2.Format("01/02/2006"))

	//---------------------------------------------------------------
	// Start by printing out the column headers...
	//---------------------------------------------------------------
	for i := 0; i < len(cols); i++ {
		s := cols[i]
		if i < len(cols)-2 {
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
	lastRow := 1
	i := 1
	for dt := d1; dt.Before(d2); dt = dt.Add(24 * time.Hour) {
		// compare dt to the date in record[i]
		d1, err = time.Parse("1/02/2006", records[i][0])
		if err != nil {
			fmt.Println("Error parsing date:", err)
			os.Exit(1)
		}
		fmt.Printf("%s", dt.Format("1/02/2006"))
		idx := i // assume this is a new row
		if dt.After(d1) {
			idx = lastRow // adjust, fill using value from previous record
		} else {
			lastRow = i // mark that we are on a new row
		}
		//-------------------------------------
		// print all columns after the date...
		//-------------------------------------
		for j := 1; j < len(records[idx]); j++ {
			fmt.Printf(",%s", records[lastRow][j])
		}
		fmt.Printf("\n")
	}
}
