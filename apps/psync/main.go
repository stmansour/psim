package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Trading Economics API
//
//		 key urls:
//		       https://api.tradingeconomics.com/indicators  - provides a list of all indicators, though many don't work
//
//		       https://api.tradingeconomics.com/country/{c}/indicator/{i}[/{dt1}/dt2/]?f=json&c=APIKey   -  generic url for indicators
//					c = country; "united states", "germany", "sweden", ...
//					i = indicator from the list of available indicators above
//					dt1 = start date, this is an optional field, with no date range it returns the most recent data
//					dt2 = end date, this is an optional field
//
//
//	         https://api.tradingeconomics.com/markets/historical/USDJPY:CUR,AUDUSD:CUR?d1=2024-04-01&d2=2024-04-16&c=APIKey&f=json
//
// ------------------------------------------------------------------------------------------------
var app = struct {
	APIKey            string
	Dt1               string
	Dt2               string
	StartDate         time.Time
	StopDate          time.Time
	SQLDB             *newdata.Database
	cfg               *util.AppConfig
	cfName            string
	extres            *util.ExternalResources
	countries         []string // the countries we'll pull data for
	metricsSrc        string   // Trading Economics API
	MSID              int      // metrics source unique id
	HTTPGetCalls      int      // how many times we've called http.Get
	HTTPGetErrs       int      // how many times we've gotten errors
	Verified          int      // how many metrics were verified
	Miscompared       int      // how many metrics were miscompared
	Tolerance         float64  // tolerance for miscomparison
	Verbose           bool     // show all data found and the actions taken
	APIFixMiscompares bool     // fix miscompares by using the API values
	Corrected         int      // how many metrics were corrected
	SkipIndicators    bool     // if true, skip the indicators
	SkipForex         bool     // if true, skip the forex
	SingleMetric      string   // if specified, only update this metric
}{
	APIKey: "",
}

