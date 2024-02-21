package newcore

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// SimulationStatistics contains relevant metrics for each generation simulated
// ------------------------------------------------------------------------------
type SimulationStatistics struct {
	ProfitableInvestors  int       // number of Investors that were profitable in this generation
	AvgProfit            float64   // avg profitability for profitable Investors in this generation
	MaxProfit            float64   // largest profit Investor in this generation
	TotalBuys            int       // total number of "buy" decisions made by the investor
	ProfitableBuys       int       // total number of buys that were profitable
	MaxProfitDNA         string    // DNA of the Investor making the highest profit this generation
	TotalNilDataRequests int       // total number of nildata errors that occurred across all Influencers
	DtGenStart           time.Time // first date of this generation
	DtGenStop            time.Time // target last date of this generation
	TotalHoldingC2       int       // total number of Investors still holding C2 after simulation stop date
	DtActualStop         time.Time // the date we actually stopped the simulation after trying to settle remaining C2 after DtGenStop
	UnsettledC2          float64   // the amount of C2 held across all Investors when simulation stopped.
	EndOfDataReached     bool      // true if current day was reached before all C2 was sold
}

// TopInvestor maintains the subset of information we need to keep for top investors
// in order to generate the financial report.
// ------------------------------------------------------------------------------------
type TopInvestor struct {
	DtPV           time.Time // the date all C2 was settled if settled after the simalation end date
	PortfolioValue float64   // value of the Investor's funds on the day the simulation ended
	DNA            string    // DNA string to recreate this Investor
	GenNo          int       // which generation did this Investor come from
	BalanceC1      float64   // Investor's C1 balance on simulation end date
	BalanceC2      float64   // Investor's C2 balance on simulation end date
}

// Simulator is a simulator object
type Simulator struct {
	cfg                        *util.AppConfig          // system-wide configuration info
	factory                    Factory                  // used to create Influencers
	db                         *newdata.Database        // database to use in this simulation
	mim                        *MetricInfluencerManager // metric influencer manager
	Investors                  []Investor               // the population of the current generation
	dayByDay                   bool                     // show day by day results, debug feature
	dumpTopInvestorInvestments bool                     // dump the investment list for top investor at the end of each generation
	maxProfitThisRun           float64                  // the largest profit made by any investor during this simulation run
	maxPredictions             map[string]int           // max predictions indexed by subclass
	maxProfitInvestor          int                      // the investor that had the max profit for this generation
	maxFitnessScore            float64                  // maximum fitness score seen in this generation
	GensCompleted              int                      // the current count of the number of generations completed in the simulation
	SimStats                   []SimulationStatistics   // keep track of what happened
	SimStart                   time.Time                // timestamp for simulation start
	SimStop                    time.Time                // timestamp for simulation stop
	StopTimeSet                bool                     // set to true once SimStop is set. If it's false either the simulation is still in progress or did not complete
	WindDownInProgress         bool                     // initially false, set to true when we have a C2 balance on or after cfg.DtStop, when all C2 is sold this will return to being false
	FinRpt                     *FinRep                  // Financial Report generator
	TopInvestors               []TopInvestor            // the top n Investors across all generations
	ReportTimestamp            string                   // use this timestamp in the filenames we generate
}

// ResetSimulator is primarily to support tests. It resets the simulator
// object to its initial state for successive simulator runs from different
// test functions
// -----------------------------------------------------------------------------
func (s *Simulator) ResetSimulator() {
	var f Factory
	s.cfg = nil
	s.mim = nil
	s.db = nil
	s.factory = f
	s.Investors = nil
	s.dayByDay = false
	s.dumpTopInvestorInvestments = false
	s.maxProfitThisRun = 0
	s.maxPredictions = make(map[string]int)
	s.maxProfitInvestor = 0
	s.maxFitnessScore = 0
	s.GensCompleted = 0
	s.SimStats = make([]SimulationStatistics, 0)
	s.StopTimeSet = false
	s.WindDownInProgress = false
}

// SetAppConfig simply sets the simulators pointer to the AppConfig struct
// -----------------------------------------------------------------------------
func (s *Simulator) SetAppConfig(cfg *util.AppConfig) {
	s.cfg = cfg
}

// GetFactory simply returns the simulator's factory
// -----------------------------------------------------------------------------
func (s *Simulator) GetFactory() *Factory {
	return &s.factory
}

