package newcore

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/sqlt"
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
	Cfg                          *util.AppConfig        // system-wide configuration info
	factory                      Factory                // used to create Influencers
	db                           *newdata.Database      // database to use in this simulation
	crucible                     *Crucible              // must not be nil when cfg.CrucibleMode is true, pointer to crucible object
	SqltDB                       *sql.DB                // the sqlite3 database used for Investor ids
	ir                           *InvestorReport        // investment table of investors
	Investors                    []Investor             // the population of the current generation
	DayByDay                     bool                   // show day by day results, debug feature
	ReportTopInvestorInvestments bool                   // dump the investment list for top investor at the end of each generation
	maxProfitThisRun             float64                // the largest profit made by any investor during this simulation run
	maxPredictions               map[string]int         // max predictions indexed by subclass
	maxProfitInvestor            int                    // the investor that had the max profit for this generation
	GensCompleted                int                    // the current count of the number of generations completed in the simulation
	LoopsCompleted               int                    // the current loop being executed in the simulation
	GenStats                     []SimulationStatistics // keep track of what happened
	SimStart                     time.Time              // timestamp for simulation start
	SimStop                      time.Time              // timestamp for simulation stop
	StopTimeSet                  bool                   // set to true once SimStop is set. If it's false either the simulation is still in progress or did not complete
	WindDownInProgress           bool                   // initially false, set to true when we have a C2 balance on or after cfg.DtStop, when all C2 is sold this will return to being false
	FinRpt                       *FinRep                // Financial Report generator
	TopInvestors                 []TopInvestor          // the top n Investors across all generations
	ReportTimestamp              string                 // use this timestamp in the filenames we generate
	GenInfluencerDistribution    bool                   // show Influencer distribution for each generation
	FitnessScores                bool                   // save the fitness scores for each generation to dbgFitnessScores.csv
	T3ForThreadPool              time.Time              // timestamp to be used by thread pool
	WorkerThreads                int                    // number of worker threads in the thread pool
	TraceTiming                  bool                   // show the timing of the various parts of the simulation
	TrackingGenStart             time.Time              // the start time of the current generation
	TrackingGenStop              time.Time              // the stop time of the current generation
	Simtalkport                  int                    // the port on which the simulator is listening for external commands
	HashDuplicates               int64                  // the count of duplicate Investors encountered
}

// ResetSimulator is primarily to support tests. It resets the simulator
// object to its initial state for successive simulator runs from different
// test functions
// -----------------------------------------------------------------------------
func (s *Simulator) ResetSimulator() {
	var f Factory
	s.Cfg = nil
	s.db = nil
	s.factory = f
	s.Investors = nil
	s.DayByDay = false
	s.ReportTopInvestorInvestments = false
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
	s.Cfg = cfg
}

// GetFactory simply returns the simulator's factory
// -----------------------------------------------------------------------------
func (s *Simulator) GetFactory() *Factory {
	return &s.factory
}

// GetSimulationRunTime returns a printable string and a duration with the run
// time for this simulation
// ----------------------------------------------------------------------------
func (s *Simulator) GetSimulationRunTime() string {
	if !s.StopTimeSet {
		return "Simulation has not completed"
	}
	return util.ElapsedTime(s.SimStart, s.SimStop)
	// elapsed = s.SimStop.Sub(s.SimStart) // calculate elapsed time

	// return fmt.Sprintf("Simulation took %d hr, %d min, %d sec, and %d msec", int(elapsed.Hours()), int(elapsed.Minutes())%60, int(elapsed.Seconds())%60, int(elapsed.Milliseconds())%1000), elapsed
}

