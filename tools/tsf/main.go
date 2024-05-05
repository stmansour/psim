package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/stmansour/psim/util"
)

// tsf - Time Series Filler is designed to process a CSV file containing
// time-series data, where each row represents a single day's metrics.
// The program's primary goal is to generate an updated version of the
// input CSV file, ensuring that there is one row for every date, even if
// the input file is missing data for certain dates.

// Here's a brief description of the program's functionality:

// Input File Format: The input CSV file must have the following
// characteristics:
//    1. The first row contains column names.
//    2.The first column represents the date, and the dates are in ascending
//      order (one row per date).
//    3. Some columns may be empty (missing data for that date).
//
// Output File Generation:
// 1. The output file will contain one row for every date, starting from the
//    earliest date in the input file and ending on the latest date.
// 2. If a date is missing from the input file, the program will generate a new
//    row for that date, populating the columns with the last known values from
//    the previous row. If a column has no previously known value, it will be
//    left blank in the output row.
// 3. The first column of the output file will be the date, followed by all
//    the columns from the input file, preserving their order.
//
// Data Consistency:
// The program ensures that the output file has a consistent date range, with
// one row per day, even if the input file has gaps or missing dates. For
// columns with missing data in the input file, the program carries forward
// the last known value from the previous row, ensuring data continuity in the
// output file. The program does not generate new data; it only propagates the
// last known values for existing columns.
//
// Output File Termination:
// The last row in the output file corresponds to the latest date present in
// the input file. The program does not add any additional rows beyond the
// last date in the input file.

// Usage: tsf <input_csv_file>
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tsf <input_csv_file>")
		os.Exit(1)
	}

	inputFilePath := os.Args[1]
	outputFilePath := strings.TrimSuffix(inputFilePath, filepath.Ext(inputFilePath)) + "-filled.csv"
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	// Create a CSV reader
	reader := csv.NewReader(inputFile)
	reader.TrimLeadingSpace = true

	// Read the header
	header, err := reader.Read()
	if err != nil {
		panic(err)
	}

	// Map to keep the last known values
	lastKnown := make(map[string]string)
	// Slice to hold all rows for output
	var allRows [][]string

	// Start and end date tracker
	var startDate, endDate time.Time
	firstRow := true

	// Read each row
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		// Parse the date from the first column using the specified format
		currentDate, err := util.StringToDate(row[0])
		if err != nil {
			panic(fmt.Errorf("date parsing error: %v for %v", err, row[0]))
		}
		if firstRow {
			startDate = currentDate
			endDate = currentDate
			firstRow = false
		}
		endDate = currentDate

		// Update last known values and prepare the output row
		outputRow := make([]string, len(header))
		for i, val := range row {
			if val != "" {
				lastKnown[header[i]] = val
			}
			outputRow[i] = lastKnown[header[i]]
		}
		allRows = append(allRows, outputRow)
	}

	// Fill in missing dates and values
	var filledRows [][]string
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("1/2/2006")
		found := false
		for _, row := range allRows {
			if row[0] == dateStr {
				filledRows = append(filledRows, row)
				found = true
				break
			}
		}
		if !found {
			newRow := make([]string, len(header))
			newRow[0] = dateStr
			for i := 1; i < len(header); i++ {
				newRow[i] = lastKnown[header[i]]
			}
			filledRows = append(filledRows, newRow)
		}
	}

	// Create and write to the output file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the header
	writer.Write(header)

	// Write the filled rows
	for _, row := range filledRows {
		writer.Write(row)
	}
}