// GetSimulationRunTime returns a printable string and a duration with the run
// time for this simulation
// ----------------------------------------------------------------------------
func (s *Simulator) GetSimulationRunTime() (string, time.Duration) {
	var elapsed time.Duration
	if !s.StopTimeSet {
		return "Simulation has not completed", elapsed
	}
	elapsed = s.SimStop.Sub(s.SimStart) // calculate elapsed time

	return fmt.Sprintf("Simulation took %d hours, %d minutes and %d seconds", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60), elapsed
}

// Init initializes the simulation system, it also creates Investors and
// calls their init functions.
// ----------------------------------------------------------------------------
func (s *Simulator) Init(cfg *util.AppConfig, db *newdata.Database, mim *MetricInfluencerManager, dayByDay, dumpTopInvestorInvestments bool) error {
	s.cfg = cfg
	s.db = db
	s.mim = mim
	s.dayByDay = dayByDay
	s.dumpTopInvestorInvestments = dumpTopInvestorInvestments

	s.factory.Init(s.cfg, db, mim)
	s.FinRpt = &FinRep{}

	//------------------------------------------------------------------------
	// Create an initial population of investors with just 1 investor for now
	//------------------------------------------------------------------------
	var err error
	if err = s.NewPopulation(); err != nil {
		log.Panicf("*** ERROR ***  NewPopulation returned error: %s\n", err.Error())
	}

	return nil
}

// NewPopulation create a new population. If this is generation 0, it will be
// a random population.  If
//
// ----------------------------------------------------------------------------
func (s *Simulator) NewPopulation() error {
	//----------------------------------------------------------------------------
	// First generation is random.  Also, if an entire generation completed
	// and the max fitness score is 0, then treat it like the first generation...
	// In other words, just make it a random population.
	//----------------------------------------------------------------------------
	if s.GensCompleted == 0 || s.maxFitnessScore == 0 {
		s.Investors = make([]Investor, 0)
		for i := 0; i < s.cfg.PopulationSize; i++ {
			var v Investor
			if s.cfg.SingleInvestorMode {
				v = s.factory.NewInvestorFromDNA(s.cfg.SingleInvestorDNA)
			} else {
				v.ID = s.factory.GenerateInvestorID()
			}
			s.Investors = append(s.Investors, v)
		}
		return nil
	}

	//-----------------------------------------------------------------------
	// If we have run a full simulation cycle, the next generation is based
	// on the genetic algorithm.
	//-----------------------------------------------------------------------
	var err error
	var newPop []Investor

	if newPop, err = s.factory.NewPopulation(s.Investors); err != nil {
		log.Panicf("*** PANIC ERROR ***  NewPopulation returned error: %s\n", err)
	}
	s.Investors = newPop

	return nil
}

