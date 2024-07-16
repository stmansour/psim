package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/stmansour/psim/newcore"
	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/sqlt"
	"github.com/stmansour/psim/util"

	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

// SimApp is the main application
type SimApp struct {
	ReportTopInvestorInvestments bool
	DayByDay                     bool
	// showAllInvestors             bool // adds all investors to the output in the simulation results
	sim                       newcore.Simulator
	randNano                  int64
	InfPredDebug              bool
	trace                     bool // traces voting activity of Influencers and Investors buy,hold,sell decisions
	traceTiming               bool // traces timing of simulation phase and next creating a new generation
	version                   bool
	cfName                    string // override default config filename (config.json5) with this
	cfg                       *util.AppConfig
	extres                    *util.ExternalResources
	db                        *newdata.Database
	archiveBaseDir            string        // where archives go
	archiveMode               bool          // if true it copies the config file to an archive directory, places simstats and finrep there as well
	CrucibleMode              bool          // normal or crucible
	DNALog                    bool          // generate dnalog when true and CrucibleMode is true
	GenInfluencerDistribution bool          // show Influencer distribution for each generation
	FitnessScores             bool          // save the fitness scores for each generation to dbgFitnessScores.csv
	dbfilename                string        // override database name with this name
	CPUProfile                string        // where is time being spent?
	MemProfile                string        // where is memory being consumed?
	basePort                  int           // Starting port
	maxPort                   int           // Upper limit for trying different ports
	Simtalkport               int           // current port being used
	notalk                    bool          // if true, the simulator does not start up an HTTP listener
	SQLiteFileName            string        // where we keep the Investor cache
	SQLiteDB                  *sql.DB       // the sqlite3 database used for Investor ids
	AllowDuplicateInvestors   bool          // whether to check for duplicate investors or not
	ProgramStarted            time.Time     // when the program started
	SID                       int64         // simulation ID, from the dispatcher
	URL                       string        // URL of the simulator
	DispatcherStatusChannel   chan struct{} // created only when we're sending status to the dispatcher
	HTTPHdrsDbg               bool          // print HTTP headers
	HexASCIIDbg               bool          // print hex and ASCII data from the HTTP request and response
	DispatcherURL             string        // where to reach dispatcher, simd will supply it
	MachineID                 string        // unique id for this machine
}

var app SimApp

func dateIsInDataRange(a time.Time) string {
	switch app.db.Datatype {
	case "CSV":
		if a.Before(app.db.CSVDB.DtStart) {
			return "prior to Discount Rate data range"
		}
		if a.After(app.db.CSVDB.DtStop) {
			return "after to Discount Rate data range"
		}
	case "SQL":
		if a.Before(app.db.SQLDB.DtStart) {
			return "prior to Discount Rate data range"
		}
		if a.After(app.db.SQLDB.DtStop) {
			return "after to Discount Rate data range"
		}
	}
	return "√"
}

func readCommandLineArgs() {
	flag.StringVar(&app.archiveBaseDir, "adir", "", "base archive directory, default is current directory")
	flag.BoolVar(&app.archiveMode, "ar", false, "create archive directory for config file, finrep, simstats, and all other reports. Also see -adir.")
	flag.StringVar(&app.cfName, "c", "", "configuration file to use (instead of config.json5)")
	flag.BoolVar(&app.CrucibleMode, "C", false, "Crucible mode.")
	flag.StringVar(&app.CPUProfile, "cpuprofile", "", "write cpu profile to file")
	flag.BoolVar(&app.InfPredDebug, "D", false, "show prediction debug info - dumps a lot of data, use on short simulations, with minimal Influencers")
	flag.BoolVar(&app.DayByDay, "d", false, "show day-by-day results")
	flag.StringVar(&app.dbfilename, "db", "", "override CSV datatbase name with this name. All CSV database files are assumed to be in the same directory.")
	flag.BoolVar(&app.DNALog, "dnalog", false, "generate DNA log, only relevant when CrucibleMode is enabled.")
	flag.BoolVar(&app.AllowDuplicateInvestors, "dup", false, "Allow duplicate investors within a population.")
	flag.BoolVar(&app.FitnessScores, "fit", false, "generate a Fitness Report that shows the fitness of all Investors for each generation")
	//  flag.BoolVar(&app.showAllInvestors, "i", false, "show all investors in the simulation results")
	flag.BoolVar(&app.GenInfluencerDistribution, "idist", false, "report Influencer Distribution each time a generation completes")
	flag.BoolVar(&app.ReportTopInvestorInvestments, "inv", false, "for each generation, write top investors Investment List to invrep.csv")
	flag.StringVar(&app.MemProfile, "memprofile", "", "write memory profile to this file")
	flag.BoolVar(&app.notalk, "notalk", false, "if true, the simulator does not start up an HTTP listener")
	flag.Int64Var(&app.randNano, "r", -1, "random number seed. ex: ./simulator -r 1687802336231490000")
	flag.Int64Var(&app.SID, "SID", 0, "SID from dispatcher. Should only used by simd or dispatcher")
	flag.BoolVar(&app.trace, "trace", false, "trace decision-making process every day, all investors")
	flag.BoolVar(&app.traceTiming, "tracetime", false, "shows timing of simulation phase and next creating a new generation")
	flag.StringVar(&app.DispatcherURL, "DISPATCHER", "", "Network address for dispatcher. Should only used by simd")
	flag.BoolVar(&app.version, "v", false, "print the program version string")
	flag.Parse()
}

