package main

import (
	"fmt"
	"os"
	"testing"
)

func TestGenDur(t *testing.T) {
	app.randNano = -1
	app.cfName = "gendur.json5"
	setDBLocation()
	s := &app.sim
	s.ResetSimulator()

	doSimulation()
}

func TestCrucible(t *testing.T) {
	app.randNano = -1
	app.cfName = "confcru.json5"
	setDBLocation()
	s := &app.sim
	s.ResetSimulator()

	doSimulation()
}

func TestInvestorDNA(t *testing.T) {
	app.randNano = -1
	app.cfName = "configsmall.json5"
	app.dumpTopInvestorInvestments = true
	setDBLocation()
	s := &app.sim
	s.ResetSimulator()

	doSimulation()
}

func TestLinguisticDNA(t *testing.T) {
	app.randNano = -1
	app.cfName = "linguistics.json5"
	setDBLocation()
	s := &app.sim
	app.trace = false
	s.ResetSimulator()
	doSimulation()
}
func TestWTInfl(t *testing.T) {
	app.randNano = -1
	app.cfName = "wtconfig.json5"
	setDBLocation()
	s := &app.sim
	app.trace = false
	s.ResetSimulator()
	doSimulation()
}

// setDBLocation sets the correct db location.  When running tests from within
// VS Code, it creates an executable in some location under /var . It's not the
// directory we expect, so it doesn't have the data in it.
func setDBLocation() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	app.dbfilename = dir + "/data/platodb.csv"
}