// Run loops through the simulation day by day, first handling any conversions
// from C2 to C1 on that day, and then having each Investor consult its
// Influencers and deciding whether or not to convert C1 to C2. At the end
// of each day it prints out a message indicating where it is at in the
// simulation and some indicators as to how things are progressing.
// ----------------------------------------------------------------------------
func (s *Simulator) Run() {
	var thisGenDtStart time.Time
	var thisGenDtEnd time.Time
	var EndOfDataReached bool
	now := time.Now()
	iteration := 0
	s.SimStart = time.Now()
	// var DateSettled time.Time

	for lc := 0; lc < s.cfg.LoopCount; lc++ {

		dtStop := time.Time(s.cfg.DtStop)
		// DateSettled = dtStop
		isGenDur := len(s.cfg.GenDurSpec) > 0
		genStart := time.Time(s.cfg.DtStart)

		if isGenDur {
			genDays := s.getGenerationDays(s.cfg.GenDur)   // number of days in one generation
			totalDays := dtStop.Sub(genStart).Hours() / 24 // number of generations, rounding up
			s.cfg.Generations = int(float64(totalDays) / float64(genDays))
			if float64(totalDays)/float64(genDays) > float64(s.cfg.Generations) {
				s.cfg.Generations++
			}
		}

		//-------------------------------------------------------------------------
		// Iterate day-by-day through the simulation.
		//-------------------------------------------------------------------------
		var d time.Time
		var dtGenEnd time.Time
		for g := 0; g < s.cfg.Generations; g++ {
			T3 := genStart
			if isGenDur {
				dtGenEnd = T3.AddDate(s.cfg.GenDur.Years, s.cfg.GenDur.Months, s.cfg.GenDur.Weeks*7+s.cfg.GenDur.Days) // end of this generation
			} else {
				dtGenEnd = dtStop
			}
			if dtGenEnd.After(dtStop) || !isGenDur {
				dtGenEnd = dtStop
			}
			EndOfDataReached = false

			for T3.Before(dtGenEnd) || T3.Equal(dtGenEnd) || s.WindDownInProgress {
				iteration++

				if len(s.Investors) > s.cfg.PopulationSize {
					log.Panicf("Population size should be %d, len(Investors) = %d", s.cfg.PopulationSize, len(s.Investors))
				}

				//*********************** BEGIN SIMULATOR DAILY LOOP ***********************

				SettleC2 := 0 // if past simulation end date, we'll count the Investors that still have C2
				//-----------------------------------------------
				// Ask each investor to do their daily run...
				//-----------------------------------------------
				for j := 0; j < len(s.Investors); j++ {
					err := s.Investors[j].DailyRun(T3, s.WindDownInProgress)
					if err != nil {
						fmt.Printf("Investors[%d].DailyRun() returned: %s\n", j, err.Error())
					}

					//------------------------------------------------------------------------------
					// If we're in MarkToEnd mode, check to see if all Investors have < 1.00 of C2
					//------------------------------------------------------------------------------
					if s.WindDownInProgress {
						if s.Investors[j].BalanceC2 > 1.00 {
							SettleC2++ // another investor needs to settle C2
						}
					}
				}

				//-------------------------------------------------------------------
				// Terminate MarkToEnd if all Investors have settled their C2
				//-------------------------------------------------------------------
				if s.WindDownInProgress && SettleC2 == 0 {
					s.WindDownInProgress = false
				}
				//*********************** END SIMULATOR DAILY LOOP ***********************

				//---------------------------------------------------------------------------
				// We may update the portfolio value later, but for now we'll always set it
				// on the last day of the simulation.
				//---------------------------------------------------------------------------
				if T3.Equal(dtGenEnd) {
					for k := 0; k < len(s.Investors); k++ {
						s.Investors[k].PortfolioValueC1 = s.Investors[k].PortfolioValue(T3)
						s.Investors[k].DtPortfolioValue = T3
					}
				}

				d = T3
				T3 = T3.AddDate(0, 0, 1)

				if !s.WindDownInProgress && !T3.Before(dtGenEnd) && !s.cfg.EnforceStopDate {
					//----------------------------------------------------------------------
					// if we get to this point, it means that we're about to drop out of
					// the daily loop. See if any investor is holding C2. If so, we will
					// set s.WindDownInProgress to true in order to sell all remaining C2.
					//----------------------------------------------------------------------
					for j := 0; j < len(s.Investors) && !s.WindDownInProgress; j++ {
						if s.Investors[j].BalanceC2 > 1 {
							s.WindDownInProgress = true
						}
					}
				}

				// Don't let it go past today's date.
				if !T3.Before(now) {
					s.WindDownInProgress = false
					EndOfDataReached = true
				}
			}
			T3 = T3.AddDate(0, 0, -1)
			s.GensCompleted++ // we have just concluded another generation
			if g+1 == s.cfg.Generations || !isGenDur {
				d = T3
			}
			thisGenDtStart = genStart
			thisGenDtEnd = d
			unsettled := float64(0)
			for j := 0; j < len(s.Investors); j++ {
				unsettled += s.Investors[j].BalanceC2
			}
			fmt.Printf("Completed generation %d, %s - %s,  unsettled = %12.2f %s\n", s.GensCompleted, thisGenDtStart.Format("Jan _2, 2006"), d.Format("Jan _2, 2006"), unsettled, s.cfg.C2)
			if isGenDur {
				genStart = dtGenEnd // Start next generation from the end of the last
			}

			//----------------------------------------------------------------------
			// Compute scores and stats
			//----------------------------------------------------------------------
			s.CalculateMaxVals()
			s.CalculateAllFitnessScores()
			s.SaveStats(thisGenDtStart, thisGenDtEnd, T3, EndOfDataReached)
			if s.dumpTopInvestorInvestments {
				if err := s.InvestmentsToCSV(&s.Investors[s.maxProfitInvestor]); err != nil {
					log.Printf("ERROR: InvestmentsToCSV returned: %s\n", err)
				}
			}
			s.UpdateTopInvestors() // used by financial report

			//----------------------------------------------------------------------------------------------
			// Now replace current generation with next generation unless this is the last generation...
			//----------------------------------------------------------------------------------------------
			if s.GensCompleted < s.cfg.Generations || lc+1 < s.cfg.LoopCount {
				if err := s.NewPopulation(); err != nil {
					log.Panicf("*** PANIC ERROR *** NewPopulation returned error: %s\n", err)
				}
				s.maxPredictions = make(map[string]int, 0)
			}
			s.WindDownInProgress = false
		}
		fmt.Printf("loop %d completed.  %s - %s\n", lc, thisGenDtStart.Format("Jan _2, 2006"), thisGenDtEnd.Format("Jan _2, 2006"))
	}

	s.SimStop = time.Now()
	s.StopTimeSet = true
	s.ReportTimestamp = s.SimStop.Format("2006-01-02T15-04-05")
}