// Init initializes the simulation system, it also creates Investors and
// calls their init functions.
// ----------------------------------------------------------------------------
func (s *Simulator) Init(cfg *util.AppConfig, db *newdata.Database, crucible *Crucible, DayByDay, ReportTopInvestorInvestments bool) error {
	s.Cfg = cfg
	s.db = db
	s.crucible = crucible
	s.DayByDay = DayByDay
	s.ReportTopInvestorInvestments = ReportTopInvestorInvestments
	if s.crucible != nil {
		s.crucible.DayByDay = DayByDay
		s.crucible.ReportTopInvestorInvestments = ReportTopInvestorInvestments
	}
	s.ir = NewInvestorReport(s)
	s.factory.Init(s.Cfg, db, s.SqltDB, s)
	s.FinRpt = &FinRep{}

	if s.Cfg.PreserveElite {
		s.Cfg.EliteCount = int(s.Cfg.PreserveElitePct*float64(s.Cfg.PopulationSize)/100 + 0.5)
	}

	//------------------------------------------------------------------------
	// Create an initial population of investors with just 1 investor for now
	//------------------------------------------------------------------------
	var err error
	if err = s.NewPopulation(); err != nil {
		if err.Error() != "hash exists" {
			log.Panicf("*** ERROR ***  NewPopulation returned error: %s\n", err.Error())
		}
	}

	return nil
}

// CheckAndAddNewInvestor checks if the hash of the investor already exists
// in the database. If it does not exist, it will be inserted into Simulator's list
// of investors
// ------------------------------------------------------------------------------
func (s *Simulator) CheckAndAddNewInvestor(v *Investor) error {
	if !s.Cfg.AllowDuplicateInvestors {
		found, err := sqlt.CheckAndInsertHash(s.SqltDB, v.ID, v.Elite)
		if err != nil {
			return fmt.Errorf("error checking/inserting hash: %s", err)
		}
		if found {
			s.HashDuplicates++
			return fmt.Errorf("hash exists")
		}
	}
	s.Investors = append(s.Investors, *v)
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

		//------------------------------------------
		// introduce Gen0Elites if requested...
		//------------------------------------------
		gen0elites := 0
		if s.Cfg.Gen0Elites {
			gen0elites = len(s.Cfg.TopInvestors)
			for i := 0; i < gen0elites; i++ {
				v := s.factory.NewInvestorFromDNA(s.Cfg.TopInvestors[i].DNA)
				if err := s.CheckAndAddNewInvestor(&v); err != nil {
					if err.Error() == "hash exists" {
						log.Printf("duplicate hash: %s\n", v.ID)
					}
					return err
				}
			}
		}
		for i := 0; i < s.Cfg.PopulationSize-gen0elites; i++ {
			var v Investor
			if s.Cfg.SingleInvestorMode {
				v = s.factory.NewInvestorFromDNA(s.Cfg.SingleInvestorDNA)
			} else {
				v.Init(s.Cfg, &s.factory, s.db)
				v.EnsureID()
			}
			if err := s.CheckAndAddNewInvestor(&v); err != nil {
				if err.Error() == "hash exists" {
					log.Printf("duplicate hash: %s\n", v.ID)
				}
				return err
			}
		}
		return nil
	}

	//-----------------------------------------------------------------------
	// If we're in PreserveElite mode, save the elite members of the current
	// generation now.  The Investors have already been sorted so that the
	// top investors start at Investors[0]
	//-----------------------------------------------------------------------
	elite := []Investor{}
	if s.Cfg.PreserveElite {
		elite = s.Investors[0:s.Cfg.EliteCount]
		for i := 0; i < len(elite); i++ {
			elite[i].Elite = true
		}
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
	if s.Cfg.PreserveElite {
		//---------------------------------------------------------------------
		// They may be elite, but they cannot carry their balance forward :-)
		//---------------------------------------------------------------------
		for k := 0; k < len(elite); k++ {
			elite[k].BalanceC1 = s.Cfg.InitFunds
			elite[k].BalanceC2 = 0
			elite[k].PortfolioValueC1 = 0
		}
		//--------------------------------------
		// add the elites to the new population
		//--------------------------------------
		popCount := s.Cfg.PopulationSize - s.Cfg.EliteCount
		newPop = newPop[0:popCount]
		newPop = append(newPop, elite...)

		//----------------------------------------------------------------------
		// Now that the new population is set, we can remove the Elite flag.
		// The elites need to earn their spot each generation.
		//----------------------------------------------------------------------
		for k := 0; k < len(newPop); k++ {
			newPop[k].Elite = false
		}
	}
	if s.GenInfluencerDistribution {
		s.printNewPopStats(newPop)
	}

	s.Investors = newPop
	return nil
}

