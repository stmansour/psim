package newcore

import (
	"fmt"
	"log"
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
	StopLossCount        int       // how many times this investor invoked stoploss
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
	StopLossCount  int       // number of times the investor invoked StopLoss
}

// Simulator is a simulator object
type Simulator struct {
	cfg                        *util.AppConfig        // system-wide configuration info
	factory                    Factory                // used to create Influencers
	db                         *newdata.Database      // database to use in this simulation
	crucible                   *Crucible              // must not be nil when cfg.CrucibleMode is true, pointer to crucible object
	Investors                  []Investor             // the population of the current generation
	dayByDay                   bool                   // show day by day results, debug feature
	dumpTopInvestorInvestments bool                   // dump the investment list for top investor at the end of each generation
	maxProfitThisRun           float64                // the largest profit made by any investor during this simulation run
	maxPredictions             map[string]int         // max predictions indexed by subclass
	maxProfitInvestor          int                    // the investor that had the max profit for this generation
	GensCompleted              int                    // the current count of the number of generations completed in the simulation
	GenStats                   []SimulationStatistics // keep track of what happened
	SimStart                   time.Time              // timestamp for simulation start
	SimStop                    time.Time              // timestamp for simulation stop
	StopTimeSet                bool                   // set to true once SimStop is set. If it's false either the simulation is still in progress or did not complete
	WindDownInProgress         bool                   // initially false, set to true when we have a C2 balance on or after cfg.DtStop, when all C2 is sold this will return to being false
	FinRpt                     *FinRep                // Financial Report generator
	TopInvestors               []TopInvestor          // the top n Investors across all generations
	ReportTimestamp            string                 // use this timestamp in the filenames we generate
	GenInfluencerDistribution  bool                   // show Influencer distribution for each generation
	FitnessScores              bool                   // save the fitness scores for each generation to dbgFitnessScores.csv
}

