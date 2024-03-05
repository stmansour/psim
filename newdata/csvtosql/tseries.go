package main

import (
	"fmt"
	"time"
)

// MigrateTimeSeriesData migrates the CSV time-series data into sql tables
func MigrateTimeSeriesData() error {
	// grab a record from the csv file, something that's fully populated...
	dt := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	fields := []string{}                     // an empty slice
	rec, err := app.csvdb.Select(dt, fields) // empty slice gets all fields
	if err != nil {
		return err
	}

	// Write it to the database...
	if err = app.sqldb.Insert(rec); err != nil {
		return err
	}

	// now read it back and make sure we have the same values...
	rec1, err := app.sqldb.Select(dt, fields)
	if err != nil {
		return err
	}

	// compare...
	count := 0
	if !rec1.Date.Equal(rec.Date) {
		fmt.Printf("Dates miscompare!\n")
		count++
	}
	for k, v := range rec1.Fields {
		if rec.Fields[k] != v {
			fmt.Printf("Metric Values for %s miscompare!\n", k)
			count++
		}
	}
	fmt.Printf("miscompares found: %d\n", count)

	return nil
}
