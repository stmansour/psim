package main

import "testing"

func TestInvestorDNA(t *testing.T) {
	app.randNano = -1
	app.dumpTopInvestorInvestments = true
	// app.trace = true

	// var f core.Factory
	// app.sim.cfg = nil
	// app.sim.factory = f
	// app.sim.Investors = nil
	// app.sim.dayByDay = false
	// app.sim.dumpTopInvestorInvestments = false
	// app.sim.maxProfitThisRun = 0
	// app.sim.maxPredictions = make(map[string]int)
	// app.sim.maxProfitInvestor = 0
	// app.sim.maxFitnessScore = 0
	// app.sim.GensCompleted = 0
	// app.sim.SimStats = make([]SimulationStatistics, 0)
	// app.sim.StopTimeSet = false
	// app.sim.WindDownInProgress = false

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

func TestBug(t *testing.T) {
	app.randNano = -1
	app.cfName = "dna-0-jun-2022.json5"
	s := &app.sim
	app.trace = false
	s.ResetSimulator()
	doSimulation()
}