// Helper function to calculate the total days in a generation.  It is not
// always the same because the GenerationDuration spec can include months
// which have a varying number of days. It can also span years which can
// introduce leap years which have an extra day.
//
// INPUTS
//
//	gd - the parsed generation duration struct
//
// --------------------------------------------------------------------------
func (s *Simulator) getGenerationDays(gd *util.GenerationDuration) int {
	dtTmp := time.Date(0, time.January, 0, 0, 0, 0, 0, time.UTC)       // Using a zero year for calculation purposes
	dtGenEnd := dtTmp.AddDate(gd.Years, gd.Months, gd.Weeks*7+gd.Days) // end date of this generation
	return int(dtGenEnd.Sub(dtTmp).Hours() / 24)
}

// SettleC2Balance - At the end of a simulation, we'll cash out all C2 for
//
//	the amount of C1 it gets on the last day of the simulation.
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
// func (s *Simulator) SettleC2Balance() {
// 	for i := 0; i < len(s.Investors); i++ {
// 		err := s.Investors[i].SettleC2Balance()
// 		if err != nil {
// 			log.Panicf("SettleC2Balance error from Investor %d: %s\n", i, err)
// 		}
// 	}
// }

// CalculateAllFitnessScores - calculates values over all the Influncers and Investors
//
//	that are needed to compute FitnesScores.
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) CalculateAllFitnessScores() {
	//----------------------------------------------------
	// Investor fitness scores. Then call each Influencer
	// to compute its score.
	//----------------------------------------------------
	max := float64(0)
	for i := 0; i < len(s.Investors); i++ {
		x := s.Investors[i].CalculateFitnessScore()
		if x > max {
			max = x
		}
		for j := 0; j < len(s.Investors[i].Influencers); j++ {
			s.Investors[i].Influencers[j].CalculateFitnessScore()
		}
	}
	s.maxFitnessScore = max
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

// UpdateTopInvestors saves the best investors in s.TopInvestors
// -----------------------------------------------------------------------------
func (s *Simulator) UpdateTopInvestors() {
	n := s.cfg.TopInvestorCount

	//----------------------------------------------------------------
	// First, sort Investors by PortfolioValueC1 in descending order
	//----------------------------------------------------------------
	sort.Slice(s.Investors, func(i, j int) bool {
		return s.Investors[i].PortfolioValueC1 > s.Investors[j].PortfolioValueC1
	})

	//------------------------------------------------------------------------------------------
	// Now, convert only the top n Investors (or less, if there are fewer than 'n' Investors)
	// to TopInvestors and append them to a new slice
	//------------------------------------------------------------------------------------------
	newTopInvestors := make([]TopInvestor, 0, n)
	for i := 0; i < len(s.Investors) && i < n; i++ {
		newTopInvestor := TopInvestor{
			DtPV:           s.Investors[i].DtPortfolioValue,
			PortfolioValue: s.Investors[i].PortfolioValueC1,
			BalanceC1:      s.Investors[i].BalanceC1,
			BalanceC2:      s.Investors[i].BalanceC2,
			DNA:            s.Investors[i].DNA(),
			GenNo:          s.GensCompleted,
		}
		newTopInvestors = append(newTopInvestors, newTopInvestor)
	}

	//---------------------------------------------------------------
	// Combine newTopInvestors with s.TopInvestors for comparison
	//---------------------------------------------------------------
	combinedTopInvestors := append(s.TopInvestors, newTopInvestors...)

	//------------------------------------------------------------------------------------------
	// Sort combinedTopInvestors by PortfolioValue in descending order, if needed
	// Note: Depending on your logic, you may only need to sort if s.TopInvestors were not already sorted
	//------------------------------------------------------------------------------------------
	sort.Slice(combinedTopInvestors, func(i, j int) bool {
		return combinedTopInvestors[i].PortfolioValue > combinedTopInvestors[j].PortfolioValue
	})

	//-----------------------------------------------------------------
	// Truncate combinedTopInvestors to keep only the top 'n' entries
	//-----------------------------------------------------------------
	if len(combinedTopInvestors) > n {
		combinedTopInvestors = combinedTopInvestors[:n]
	}
	s.TopInvestors = combinedTopInvestors // Update s.TopInvestors with the newly determined top 'n' performers
}

// SaveStats - dumps the top investor for the current generation into an array
//
//	to be used in SimStats.csv when the simulation completes
//
// INPUTS
//
//	dtStart, dtStop = the start and stop time of the generation being saved
//
// RETURNS
// ----------------------------------------------------------------------------
func (s *Simulator) SaveStats(dtStart, dtStop, dtSettled time.Time, eodr bool) {
	//----------------------------------------------------
	// Compute average investor profit this generation...
	//----------------------------------------------------
	prof := 0
	maxProfit := float64(0)
	avgProfit := float64(0)
	maxProfitDNA := ""
	totalHoldingC2 := 0
	totalC2 := float64(0)

	for i := 0; i < len(s.Investors); i++ {
		if s.Investors[i].BalanceC1 > s.cfg.InitFunds {
			prof++
			profit := s.Investors[i].BalanceC1 - s.cfg.InitFunds
			avgProfit += profit
			if profit > maxProfit {
				maxProfit = profit
				maxProfitDNA = s.Investors[i].DNA()
				s.maxProfitInvestor = i
			}
		}
		if s.Investors[i].BalanceC2 >= 1.0 {
			totalHoldingC2++
			totalC2 += s.Investors[i].BalanceC2
		}
	}
	if prof > 0 {
		avgProfit = avgProfit / float64(prof) // average profit among the profitable
	}

	// Compute the total number of nildata errors across all Influencers

	totNil := 0
	for j := 0; j < len(s.Investors); j++ {
		inf := s.Investors[j].Influencers
		for k := 0; k < len(inf); k++ {
			if inf[k].GetNilDataCount() > 0 {
				totNil += inf[k].GetNilDataCount()
			}
		}
	}

	//----------------------------------------------------
	// Compute details about Investor with max profit...
	//----------------------------------------------------
	idx := s.maxProfitInvestor
	tot := 0
	pro := 0
	for _, investment := range s.Investors[idx].Investments {
		if investment.Completed {
			tot++
		}

		//----------------------------------------------------------------
		// Note that when we sell, we try to sell at a loss first. So
		// this might not be a good way to determine profitable buys
		//----------------------------------------------------------------
		for j := 0; j < len(investment.Profitable); j++ {
			if investment.Profitable[j] {
				pro++
			}
		}
	}

	ss := SimulationStatistics{
		ProfitableInvestors:  prof,
		AvgProfit:            avgProfit,
		MaxProfit:            maxProfit,
		MaxProfitDNA:         maxProfitDNA,
		TotalBuys:            tot,
		ProfitableBuys:       pro,
		TotalNilDataRequests: totNil,
		DtGenStart:           dtStart,
		DtGenStop:            dtStop,
		TotalHoldingC2:       totalHoldingC2,
		DtActualStop:         dtSettled,
		UnsettledC2:          totalC2,
		EndOfDataReached:     eodr,
	}
	s.SimStats = append(s.SimStats, ss)
}

// DumpStats - dumps the top investor to a file after the simulation.
//
// RETURNS
//
//	any error encountered
//
// ----------------------------------------------------------------------------
func (s *Simulator) DumpStats() error {
	fname := "simstats-" + s.ReportTimestamp + ".csv"
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	et, _ := s.GetSimulationRunTime()
	a := time.Time(s.cfg.DtStart)
	b := time.Time(s.cfg.DtStop)
	c := b.AddDate(0, 0, 1)

	// context information
	fmt.Fprintf(file, "%q\n", "PLATO Simulator Results")
	fmt.Fprintf(file, "\"Configuration File:  %s\"\n", s.cfg.Filename)
	fmt.Fprintf(file, "\"Run Date: %s\"\n", time.Now().Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Start Date: %s\"\n", a.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Stop Date: %s\"\n", b.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))

	if s.cfg.SingleInvestorMode {
		fmt.Fprintf(file, "\"Single Investor Mode\"\n")
		fmt.Fprintf(file, "\"DNA: %s\"\n", s.cfg.SingleInvestorDNA)
	} else {
		fmt.Fprintf(file, "\"Generations: %d\"\n", s.GensCompleted)
		if len(s.cfg.GenDurSpec) > 0 {
			fmt.Fprintf(file, "\"Generation Lifetime: %s\"\n", util.FormatGenDur(s.cfg.GenDur))
		}
		fmt.Fprintf(file, "\"Simulation Loop Count: %d\"\n", s.cfg.LoopCount)
		fmt.Fprintf(file, "\"Simulation Time Duration: %s\"\n", util.DateDiffString(a, c))
	}
	fmt.Fprintf(file, "\"C1: %s\"\n", s.cfg.C1)
	fmt.Fprintf(file, "\"C2: %s\"\n", s.cfg.C2)

	fmt.Fprintf(file, "\"Population: %d\"\n", s.cfg.PopulationSize)
	fmt.Fprintf(file, "\"COA Strategy: %s\"\n", s.cfg.COAStrategy)

	s.influencersToCSV(file)
	// s.influencerMissingData(file)

	omr := float64(0)
	if s.factory.MutateCalls > 0 {
		omr = 100.0 * float64(s.factory.Mutations) / float64(s.factory.MutateCalls)
	}
	fmt.Fprintf(file, "\"Observed Mutation Rate: %6.3f%%\"\n", omr)
	fmt.Fprintf(file, "\"Elapsed Run Time: %s\"\n", et)
	fmt.Fprintf(file, "\"\"\n")

	// the header row   0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15
	fmt.Fprintf(file, "%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q\n",
		"Generation",             // 0
		"Gen Start",              // 1
		"Gen Stop",               // 2
		"Profitable Investors",   // 3
		"% Profitable Investors", // 4
		"Average Profit",         // 5
		"Max Profit",             // 6
		"Total Buys",             // 7
		"Profitable Buys",        // 8
		"% Profitable Buys",      // 9
		"Nil Data Requests",      // 10
		"Investors Holding C2",   // 11
		"Total Unsettled C2",     // 12
		"Actual Stop Date",       // 13
		"All Investors Settled",  // 14
		"DNA")                    // 15

	// investment rows
	for i := 0; i < len(s.SimStats); i++ {
		pctProfPred := float64(0)
		if s.SimStats[i].TotalBuys > 0 {
			pctProfPred = 100.0 * float64(s.SimStats[i].ProfitableBuys) / float64(s.SimStats[i].TotalBuys)
		}
		settled := "no"
		if !s.SimStats[i].EndOfDataReached {
			settled = "yes"
		}
		fmt.Fprintf(file, "%d,%q,%q,%d,%8.2f%%,%12.2f,%12.2f,%d,%d,%4.2f%%,%d,%d,%12.2f,%q,%q,%q\n",
			i, // 0
			s.SimStats[i].DtGenStart.Format("1/2/2006"),                                    // 1
			s.SimStats[i].DtGenStop.Format("1/2/2006"),                                     // 2
			s.SimStats[i].ProfitableInvestors,                                              // 3
			100.0*float64(s.SimStats[i].ProfitableInvestors)/float64(s.cfg.PopulationSize), // 4
			s.SimStats[i].AvgProfit,                                                        // 5
			s.SimStats[i].MaxProfit,                                                        // 6
			s.SimStats[i].TotalBuys,                                                        // 7
			s.SimStats[i].ProfitableBuys,                                                   // 8
			pctProfPred,                                                                    // 9
			s.SimStats[i].TotalNilDataRequests,                                             // 10
			s.SimStats[i].TotalHoldingC2,                                                   // 11
			s.SimStats[i].UnsettledC2,                                                      // 12
			s.SimStats[i].DtActualStop.Format("1/2/2006"),                                  // 13
			settled,                    // 14
			s.SimStats[i].MaxProfitDNA) // 15
	}
	return nil
}

