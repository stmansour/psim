package main

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Create a new CSV database from the SQL database

// Application is a struct that holds key application resources
type Application struct {
	sqldb       *newdata.Database
	csvdb       *newdata.Database
	cfg         *util.AppConfig
	extres      *util.ExternalResources
	BucketCount int
	DtStart     time.Time
	DtStop      time.Time
}

var app Application

func main() {
	var err error
	start := time.Now()

	//----------------------------------------------------------------------
	// Get ancillary data...
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

	//---------------------------------------------------------------------
	// open the SQL database
	//---------------------------------------------------------------------
	if app.sqldb, err = newdata.NewDatabase("SQL", app.cfg, app.extres); err != nil {
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

	//----------------------------------------------------------------------
	// Create the CSV database from sql
	//----------------------------------------------------------------------
	if app.csvdb, err = newdata.NewDatabase("CSV", app.cfg, nil); err != nil {
		log.Fatalf("*** FATAL ERROR ***  NewDatabase returned error: %s\n", err)
	}

	if app.csvdb.CSVDB.DBPath, err = app.csvdb.CSVDB.EnsureDataDirectory(); err != nil {
		log.Fatalf("*** FATAL ERROR ***  Could not create data directory: %s\n", err)
	}
	if err = app.csvdb.Open(); err != nil {
		log.Fatalf("*** FATAL ERROR ***  db.Init returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	// Copy the tables one by one...
	//   1. MetricsSources
	//----------------------------------------------------------------------
	if err = app.csvdb.InsertMetricsSources(app.sqldb.SQLDB.MetricSrcCache); err != nil {
		log.Fatalf("*** FATAL ERROR ***  InsertMetricsSources returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	//   2. Locales
	//----------------------------------------------------------------------
	if err = app.csvdb.CSVDB.WriteLocalesToCSV(app.sqldb.SQLDB.LocaleCache); err != nil {
		log.Fatalf("*** FATAL ERROR ***  InsertMetricsSources returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	//   3. MISubclasses
	//----------------------------------------------------------------------
	if err = app.csvdb.CSVDB.WriteMISubclassesToCSV(app.sqldb.Mim.MInfluencerSubclasses); err != nil {
		log.Fatalf("*** FATAL ERROR ***  InsertMetricsSources returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	//   4. Exchange Rates
	//----------------------------------------------------------------------

	//----------------------------------------------------------------------
	// All done, let the user know where it was created
	//----------------------------------------------------------------------
	fmt.Printf("Completed. CSV files were created in directory: %s/\n", app.csvdb.CSVDB.DBPath)
	end := time.Now()
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
