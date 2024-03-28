package newcore

import (
	"fmt"
	"os"
	"time"

	"github.com/stmansour/psim/util"
)

// This module generates the Financial Report for a simulation

// InvStats contains the information for columns in the financial report
type InvStats struct {
	Generation     int64   // what generation did this Investor appear in?
	DNA            string  // Investor's DNA
	FundStart      float64 // what was its initial fund
	PortfolioValue float64 // what was the Portfolio Value of the fund on the simulation stop date
}

// FinRep is the struct that contains context information for the Financial Report
type FinRep struct {
	InvStatsList []InvStats
	Sim          *Simulator
	file         *os.File
}

// GenerateFinRep generates the simulation's financial report.
// The report is generated as a CSV file
// -----------------------------------------------------------------------------
func (f *FinRep) GenerateFinRep(sim *Simulator, dirname string) error {
	var err error
	f.Sim = sim
	fname := f.Sim.generateFName("finrep")

	f.file, err = os.Create(fname)
	if err != nil {
		return err
	}
	defer f.file.Close()
	f.GenerateHeader()
	f.GenerateRows()

	return nil
}

// GenerateHeader prints the header lines to the csv file
// ----------------------------------------------------------
func (f *FinRep) GenerateHeader() error {
	fmt.Fprintf(f.file, "%q\n", "PLATO Simulator Financial Results")
	f.Sim.ReportHeader(f.file, true)
	return nil
}

// GenerateRows prints the header lines to the csv file
// ----------------------------------------------------------
func (f *FinRep) GenerateRows() error {
	c1b := fmt.Sprintf("C1 Balance (%s)", f.Sim.cfg.C1)
	c2b := fmt.Sprintf("C2 Balance (%s)", f.Sim.cfg.C2)
	cols := []string{
		"Rank",
		"Date",
		"Generation",
		"Portfolio Value",
		"Annualized Return",
		"Stop Loss Count",
		c1b,
		c2b,
		"DNA",
	}

	//------------------------------------------------------------------------
	// WRITE COLUMN HEADERS...
	//------------------------------------------------------------------------
	fmtstr := ""
	for i := 0; i < len(cols); i++ {
		fmtstr += "%q"
		if i != len(cols)-1 {
			fmtstr += ","
		}
	}
	fmtstr += "\n"
	args := make([]interface{}, len(cols))
	for i, v := range cols {
		args[i] = v
	}
	fmt.Fprintf(f.file, fmtstr, args...)

	//------------------------------------------------------------------------
	// WRITE COLUMNS...
	//------------------------------------------------------------------------
	for i, t := range f.Sim.TopInvestors {
		ar, err := util.AnnualizedReturn(f.Sim.cfg.InitFunds, t.PortfolioValue, time.Time(f.Sim.cfg.DtStart), time.Time(f.Sim.cfg.DtStop).AddDate(0, 0, 1))
		if err != nil {
			fmt.Printf("Error calculating annualized return: %s\n", err.Error())
		}
		fmt.Fprintf(f.file, "%d,%s,%d,%12.2f,%.2f,%d,%12.2f,%12.2f,%q\n",
			i+1,                       // rank
			t.DtPV.Format("1/2/2006"), // date
			t.GenNo,                   // generation number
			t.PortfolioValue,          // portfolio
			ar*100,                    // annualized return
			t.StopLossCount,           // count of stoploss invocations
			t.BalanceC1,               // C1
			t.BalanceC2,               // C2
			t.DNA,
		)
	}

	return nil
}