// InvestmentsToCSV - dumps the top investor to a file after the simulation.
//
// RETURNS
//
//	any error encountered
//
// ----------------------------------------------------------------------------
func (s *Simulator) InvestmentsToCSV(inv *Investor) error {
	gen := s.GensCompleted - 1 // the generation number has already been incremented
	fname := fmt.Sprintf("IList-Gen-%d.csv", gen)
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	a := time.Time(s.cfg.DtStart)
	b := time.Time(s.cfg.DtStop)
	c := b.AddDate(0, 0, 1)
	//------------------------------------------------------------------------
	// context information
	//------------------------------------------------------------------------
	fmt.Fprintf(file, "%q\n", "PLATO Simulator - Investor Investment List")
	fmt.Fprintf(file, "\"Configuration File:  %s\"\n", s.cfg.Filename)
	fmt.Fprintf(file, "\"Run Date: %s\"\n", time.Now().Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Start Date: %s\"\n", a.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Stop Date: %s\"\n", c.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Loop Count: %d\"\n", s.cfg.LoopCount)
	fmt.Fprintf(file, "\"Simulation Settle Date: %s\"\n", s.cfg.DtSettle.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"C1: %s\"\n", s.cfg.C1)
	fmt.Fprintf(file, "\"C2: %s\"\n", s.cfg.C2)
	fmt.Fprintf(file, "\"Generation: %d\"\n", gen)
	fmt.Fprintf(file, "\"Initial Funds: %10.2f\"\n", s.cfg.InitFunds)
	fmt.Fprintf(file, "\"Ending Funds: %10.2f %s\"\n", inv.BalanceC1, inv.cfg.C1)
	// fmt.Fprintf(file, "\"Settled Funds: %10.2f %s  (converted to C1 due to simulation end prior to T4)\"\n", inv.BalanceSettled, inv.cfg.C1)
	fmt.Fprintf(file, "\"Random Seed: %d\"\n", s.cfg.RandNano)
	fmt.Fprintf(file, "\"COA Strategy: %s\"\n", s.cfg.COAStrategy)

	//------------------------------------------------------------------------
	// Influencers for this investor.
	//------------------------------------------------------------------------
	s.influencersToCSV(file)

	// the header row                                         0          1                2        3                  4                     5                      6                 7                 8             9       10      11                 12
	fmt.Fprintf(file, "%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q\n", "T3", "Exchange Rate (T3)", "T4", "Exchange Rate (T4)", "Purchase Amount C1", "Purchase Amount (C2)", "BalanceC1 (T3)", "BalanceC2 (T3)", "T4 C2 Sold", "T4 C1", "Gain", "Balance C1 (T4)", "Balance C2 (T4)")

	// investment rows
	for i := 0; i < len(inv.Investments); i++ {
		m := inv.Investments[i]
		//                 0  1     2  3     4     5     6     7     8      9    10
		//                 t3       t4       t3c1  buyc2 sellc2 balc1 balc2   t4c1  net
		fmt.Fprintf(file, "%s,%12.2f,%s,%12.2f,%12.2f,%12.2f,%12.2f,%12.2f,%12.2f,%12.2f,%12.2f,%12.2f,%12.2f\n",
			m.T3.Format("1/2/2006"), // 0 - date on which purchase of C2 was made
			m.ERT3,                  // 1 - the exchange rate on T3
			m.T4.Format("1/2/2006"), // 2 - date on which C2 will be exchanged for C1
			m.ERT4,                  // 3 - the exchange rate on T4
			m.T3C1,                  // 4 - amount of C1 exchanged for C2 on T3
			m.T3C2Buy,               // 5 - the amount of currency in C2 that T3C1 purchased on T3
			m.T3BalanceC1,           // 6 - C1 balance after exchange on T3
			m.T3BalanceC2,           // 7 - C2 balance after exchange on T3
			m.T4C2Sold,              // 8 - for now, this is always going to be the same as T3C2Buy
			m.T4C1,                  // 9 - amount of currency C1 we were able to purchase with C2 on T4 at exchange rate ERT4
			m.T4C1-m.T3C1,           // 10 - profit (or loss if negative) on this investment
			m.T4BalanceC1,           // 11 - C1 balance after exchange on T4
			m.T4BalanceC2,           // 12 - C2 balance after exchange on T4
		)
	}
	return nil
}

