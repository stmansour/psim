package main

import "testing"

func TestSingleInvestorMode(t *testing.T) {
	app.randNano = -1
	app.dumpTopInvestorInvestments = true
	app.trace = true
	// app.cfName = "singleInvestor.json5"
	app.cfName = "si2.json5"
	s := &app.sim
	s.ResetSimulator()
	doSimulation()
}

func TestLinguisticsInfluencers(t *testing.T) {
	app.randNano = -1
	app.dumpTopInvestorInvestments = false
	app.trace = false
	app.cfName = "linguistics.json5"
	s := &app.sim
	s.ResetSimulator()
	doSimulation()
}

func TestFinRep(t *testing.T) {
	app.randNano = -1
	app.dumpTopInvestorInvestments = false
	app.trace = false
	app.cfName = "finrep.json5"
	s := &app.sim
	s.ResetSimulator()
	doSimulation()
}
