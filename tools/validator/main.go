package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Application is a struct that holds key application resources
type Application struct {
	sqldb       *newdata.Database
	csvdb       *newdata.Database
	cfg         *util.AppConfig
	extres      *util.ExternalResources
	BucketCount int
}

var app Application

func main() {
	var err error
	start := time.Now()

	//----------------------------------------------------------------------
	// Now get any other info we need for the databases
	//----------------------------------------------------------------------
	app.extres, err = util.ReadExternalResources()
	if err != nil {
		log.Fatalf("ReadExternalResources returned error: %s\n", err.Error())
	}
	cfg, err := util.LoadConfig("")
	if err != nil {
		log.Fatalf("failed to read config file: %v\n", err)
	}
	app.cfg = cfg

	//----------------------------------------------------------------------
	// open the CSV database from which we'll be pulling data
	//----------------------------------------------------------------------
	app.csvdb, err = newdata.NewDatabase("CSV", app.cfg, nil)
	if err != nil {
		log.Panicf("*** PANIC ERROR ***  NewDatabase returned error: %s\n", err)
	}
	if err := app.csvdb.Open(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
	}
	if err := app.csvdb.Init(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
	}

	//---------------------------------------------------------------------
	// open the SQL database
	//---------------------------------------------------------------------
	app.sqldb, err = newdata.NewDatabase("SQL", app.cfg, app.extres)
	if err != nil {
		log.Fatalf("Error creating database: %s\n", err.Error())
	}
	if err = app.sqldb.Open(); err != nil {
		log.Fatalf("db.Open returned error: %s\n", err.Error())
	}

	if err = app.sqldb.Init(); err != nil {
		log.Fatalf("db.Init returned error: %s\n", err.Error())
	}

	defer app.sqldb.SQLDB.DB.Close()
	if err = app.sqldb.Init(); err != nil {
		log.Fatalf("db.Init returned error: %s\n", err.Error())
	}

	//---------------------------------------------------------------------
	// loop through the data and read records from both sources
	// compare and notify on any discrepancy
	//---------------------------------------------------------------------
	fields := []newdata.FieldSelector{} // an empty slice
	dtStart := time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
	dtEnd := app.csvdb.CSVDB.DtStop.AddDate(0, 0, 1)
	count := 0
	unrecognized := map[string]bool{}

	for dt := dtStart; dt.Before(dtEnd); dt = dt.AddDate(0, 0, 1) {
		rec, err := app.csvdb.Select(dt, fields) // empty slice gets all fields
		if err != nil {
			log.Fatalf("csvdb.Select returned error: %s\n", err.Error())
		}
		//---------------------------------------------------------------------
		// For the SQL db we must supply the fields.  We can build the fields
		// from the return value of the CSV select...
		//---------------------------------------------------------------------
		fields = app.sqldb.SQLDB.FieldSelectorsFromRecord(rec)
		rec1, err := app.sqldb.Select(dt, fields)
		if err != nil && err != sql.ErrNoRows {
			log.Fatalf("sqldb.Select returned error: %s\n", err.Error())
		}
		if len(rec.Fields) == 0 {
			fmt.Printf("No data found for date: %s\n", dt.Format("Jan 2, 2006"))
		}
		if !rec.Date.Equal(rec1.Date) {
			fmt.Printf("Date miscompare. csv: %s, sql %s\n", rec.Date.Format("Jan 2, 2006"), rec1.Date.Format("Jan 2, 2006"))
			count++
		}
		for k, v := range rec.Fields {
			// first, make sure we recognize this metric...
			if !strings.Contains(k, "EXClose") {
				f := newdata.FieldSelector{}
				app.sqldb.SQLDB.FieldSelectorFromCSVColName(k, &f)
				if _, ok := app.sqldb.Mim.MInfluencerSubclasses[f.Metric]; !ok {
					// have we seen this metric before?
					if _, ok = unrecognized[k]; !ok {
						unrecognized[k] = true
						fmt.Printf("Unrecognized metric: %s\n", k)
					}
					continue // no matter what, we don't have this metric, so just keep going
				}
			}

			delta := rec1.Fields[k].Value - v.Value
			if delta < 0 {
				delta = -delta
			}
			if delta > .0001 {
				fmt.Printf("Miscompare - %s: Metric = %s, csv: %12.6f, sql: %12.6f, delta: %12.6f\n", dt.Format("Jan 2, 2006"), k, v.Value, rec1.Fields[k].Value, delta)
				count++
			}
		}
	}
	end := time.Now()
	fmt.Printf("Done!\nMiscompares: %d\n", count)
	FormatDuration(start, end)
}

// FormatDuration prints the duration between two times
func FormatDuration(start, end time.Time) {
	duration := end.Sub(start)

	// Print the duration in a human-readable format
	fmt.Println("Duration:", duration)

	// For more control over the format, you can use the individual components of the duration:
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second
	duration -= seconds * time.Second
	milliseconds := duration / time.Millisecond
	fmt.Printf("Elapsed: %02d hr %02d min %02d sec %03d msec\n", hours, minutes, seconds, milliseconds)
}