// influencersToCSV - single place to call to dump Influencers to CSV file
// ---------------------------------------------------------------------------
func (s *Simulator) influencersToCSV(file *os.File) {
	t := "Influencers: "
	fmt.Fprintf(file, "%s", t)
	n := len(t)
	namesThisLine := 0
	for i := 0; i < len(util.InfluencerSubclasses); i++ {
		subclass := util.InfluencerSubclasses[i]
		if namesThisLine > 0 {
			fmt.Fprintf(file, " ")
			n++
		}
		if n+len(subclass) > 77 {
			t = "        "
			fmt.Fprintf(file, "\n%s", t)
			n = len(t)
			namesThisLine = 0
		}
		fmt.Fprintf(file, "%s", subclass)
		n += len(subclass)
		namesThisLine++
	}
	fmt.Fprintf(file, "\n\n")
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
	return nil
}

// ResultsByInvestor - dumps results of each investor
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------
func (s *Simulator) ResultsByInvestor() {
	largestBalance := -100000000.0 // a very low number
	profitable := 0                // count of profitable investors in this population
	idx := -1

	for i := 0; i < len(s.Investors); i++ {
		fmt.Printf("Investor %3d. DNA: %s\n", i, s.Investors[i].DNA())
		fmt.Printf("%s\n", s.ResultsForInvestor(i, &s.Investors[i]))
		if s.Investors[i].BalanceC1 > s.cfg.InitFunds {
			profitable++
		}
		if s.Investors[i].BalanceC1 > largestBalance {
			idx = i
			largestBalance = s.Investors[i].BalanceC1
		}
	}
	fmt.Printf("-------------------------------------------------------------------------\n")
	fmt.Printf("Profitable Investors:  %d / %d  (%6.3f%%)\n", profitable, s.cfg.PopulationSize, float64(profitable*100)/float64(s.cfg.PopulationSize))
	fmt.Printf("Best Performer:  Investor %d.  Ending balance = %12.2f %s\n", idx, largestBalance, s.cfg.C1)
}

