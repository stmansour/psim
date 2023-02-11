package core

import "time"

type Simulator struct {
	startDate   time.Time  // processing begins on this date
	stopDate    time.Time  // processing ends on this date
	popSize     int        // how many investors are in this population
	maxInf      int        // maximum number of influencers for any Investor
	minInf      int        // minimum number of influencers for any Investor
	funds       float64    // amount of funds each Investor is "staked" at the outset of the simulation
	tradingDay  int        // this needs to be completely re-thought -- it's like a recurrence rule
	tradingTime time.Time  // time of day when buy/sell is executed
	generations int        // current generation in the simulator
	Investors   []Investor // the population of the current generation

}

// initializePopulation - create a population of investors with random settings
//
// -------------------------------------------------------------------------------
func (s *Simulator) initializePopulation() {

}

// runGeneration - run the current generation from start to stop date
//
// RETURNS
//
//	nil       - no problems, the generation ran to completion without error
//	otherwise - reason the generation did not run or that the run was stopped
//	              * stopDate has been exceeded
//	              * all Investors are out of funds
//
// -------------------------------------------------------------------------------
func (s *Simulator) runGeneration() {
	// evaluate the fitness of each investor at the end of the generation
}

// runSimulation - run all generations
//
// -------------------------------------------------------------------------------
func (s *Simulator) runSimulation() []Investor {
	s.initializePopulation()

	// loop: run the generations until stop criteria is met
	// stop criteria:
	//		1. stopDate reached
	//		2. no Investors have any funds remaining

}
