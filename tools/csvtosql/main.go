package main

import (
	"fmt"
	"log"
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

//	main - this function creates a new db from scratch. It uses the
//	data from platodb.csv
//
// -----------------------------------------------------------------------------
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
	app.cfg = &cfg

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

	//------------------------------------
	// We delete the database first...
	//------------------------------------
	// db, err := newdata.NewDatabase("SQL", app.cfg, app.extres)
	// if err != nil {
	// 	log.Fatalf("Error from NewDatabase: %s\n", err.Error())

	// }
	// if err = db.DropDatabase(); err != nil {
	// 	log.Fatalf("Error from DropDatabase: %s\n", err.Error())
	// }

	//---------------------------------------------------------------------
	// open the MySQL database
	//---------------------------------------------------------------------
	app.sqldb, err = newdata.NewDatabase("SQL", app.cfg, app.extres)
	if err != nil {
		log.Fatalf("Error creating database: %s\n", err.Error())
	}
	if err = app.sqldb.Open(); err != nil {
		log.Fatalf("db.Open returned error: %s\n", err.Error())
	}

	defer app.sqldb.SQLDB.DB.Close()

	if err = app.sqldb.DropDatabase(); err != nil {
		log.Fatalf("DropDatabase returned error: %s\n", err.Error())
	}
	app.sqldb.SQLDB.ParentDB = app.sqldb // we will need this even before we call Init()

	if err = app.sqldb.CreateDatabasePart1(); err != nil {
		log.Fatalf("CreateDatabasePart1 returned error: %s\n", err.Error())
	}

	//----------------------------------------------------------------------
	// We have a new sql database now. Tables are defined, but contain
	// no data at this point. First thing to do is populate the ancillary
	// SQL tables.
	//----------------------------------------------------------------------
	if err = PopulateLocales(); err != nil {
		log.Fatalf("Error from PopulateLocales: %s\n", err.Error())
	}
	// now we need to load the sqldb's locale cache. It's needed by MigrateTimeSeriesData
	if err = app.sqldb.SQLDB.LoadLocaleCache(); err != nil {
		log.Fatalf("Error from LoadLocalCache: %s\n", err.Error())
	}
	if err = CopyCsvMISubclassesToSQL(); err != nil {
		log.Fatalf("Error from CopyCsvMISubclassesToSql: %s\n", err.Error())
	}
	// now that the MISubclasses table has been loaded, we'll need to cache it for use in MigrateTimeSeriesData
	app.sqldb.Mim.ParentDB = app.sqldb
	if err = app.sqldb.Mim.LoadMInfluencerSubclasses(); err != nil {
		log.Fatalf("Error from LoadMInfluencerSubclasses: %s\n", err.Error())
	}
	if err = MigrateTimeSeriesData(); err != nil {
		log.Fatalf("Error from MigrateTimeSeriesData: %s\n", err.Error())
	}
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
