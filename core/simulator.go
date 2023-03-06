package core

import (
	"fmt"
	"psim/util"
	"time"
)

// Simulator is a simulator object
type Simulator struct {
	cfg       *util.AppConfig // system-wide configuration info
	Investors []Investor      // the population of the current generation
}

// Init initializes the simulation system, it also creates Investors and
// calls their init functions.
// ----------------------------------------------------------------------------
func (s *Simulator) Init(cfg *util.AppConfig) error {
	s.cfg = cfg

	//------------------------------------------------------------------------
	// Create an initial population of investors with just 1 investor for now
	//------------------------------------------------------------------------
	var i Investor
	i.Init(s.cfg)
	s.Investors = append(s.Investors, i)

	//------------------------------------------------------------------------
	// Initialize all Investors...
	//------------------------------------------------------------------------
	for i := 0; i < len(s.Investors); i++ {
		s.Investors[i].Init(cfg)
	}
	return nil
}

// Run loops through the simulation day by day, first handling any conversions
// from C2 to C1 on that day, and then having each Investor consult its
// Influencers and deciding whether or not to convert C1 to C2. At the end
// of each day it prints out a message indicating where it is at in the
// simulation and some indicators as to how things are progressing.
// ----------------------------------------------------------------------------
func (s *Simulator) Run() {
	dt := time.Time(s.cfg.DtStart)
	dtStop := time.Time(s.cfg.DtStop)

	it := 0
	for dt.Before(dtStop) || dt.Equal(dtStop) {
		it++
		// Call SellConversion for each investor
		for _, inv := range s.Investors {
			inv.SellConversion(dt)
		}

		// Call BuyConversion for each investor
		for _, inv := range s.Investors {
			inv.BuyConversion(dt)
		}

		// Print out iteration number, date, and count of investors with balance > 0
		var count int
		for _, inv := range s.Investors {
			if inv.BalanceC1 > 0 {
				count++
			}
		}
		fmt.Printf("Iteration %d: Date: %s, Investors with balance > 0: %d\n", it, dt.Format("2006-01-02"), count)

		dt = dt.AddDate(0, 0, 1)
	}
}
