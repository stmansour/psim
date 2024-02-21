package main

import "testing"

func TestInvestorDNA(t *testing.T) {
	app.randNano = -1
	app.cfName = "configsmall.json5"
	app.dumpTopInvestorInvestments = true
	s := &app.sim
	s.ResetSimulator()

	doSimulation()
}

func TestLinguisticDNA(t *testing.T) {
	app.randNano = -1
	app.cfName = "linguistics.json5"
	s := &app.sim
	app.trace = false
	s.ResetSimulator()
	doSimulation()
}
func TestWTInfl(t *testing.T) {
	app.randNano = -1
	app.cfName = "wtconfig.json5"
	s := &app.sim
	app.trace = false
	s.ResetSimulator()
	doSimulation()
}
