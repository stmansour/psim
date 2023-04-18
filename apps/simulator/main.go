package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"psim/core"
	"psim/data"
	"psim/util"
	"time"
)

var app struct {
	showTopInvestor bool
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
	fmt.Printf("**********  S I M U L A T I O N   D E T A I L S  **********\n")
	a := time.Time(cfg.DtStart)
	b := time.Time(cfg.DtStop)
	b = b.AddDate(0, 0, 1)
	fmt.Printf("Start:    %s\tvalid: %s\n", a.Format("Jan 2, 2006"), dateIsInDataRange(a))
	fmt.Printf("Stop:     %s\tvalid: %s\n", b.Format("Jan 2, 2006"), dateIsInDataRange(b))

	if a.After(b) {
		fmt.Printf("*** ERROR *** Start date is after Stop ")
		os.Exit(2)
	}
	fmt.Printf("Duration: %s\n", util.DateDiffString(a, b))
	fmt.Printf("***********************************************************\n\n")
}

func readCommandLineArgs() {
	stiptr := flag.Bool("i", false, "write top investor profile to investorProfile.txt and its investments to investments.csv")
	flag.Parse()
	app.showTopInvestor = *stiptr
}

func main() {
	util.Init()
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	readCommandLineArgs()

	data.Init()

	displaySimulationDetails(&cfg)

	var sim core.Simulator
	sim.Init(&cfg)
	sim.Run()

	if app.showTopInvestor {
		err := sim.ShowTopInvestor()
		if err != nil {
			fmt.Printf("Error writing Top Investor profile: %s\n", err.Error())
		}
	}
}
