package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"psim/core"
	"psim/data"
	"psim/util"
	"time"
)

var app struct {
	showTopInvestor     bool
	dayByDayResults     bool
	dumpInvestmentTable bool
	sim                 core.Simulator
}

func dateIsInDataRange(a time.Time) string {
	if a.Before(data.ER.DtStart) {
		return "prior to Exchange Rate data range"
	}
	if a.After(data.ER.DtStop) {
		return "after to Exchange Rate data range"
	}
	if a.Before(data.DR.DtStart) {
		return "prior to Discount Rate data range"
	}
	if a.After(data.DR.DtStop) {
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

	if a.After(b) {
		fmt.Printf("*** ERROR *** Start date is after Stop ")
		os.Exit(2)
	}
	fmt.Printf("Duration:        %s\n", util.DateDiffString(a, c))
	fmt.Printf("Population Size: %d\n", cfg.PopulationSize)
	fmt.Printf("*******************************************************************\n\n")
}

func displaySimulationResults(cfg *util.AppConfig) {
	fmt.Printf("\n**************  S I M U L A T I O N   R E S U L T S  **************\n")
	(&app.sim).ResultsByInvestor()

}

func readCommandLineArgs() {
	stiptr := flag.Bool("i", false, "write top investor profile to investorProfile.txt and its investments to investments.csv")
	dptr := flag.Bool("d", false, "show day-by-day results")
	invptr := flag.Bool("v", false, "dump remaining Investments at simulation end")
	flag.Parse()
	app.showTopInvestor = *stiptr
	app.dayByDayResults = *dptr
	app.dumpInvestmentTable = *invptr
}

func main() {
	util.Init()
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	readCommandLineArgs()
	rand.Seed(time.Now().UnixNano())

	data.Init()

	displaySimulationDetails(&cfg)
	app.sim.Init(&cfg, app.dayByDayResults, app.dumpInvestmentTable)
	app.sim.Run()

	app.sim.CalculateFitness()
	displaySimulationResults(&cfg)

	if app.showTopInvestor {
		err := app.sim.ShowTopInvestor()
		if err != nil {
			fmt.Printf("Error writing Top Investor profile: %s\n", err.Error())
		}
	}
}
