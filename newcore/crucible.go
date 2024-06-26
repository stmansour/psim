package newcore

import (
	"fmt"
	"os"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
	"gonum.org/v1/gonum/stat"
)

// Crucible is the class that implements and reports on a crucible... a
// series of simulations on a list of Investor DNA
// ---------------------------------------------------------------------------
type Crucible struct {
	cfg                          *util.AppConfig
	db                           *newdata.Database
	sim                          *Simulator
	CreateDLog                   bool    // if true generate DNAlog
	dlog                         *DNALog // if DNALog is true
	idx                          int     // index of currently running investor
	jdx                          int     // index of current time span
	fname                        string  // name of the crucible report file
	ReportTopInvestorInvestments bool
	DayByDay                     bool
	AnnualizedReturnList         []float64 // annualized for each crucible period
	// The next field is best explained by example:
	//     "CruciblePeriods": [
	//         {"Index":  0, "Duration": "1w", "Ending": "yesterday"},
	//         {"Index":  1, "Duration": "2w", "Ending": "yesterday"},
	//         {"Index":  2, "Duration": "1y", "Ending": "yesterday"},
	//     ]
	// The list of day-by-day returns for period 0 is InvestorHistory[0].  The
	// annualized return for the 3rd day of period 0 is InvestorHistory[0][2].
	// The annualized return for the 7th day of the 1year period is InvestorHistory[2][6].
	InvestorHistory [][]float64 // history of day-by-day annualized returns indexed by crucible-period index.
}

// NewCrucible returns a pointer to a new crucible object
func NewCrucible() *Crucible {
	c := Crucible{}
	return &c
}

