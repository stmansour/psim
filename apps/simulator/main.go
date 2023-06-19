package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/stmansour/psim/core"

	"github.com/stmansour/psim/util"

	"github.com/stmansour/psim/data"
)

var app struct {
	showTopInvestor     bool
	dayByDayResults     bool
	dumpInvestmentTable bool
	showAllInvestors    bool // adds all investors to the output in the simulation results
	sim                 core.Simulator
	randNano            int64
}

func dateIsInDataRange(a time.Time) string {
	// if a.Before(data.ER.DtStart) {
	// 	return "prior to Exchange Rate data range"
	// }
	// if a.After(data.ER.DtStop) {
	// 	return "after to Exchange Rate data range"
	// }
	if a.Before(data.DInfo.DtStart) {
		return "prior to Discount Rate data range"
	}
	if a.After(data.DInfo.DtStop) {
		return "after to Discount Rate data range"
	}
	return "√"
}

func displaySimulationDetails(cfg *util.AppConfig) {
	fmt.Printf("**************  S I M U L A T I O N   D E T A I L S  **************\n")
	a := time.Time(cfg.DtStart)
	b := time.Time(cfg.DtStop)
	c := b.AddDate(0, 0, 1)
	fmt.Printf("Start:           %s\tvalid: %s\n", a.Format("Jan 2, 2006"), dateIsInDataRange(a))
	fmt.Printf("Stop:            %s\tvalid: %s\n", b.Format("Jan 2, 2006"), dateIsInDataRange(b))
	fmt.Printf("C1:              %s\n", cfg.C1)
	fmt.Printf("C2:              %s\n", cfg.C2)

	if a.After(b) {
		fmt.Printf("*** ERROR *** Start date is after Stop ")
		os.Exit(2)
	}
	fmt.Printf("Duration:        %s\n", util.DateDiffString(a, c))
	fmt.Printf("Population Size: %d\n", cfg.PopulationSize)
	fmt.Printf("*******************************************************************\n\n")
}

func displaySimulationResults(cfg *util.AppConfig) {
	f := app.sim.GetFactory()
	fmt.Printf("\n**************  S I M U L A T I O N   R E S U L T S  **************\n")
	fmt.Printf("Number of generations: %d\n", app.sim.GensCompleted)
	fmt.Printf("Observed Mutation Rate: %6.3f%%\n", 100.0*float64(f.Mutations)/float64(f.MutateCalls))
	s, _ := app.sim.GetSimulationRunTime()
	fmt.Printf("Elapsed time: %s\n", s)
	err := (&app.sim).DumpStats()
	if err != nil {
		fmt.Printf("Simulator DumpSimStats returned error: %s\n", err)
	}
	if app.showAllInvestors {
		(&app.sim).ResultsByInvestor()
	}
}

func readCommandLineArgs() {
	dptr := flag.Bool("d", false, "show day-by-day results")
	stiptr := flag.Bool("t", false, "write top investor profile to investorProfile.txt and its investments to investments.csv")
	diptr := flag.Bool("i", false, "show 	all investors in the simulation results")
	invptr := flag.Bool("v", false, "dump remaining Investments at simulation end")
	rndptr := flag.Int64("r", -1, "random number seed")
	flag.Parse()
	app.showTopInvestor = *stiptr
	app.dayByDayResults = *dptr
	app.dumpInvestmentTable = *invptr
	app.showAllInvestors = *diptr
	app.randNano = *rndptr
}

func main() {
	app.randNano = -1
	util.Init(app.randNano)
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	if err = util.ValidateConfig(&cfg); err != nil {
		fmt.Printf("Please fix errors in the simulator configuration file, config.json5, and try again\n")
		os.Exit(1)
	}
	readCommandLineArgs()
	rand.Seed(time.Now().UnixNano())

	if err = data.Init(&cfg); err != nil {
		log.Fatalf("Error initilizing data subsystem: %s\n", err)
	}

	displaySimulationDetails(&cfg)
	app.sim.Init(&cfg, app.dayByDayResults, app.dumpInvestmentTable)
	app.sim.Run()

	// app.sim.CalculateFitness()
	displaySimulationResults(&cfg)

	if app.showTopInvestor {
		err := app.sim.ShowTopInvestor()
		if err != nil {
			fmt.Printf("Error writing Top Investor profile: %s\n", err.Error())
		}
	}
}
