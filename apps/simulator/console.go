package main

import (
	"fmt"
	"os"
	"time"

	"github.com/stmansour/psim/util"
)

func displaySimulationDetails(cfg *util.AppConfig) {
	fmt.Printf("**************  S I M U L A T I O N   D E T A I L S  **************\n")
	a := time.Time(cfg.DtStart)
	b := time.Time(cfg.DtStop)
	c := b.AddDate(0, 0, 1)
	fmt.Printf("Version:             %s\n", util.Version())
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

	arch, err := archiveResults(cfg.Filename, app.archiveBaseDir)
	if err != nil {
		fmt.Printf("archiveResults returned error: %s\n", err)
	}
	fmt.Printf("Archive directory: %s\n", arch)

	// GENERATE  simstats.csv
	err = (&app.sim).DumpStats(arch)
	if err != nil {
		fmt.Printf("Simulator DumpSimStats returned error: %s\n", err)
	}

	// GENERATE  finrep.csv
	err = (&app.sim).FinRpt.GenerateFinRep(&app.sim, arch)
	if err != nil {
		fmt.Printf("Simulator FinRep returned error: %s\n", err)
	}

}

func archiveResults(configFilePath, baseDir string) (string, error) {
	newDir, err := util.CreateTimestampedDir(baseDir)
	if err != nil {
		return "", fmt.Errorf("error creating archive directory: %s", err.Error())
	}

	err = util.FileCopy(configFilePath, newDir)
	if err != nil {
		return newDir, fmt.Errorf("error copying file: %s", err.Error())

	}
	return newDir, nil
}
