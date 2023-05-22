package core

import (
	"fmt"
	"psim/data"
	"psim/util"
	"time"
)

// Simulator is a simulator object
type Simulator struct {
	cfg              *util.AppConfig // system-wide configuration info
	Investors        []Investor      // the population of the current generation
	dayByDay         bool            // show day by day results, debug feature
	invTable         bool            // dump the investment table at the end of the simulation
	maxProfitThisRun float64         // the largest profit made by any investor during this simulation run
	maxPredictions   map[string]int  // max predictions indexed by subclass
}

// Init initializes the simulation system, it also creates Investors and
// calls their init functions.
// ----------------------------------------------------------------------------
func (s *Simulator) Init(cfg *util.AppConfig, dayByDay, invTable bool) error {
	s.cfg = cfg
	s.dayByDay = dayByDay
	s.invTable = invTable

	//------------------------------------------------------------------------
	// Create an initial population of investors with just 1 investor for now
	//------------------------------------------------------------------------
	for i := 0; i < s.cfg.PopulationSize; i++ {
		var v Investor
		s.Investors = append(s.Investors, v)
	}

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

	//-------------------------------------------------------------------------
	// Iterate day-by-day through the simulation.
	//-------------------------------------------------------------------------
	iteration := 0
	for dtStop.After(dt) || dtStop.Equal(dt) {
		iteration++
		SellCount := 0
		BuyCount := 0

		//-----------------------------------------
		// Call SellConversion for each investor
		//-----------------------------------------
		for j := 0; j < len(s.Investors); j++ {
			sc, err := (&s.Investors[j]).SellConversion(dt)
			if err != nil {
				fmt.Printf("SellConversion returned: %s\n", err.Error())
			}
			SellCount += sc
		}

		//-----------------------------------------
		// Call BuyConversion for each investor
		//-----------------------------------------
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

		//============== DEBUG --------------------------------------------------------
		if s.dayByDay {
			count := 0
			invPending := 0
			for j := 0; j < len(s.Investors); j++ {
				if s.Investors[j].BalanceC1 > 0 {
					count++
				}
				for k := 0; k < len(s.Investors[j].Investments); k++ {
					if !s.Investors[j].Investments[k].Completed {
						invPending++
					}
				}
			}
			fmt.Printf("%4d. Date: %s, Buys: %d, Sells %d,\n      investors remaining: %d, investments pending: %d\n",
				iteration, dt.Format("2006-Jan-02"), BuyCount, SellCount, count, invPending)
		}
		//============== DEBUG --------------------------------------------------------

		dt = dt.AddDate(0, 0, 1)
	}

	//----------------------------------------------------------------------
	// Compute fitness scores...
	//----------------------------------------------------------------------
	s.CalculateMaxVals()

	//====================== DEBUG =========================================
	util.DPrintf("maxProfitThisRun = %8.2f\n", s.maxProfitThisRun)
	util.DPrintf("Max Buy Predictions by subclass:\n")
	for key, value := range s.maxPredictions {
		util.DPrintf("\t%20s: %6d\n", key, value)
	}
	//====================== DEBUG =========================================

	s.CalculateInvestorFitnessScores()

}

// CalculateInvestorFitnessScores - calculates values over all the Influncers and Investors
//
//	that are needed to compute FitnesScores.
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) CalculateInvestorFitnessScores() {
	//----------------------------------------------------
	// Investor fitness scores
	//----------------------------------------------------
	util.DPrintf("Investor Fitness Scores\n")
	for i := 0; i < len(s.Investors); i++ {
		score := s.Investors[i].FitnessScore()
		util.DPrintf("%3d.  %5.3f\n", i, score)
		for j := 0; j < len(s.Investors[i].Influencers); j++ {
			iscore := s.Investors[i].Influencers[j].FitnessScore()
			util.DPrintf("      %s: %5.3f\n", s.Investors[i].Influencers[j].Subclass(), iscore)
		}
	}
}

