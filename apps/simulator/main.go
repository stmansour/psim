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
	cfName                     string // override default with this file
	cfg                        *util.AppConfig
	db                         *newdata.Database
	mim                        *newcore.MetricInfluencerManager
}

func dateIsInDataRange(a time.Time) string {
	if a.Before(app.db.CSVData.DtStart) {
		return "prior to Discount Rate data range"
	}
	if a.After(app.db.CSVData.DtStop) {
		return "after to Discount Rate data range"
	}
	return "√"
}

func displaySimulationDetails(cfg *util.AppConfig) {
	fmt.Printf("**************  S I M U L A T I O N   D E T A I L S  **************\n")
	a := time.Time(cfg.DtStart)
	b := time.Time(cfg.DtStop)
	c := b.AddDate(0, 0, 1)
	fmt.Printf("Configuration File:  %s\n", app.cfName)
	fmt.Printf("Start:               %s\tvalid: %s\n", a.Format("Jan 2, 2006"), dateIsInDataRange(a))
	fmt.Printf("Stop:                %s\tvalid: %s\n", b.Format("Jan 2, 2006"), dateIsInDataRange(b))
	if len(cfg.GenDurSpec) > 0 {
		fmt.Printf("Generation Lifetime: %s\n", util.FormatGenDur(cfg.GenDur))
	}
	fmt.Printf("Loop count:          %d\n", cfg.LoopCount)

	fmt.Printf("C1:                  %s\n", cfg.C1)
	fmt.Printf("C2:                  %s\n", cfg.C2)

	if a.After(b) {
		fmt.Printf("*** ERROR *** Start date is after Stop ")
		os.Exit(2)
	}
	fmt.Printf("Duration:            %s\n", util.DateDiffString(a, c))
	fmt.Printf("Population Size:     %d\n", cfg.PopulationSize)
	fmt.Printf("COA Strategy:        %s\n", cfg.COAStrategy)
	s := "Influencers:     "
	fmt.Printf("%s", s)
	n := len(s)
	namesThisLine := 0
	for i := 0; i < len(util.InfluencerSubclasses); i++ {
		subclass := util.InfluencerSubclasses[i]
		if namesThisLine > 0 {
			fmt.Printf(", ")
			n += 2
		}
		if n+len(subclass) > 77 {
			s = "                 "
			fmt.Printf("\n%s", s)
			n = len(s)
			namesThisLine = 0
		}
		fmt.Printf("%s", subclass)
		n += len(subclass)
		namesThisLine++
	}
	fmt.Printf("\n")
	fmt.Printf("*******************************************************************\n\n")
}

func displaySimulationResults(cfg *util.AppConfig) {
	f := app.sim.GetFactory()
	omr := float64(0)
	if f.MutateCalls > 0 {
		omr = 100.0 * float64(f.Mutations) / float64(f.MutateCalls)
	}
	fmt.Printf("\n**************  S I M U L A T I O N   R E S U L T S  **************\n")
	fmt.Printf("Number of generations: %d\n", app.sim.GensCompleted)
	fmt.Printf("Observed Mutation Rate: %6.3f%%\n", omr)
	s, _ := app.sim.GetSimulationRunTime()
	fmt.Printf("Elapsed time: %s\n", s)

	// GENERATE  simstats.csv
	err := (&app.sim).DumpStats()
	if err != nil {
		fmt.Printf("Simulator DumpSimStats returned error: %s\n", err)
	}

	// GENERATE  finrep.csv
	err = (&app.sim).FinRpt.GenerateFinRep(&app.sim)
	if err != nil {
		fmt.Printf("Simulator FinRep returned error: %s\n", err)
	}

}

func readCommandLineArgs() {
	dptr := flag.Bool("d", false, "show day-by-day results")
	Dptr := flag.Bool("D", false, "show prediction debug info - dumps a lot of data, use on short simulations, with minimal Influencers")
	stiptr := flag.Bool("t", false, "for each generation, write top investor Investment List to IList-Gen-n.csv")
	traceptr := flag.Bool("trace", false, "trace decision-making process every day, all investors")
	diptr := flag.Bool("i", false, "show all investors in the simulation results")
	rndptr := flag.Int64("r", -1, "random number seed. ex: ./simulator -r 1687802336231490000")
	cfptr := flag.String("c", "", "configuration file to use (instead of config.json)")
	flag.Parse()
	app.dumpTopInvestorInvestments = *stiptr
	app.dayByDayResults = *dptr
	app.showAllInvestors = *diptr
	app.randNano = *rndptr
	app.InfPredDebug = *Dptr
	app.trace = *traceptr
	app.cfName = *cfptr
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
	app.sim.Run()

	displaySimulationResults(&cfg)

	if app.dumpTopInvestorInvestments {
		err := app.sim.ShowTopInvestor()
		if err != nil {
			fmt.Printf("Error writing Top Investor profile: %s\n", err.Error())
		}
	}
}

func crucible() {
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
	doSimulation()
}
