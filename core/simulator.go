package core

import (
	"fmt"
	"psim/data"
	"psim/util"
	"time"
)

// Simulator is a simulator object
type Simulator struct {
	cfg       *util.AppConfig // system-wide configuration info
	Investors []Investor      // the population of the current generation
	dayByDay  bool            // show day by day results, debug feature
	invTable  bool            // dump the investment table at the end of the simulation
}

// Init initializes the simulation system, it also creates Investors and
// calls their init functions.
// ----------------------------------------------------------------------------
func (s *Simulator) Init(cfg *util.AppConfig, dayByDay, invTable bool) error {
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
	for dtStop.After(dt) || dtStop.Equal(dt) {
		iteration++
		SellCount := 0
		BuyCount := 0
		// Call SellConversion for each investor
		for j := 0; j < len(s.Investors); j++ {
			x, sc, err := (&s.Investors[j]).SellConversion(dt)
			if err != nil {
				fmt.Printf("SellConversion returned: %s\n", err.Error())
			}
			SellCount += sc
			// util.DPrintf("Simulator.Run -- SellConversion returned:C1 Bal: %8.2f %s, C2 Bal: %8.2f %s\n", x.BalanceC1, x.cfg.C1, x.BalanceC2, x.cfg.C2)
			s.Investors[j] = x
			// util.DPrintf("Simulator.Run -- AFTER SellConversion Investor[0] info: C1 Bal: %8.2f %s, C2 Bal: %8.2f %s\n",
			// 	s.Investors[j].BalanceC1, s.Investors[j].cfg.C1,
			// 	s.Investors[j].BalanceC2, s.Investors[j].cfg.C2)
		}

		// Call BuyConversion for each investor
		check := false
		for j := 0; j < len(s.Investors); j++ {
			bc, err := (&s.Investors[j]).BuyConversion(dt)
			if err != nil {
				fmt.Printf("BuyConversion returned: %s\n", err.Error())
			}
			if len(s.Investors[j].Investments) > 0 {
				check = true
			}
			BuyCount += bc
		}
		if check {
			x := 0
			for j := 0; j < len(s.Investors); j++ {
				x += len(s.Investors[j].Investments)
			}
		}

		// debug --------------------------------------------------------
		if s.dayByDay {
			count := 0
			txns := 0
			for j := 0; j < len(s.Investors); j++ {
				if s.Investors[j].BalanceC1 > 0 {
					count++
				}
				txns += len(s.Investors[j].Investments)
			}
			fmt.Printf("%4d. Date: %s, Buys: %d, Sells %d,\n      investors remaining: %d, investments pending: %d\n",
				iteration, dt.Format("2006-Jan-02"), BuyCount, SellCount, count, txns)
			// util.DPrintf("Simulator.Run -- Investor[0] info: C1 Bal: %8.2f %s, C2 Bal: %8.2f %s\n",
			// 	s.Investors[0].BalanceC1, s.Investors[0].cfg.C1,
			// 	s.Investors[0].BalanceC2, s.Investors[0].cfg.C2)
		}

		if s.invTable {
			for j := 0; j < len(s.Investors); j++ {
				s.Investors[j].OutputInvestments()
			}
		}
		// debug --------------------------------------------------------

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

func (s *Simulator) ResultsByInvestor() {
	for i := 0; i < len(s.Investors); i++ {
		fmt.Printf("%s\n", s.ResultsForInvestor(i, &s.Investors[i]))
	}
}

func (s *Simulator) ResultsForInvestor(i int, v *Investor) string {
	c1Amt := float64(0)
	dt := time.Time(s.cfg.DtStop)
	pending := len(v.Investments)
	//-------------------------------------------------------------------------
	// Determine the amount of C1 currency that is still invested in C2...
	// store in:  amt
	//-------------------------------------------------------------------------
	amt := float64(0)
	for j := 0; j < pending; j++ {
		if !v.Investments[j].Completed {
			amt += v.Investments[j].BuyC2
		}
	}
	str := fmt.Sprintf("Investor %3d.   C1: %8.2f,  C2 %8.2f\n", i, v.BalanceC1, v.BalanceC2)

	//-------------------------------------------------------------------------
	// Convert amt to C1 currency on the day of the simulation end...
	// store in:   c1Amt
	//-------------------------------------------------------------------------
	if amt > 0 {
		er4 := data.ERFindRecord(dt) // get the exchange rate on t4
		if er4 == nil {
			err := fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found", dt.Format("1/2/2006"))
			return err.Error()
		}
		c1Amt = amt / er4.Close
		str += fmt.Sprintf("                Pending Investments: %d, value: %8.2f %s  =  %8.2f %s\n", pending, amt, s.cfg.C2, c1Amt, s.cfg.C1)
	}
	str += fmt.Sprintf("                Initial Stake: %8.2f %s,  End Balance: %8.2f\n", s.cfg.InitFunds, s.cfg.C1, v.BalanceC1+c1Amt)

	endingC1Balance := c1Amt + v.BalanceC1
	netGain := endingC1Balance - s.cfg.InitFunds
	pctGain := netGain / s.cfg.InitFunds
	str += fmt.Sprintf("                Initial Balance: %8.2f,   Ending Balance: %8.2f,    Net Gain:  %8.2f   (%3.1f%%)\n",
		s.cfg.InitFunds, endingC1Balance, netGain, pctGain)
	return str
}