// CalculateMaxVals - calculates values over all the Influncers and Investors
//
//	that are needed to compute FitnesScores.
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) CalculateMaxVals() {

	//----------------------------------------------------
	// Max investor profit needed for normalization...
	//----------------------------------------------------
	maxInvestorProfit := float64(-100000000) // a large negative amount
	for i := 0; i < len(s.Investors); i++ {
		profit := s.Investors[i].BalanceC1 - s.cfg.InitFunds
		if profit > maxInvestorProfit {
			maxInvestorProfit = profit
		}
	}
	s.maxProfitThisRun = maxInvestorProfit

	//-------------------------------------------------------
	// Max number of buy recommendations that can be indexed
	// by Influencer subclass...
	// a map where keys are strings and values are float64
	//-------------------------------------------------------
	maxBuyRecommendations := make(map[string]int)
	for i := 0; i < len(s.Investors); i++ {
		for j := 0; j < len(s.Investors[i].Influencers); j++ {
			subclass := s.Investors[i].Influencers[j].Subclass()                  // subclass of this investor
			buyPredictions := s.Investors[i].Influencers[j].GetLenMyPredictions() // only "buy" predictions are saved
			if value, exists := maxBuyRecommendations[subclass]; exists {
				// value exists, see if buyPredictions is larger
				if buyPredictions > value {
					maxBuyRecommendations[subclass] = buyPredictions // this is the largest so far
				}
			} else {
				maxBuyRecommendations[subclass] = buyPredictions // this is the initial value in the map
			}
		}

	}
	s.maxPredictions = maxBuyRecommendations

	//---------------------------------------------------------------------------
	// We need to let all the Investors know the maximum # of buy
	// recommendations during this cycle so that they can calculate their
	// fitness scores.  Additionally, they need the maxPredictions by subclass
	// so that their Influencers can calculate their fitness.
	// Here is where we give them that information...
	//---------------------------------------------------------------------------
	for i := 0; i < len(s.Investors); i++ {
		s.Investors[i].maxProfit = s.maxProfitThisRun
		s.Investors[i].maxPredictions = s.maxPredictions
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
	return s.Investors[topInvestorIdx].OutputInvestments(topInvestorIdx)
}

// ResultsByInvestor - dumps results of each investor
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) CalculateFitness() {

}

// ResultsByInvestor - dumps results of each investor
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) ResultsByInvestor() {
	var err error
	for i := 0; i < len(s.Investors); i++ {
		fmt.Printf("Investor %3d: %s\n", i, s.Investors[i].DNA())
		fmt.Printf("%s\n", s.ResultsForInvestor(i, &s.Investors[i]))
		if s.invTable {
			if err = s.Investors[i].OutputInvestments(i); err != nil {
				fmt.Printf("*** ERROR *** outputting investments for Investor[%d]: %s\n", i, err.Error())
			}
		}
	}
}

// ResultsForInvestor - dumps results of investor [i]
//
// INPUTS
//
//	n =      The index of this investor in the list
//	inv =    Pointer to the investor
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) ResultsForInvestor(n int, v *Investor) string {
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

	str := ""

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
		str += fmt.Sprintf("Pending Investments: %d, value: %8.2f %s  =  %8.2f %s\n", pending, amt, s.cfg.C2, c1Amt, s.cfg.C1)
	}
	str += fmt.Sprintf("\t\tInitial Stake: %8.2f %s,  End Balance: %8.2f %s\n", s.cfg.InitFunds, s.cfg.C1, v.BalanceC1+c1Amt, s.cfg.C1)

	endingC1Balance := c1Amt + v.BalanceC1
	netGain := endingC1Balance - s.cfg.InitFunds
	pctGain := netGain / s.cfg.InitFunds
	str += fmt.Sprintf("\t\tNet Gain:  %8.2f %s  (%3.3f%%)\n", netGain, s.cfg.C1, pctGain)

	//-------------------------------------------------------------------------
	// When this investor made a buy prediction, how often was it correct...
	//-------------------------------------------------------------------------
	m := 0 // number of times the prediction was "correct" (resulted in a profit)
	for i := 0; i < len(v.Investments); i++ {
		if v.Investments[i].Profitable {
			m++
		}
	}
	str += fmt.Sprintf("\t\tPrediction Accuracy:  %d / %d  = %3.3f%%\n", m, len(v.Investments), (float64(m*100) / float64(len(v.Investments))))

	str += fmt.Sprintf("\t\tFitness Score:       %6.2f\n", v.FitnessScore())
	str += fmt.Sprintf("\t\tInfluencer Fitness Scores:\n")
	for i := 0; i < len(v.Influencers); i++ {
		str += fmt.Sprintf("\t\t    %d: [%s] %6.2f\n", i, v.Influencers[i].DNA(), v.Influencers[i].FitnessScore())
	}

	return str
}
