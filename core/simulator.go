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

	iteration := 0
	for dt.Before(dtStop) || dt.Equal(dtStop) {
		iteration++
		// Call SellConversion for each investor
		for j := 0; j < len(s.Investors); j++ {
			s.Investors[j].SellConversion(dt)
		}

		// Call BuyConversion for each investor
		check := false
		for j := 0; j < len(s.Investors); j++ {
			s.Investors[j].BuyConversion(dt)
			if len(s.Investors[j].Investments) > 0 {
				check = true
			}
		}
		if check {
			x := 0
			for j := 0; j < len(s.Investors); j++ {
				x += len(s.Investors[j].Investments)
			}
		}

		// debug
		count := 0
		txns := 0
		for j := 0; j < len(s.Investors); j++ {
			if s.Investors[j].BalanceC1 > 0 {
				count++
			}
			txns += len(s.Investors[j].Investments)
		}
		fmt.Printf("%4d. Date: %s, investors remaining: %d, investments pending: %d\n", iteration, dt.Format("2006-Jan-02"), count, txns)

		dt = dt.AddDate(0, 0, 1)
	}
}

// ShowTopInvestor - dumps the top investor to a file after the simulation.
//
// RETURNS
//
//	nil = success
//	otherwise = error encountered
//
// ----------------------------------------------------------------------------
func (s *Simulator) ShowTopInvestor() error {
	if len(s.Investors) < 1 {
		return fmt.Errorf("Simulator has 0 Investors")
	}
	topBalance := s.Investors[0].BalanceC1
	topInvestorIdx := 0
	for i := 1; i < len(s.Investors); i++ {
		if s.Investors[i].BalanceC1 > topBalance {
			topBalance = s.Investors[i].BalanceC1
			topInvestorIdx = i
		}
	}
	if err := s.Investors[topInvestorIdx].InvestorProfile(); err != nil {
		return err
	}
	return s.Investors[topInvestorIdx].OutputInvestments()
}
