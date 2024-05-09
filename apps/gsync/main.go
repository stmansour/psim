package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// app holds the configuration and global options for the application
// ---------------------------------------------------------------------
var app struct {
	cfg            *util.AppConfig         // application configuration
	cfName         string                  // configuration file name
	PathProcessed  string                  // what basedirectory/directory/file did we process
	StartDate      time.Time               // if a date range for update was specified, this is the start date
	StopDate       time.Time               // if a date range for update was specified, this is the end date
	FixMiscompares bool                    // fix miscompares by using the GCAM values calculated in this run
	Verbose        bool                    // if true, show all data found and the actions taken
	extractURLmode bool                    // if true we just extract the URLS, print them and exit
	gdeltbasedir   string                  // base directory for GDELT
	gdeltfile      string                  // filename of the GDELT csv file to process
	masterlist     string                  // master list of all GDELT csv files if we need to download a list of them
	SQLDB          *newdata.Database       // SQL database
	extres         *util.ExternalResources // credentials, etc
	dt             time.Time               // the date for the metrics computed here
	MSID           int                     // metrics source unique id
	MetricSource   string                  // GDELT
	MetricsTotal   int                     // total number of metrics seed / processed
	Tolerance      float64                 // tolerance for miscomparison
	Miscompared    int
	Corrected      int
	Verified       int
	locs           []LocInfo
}