// Init initializes the crucible object
func (c *Crucible) Init(cfg *util.AppConfig, db *newdata.Database, sim *Simulator) {
	cfg.AllowDuplicateInvestors = true
	c.cfg = cfg
	c.db = db
	c.sim = sim
	c.sim.Cfg = cfg // required for generateFName
	c.sim.SetReportDirectory()
	c.sim.db = db
	c.fname = c.cfg.GenerateFName("crep")
	file, err := os.Create(c.fname)
	if err != nil {
		fmt.Printf("error creating %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()
	//---------------------------------------------------------------------------
	// crep report header...
	//---------------------------------------------------------------------------
	fmt.Fprintf(file, "\"PLATO - Crucible Report\"\n")
	c.sim.ReportHeader(file, false)

	//---------------------------------------------------------------------------
	// DNALog create spreadsheet and initialize column headers
	//---------------------------------------------------------------------------
	if c.CreateDLog {
		c.dlog = NewDNALog()
		c.dlog.Init(c, cfg, sim)
		c.dlog.WriteHeader()
		c.InvestorHistory = make([][]float64, len(c.cfg.CrucibleSpans))
	}
}

// Run sends the crucible report to a csv file
func (c *Crucible) Run() {
	ir := NewInvestorReport(c.sim)
	ir.Cru = c
	ir.CrucibleMode = true
	for i := 0; i < len(c.cfg.TopInvestors); i++ {
		c.idx = i // mark the investor we're testing
		c.SubHeader()
		for j := 0; j < len(c.cfg.CrucibleSpans); j++ {
			c.jdx = j // mark the time span we're testing
			c.sim.ResetSimulator()
			c.cfg.DtStart = util.CustomDate(c.cfg.CrucibleSpans[j].DtStart)
			c.cfg.DtStop = util.CustomDate(c.cfg.CrucibleSpans[j].DtStop)
			c.cfg.SingleInvestorDNA = c.cfg.TopInvestors[i].DNA
			c.cfg.SingleInvestorMode = true
			c.cfg.PopulationSize = 1
			c.cfg.LoopCount = 1
			c.cfg.Generations = 1
			c.sim.Init(c.cfg, c.db, c, c.DayByDay, c.ReportTopInvestorInvestments)
			c.sim.ir = ir // we need to override the simulators new creation of this with our ongoing report
			if c.CreateDLog {
				c.InvestorHistory[c.jdx] = make([]float64, 0)
				c.cfg.DNALog = true
			}
			c.sim.Run()
		}
		c.DumpSuccessCoefficient()
		if c.CreateDLog {
			c.dlog.WriteRow()
		}
	}
	//--------------------------------------------
	// Now do todays recommendation if requested...
	//--------------------------------------------
	if c.cfg.Recommendation {
		c.cfg.PredictionMode = true
		fmt.Printf("Today's recommendation\n")
		for i := 0; i < len(c.cfg.TopInvestors); i++ {
			c.cfg.DtStart = util.CustomDate(util.UTCDate(time.Now()))
			c.cfg.DtStop = c.cfg.DtStart
			c.idx = i // mark the investor we're testing
			c.sim.ResetSimulator()
			c.cfg.SingleInvestorDNA = c.cfg.TopInvestors[i].DNA
			c.cfg.SingleInvestorMode = true
			c.cfg.Trace = true
			c.cfg.PopulationSize = 1
			c.cfg.LoopCount = 1
			c.cfg.Generations = 1
			c.sim.Init(c.cfg, c.db, c, c.DayByDay, c.ReportTopInvestorInvestments)
			c.sim.ir = nil
			c.sim.Run()
		}
	}
	fmt.Printf("Crucible run completed\n")
	fmt.Printf("Crucible report is: %s\n", c.fname)
	if c.CreateDLog {
		fmt.Printf("DNA Log report is: %s\n", c.dlog.filename)
	}
}

// SaveInvestorPortfolioValue saves the annualized return value for the day in
// the current crucible time span
// --------------------------------------------------------------------------
func (c *Crucible) SaveInvestorPortfolioValue(ar float64) {
	c.InvestorHistory[c.jdx] = append(c.InvestorHistory[c.jdx], ar)
}

// SubHeader is used to identify a new DNA for the crucible.
//
//	This is called at the start of each DNA so we'll empty the list of
//	ROI and start over
//
// --------------------------------------------------------------------------
func (c *Crucible) SubHeader() {
	file, err := os.OpenFile(c.fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()
	fmt.Fprintf(file, "\n\"DNA Name: %s\",,,,,%q\n", c.cfg.TopInvestors[c.idx].Name, c.cfg.TopInvestors[c.idx].DNA)
	fmt.Fprintf(file, "%q,%q,%q,%q,%q\n", "Start", "End", "Opening Portfolio Value", "Ending Portfolio Value", "Annualized Return")

	c.AnnualizedReturnList = make([]float64, 0) // reset the list
}

// DumpResults sends the crucible report to a csv file.
//
//	This is called upon the completion of a generation.  So we'll save the annualized return
func (c *Crucible) DumpResults() {
	file, err := os.OpenFile(c.fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()
	dtStart := time.Time(c.cfg.DtStart)
	dtStop := time.Time(c.cfg.DtStop)
	pv := float64(0)
	roi := float64(0)
	if len(c.sim.Investors) > 0 {
		pv = c.sim.Investors[0].PortfolioValueC1
		roi, err = util.AnnualizedReturn(c.cfg.InitFunds, pv, dtStart, dtStop)
		if err != nil {
			fmt.Printf("error computing AnnualizedReturn: %s\n", err.Error())
			os.Exit(1)
		}
	}
	fmt.Fprintf(file, "%q,%q,%9.2f,%9.2f,%5.2f%%\n", dtStart.Format("1/2/2006"), dtStop.Format("1/2/2006"), c.cfg.InitFunds, pv, roi*100)
	c.AnnualizedReturnList = append(c.AnnualizedReturnList, roi)
}

// DumpSuccessCoefficient calculates the success coefficient and adds it to the report
// -------------------------------------------------------------------------------------
func (c *Crucible) DumpSuccessCoefficient() {
	file, err := os.OpenFile(c.fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()

	// Consistency coefficient = 1 - stddev(of all annualized returns)
	// mean annualized return = SUM(annualized returns) / COUNT(annualized returns)
	// SUCCESS coefficient = (mean annualized return)* consistency
	mean, stddev := stat.MeanStdDev(c.AnnualizedReturnList, nil)
	consistency := 1 - stddev
	sc := mean * consistency

	fmt.Fprintf(file, "%s  |||  mean: %.4f  stddev: %.4f   consistency: %.4f   success coefficient: %.4f\n", c.cfg.CrucibleName,
		mean, stddev, consistency*100, sc*100)

}
