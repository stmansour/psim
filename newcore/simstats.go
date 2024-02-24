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
func (s *Simulator) dumpGeneticParentMap(dirname string) error {
	fname := "dbgGeneticParentMap.csv"
	var file *os.File
	var err error
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

// DumpStats - dumps the top investor to a file after the simulation.
//
// RETURNS
//
//	any error encountered
//
// ----------------------------------------------------------------------------
func (s *Simulator) DumpStats(dirname string) error {
	fname := "simstats-" + s.ReportTimestamp + ".csv"
	if len(dirname) > 0 {
		fname = dirname + "/" + fname
	}
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
	fmt.Fprintf(file, "\"Program Version:  %s\"\n", util.Version())
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

	// s.influencersToCSV(file)
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
	// s.influencersToCSV(file)

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
			err := fmt.Errorf("*** ERROR *** s.db.Select error: %s", err.Error())
			return err.Error()
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
