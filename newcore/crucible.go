package newcore

import (
	"fmt"
	"os"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Crucible is the class that implements and reports on a crucible... a
// series of simulations on a list of Investor DNA
// ---------------------------------------------------------------------------
type Crucible struct {
	cfg   *util.AppConfig
	db    *newdata.Database
	mim   *MetricInfluencerManager
	sim   *Simulator
	idx   int    // index of currently running investor
	fname string // name of the crucible report file
	// file     *os.File // crucible file
}

// NewCrucible returns a pointer to a new crucible object
func NewCrucible() *Crucible {
	c := Crucible{}
	return &c
}

// Init initializes the crucible object
func (c *Crucible) Init(cfg *util.AppConfig, db *newdata.Database, mim *MetricInfluencerManager, sim *Simulator) {
	c.cfg = cfg
	c.db = db
	c.mim = mim
	c.sim = sim
	c.sim.cfg = cfg // required for generateFName
	c.fname = c.sim.generateFName("crep")
	file, err := os.Create(c.fname)
	if err != nil {
		fmt.Printf("error creating %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()
	fmt.Fprintf(file, "PLATO - Crucible Report\n")
	fmt.Fprintf(file, "\"Program Version:  %s\"\n", util.Version())
	fmt.Fprintf(file, "\"Configuration File:  %s\"\n", c.cfg.Filename)
	fmt.Fprintf(file, "\"Run Date: %s\"\n", time.Now().Format("Mon, Jan 2, 2006 - 15:04:05 MST"))
}

// Run sends the crucible report to a csv file
func (c *Crucible) Run() {
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
			c.sim.Init(c.cfg, c.db, c.mim, c, false, false)
			c.sim.Run()
		}
	}
	fmt.Printf("Crucible run completed\n")
	fmt.Printf("Output file is: %s\n", c.fname)
}

// SubHeader is used to identify a new DNA for the crucible
func (c *Crucible) SubHeader() {
	file, err := os.OpenFile(c.fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("error opening %s: %s\n", c.fname, err.Error())
		os.Exit(1)
	}
	defer file.Close()
	fmt.Fprintf(file, "\n\"DNA Name: %s\"\n", c.cfg.TopInvestors[c.idx].Name)
	fmt.Fprintf(file, "%q,%q,%q,%q,%q\n", "Start", "End", "Opening Portfolio Value", "Ending Portfolio Value", "Annualized Return")
}

// DumpResults sends the crucible report to a csv file
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
}