// ResetSimulator is primarily to support tests. It resets the simulator
// object to its initial state for successive simulator runs from different
// test functions
// -----------------------------------------------------------------------------
func (s *Simulator) ResetSimulator() {
	var f Factory
	s.cfg = nil
	s.db = nil
	s.factory = f
	s.Investors = nil
	s.dayByDay = false
	s.dumpTopInvestorInvestments = false
	s.maxProfitThisRun = 0
	s.maxPredictions = make(map[string]int)
	s.maxProfitInvestor = 0
	// s.maxFitnessScore = 0
	s.GensCompleted = 0
	s.GenStats = make([]SimulationStatistics, 0)
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
func (s *Simulator) Init(cfg *util.AppConfig, db *newdata.Database, crucible *Crucible, dayByDay, dumpTopInvestorInvestments bool) error {
	s.cfg = cfg
	s.db = db
	s.crucible = crucible
	s.dayByDay = dayByDay
	s.dumpTopInvestorInvestments = dumpTopInvestorInvestments

	s.factory.Init(s.cfg, db)
	s.FinRpt = &FinRep{}

	if s.cfg.PreserveElite {
		s.cfg.EliteCount = int(s.cfg.PreserveElitePct*float64(s.cfg.PopulationSize)/100 + 0.5)
	}

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
// a random population.
// ----------------------------------------------------------------------------
func (s *Simulator) NewPopulation() error {
	//----------------------------------------------------------------------------
	// First generation is random.  Also, if an entire generation completed
	// and the max fitness score is 0, then treat it like the first generation...
	// In other words, just make it a random population.
	//----------------------------------------------------------------------------
	if s.GensCompleted == 0 {
		s.Investors = make([]Investor, 0)
		for i := 0; i < s.cfg.PopulationSize; i++ {
			var v Investor
			if s.cfg.SingleInvestorMode {
				v = s.factory.NewInvestorFromDNA(s.cfg.SingleInvestorDNA)
			} else {
				v.ID = s.factory.GenerateInvestorID()
				v.Init(s.cfg, &s.factory, s.db)
			}
			s.Investors = append(s.Investors, v)
		}
		return nil
	}

	//-----------------------------------------------------------------------
	// If we're in PreserveElite mode, save the elite members of the current
	// generation now.  The Investors have already been sorted so that the
	// top investors start at Investors[0]
	//-----------------------------------------------------------------------
	elite := []Investor{}
	if s.cfg.PreserveElite {
		elite = s.Investors[0:s.cfg.EliteCount]
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

	//-------------------------------------
	// Dump any reports requested...
	//-------------------------------------
	if s.FitnessScores {
		s.dumpFitnessScores()
	}
	if s.cfg.PreserveElite {
		//---------------------------------------------------------------------
		// They may be elite, but they cannot carry their balance forward :-)
		//---------------------------------------------------------------------
		for k := 0; k < len(elite); k++ {
			elite[k].BalanceC1 = s.cfg.InitFunds
			elite[k].BalanceC2 = 0
			elite[k].PortfolioValueC1 = 0
		}
		//--------------------------------------
		// add the elites to the new population
		//--------------------------------------
		popCount := s.cfg.PopulationSize - s.cfg.EliteCount
		newPop = newPop[0:popCount]
		newPop = append(newPop, elite...)
	}
	if s.GenInfluencerDistribution {
		s.printNewPopStats(newPop)
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
	s.SetReportDirectory()

	// var DateSettled time.Time

	for lc := 0; lc < s.cfg.LoopCount; lc++ {
		for k, v := range s.Investors {
			if v.BalanceC1 > s.cfg.InitFunds || v.BalanceC2 != 0 {
				fmt.Printf("Investor %d has C1 = %8.2f and C2 = %8.2f\n", k, v.BalanceC1, v.BalanceC2)
			}
		}

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
			if !s.cfg.CrucibleMode {
				fmt.Printf("Completed generation %d, %s - %s,  unsettled = %12.2f %s\n", s.GensCompleted, thisGenDtStart.Format("Jan _2, 2006"), d.Format("Jan _2, 2006"), unsettled, s.cfg.C2)
			}
			if isGenDur {
				genStart = dtGenEnd // Start next generation from the end of the last
			}

			//----------------------------------------------------------------------
			// Compute scores and stats
			//----------------------------------------------------------------------
			s.CalculateMaxVals(T3)
			s.CalculateAllFitnessScores()
			s.SaveStats(thisGenDtStart, thisGenDtEnd, T3, EndOfDataReached)

			//----------------------------------------------------------------------
			// handle any reports...
			//----------------------------------------------------------------------
			if s.dumpTopInvestorInvestments {
				if err := s.InvestmentsToCSV(&s.Investors[s.maxProfitInvestor]); err != nil {
					log.Printf("ERROR: InvestmentsToCSV returned: %s\n", err)
				}
			}
			s.UpdateTopInvestors() // used by financial report
			if s.cfg.CrucibleMode {
				s.crucible.DumpResults()
			}

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
		if !s.cfg.CrucibleMode {
			fmt.Printf("loop %d completed.  %s - %s\n", lc, thisGenDtStart.Format("Jan _2, 2006"), thisGenDtEnd.Format("Jan _2, 2006"))
		}
	}

	s.SimStop = time.Now()
	s.StopTimeSet = true
}

// SetReportDirectory ensures that all the directory and file information for reports is
// set in s.cfg.
// ----------------------------------------------------------------------------------------
func (s *Simulator) SetReportDirectory() {
	if !s.cfg.ReportDirSet {
		if s.SimStart.Year() < 1900 {
			s.SimStart = time.Now()
		}
		s.cfg.ReportTimestamp = s.SimStart.Format("2006-01-02T15-04-05.05.000000000")
		s.cfg.ReportDirectory = s.cfg.ArchiveBaseDir
		if s.cfg.ArchiveMode {
			s.cfg.ReportDirectory += "/" + s.cfg.ReportTimestamp
		}
		if len(s.cfg.ReportDirectory) > 0 {
			_, err := util.VerifyOrCreateDirectory(s.cfg.ReportDirectory)
			if err != nil {
				log.Fatalf("Could not create directory %s, err = %s\n", s.cfg.ReportDirectory, err.Error())
			}
		}
		s.cfg.ReportDirSet = true
	}
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

// CreateArchiveDirectory ensures that the directory exists
func (s *Simulator) CreateArchiveDirectory(baseDir string) (string, error) {
	newDir, err := util.VerifyOrCreateDirectory(baseDir)
	if err != nil {
		return "", fmt.Errorf("error creating archive directory: %s", err.Error())
	}
	return newDir, err
}

// ArchiveResults creates an archive directory if needed and copies the config file there
func (s *Simulator) ArchiveResults(configFilePath string) (string, error) {
	newDir, err := s.CreateArchiveDirectory(s.cfg.ReportDirectory)
	if err != nil {
		return newDir, err
	}
	err = util.FileCopy(configFilePath, newDir)
	if err != nil {
		return newDir, fmt.Errorf("error copying file: %s", err.Error())

	}
	return newDir, nil
}

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
	min := float64(99999999)
	max := float64(-99999999)
	for i := 0; i < len(s.Investors); i++ {
		x := s.Investors[i].CalculateFitnessScore()
		if x < min {
			min = x
		}
		if x > max {
			max = x
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
func (s *Simulator) CalculateMaxVals(t3 time.Time) {
	// set the portfolio values for all investors
	//----------------------------------------------------
	if err := s.SetAllPortfolioValues(t3); err != nil {
		log.Panicf("Error from SetAllPortfolioValues: %s\n", err.Error())
	}
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

	//---------------------------------------------------------------------------
	// We need to let all the Investors know the maximum # of buy
	// recommendations during this cycle so that they can calculate their
	// fitness scores.  Additionally, they need the maxPredictions by subclass
	// so that their Influencers can calculate their fitness.
	// Here is where we give them that information...
	//---------------------------------------------------------------------------
	for i := 0; i < len(s.Investors); i++ {
		s.Investors[i].maxProfit = s.maxProfitThisRun
		// s.Investors[i].maxPredictions = s.maxPredictions
	}
}

// SetAllPortfolioValues returns the value of the Investors portfolio at time t. The
// portfolio value is returned in terms of C1 and it is the current BalanceC1
// plus BalanceC2 converted to C1 at t.
// ------------------------------------------------------------------------------
func (s *Simulator) SetAllPortfolioValues(t time.Time) error {
	// First, get today's closing price
	field := newdata.FieldSelector{
		Metric:  "EXClose",
		Locale:  s.cfg.C1,
		Locale2: s.cfg.C2,
	}
	// field := fmt.Sprintf("%s%sEXClose", s.cfg.C1, s.cfg.C2)
	// ss := []string{field}
	ss := []newdata.FieldSelector{}
	ss = append(ss, field)
	er, err := s.db.Select(t, ss) // exchange rate for C2 at time t
	if err != nil {
		log.Fatalf("Error getting exchange close rate")
	}
	exch := er.Fields[field.FQMetric()] // exchange rate at time t
	if exch < 0.0001 {
		log.Panicf("exch = %12.6f\n", exch)
	}

	for i := 0; i < len(s.Investors); i++ {
		if s.Investors[i].BalanceC2 == 0 {
			continue
		}
		C2 := s.Investors[i].BalanceC2 / exch // amount of C1 we get for BalanceC2 at this exchange rate
		s.Investors[i].PortfolioValueC1 = s.Investors[i].BalanceC1 + C2
		s.Investors[i].DtPortfolioValue = t
	}
	return nil
}
