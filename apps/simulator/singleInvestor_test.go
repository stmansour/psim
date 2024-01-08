package main

import "testing"

func TestSingleInvestorMode(t *testing.T) {
	app.randNano = -1
	app.dumpTopInvestorInvestments = true
	app.trace = true
	app.cfName = "singleInvestor.json5"

	s := &app.sim
	s.ResetSimulator()

	doSimulation()
}
