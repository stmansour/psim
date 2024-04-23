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
	DtStart     time.Time
	DtStop      time.Time
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
	app.cfg = cfg
	app.DtStart = time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
	app.DtStop = time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC)

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
	defer app.sqldb.SQLDB.DB.Close()
	if app.DtStop.After(app.csvdb.CSVDB.DtStop) {
		fmt.Printf("Stop date provided goes beyond the end of the data available in platodb.csv (which ends %s)\n", app.csvdb.CSVDB.DtStop.Format("2006-Jan-02"))
	}

	//------------------------------------------------------------------
	//  DELETE THE CURRENT SQL DATABASE.   THIS PROGRAM SHOULD ONLY
	//  BE RUN IF YOU REALLY KNOW WHAT YOU ARE DOING...
	//------------------------------------------------------------------
	if err = app.sqldb.DropDatabase(); err != nil {
		log.Fatalf("DropDatabase returned error: %s\n", err.Error())
	}
	app.sqldb.SQLDB.ParentDB = app.sqldb // we will need this even before we call Init()

	//-------------------------
	// Create TABLES
	//-------------------------
	if err = app.sqldb.CreateDatabaseTables(); err != nil {
		log.Fatalf("CreateDatabaseTables returned error: %s\n", err.Error())
	}

	//----------------------------------------------------------------------
	// We have a new sql database now. Tables are defined, but contain
	// no data at this point.  So, now we populate the tables:
	//
	//         Tables             Description
	//      -----------           -------------------------------
	//       1. Locales           locale names:  USA USD, JPN JPY, etc..
	//       2. MISubclasses      Metric Influencers
	//       3  MetricsSources    Where the metrics come from
	//       4. Exchange Rate     Currency exchange rates
	//       4. Metrics_n_decade  all metrics
	//----------------------------------------------------------------------
	if err = CopyCsvLocalesToSQL(); err != nil {
		log.Fatalf("Error from CopyCsvLocalesToSQL: %s\n", err.Error())
	}
	//-------------------------------------------------------------------------------------
	// Now we need to load the sqldb's locale cache. It's needed by MigrateTimeSeriesData.
	// This normally happens in Init, but we don't use Init in this case only
	//-------------------------------------------------------------------------------------
	if err = app.sqldb.SQLDB.LoadLocaleCache(); err != nil {
		log.Fatalf("Error from LoadLocalCache: %s\n", err.Error())
	}
	if err = CopyCsvMISubclassesToSQL(); err != nil {
		log.Fatalf("Error from CopyCsvMISubclassesToSql: %s\n", err.Error())
	}

	//-------------------------------------------------------------------------------------
	// This may have caused IDs to change. Let's reload them now so
	// we can be assured that the ids are correct.  We do this by
	// loading the SQL database cache into the CSV database.
	//-------------------------------------------------------------------------------------
	if err = app.csvdb.CSVDB.LoadMetricsSourceCache(); err != nil {
		log.Fatalf("Error from LoadMetricsSourceCache: %s\n", err.Error())
	}

	//-------------------------------------------------------------------------------------
	// now cache it for the SQL db...
	//-------------------------------------------------------------------------------------
	app.sqldb.Mim.ParentDB = app.sqldb
	if err = app.sqldb.Mim.LoadMInfluencerSubclasses(); err != nil {
		log.Fatalf("Error from LoadMInfluencerSubclasses: %s\n", err.Error())
	}

	//-------------------------------------------------------------------------------------
	// and now we write the metrics...
	//-------------------------------------------------------------------------------------
	if err = app.sqldb.WriteMetricsSources(app.csvdb.CSVDB.MetricSrcCache); err != nil {
		log.Fatalf("Error from WriteMetricsSources: %s\n", err.Error())
	}
	if err = MigrateTimeSeriesData(); err != nil {
		log.Fatalf("Error from MigrateTimeSeriesData: %s\n", err.Error())
	}

	end := time.Now()
	fmt.Printf("Elapsed time: %s\n", util.ElapsedTime(start, end))
	// FormatDuration(start, end)
}
