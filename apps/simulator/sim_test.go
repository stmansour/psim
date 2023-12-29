package main

import "testing"

func TestInvestorDNA(t *testing.T) {
	app.randNano = -1
	app.dumpTopInvestorInvestments = true
	// app.trace = true

	doSimulation()
}
