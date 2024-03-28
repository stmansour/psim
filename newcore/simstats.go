package newcore

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/stmansour/psim/util"
)

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
			StopLossCount:  s.Investors[i].StopLossCount,
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
		if s.Investors[i].PortfolioValueC1 > s.cfg.InitFunds {
			prof++
			profit := s.Investors[i].PortfolioValueC1 - s.cfg.InitFunds
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

	// Compute the total number of nildata errors across all Influencers.
	// Also compute the total number of stoploss invocations in the generation.
	totNil := 0
	stoploss := 0
	for j := 0; j < len(s.Investors); j++ {
		inf := s.Investors[j].Influencers
		stoploss += s.Investors[j].StopLossCount
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
	pro := 0
	for _, investment := range s.Investors[idx].Investments {
		//----------------------------------------------------------------
		// Note that when we sell, we try to sell at a loss first. So
		// this might not be a good way to determine profitable buys
		//----------------------------------------------------------------
		chunkProfit := float64(0)
		for j := 0; j < len(investment.Chunks); j++ {
			if investment.Chunks[j].Profitable {
				chunkProfit += investment.Chunks[j].ChunkProfit
			}
		}
		if chunkProfit > 0 {
			pro++
		}
	}

	ss := SimulationStatistics{
		ProfitableInvestors:  prof,
		AvgProfit:            avgProfit,
		MaxProfit:            maxProfit,
		MaxProfitDNA:         maxProfitDNA,
		TotalBuys:            len(s.Investors[idx].Investments),
		ProfitableBuys:       pro,
		TotalNilDataRequests: totNil,
		DtGenStart:           dtStart,
		DtGenStop:            dtStop,
		TotalHoldingC2:       totalHoldingC2,
		DtActualStop:         dtSettled,
		UnsettledC2:          totalC2,
		EndOfDataReached:     eodr,
		StopLossCount:        stoploss,
	}
	s.GenStats = append(s.GenStats, ss)
}

func (s *Simulator) generateFName(basename string) string {
	fname := ""
	if len(s.cfg.ReportDirectory) > 0 {
		fname = s.cfg.ReportDirectory + "/"
	}
	fname += basename
	if s.cfg.ArchiveMode {
		fname += s.cfg.ReportTimestamp
	}
	fname += ".csv"
	return fname
}

// dumpFitnessScores - dumps Investor fitness Scores for each generation
// to a file
//
// RETURNS
//
//	any error encountered
//
// ----------------------------------------------------------------------------
func (s *Simulator) dumpFitnessScores() error {
	var file *os.File
	var err error
	fname := s.generateFName("fitnessScores")
	if s.GensCompleted == 1 {
		file, err = os.Create(fname)
	} else {
		file, err = os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	if s.GensCompleted == 1 {
		fmt.Fprintf(file, "%q,%q,%q,%q,%q\n", "Generation", "Portfolio Value", "Fitness Score", "Parented", "DNA")
	}

	for _, inv := range s.Investors {
		fmt.Fprintf(file, "%d,%9.2f,%9.4f,%d,%q\n", s.GensCompleted, inv.PortfolioValueC1, inv.CalculateFitnessScore(), inv.Parented, inv.DNA())
	}
	return nil
}

// printNewPopStats - total up all the counts for different types of influencers
// and print.
func (s *Simulator) printNewPopStats(newpop []Investor) {
	m := map[string]int{}
	tot := 0
	for _, v := range newpop {
		for _, inf := range v.Influencers {
			m[inf.GetMetric()]++
		}
		tot += len(v.Influencers)
	}
	avg := float64(tot) / float64(len(newpop))
	fmt.Printf("------------------------------\n")
	fmt.Printf("New Population:  size: %d,  avg # Infl: %5.2f,   unique: %d\n", len(newpop), avg, len(m))

	// Create a slice to hold the keys
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort the keys slice
	sort.Strings(keys)

	// Iterate over the sorted keys to print key-value pairs
	fmt.Printf("%15s %s  %s\n", "Metric", "Count", "Percent")
	for _, k := range keys {
		count := m[k]
		fmt.Printf("%15s %5d %8.2f\n", k, count, float64(count*100)/float64(tot))
	}
	fmt.Printf("------------------------------\n")
}

// SimStats - dumps the top investor to a file after the simulation.
//
// RETURNS
//
//	any error encountered
//
// ----------------------------------------------------------------------------
func (s *Simulator) SimStats(d string) error {
	fname := s.generateFName("simstats")
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "%q\n", "PLATO Simulator Results")
	// s.influencersToCSV(file)
	// s.influencerMissingData(file)
	s.ReportHeader(file, true)

	// the header row   0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16
	fmt.Fprintf(file, "%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q\n",
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
		"Stop Loss",              // 10
		"Nil Data Requests",      // 11
		"Investors Holding C2",   // 12
		"Total Unsettled C2",     // 13
		"Actual Stop Date",       // 14
		"All Investors Settled",  // 15
		"DNA")                    // 16

	// investment rows
	for i := 0; i < len(s.GenStats); i++ {
		pctProfPred := float64(0)
		if s.GenStats[i].TotalBuys > 0 {
			pctProfPred = 100.0 * float64(s.GenStats[i].ProfitableBuys) / float64(s.GenStats[i].TotalBuys)
		}
		settled := "no"
		if !s.GenStats[i].EndOfDataReached {
			settled = "yes"
		}
		fmt.Fprintf(file, "%d,%q,%q,%d,%8.2f%%,%12.2f,%12.2f,%d,%d,%4.2f%%,%d,%d,%d,%12.2f,%q,%q,%q\n",
			i, // 0
			s.GenStats[i].DtGenStart.Format("1/2/2006"),                                    // 1
			s.GenStats[i].DtGenStop.Format("1/2/2006"),                                     // 2
			s.GenStats[i].ProfitableInvestors,                                              // 3
			100.0*float64(s.GenStats[i].ProfitableInvestors)/float64(s.cfg.PopulationSize), // 4
			s.GenStats[i].AvgProfit,                                                        // 5
			s.GenStats[i].MaxProfit,                                                        // 6
			s.GenStats[i].TotalBuys,                                                        // 7
			s.GenStats[i].ProfitableBuys,                                                   // 8
			pctProfPred,                                                                    // 9
			s.GenStats[i].StopLossCount,                                                    // 10
			s.GenStats[i].TotalNilDataRequests,                                             // 11
			s.GenStats[i].TotalHoldingC2,                                                   // 12
			s.GenStats[i].UnsettledC2,                                                      // 13
			s.GenStats[i].DtActualStop.Format("1/2/2006"),                                  // 14
			settled,                    // 15
			s.GenStats[i].MaxProfitDNA) // 16
	}
	return nil
}

// ReportHeader - prints out the header information for Plato reports.
// The caller should print the name of the report first, then call this method.
//
// INPUTS
//
//		file - the file to print to
//	 bSim - true if this is the simstats report
//
// -------------------------------------------------------------------------------
func (s *Simulator) ReportHeader(file *os.File, bSim bool) {
	et, _ := s.GetSimulationRunTime()
	a := time.Time(s.cfg.DtStart)
	b := time.Time(s.cfg.DtStop)
	c := b.AddDate(0, 0, 1)
	fmt.Fprintf(file, "\"Program Version:  %s\"\n", util.Version())
	fmt.Fprintf(file, "\"Run Date: %s\"\n", time.Now().Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Configuration File:  %s\"\n", s.cfg.Filename)
	if s.db.Datatype == "CSV" {
		fmt.Fprintf(file, "\"Database: %s\"\n", s.db.CSVDB.DBFname)
	} else {
		fmt.Fprintf(file, "\"Database: %s  (SQL)\"\n", s.db.SQLDB.Name)
	}
	fmt.Fprintf(file, "\"Simulation Start Date: %s\"\n", a.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Stop Date: %s\"\n", b.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	if bSim {
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
	}
	fmt.Fprintf(file, "\"C1: %s\"\n", s.cfg.C1)
	fmt.Fprintf(file, "\"C2: %s\"\n", s.cfg.C2)

	fmt.Fprintf(file, "\"Population: %d\"\n", s.cfg.PopulationSize)
	fmt.Fprintf(file, "\"Influencers: min %d,  max %d\"\n", s.cfg.MinInfluencers, s.cfg.MaxInfluencers)
	fmt.Fprintf(file, "\"Initial Funds: %.2f %s\"\n", s.cfg.InitFunds, s.cfg.C1)
	fmt.Fprintf(file, "\"Standard Investment: %.2f %s\"\n", s.cfg.StdInvestment, s.cfg.C1)
	fmt.Fprintf(file, "\"Stop Loss: %.2f%%\"\n", s.cfg.StopLoss*100)
	fmt.Fprintf(file, "\"Preserve Elite: %v  (%5.2f%%)\"\n", s.cfg.PreserveElite, s.cfg.PreserveElitePct)
	fmt.Fprintf(file, "\"Transaction Fee: %.2f (flat rate)  %5.1f bps\"\n", s.cfg.TxnFee, s.cfg.TxnFeeFactor*10000)
	fmt.Fprintf(file, "\"Investor Bonus Plan: %v\"\n", s.cfg.InvestorBonusPlan)
	fmt.Fprintf(file, "\"Gen 0 Elites: %v\"\n", s.cfg.Gen0Elites)

	omr := float64(0)
	if s.factory.MutateCalls > 0 {
		omr = 100.0 * float64(s.factory.Mutations) / float64(s.factory.MutateCalls)
	}
	fmt.Fprintf(file, "\"Observed Mutation Rate: %6.3f%%\"\n", omr)
	if !s.cfg.CrucibleMode {
		fmt.Fprintf(file, "\"Elapsed Run Time: %s\"\n", et)
	}
	fmt.Fprintf(file, "\"\"\n")
}