func initSimulation() {
	var err error
	app.randNano = util.Init(app.randNano)
	// fmt.Printf("cfName = %s\n", app.cfName)
	app.extres, err = util.ReadExternalResources()
	if err != nil {
		log.Fatalf("ReadExternalResources returned error: %s\n", err.Error())
	}

	cfg, err := util.LoadConfig(app.cfName)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	app.cfName = cfg.ConfigFilename // it may have been set by LoadConfig
	cfg.InfPredDebug = app.InfPredDebug
	if err = util.ValidateConfig(cfg); err != nil {
		fmt.Printf("Please fix errors in the simulator configuration file, config.json5, and try again\n")
		os.Exit(1)
	}
	cfg.Trace = app.trace
	cfg.ArchiveBaseDir = app.archiveBaseDir
	cfg.ArchiveMode = app.archiveMode
	cfg.AllowDuplicateInvestors = app.AllowDuplicateInvestors //whether to check for need this if swhether to check for support unit testing or not
	if !cfg.CrucibleMode {
		cfg.CrucibleMode = app.CrucibleMode
	}
	cfg.SID = app.SID
	app.cfg = cfg

	//--------------------------
	// OPEN THE DATABASE...
	//--------------------------
	app.db, err = newdata.NewDatabase(cfg.DBSource, cfg, app.extres)
	if err != nil {
		log.Panicf("*** PANIC ERROR ***  NewDatabase returned error: %s\n", err)
	}
	app.db.SetCSVFilename(app.dbfilename) // This call is not actually necessary, but this is when you'd set the override filename if you need to
	if err := app.db.Open(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Open returned error: %s\n", err)
	}
	if err := app.db.Init(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
	}

}

func doSimulation() {
	var err error
	initSimulation()
	app.MachineID, err = util.GetMachineUUID()
	if err != nil {
		log.Panicf("*** PANIC ERROR ***  GetMachineUUID returned error: %s\n", err)
	}

	//---------------------------------------------------------------------------------
	// OPEN SQLITE Cache DB
	// This is used to store the hashes of all the Investors we create so that
	// we don't have duplicates in our simulations. The exception is for PreserveElite
	//---------------------------------------------------------------------------------
	if app.SQLiteFileName, err = sqlt.GenerateDBFileName(); err != nil {
		log.Panicf("*** PANIC ERROR ***  GenerateDBFileName returned error: %s\n", err)
	}
	if app.SQLiteDB, err = sql.Open("sqlite3", app.SQLiteFileName); err != nil {
		log.Fatal(err)
	}
	app.sim.SqltDB = app.SQLiteDB
	sqlt.CreateSchema(app.SQLiteDB)
	defer func() {
		if err := app.SQLiteDB.Close(); err != nil {
			fmt.Printf("Error closing database: %s\n", err)
		}
		if err := os.Remove(app.SQLiteFileName); err != nil {
			fmt.Printf("Error deleting database file: %s\n", err)
		}
	}()

	//##########################################################################################
	//  ON WITH THE SIMULATION...
	//##########################################################################################

	if app.cfg.CrucibleMode {
		c := newcore.NewCrucible()
		c.ReportTopInvestorInvestments = app.ReportTopInvestorInvestments
		c.DayByDay = app.DayByDay
		c.CreateDLog = app.DNALog
		c.Init(app.cfg, app.db, &app.sim)
		c.Run()
		return
	}

	displaySimulationDetails(app.cfg)
	app.sim.Init(app.cfg, app.db, nil, app.DayByDay, app.ReportTopInvestorInvestments)
	app.sim.GenInfluencerDistribution = app.GenInfluencerDistribution
	app.sim.FitnessScores = app.FitnessScores
	app.sim.TraceTiming = app.traceTiming
	app.sim.Simtalkport = app.Simtalkport
	app.sim.Run()

	displaySimulationResults(app.cfg, app.db)
}