// main is the entry point for the application
// ----------------------------------------------------------------------
func main() {
	var err error
	time0 := time.Now()

	app.extres, err = util.ReadExternalResources()
	if err != nil {
		log.Fatal(err)
	}

	app.cfg, err = util.LoadConfig(app.cfName)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	defaultgdeltdir := time0.AddDate(0, 0, -1).Format("2006-01-02")
	StartDate := flag.String("d1", "", "Start date in the format YYYY-MM-DD")
	StopDate := flag.String("d2", "", "End date in the format YYYY-MM-DD")
	flag.BoolVar(&app.FixMiscompares, "F", false, "Fix miscompares by overwriting miscompared values with the API values")
	flag.BoolVar(&app.Verbose, "verbose", false, "Verbose mode, show all data found and the actions taken")
	flag.StringVar(&app.cfName, "c", "", "configuration file to use (instead of config.json)")
	flag.StringVar(&app.masterlist, "gml", "masterlist.txt", "GDELT MasterList filename")
	flag.StringVar(&app.gdeltfile, "gf", defaultgdeltdir, "GDELT csv filename")
	flag.StringVar(&app.gdeltbasedir, "gd", "./gdelt", "GDELT base directory")
	flag.Parse()

	//---------------------------------------------------------------------
	// open the SQL database
	//---------------------------------------------------------------------
	if app.SQLDB, err = newdata.NewDatabase("SQL", app.cfg, app.extres); err != nil {
		log.Fatalf("Error creating database: %s\n", err.Error())
	}
	if err = app.SQLDB.Open(); err != nil {
		log.Fatalf("db.Open returned error: %s\n", err.Error())
	}
	if err = app.SQLDB.Init(); err != nil {
		log.Fatalf("db.Init returned error: %s\n", err.Error())
	}
	defer app.SQLDB.SQLDB.DB.Close()

	//---------------------
	// Set the MSID...
	//---------------------
	app.MetricSource = "GDELT"
	app.MSID = -1
	for i := 0; i < len(app.SQLDB.SQLDB.MetricSrcCache); i++ {
		if strings.Contains(app.SQLDB.SQLDB.MetricSrcCache[i].Name, app.MetricSource) {
			app.MSID = app.SQLDB.SQLDB.MetricSrcCache[i].MSID
			break
		}
	}
	if app.MSID == -1 {
		fmt.Printf("Could not find MSID for %s\n", app.MetricSource)
		os.Exit(1)
	}

	//---------------------------------
	// Parse the start and end dates
	//---------------------------------
	if len(*StartDate) > 0 {
		app.StartDate, err = time.Parse("2006-01-02", *StartDate)
		if err != nil {
			fmt.Printf("Error parsing start date: %v\n", err)
			return
		}
	}
	if len(*StopDate) > 0 {
		app.StopDate, err = time.Parse("2006-01-02", *StopDate)
		if err != nil {
			fmt.Printf("Error parsing end date: %v\n", err)
			return
		}
	}

	//--------------------------------------------------------------------------
	// Determine the purpose of this run:  URL Extraction or database update...
	//--------------------------------------------------------------------------
	if len(*StartDate) > 0 && len(*StopDate) > 0 && (app.StartDate.Before(app.StopDate) || app.StartDate.Equal(app.StopDate)) {
		app.extractURLmode = true
		if err := extractURLs(); err != nil {
			fmt.Printf("Error extracting URLs: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	//--------------------------------------------------------------------------
	// process a gdelt csv file
	//--------------------------------------------------------------------------
	if len(app.gdeltfile) > 0 {
		if app.dt, err = ParseDate(app.gdeltfile); err != nil {
			fmt.Printf("Error parsing GDELT CSV filename: %v\n", err)
			os.Exit(1)
		}
		f := util.Stripchars(app.gdeltfile, "-")
		f = app.gdeltbasedir + "/" + f + "/" + f + ".csv"
		app.PathProcessed = f
		if err = ProcessGDELTCSV(app.PathProcessed); err != nil {
			fmt.Printf("Error processing GDELT CSV file: %s\n", err.Error())
			os.Exit(1)
		}
	}

	time1 := time.Now()

	//----------------------------------
	// Results Report...
	//----------------------------------
	fmt.Printf("Program Started......: %s\n", time0.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Processing...........: %s\n", app.PathProcessed)
	fmt.Printf("Total countries......: %d\n", len(app.locs))
	fmt.Printf("Total metrics seen...: %d\n", app.MetricsTotal)
	fmt.Printf("Total SQL inserts....: %d\n", app.SQLDB.SQLDB.InsertCount)
	fmt.Printf("Total SQL updates....: %d\n", app.SQLDB.SQLDB.UpdateCount)
	fmt.Printf("Verified Correct.....: %d\n", app.Verified)
	fmt.Printf("Miscompared..........: %d\n", app.Miscompared)
	fmt.Printf("Corrected............: %d\n", app.Corrected)
	fmt.Printf("Program Finished.....: %s\n", time1.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Elapsed time.........: %s\n", util.ElapsedTime(time0, time1))

}

// ParseDate takes a string in the format "YYYYMMDD" and converts it to a time.Time
// ------------------------------------------------------------------------------------
func ParseDate(dateStr string) (time.Time, error) {
	if len(dateStr) != 8 {
		return time.Time{}, fmt.Errorf("invalid date length: %s", dateStr)
	}

	year, err := strconv.Atoi(dateStr[:4]) // Parse year
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year format: %w", err)
	}

	month, err := strconv.Atoi(dateStr[4:6]) // Parse month
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month format: %w", err)
	}

	day, err := strconv.Atoi(dateStr[6:]) // Parse day
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day format: %w", err)
	}

	//-----------------------------------------------------
	// Create time.Date with parsed year, month, and day
	//-----------------------------------------------------
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return date, nil
}

// extractURLs extracts the URLs from the master list file. The user has indicated
// a date range for the update. If the URL is in the date range, it is printed.
// ---------------------------------------------------------------------------------
func extractURLs() error {
	//---------------------------------
	// Open the masterlist file
	//---------------------------------
	file, err := os.Open(app.masterlist)
	if err != nil {
		fmt.Printf("Error opening master list file: %v\n", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) >= 3 && strings.Contains(fields[2], "gkg.csv.zip") {
			u := fields[2]
			// dateStr := url[49:57] // Extract the date part from the URL
			parsedURL, err := url.Parse(u)
			if err != nil {
				return err
			}

			// Extract the filename from the path
			filename := path.Base(parsedURL.Path)
			dateStr := filename[0:8]

			fileDate, err := time.Parse("20060102", dateStr)
			if err != nil {
				fmt.Printf("Error parsing date from URL: %v\n", err)
				continue
			}

			if (fileDate.Equal(app.StartDate) || fileDate.After(app.StartDate)) &&
				(fileDate.Equal(app.StopDate) || fileDate.Before(app.StopDate)) {
				fmt.Println(u)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
	return nil
}