// SortInvestors calls on each investor to sort itself and its influencers
// in a consistent order.
// ----------------------------------------------------------------------------
func (s *Simulator) SortInvestors() {
	for _, v := range s.Investors {
		v.SortInfluencers()
	}
}

// workerPoolSize returns the number of CPU cores, which we will use for
// parallel processing. Using all the cpu cores is a reasonable default.
// This may change over time.
// ----------------------------------------------------------------------------
func (s *Simulator) workerPoolSize() int {
	numCPU := s.Cfg.WorkerPoolSize
	if s.Cfg.WorkerPoolSize < 1 {
		numCPU = runtime.NumCPU()
		if numCPU == 120 {
			numCPU = 25
		} else if numCPU > 10 {
			numCPU /= 2
		}
	}
	return numCPU
}

// worker is a goroutine that runs the dailyRun function for each Investor.
// We allocate one worker process for each available CPU core. We broadcast
// the index of each Investor to the workers over the tasks channel. One of
// the workers will receive the index and the Investor. That worker grabs it,
// runs the dailyRun function, and returns the result on the results channel.
// ----------------------------------------------------------------------------
func (s *Simulator) worker(tasks <-chan int, results chan<- error) {
	for j := range tasks {
		err := s.Investors[j].DailyRun(s.T3ForThreadPool, s.WindDownInProgress)
		results <- err
	}
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
	s.WorkerThreads = s.workerPoolSize() // for now, just use the number of CPU cores

	//-------------------------------------
	// ITERATE THROUGH THE LOOP COUNT...
	//-------------------------------------
	for lc := 0; lc < s.Cfg.LoopCount; lc++ {

		//---------------------------------
		// DO WE STILL NEED THIS?
		//---------------------------------
		for k, v := range s.Investors {
			if v.BalanceC1 > s.Cfg.InitFunds || v.BalanceC2 != 0 {
				fmt.Printf("Investor %d has C1 = %8.2f and C2 = %8.2f\n", k, v.BalanceC1, v.BalanceC2)
			}
		}

		dtStop := time.Time(s.Cfg.DtStop)
		// DateSettled = dtStop
		isGenDur := len(s.Cfg.GenDurSpec) > 0
		genStart := time.Time(s.Cfg.DtStart)

		if isGenDur {
			genDays := s.getGenerationDays(s.Cfg.GenDur)   // number of days in one generation
			totalDays := dtStop.Sub(genStart).Hours() / 24 // number of generations, rounding up
			s.Cfg.Generations = int(float64(totalDays) / float64(genDays))
			if float64(totalDays)/float64(genDays) > float64(s.Cfg.Generations) {
				s.Cfg.Generations++
			}
		}

		//-------------------------------------------------------------------------
		// Iterate through the GENERATIONS
		//-------------------------------------------------------------------------
		var d time.Time
		var dtGenEnd time.Time
		for g := 0; g < s.Cfg.Generations; g++ {
			dtGenStartTrace := time.Now()
			T3 := genStart
			if isGenDur {
				dtGenEnd = T3.AddDate(s.Cfg.GenDur.Years, s.Cfg.GenDur.Months, s.Cfg.GenDur.Weeks*7+s.Cfg.GenDur.Days) // end of this generation
			} else {
				dtGenEnd = dtStop
			}
			if dtGenEnd.After(dtStop) || !isGenDur {
				dtGenEnd = dtStop
			}
			EndOfDataReached = false

			// Let Investors now a new generation is starting...
			//-----------------------------------------------
			for _, v := range s.Investors {
				v.TraceInit()
			}

			//-------------------------------------------------------------------------
			// Iterate day-by-day through this generation's start to end dates...
			//-------------------------------------------------------------------------
			for T3.Before(dtGenEnd) || T3.Equal(dtGenEnd) || s.WindDownInProgress {
				iteration++

				if len(s.Investors) > s.Cfg.PopulationSize {
					log.Panicf("Population size should be %d, len(Investors) = %d", s.Cfg.PopulationSize, len(s.Investors))
				}

				//*********************** BEGIN SIMULATOR DAILY LOOP ***********************
				//-------------------------------------------------
				// worker pools setup for each day of simulation
				//-------------------------------------------------
				if len(s.Investors) < s.WorkerThreads {
					s.WorkerThreads = len(s.Investors) // but not more than number of investors
				}
				tasks := make(chan int, len(s.Investors))     // Send indices of s.Investors to workers, the channel isbuffered to avoid blocking, enough space for every Investor
				results := make(chan error, len(s.Investors)) // Collect errors or nil if successful

				// fire up the workers!
				for w := 0; w < s.WorkerThreads; w++ {
					go s.worker(tasks, results)
				}

				//---------------------------------------------------------------
				// Dispatch tasks (the intex of each Investor) to workers
				//---------------------------------------------------------------
				s.T3ForThreadPool = T3
				for j := range s.Investors {
					tasks <- j
				}
				close(tasks) // it can hold all the messages put in the channel, it will close when the last message has been handled

				//-----------------------------------------------
				// Wait until the last Investor has finished
				//-----------------------------------------------
				for a := 0; a < len(s.Investors); a++ {
					err := <-results // each time this returns it means that an Investor has finished
					if err != nil {
						fmt.Printf("Investors.DailyRun() returned: %s\n", err.Error())
					}
				}

				SettleC2 := 0 // if past simulation end date, we'll count the Investors that still have C2

				//-----------------------------------------------
				// Ask each investor to do their daily run...
				//-----------------------------------------------
				if s.WindDownInProgress {
					for j := 0; j < len(s.Investors); j++ {
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

				if !s.WindDownInProgress && !T3.Before(dtGenEnd) && !s.Cfg.EnforceStopDate {
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

			dtGenerationStop := time.Now()
			if s.TraceTiming {
				fmt.Printf("<<<TRACE TIMING>>> generation simulation time: %s\n", util.ElapsedTime(dtGenStartTrace, dtGenerationStop))
			}

			T3 = T3.AddDate(0, 0, -1)
			s.GensCompleted++ // we have just concluded another generation
			if g+1 == s.Cfg.Generations || !isGenDur {
				d = T3
			}
			thisGenDtStart = genStart
			thisGenDtEnd = d
			unsettled := float64(0)
			for j := 0; j < len(s.Investors); j++ {
				unsettled += s.Investors[j].BalanceC2
			}
			if !s.Cfg.CrucibleMode {
				fmt.Printf("Completed generation %d, %s - %s,  unsettled = %12.2f %s\n", s.GensCompleted, thisGenDtStart.Format("Jan _2, 2006"), d.Format("Jan _2, 2006"), unsettled, s.Cfg.C2)
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
			s.UpdateTopInvestors() // NOTE: s.Investors is sorted by Portfolio value upon return

			//---------------------------------------
			// End of generation reports...
			//---------------------------------------
			if s.Cfg.CrucibleMode && !s.Cfg.PredictionMode {
				s.crucible.DumpResults()
			}
			if s.ReportTopInvestorInvestments {
				if err := s.ir.dumpTopInvestorsDetail(); err != nil {
					log.Printf("ERROR: dumpTopInvestorsDetail returned: %s\n", err)
				}
			}
			if s.Cfg.Trace && !s.Cfg.CrucibleMode {
				for j := 0; j < len(s.Investors); j++ {
					s.Investors[j].TraceWriteFile()
				}
			}

			//----------------------------------------------------------------------------------------------
			// Now replace current generation with next generation unless this is the last generation...
			//----------------------------------------------------------------------------------------------
			if s.GensCompleted < s.Cfg.Generations || lc+1 < s.Cfg.LoopCount {
				if err := s.NewPopulation(); err != nil {
					log.Panicf("*** PANIC ERROR *** NewPopulation returned error: %s\n", err)
				}
				s.maxPredictions = make(map[string]int, 0)
				dtNextGenCompleted := time.Now()
				if s.TraceTiming {
					fmt.Printf("<<<TRACE TIMING>>> next generation create time: %s\n", util.ElapsedTime(dtGenerationStop, dtNextGenCompleted))
				}
			}
			s.WindDownInProgress = false
			dtGenStopTrace := time.Now()
			s.TrackingGenStart = dtGenStartTrace
			s.TrackingGenStop = dtGenStopTrace
		}
		if !s.Cfg.CrucibleMode {
			fmt.Printf("loop %d completed.  %s - %s\n", lc, thisGenDtStart.Format("Jan _2, 2006"), thisGenDtEnd.Format("Jan _2, 2006"))
		}
		s.LoopsCompleted++
	}

	//-------------------------------------------------
	// Finalize any reports that need bottom results
	//-------------------------------------------------

	s.SimStop = time.Now()
	s.StopTimeSet = true
}

// SetReportDirectory ensures that all the directory and file information for reports is
// set in s.Cfg.
// ----------------------------------------------------------------------------------------
func (s *Simulator) SetReportDirectory() {
	if !s.Cfg.ReportDirSet {
		if s.SimStart.Year() < 1900 {
			s.SimStart = time.Now()
		}
		s.Cfg.ReportTimestamp = s.SimStart.Format("2006-01-02T15-04-05.05.000000000")
		s.Cfg.ReportDirectory = s.Cfg.ArchiveBaseDir
		if s.Cfg.ArchiveMode {
			s.Cfg.ReportDirectory += "/" + s.Cfg.ReportTimestamp
		}
		if len(s.Cfg.ReportDirectory) > 0 {
			_, err := util.VerifyOrCreateDirectory(s.Cfg.ReportDirectory)
			if err != nil {
				log.Fatalf("Could not create directory %s, err = %s\n", s.Cfg.ReportDirectory, err.Error())
			}
		}
		s.Cfg.ReportDirSet = true
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
	newDir, err := s.CreateArchiveDirectory(s.Cfg.ReportDirectory)
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
		if strings.Contains(err.Error(), "not found") {
			return
		}
		log.Panicf("Error from SetAllPortfolioValues: %s\n", err.Error())
	}
	//----------------------------------------------------
	// Max investor profit needed for normalization...
	//----------------------------------------------------
	maxInvestorProfit := float64(-100000000) // a large negative amount
	for i := 0; i < len(s.Investors); i++ {
		profit := s.Investors[i].BalanceC1 - s.Cfg.InitFunds
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
		Locale:  s.Cfg.C1,
		Locale2: s.Cfg.C2,
	}
	// field := fmt.Sprintf("%s%sEXClose", s.Cfg.C1, s.Cfg.C2)
	// ss := []string{field}
	ss := []newdata.FieldSelector{}
	ss = append(ss, field)
	er, err := s.db.Select(t, ss) // exchange rate for C2 at time t
	if err != nil {
		log.Fatalf("Error getting exchange close rate")
	}
	if er == nil {
		return fmt.Errorf("th ExchangeRate record not found for t = %s and ss.Field[0] = %s", t.Format("2006-01-02"), ss[0].FQMetric())
	}
	exch := er.Fields[field.FQMetric()].Value // exchange rate at time t
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
