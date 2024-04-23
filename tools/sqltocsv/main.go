package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	cfName      string // override default config file name with this file
	extres      *util.ExternalResources
	BucketCount int
	DtStart     time.Time
	DtStop      time.Time
	ShardMetric string
}

var app Application

func readCommandLineArgs() {
	flag.StringVar(&app.ShardMetric, "s", "", "print the shard info for the supplied metric (as seen in CSV column header)")
	flag.StringVar(&app.cfName, "c", "", "configuration file to use (instead of config.json)")
	flag.Parse()
}

func main() {
	var err error
	start := time.Now()
	readCommandLineArgs()

	//----------------------------------------------------------------------
	// Get ancillary data...
	//----------------------------------------------------------------------
	app.extres, err = util.ReadExternalResources()
	if err != nil {
		log.Fatalf("ReadExternalResources returned error: %s\n", err.Error())
	}
	cfg, err := util.LoadConfig(app.cfName)
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

	//---------------------------------------------------------------------
	// If the only thing requested was shard info for a metric
	// handle it now and exit
	//---------------------------------------------------------------------
	if len(app.ShardMetric) > 0 {
		PrintShardInfo()
		os.Exit(0)
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
	if err = app.csvdb.WriteMetricsSources(app.sqldb.SQLDB.MetricSrcCache); err != nil {
		log.Fatalf("*** FATAL ERROR ***  WriteMetricsSources returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	//   2. Locales
	//----------------------------------------------------------------------
	if err = app.csvdb.CSVDB.WriteLocalesToCSV(app.sqldb.SQLDB.LocaleCache); err != nil {
		log.Fatalf("*** FATAL ERROR ***  WriteMetricsSourcesToSQL returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	//   3. MISubclasses
	//----------------------------------------------------------------------
	if err = app.csvdb.CSVDB.WriteMISubclassesToCSV(app.sqldb.Mim.MInfluencerSubclasses); err != nil {
		log.Fatalf("*** FATAL ERROR ***  WriteMetricsSourcesToSQL returned error: %s\n", err)
	}

	//----------------------------------------------------------------------
	//   4. Sharded Metrics  (includes EXClose which is not sharded)
	//----------------------------------------------------------------------
	if err = app.csvdb.CSVDB.CopySQLRecsToCSV(app.sqldb); err != nil {
		log.Fatalf("*** FATAL ERROR ***  WriteMetricsSourcesToSQL returned error: %s\n", err)
	}

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

// PrintShardInfo prints the MID and BucketNumber for the metric in app.ShardMetric.
// This is a developer debug feature, it may not be of any value to others.
// --------------------------------------------------------------------------------------
func PrintShardInfo() {
	dt := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
	f := newdata.FieldSelector{
		Metric: app.ShardMetric,
	}
	app.sqldb.SQLDB.FieldSelectorFromCSVColName(app.ShardMetric, &f)
	app.sqldb.SQLDB.GetShardInfo(dt, &f)

	fmt.Printf(
		`            date: %s
          Metric: %s
             MID: %d
          Locale: %s
             LID: %d
         Locale2: %s
            LID2: %d
            MSID: %d
           Table: %s
    BucketNumber: %d
          FQname: %s
`,
		dt.Format("January 2, 2006"), f.Metric, f.MID, f.Locale, f.LID, f.Locale2, f.LID2, f.MSID, f.Table, f.BucketNumber, f.FQMetric())

}