func readCommandLineArgs() {
	flag.StringVar(&app.cfName, "c", "", "configuration file to use (instead of config.json)")
	flag.StringVar(&app.Dt1, "d1", "", "Start Date for data, YYYY-mm-dd, default is 7 days ago. Both d1 and d2 are required if either are specified.")
	flag.StringVar(&app.Dt2, "d2", "", "Stop Date for data, YYYY-mm-dd, default is 1 day ago. Both d1 and d2 are required if either are specified.")
	flag.StringVar(&app.SingleMetric, "metric", "", "Update data only for the supplied metric")
	flag.BoolVar(&app.SkipIndicators, "SI", false, "Skip updates and verification of indicators")
	flag.BoolVar(&app.SkipForex, "SF", false, "Skip updates and verification of forex data")
	flag.BoolVar(&app.APIFixMiscompares, "F", false, "Fix miscompares by overwriting miscompared values with the API values")
	flag.BoolVar(&app.Verbose, "verbose", false, "Verbose mode, show all data found and the actions taken")
	flag.Parse()
}

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

	app.APIKey = app.extres.TradingeconomicsAPIKey
	readCommandLineArgs()

	app.Tolerance = 0.0051 // = $0.0051, just a little over 1/2 the metric unit

	//---------------------------------------
	// Default date range... last 7 days
	//---------------------------------------
	app.StopDate = time.Now().AddDate(0, 0, -1)
	app.StartDate = app.StopDate.AddDate(0, 0, -7)
	if len(app.Dt1) > 0 {
		if app.StartDate, err = util.StringToDate(app.Dt1); err != nil {
			log.Fatal(err)
		}
		if len(app.Dt2) == 0 {
			log.Fatal("Both d1 and d2 are required.")
		}
		if app.StopDate, err = util.StringToDate(app.Dt2); err != nil {
			log.Fatal(err)
		}
	}
	app.countries = []string{"Australia", "Japan", "United States"}

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

	//---------------------------------------------------------------------
	// Build the lists of metrics:  indicators, forex, commodities
	//---------------------------------------------------------------------
	mtcs, fxrs, err := BuildMetricLists(app.StartDate, app.StopDate)
	if err != nil {
		log.Fatal(err)
	}
	if app.Verbose {
		fmt.Printf("Indicators:\n")
		for i := 0; i < len(mtcs); i++ {
			fmt.Printf("%3d. %s\n", i, mtcs[i].Handle)
		}
		fmt.Printf("Forex:\n")
		for i := 0; i < len(fxrs); i++ {
			fmt.Printf("%3d. %s\n", i, fxrs[i].Handle)
		}
	}
	fmt.Printf("Found %d forex, %d indicators\n", len(fxrs), len(mtcs))

	//-----------------------------------------------------------------
	// This is (at least for now) the Trading Economics update. So
	// we need to set the data source info in the app struct...
	//-----------------------------------------------------------------
	mn := "tradingeconomics"
	app.MSID = -1
	for _, ms := range app.SQLDB.SQLDB.MetricSrcCache {
		m := util.Stripchars(strings.ToLower(ms.Name), " ")
		if strings.Contains(m, mn) {
			app.MSID = ms.MSID
			app.metricsSrc = ms.Name
			break
		}
	}
	if app.MSID == -1 {
		log.Fatalf("Could not find metric source: %s\n", mn)
	}

	if !app.SkipIndicators && len(mtcs) > 0 {
		//----------------------------------
		// Fetch indicators.
		//----------------------------------
		fmt.Printf("Updating Indicators...\n")
		ind, err := FetchIndicators(app.StartDate, app.StopDate, mtcs)
		if err != nil {
			fmt.Println("Error fetching indicators:", err)
			return
		}
		//----------------------------------
		// Store / update the indicators
		//----------------------------------
		if err = UpdateIndicators(ind); err != nil {
			fmt.Println("Error updating indicators:", err)
			return
		}
		//----------------------------------
		// Calls are rate-limited.
		// Wait for 1 second
		//----------------------------------
		time.Sleep(time.Second)
	}

	fmt.Printf("Updating Foreign Exchange Rates...\n")

	//----------------------------------
	// Fetch forex rates.
	//----------------------------------
	if !app.SkipForex && len(fxrs) > 0 {
		rates, err := FetchForexRates(app.StartDate, app.StopDate, fxrs)
		if err != nil {
			fmt.Println("Error fetching forex rates:", err)
			return
		}
		if err = UpdateForex(rates, fxrs); err != nil {
			fmt.Println("Error updating forex rates:", err)
			return
		}
	}

	time1 := time.Now()

	//----------------------------------
	// Results Report...
	//----------------------------------
	fmt.Printf("Program Started......: %s\n", time0.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Update time period...: %s - %s\n", app.StartDate.Format("2006-01-02"), app.StopDate.Format("2006-01-02"))
	fmt.Printf("Total HTTP calls.....: %d\n", app.HTTPGetCalls)
	fmt.Printf("Total HTTP errors....: %d\n", app.HTTPGetErrs)
	fmt.Printf("Total countries......: %d\n", len(app.countries))
	fmt.Printf("Total indicators.....: %d\n", len(mtcs))
	fmt.Printf("Total forex..........: %d\n", len(fxrs))
	fmt.Printf("Total SQL inserts....: %d\n", app.SQLDB.SQLDB.InsertCount)
	fmt.Printf("Total SQL updates....: %d\n", app.SQLDB.SQLDB.UpdateCount)
	fmt.Printf("Verified Correct.....: %d\n", app.Verified)
	fmt.Printf("Miscompared..........: %d\n", app.Miscompared)
	fmt.Printf("Corrected............: %d\n", app.Corrected)
	fmt.Printf("Program Finished.....: %s\n", time1.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Elapsed time.........: %s\n", util.ElapsedTime(time0, time1))
}