// ResultsForInvestor - dumps results of investor [i]
//
// ADD: % correct predictions
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
			amt += v.Investments[j].T3C2Buy
		}
	}

	str := ""

	//-------------------------------------------------------------------------
	// Convert amt to C1 currency on the day of the simulation end...
	// store in:   c1Amt
	//-------------------------------------------------------------------------
	if amt > 0 {
		fld := s.cfg.C1 + s.cfg.C2 + "EXClose"
		ss := []string{fld}
		er4, err := s.db.Select(dt, ss) // get the exchange rate on t4
		if err != nil {

		}
		if er4 == nil {
			err := fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found", dt.Format("1/2/2006"))
			return err.Error()
		}
		c1Amt = amt / er4.Fields[fld]
		str += fmt.Sprintf("Pending Investments: %d, value: %12.2f %s  =  %12.2f %s\n", pending, amt, s.cfg.C2, c1Amt, s.cfg.C1)
	}
	str += fmt.Sprintf("\t\tInitial Stake: %12.2f %s,  End Balance: %12.2f %s\n", s.cfg.InitFunds, s.cfg.C1, v.BalanceC1+c1Amt, s.cfg.C1)

	endingC1Balance := c1Amt + v.BalanceC1
	netGain := endingC1Balance - s.cfg.InitFunds
	pctGain := netGain / s.cfg.InitFunds
	str += fmt.Sprintf("\t\tNet Gain:  %12.2f %s  (%3.3f%%)\n", netGain, s.cfg.C1, pctGain)

	//-----------------------------------z--------------------------------------
	// When this investor made a buy prediction, how often was it correct...
	//-------------------------------------------------------------------------
	m := 0 // number of times the prediction was "correct" (resulted in a profit)
	for i := 0; i < len(v.Investments); i++ {
		for _, p := range v.Investments[i].Profitable {
			if p {
				m++
			}
		}
	}
	str += fmt.Sprintf("\t\tPrediction Accuracy:  %d / %d  = %3.3f%%\n", m, len(v.Investments), (float64(m*100) / float64(len(v.Investments))))

	str += fmt.Sprintf("\t\tFitness Score:       %6.2f\n", v.CalculateFitnessScore())
	str += "\t\tInfluencer Fitness Scores:\n"
	for i := 0; i < len(v.Influencers); i++ {
		str += fmt.Sprintf("\t\t    %d: [%s] %6.2f\n", i, v.Influencers[i].DNA(), v.Influencers[i].CalculateFitnessScore())
	}

	return str
}
