package newcore

import (
	"fmt"
	"os"
	"time"
)

// InvestorReport is the class that implements and reports on an investor
type InvestorReport struct {
	s                     *Simulator
	ReportHeaderCompleted bool
	CrucibleMode          bool
	Cru                   *Crucible
}

// NewInvestorReport creates a new InvestorReport instance.
//
// It takes a pointer to a Simulator as a parameter and returns a pointer to an InvestorReport.
func NewInvestorReport(sim *Simulator) *InvestorReport {
	r := InvestorReport{
		s: sim,
	}
	return &r
}

// dumpTopInvestorsDetail - dumps Investor investments for the TopInvestors
// to a file
//
// RETURNS
//
//	any error encountered
//
// ----------------------------------------------------------------------------
func (ir *InvestorReport) dumpTopInvestorsDetail() error {
	var file *os.File
	var err error
	fname := ir.s.generateFName("invrep")
	if !ir.ReportHeaderCompleted {
		file, err = os.Create(fname)
	} else {
		file, err = os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	if !ir.ReportHeaderCompleted {
		ir.dumpInvestmentReportHeader(file)
		ir.ReportHeaderCompleted = true
	}
	lim := ir.s.cfg.TopInvestorCount
	if lim > len(ir.s.Investors) {
		lim = len(ir.s.Investors)
	}
	for i := 0; i < lim; i++ {
		inv := ir.s.Investors[i]
		name := inv.ID
		if ir.CrucibleMode && len(ir.s.cfg.TopInvestors[ir.Cru.idx].Name) > 0 {
			name = ir.s.cfg.TopInvestors[ir.Cru.idx].Name
		}
		fmt.Fprintf(file, "%d,%q,,,,,,,,,,,,,,,,%q\n", ir.s.GensCompleted, name, inv.DNA())
		for i := 0; i < len(inv.Investments); i++ {
			m := inv.Investments[i]
			//                   0  1      4      5             6      7
			//                   t3        t3c1   buyc2   fee   balc1 balc2
			fmt.Fprintf(file, ",,%s,%12.2f,%12.2f,%12.2f,%8.4f,%12.2f,%12.2f\n",
				m.T3.Format("1/2/2006"), // date on which purchase of C2 was made
				m.ERT3,                  // the exchange rate on T3
				m.T3C1,                  // amount of C1 exchanged for C2 on T3
				m.T3C2Buy,               // the amount of currency in C2 that T3C1 purchased on T3
				m.Fee,                   // fee to purchase
				m.T3BalanceC1,           // C1 balance after exchange on T3
				m.T3BalanceC2,           // C2 balance after exchange on T3
			)

			runningTotal := float64(0)
			for _, v := range m.Chunks {
				runningTotal += v.T4C1
				fmt.Fprintf(file, ",,,,,,,,,%s,%12.2f,%12.2f,%8.4f,%12.2f,%12.2f,%12.2f,%12.2f\n",
					v.T4.Format("1/2/2006"), // T4
					v.ERT4,                  // Exchange Rate on T4
					v.T4C2Sold,              // how much C2 was sold in this transaction
					v.Fee,                   // Fee for exchange
					v.T4C2Remaining,         // how much C2 is left
					v.T4C1,                  // amount of C1 we were able to purchase on T4 at exchange rate ERT4
					runningTotal,            // running total of C1 recovered by all exchanges
					v.ChunkProfit,           // profit or loss
				)
			}

		}
	}

	return nil
}

func (ir *InvestorReport) dumpInvestmentReportHeader(file *os.File) {
	a := time.Time(ir.s.cfg.DtStart)
	b := time.Time(ir.s.cfg.DtStop)
	c := b.AddDate(0, 0, 1)
	//------------------------------------------------------------------------
	// context information
	//------------------------------------------------------------------------
	fmt.Fprintf(file, "%q\n", "PLATO Simulator - Top Investor Investment Details")
	fmt.Fprintf(file, "\"Configuration File:  %s\"\n", ir.s.cfg.Filename)
	fmt.Fprintf(file, "\"Run Date: %s\"\n", time.Now().Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Start Date: %s\"\n", a.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	fmt.Fprintf(file, "\"Simulation Stop Date: %s\"\n", c.Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
	// fmt.Fprintf(file, "\"Simulation Loop Count: %d\"\n", s.cfg.LoopCount)
	fmt.Fprintf(file, "\"C1: %s\"\n", ir.s.cfg.C1)
	fmt.Fprintf(file, "\"C2: %s\"\n", ir.s.cfg.C2)
	fmt.Fprintf(file, "\"Initial Funds: %10.2f\"\n", ir.s.cfg.InitFunds)

	// the header row
	fmt.Fprintf(file, "%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q\n",
		"Generation", "Investor",
		"T3", "Exchange Rate (T3)", "Purchase Amount C1",
		"Purchase Amount (C2)", "Fee", "BalanceC1 (T3)", "BalanceC2 (T3)", "T4", "Exch Rate",
		"T4 C2", "Fee", "C2 Remaining", "C1", "Total C1", "Chunk Profit", "DNA")
}
