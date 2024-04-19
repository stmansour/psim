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
	idx                          int    // index of currently running investor
	fname                        string // name of the crucible report file
	ReportTopInvestorInvestments bool
	DayByDay                     bool
	AnnualizedReturnList         []float64 // as we list its returns, keep track of each one for computation of success coefficient
}

// NewCrucible returns a pointer to a new crucible object
func NewCrucible() *Crucible {
	c := Crucible{}
	return &c
}

// Init initializes the crucible object
func (c *Crucible) Init(cfg *util.AppConfig, db *newdata.Database, sim *Simulator) {
	c.cfg = cfg
	c.db = db
	c.sim = sim
	c.sim.Cfg = cfg // required for generateFName
	c.sim.SetReportDirectory()
	c.sim.db = db
	c.fname = c.sim.generateFName("crep")
	file, err := os.Create(c.fname)
	if err != nil {
		fmt.Printf("error creating %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()
	fmt.Fprintf(file, "\"PLATO - Crucible Report\"\n")
	c.sim.ReportHeader(file, false)
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
			c.sim.Run()
		}
		c.DumpSuccessCoefficient()

	}
	fmt.Printf("Crucible run completed\n")
	fmt.Printf("Output file is: %s\n", c.fname)
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
	roi, err := util.AnnualizedReturn(c.cfg.InitFunds, c.sim.Investors[0].PortfolioValueC1, dtStart, dtStop)
	if err != nil {
		fmt.Printf("error computing AnnualizedReturn: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(file, "%q,%q,%9.2f,%9.2f,%5.2f%%\n", dtStart.Format("1/2/2006"), dtStop.Format("1/2/2006"), c.cfg.InitFunds, c.sim.Investors[0].PortfolioValueC1, roi*100)
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
	// mean = SUM(annualized returns) / COUNT(annualized returns)
	// SUCCESS coefficient = (mean)* consistency

	mean, stddev := stat.MeanStdDev(c.AnnualizedReturnList, nil)
	consistency := 1 - stddev
	sc := mean * consistency
	fmt.Fprintf(file, "%s  |||  mean: %.4f  stddev: %.4f   consistency: %.4f   success coefficient: %.4f\n", c.cfg.CrucibleName,
		mean, stddev, consistency*100, sc*100)

}
