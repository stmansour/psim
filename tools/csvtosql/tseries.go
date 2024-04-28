package main

import (
	"fmt"
	"time"

	"github.com/stmansour/psim/newdata"
)

// MigrateTimeSeriesData migrates the CSV time-series data into sql tables
func MigrateTimeSeriesData() error {
	fields := []newdata.FieldSelector{} // an empty slice
	dtEpoch := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	dtStart := app.csvdb.CSVDB.DtStart
	if dtStart.Before(dtEpoch) {
		dtStart = dtEpoch
	}
	// dtStart := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)  // THIS IS FOR DEBUGGING
	dtEnd := app.csvdb.CSVDB.DtStop.AddDate(0, 0, 1)
	for dt := dtStart; dt.Before(dtEnd); dt = dt.AddDate(0, 0, 1) {
		rec, err := app.csvdb.Select(dt, fields) // empty slice gets all fields
		if err != nil {
			return err
		}
		if len(rec.Fields) == 0 {
			continue
		}
		//-------------------------------------------------------------------------
		// We're writing to the SQL database, so we need to set the metrics source.
		// In this case, we're sourcing from a CSV file.
		//-------------------------------------------------------------------------
		for _, v := range rec.Fields {
			v.MSID = app.MSID
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
