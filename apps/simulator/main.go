package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/stmansour/psim/newcore"
	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

var app struct {
	dumpTopInvestorInvestments bool
	dayByDayResults            bool
	showAllInvestors           bool // adds all investors to the output in the simulation results
	sim                        newcore.Simulator
	randNano                   int64
	InfPredDebug               bool
	trace                      bool
	version                    bool
	cfName                     string // override default with this file
	cfg                        *util.AppConfig
	db                         *newdata.Database
	mim                        *newcore.MetricInfluencerManager
	archiveBaseDir             string // where archives go
	archiveMode                bool   // if true it copies the config file to an archive directory, places simstats and finrep there as well
	CrucibleMode               bool   // normal or crucible
	GenInfluencerDistribution  bool   // show Influencer distribution for each generation
	FitnessScores              bool   // save the fitness scores for each generation to dbgFitnessScores.csv
}

func dateIsInDataRange(a time.Time) string {
	if a.Before(app.db.CSVDB.DtStart) {
		return "prior to Discount Rate data range"
	}
	if a.After(app.db.CSVDB.DtStop) {
		return "after to Discount Rate data range"
	}
	return "√"
}

func readCommandLineArgs() {
	flag.StringVar(&app.archiveBaseDir, "adir", "", "base archive directory, default is current directory")
	flag.BoolVar(&app.GenInfluencerDistribution, "idist", false, "report Influencer Distribution each time a generation completes")
	flag.BoolVar(&app.FitnessScores, "fit", false, "generate a Fitness Report that shows the fitness of all Investors for each generation")
	flag.BoolVar(&app.archiveMode, "ar", false, "create archive directory for config file, finrep, simstats, and all other reports. Also see -adir.")
	flag.BoolVar(&app.dayByDayResults, "d", false, "show day-by-day results")
	flag.BoolVar(&app.InfPredDebug, "D", false, "show prediction debug info - dumps a lot of data, use on short simulations, with minimal Influencers")
	flag.BoolVar(&app.dumpTopInvestorInvestments, "t", false, "for each generation, write top investor Investment List to IList-Gen-n.csv")
	flag.BoolVar(&app.trace, "trace", false, "trace decision-making process every day, all investors")
	flag.BoolVar(&app.version, "v", false, "print the program version string")
	flag.BoolVar(&app.showAllInvestors, "i", false, "show all investors in the simulation results")
	flag.Int64Var(&app.randNano, "r", -1, "random number seed. ex: ./simulator -r 1687802336231490000")
	flag.StringVar(&app.cfName, "c", "", "configuration file to use (instead of config.json)")
	flag.BoolVar(&app.CrucibleMode, "C", false, "Crucible mode.")
	flag.Parse()
}

func doSimulation() {
	app.randNano = util.Init(app.randNano)
	// fmt.Printf("cfName = %s\n", app.cfName)
	cfg, err := util.LoadConfig(app.cfName)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	cfg.InfPredDebug = app.InfPredDebug
	if err = util.ValidateConfig(&cfg); err != nil {
		fmt.Printf("Please fix errors in the simulator configuration file, config.json5, and try again\n")
		os.Exit(1)
	}
	cfg.Trace = app.trace
	cfg.ArchiveBaseDir = app.archiveBaseDir
	cfg.ArchiveMode = app.archiveMode
	app.cfg = &cfg

	app.db, err = newdata.NewDatabase("CSV", &cfg)
	if err != nil {
		log.Panicf("*** PANIC ERROR ***  NewDatabase returned error: %s\n", err)
	}
	if err := app.db.Init(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
	}

	app.mim = newcore.NewInfluencerManager()
	if err = app.mim.Init(); err != nil {
		log.Panicf("*** PANIC ERROR ***  app.mim.Init() returned error: %s\n", err)
	}

	if cfg.CrucibleMode {
		crucible()
		os.Exit(0)
	}

	displaySimulationDetails(&cfg)
	app.sim.Init(app.cfg, app.db, app.mim, app.dayByDayResults, app.dumpTopInvestorInvestments)
	app.sim.GenInfluencerDistribution = app.GenInfluencerDistribution
	app.sim.FitnessScores = app.FitnessScores
	app.sim.Run()

	displaySimulationResults(&cfg, app.db)

	if app.dumpTopInvestorInvestments {
		err := app.sim.ShowTopInvestor()
		if err != nil {
			fmt.Printf("Error writing Top Investor profile: %s\n", err.Error())
		}
	}
}

func crucible() {
	c := NewCrucible()
	c.Init(app.cfg)
	for i := 0; i < len(app.cfg.TopInvestors); i++ {
		for j := 0; j < len(app.cfg.CrucibleSpans); j++ {
			var sim newcore.Simulator
			app.cfg.DtStart = util.CustomDate(app.cfg.CrucibleSpans[j].DtStart)
			app.cfg.DtStop = util.CustomDate(app.cfg.CrucibleSpans[j].DtStop)
			app.cfg.SingleInvestorDNA = app.cfg.TopInvestors[i].DNA
			app.cfg.SingleInvestorMode = true
			sim.Init(app.cfg, app.db, app.mim, false, false)
			sim.Run()
		}
	}
}

func main() {
	app.randNano = -1
	app.cfName = "config.json5"
	readCommandLineArgs()
	if app.version {
		fmt.Printf("PLATO Simulator version %s\n", util.Version())
		os.Exit(0)
	}
	doSimulation()
}