func main() {
	var f *os.File
	var err error

	app.randNano = -1
	app.ProgramStarted = time.Now()

	readCommandLineArgs()
	if app.version {
		fmt.Printf("PLATO Simulator version %s\n", util.Version())
		fmt.Printf("app.DispatcherURL = %s\n", app.DispatcherURL)
		os.Exit(0)
	}

	//----------------------------------------------------------------------------
	// Start the HTTP server that can be used to communicate with the simulator
	//----------------------------------------------------------------------------
	ctx, cancel := context.WithCancel(context.Background()) // Create a context that can be canceled
	defer cancel()

	var wg sync.WaitGroup
	if !app.notalk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := startHTTPServer(ctx); err != nil {
				fmt.Printf("HTTP server stopped with error: %v", err)
			}
		}()
	}

	//----------------------------------------------------------------------------
	// If we need to report status to the DISPATCHER, set up the loop that
	// will perform this task every 5 mins
	//----------------------------------------------------------------------------
	if len(app.DispatcherURL) > 0 && app.SID > 0 {
		// fmt.Printf("app.DispatcherURL = %s\n", app.DispatcherURL)
		ticker := time.NewTicker(5 * time.Minute)
		app.DispatcherStatusChannel = make(chan struct{})

		// Start a goroutine that sends status updates every 5 minutes
		go func() {
			for {
                // DEBUG PRINT STATEMENT
                fmt.Printf("*** STATUS LOOP STARTING OVER ***\n")
				select {
				case <-ticker.C:
                    fmt.Printf("SENDING STATUS")
					if err = SendStatusUpdate(nil); err != nil {
						fmt.Printf("Error sending status update: %s\n", err)
					}
				case <-app.DispatcherStatusChannel:
					ticker.Stop()
					return
				}
			}
		}()
	}

	//-------------------------------------------------------------------------
	// CPU profiling.  Run it like this: ./simulator -cpuprofile cpu.prof
	// Then profile it like this: go tool pprof ./simulator cpu.prof
	//-------------------------------------------------------------------------
	if app.CPUProfile != "" {
		f, err = os.Create(app.CPUProfile)
		if err != nil {
			log.Fatalf("could not create CPU profile: %v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("could not start CPU profile: %v", err)
		}
	}

	doSimulation()
	pprof.StopCPUProfile()
	f.Close()

	//-------------------------------------------------------------------------
	// Send completion status to the DISPATCHER
	//-------------------------------------------------------------------------
	if app.SID > 0 && len(app.DispatcherURL) > 0 && app.sim.StopTimeSet {
		if err = SendStatusUpdate(&app.sim.SimStop); err != nil {
			fmt.Printf(">>>> Error sending completion status: %s\n", err)
		}
		close(app.DispatcherStatusChannel)
	}

	//-------------------------------------------------------------------------
	// Memory profiling:  Run it like this: ./simulator -memprofile mem.prof
	// Then profile it like this: go tool pprof ./simulator mem.prof
	//-------------------------------------------------------------------------
	if app.MemProfile != "" {
		f, err := os.Create(app.MemProfile)
		if err != nil {
			log.Fatalf("could not create memory profile: %v", err)
		}
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatalf("could not write memory profile: %v", err)
		}
		f.Close()
	}

	//-------------------------------------------------------------------------
	// Once the simulation is done, cancel the context to stop the HTTP server
	//-------------------------------------------------------------------------
	if !app.notalk {
		cancel()
		wg.Wait() // Wait for the HTTP server goroutine to finish
	}
	fmt.Println("simulation completed.")
}
