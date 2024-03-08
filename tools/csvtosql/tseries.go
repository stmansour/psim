package main

import (
	"fmt"

	"github.com/stmansour/psim/newdata"
)

// MigrateTimeSeriesData migrates the CSV time-series data into sql tables
func MigrateTimeSeriesData() error {
	fields := []newdata.FieldSelector{} // an empty slice
	dtEnd := app.DtStop.AddDate(0, 0, 1)
	for dt := app.DtStart; dt.Before(dtEnd); dt = dt.AddDate(0, 0, 1) {
		rec, err := app.csvdb.Select(dt, fields) // empty slice gets all fields
		if err != nil {
			return err
		}
		if len(rec.Fields) == 0 {
			continue
		}

		// Write it to the database...
		if err = app.sqldb.Insert(rec); err != nil {
			return err
		}
		fmt.Printf("%s %d\r", dt.Format("Jan"), dt.Year())
	}
	fmt.Printf("\n")

	return nil
}
