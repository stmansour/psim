package newcore

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// CourseOfAction encapsulates the elements of an Influencer's prediction
// ----------------------------------------------------------------------------
type CourseOfAction struct {
	Action     string
	ActionPct  float64
	BuyVotes   float64
	SellVotes  float64
	HoldVotes  float64
	TotalVotes float64
	Abstains   float64
}

// InvestmentStrategyMap links the name of the strategy to an index number
// ----------------------------------------------------------------------------
var InvestmentStrategyMap = map[string]int{
	"DistributedDecision": 0,
	"MajorityRules":       1,
}

// InvestmentStrategies is a slice of investment strategy names
// ----------------------------------------------------------------------------
var InvestmentStrategies = []string{
	"DistributedDecision",
	"MajorityRules",
}

// Investor is the class that manages one or more influencers to pursue an
// investment strategy in currency exchange.
// ----------------------------------------------------------------------------
type Investor struct {
	cfg               *util.AppConfig   // program wide configuration values
	factory           *Factory          // used to create Influencers
	db                *newdata.Database // where to get the data needed
	BalanceC1         float64           // total amount of currency C1
	BalanceC2         float64           // total amount of currency C2
	StopLossThreshold float64           // value of portfolio where stoploss occurs
	StopLossCount     int               // how many times stoploss was invoked
	PortfolioValueC1  float64           // the C1 value of BalanceC1 + BalanceC2 on DtPortfolioValue
	DtPortfolioValue  time.Time         // the date for which PortfolioValueC1 was calculated
	Investments       []Investment      // a record of all investments made by this investor
	Influencers       []Influencer      // all the influencerst that advise this Investor
	maxProfit         float64           // maximum profit of ALL Investors during this simulation cycle, set by simulator at the end of each simulation cycle, used when calculating fitness
	W1                float64           // weight for profit in Fitness Score
	W2                float64           // weight for correctness
	FitnessCalculated bool              // true after fitness score is calculated and stored in Fitness
	Fitness           float64           // Fitness score calculated at the end of a simulation cycle
	CreatedByDNA      bool              // some init steps must be skipped if it's created from DNA
	Strategy          int               // which strategy to use for predictions
	ID                string            // unique id for this investor
	Parented          int64             // how many times was this Investor a parent for the next gen?
	IDGenerated       bool              // true if ID was generated
	Elite             bool              // an ephemeral flag, if true it means that it may propagate the next generation if we're preserving the elites
	COATrace          Trace             // a struct to keep track of trace information
	// maxPredictions    map[string]int           // max predictions indexed by Influencer subclass, set by simulator at the end of each simulation cycle
	// maxPredictions    map[string]int    // max predictions indexed by Influencer subclass, set by simulator at the end of each simulation cycle, used when calculating fitness
}

// SellInfo is used to record all relevant information about the exchange of C2 back to C1.
// Since any given investment in C2 may be sold in chunks, this info is used to preserve
// all info about each chunk sold.
type SellInfo struct {
	T4            time.Time // date of exchange
	ERT4          float64   // exchange rate used in the exchange
	T4C2Sold      float64   // how much was sold in this chunk
	T4C2Remaining float64   // how much C2 remains?
	T4C1          float64   // amount of C1 resulting from the exchange
	ChunkProfit   float64   // amount of profit in this chunk
	Fee           float64   // cost of making this transaction
	Profitable    bool      // was this exchange profitable
}

// Investment describes a full transaction when the Investor decides to buy.
// The buy-related info is filled in at the time the purchase is made.  T4
// is also set at buy time.  When T4 arrives in the simulation, the
// transaction is completed and the remaining fields are filled in. All
// Investment structures are saved during the simulation. They can be dumped
// to a CSV file for analysis.
// ----------------------------------------------------------------------------
type Investment struct {
	id          string     // investment id
	T3          time.Time  // date on which exchange for C2 was made
	T4          time.Time  // date of exchange back to C1
	T3BalanceC1 float64    // C1 balance after exchange on T3
	T3BalanceC2 float64    // C2 balance after exchange on T3
	T4BalanceC1 float64    // C1 balance after exchange on T4
	T4BalanceC2 float64    // C2 balance after exchange on T4
	T3C1        float64    // amount of C1 exchanged for C2 on T3
	T3C2Buy     float64    // the amount of currency in C2 that T3C1 purchased on T3
	T4C2Sold    float64    // we may need to sell it off over multiple transactions. This keeps track of how much we've sold.
	ERT3        float64    // the exchange rate on T3
	ERT4        float64    // the exchange rate on T4
	T4C1        float64    // amount of currency C1 we were able to purchase with C2 on T4 at exchange rate ERT4
	Fee         float64    // Fee for converting C1 to C2, the "buy fee".  Sell fees are in the chunks
	Completed   bool       // true when the entire original buy amount of C2 has been exchanged for C1
	Chunks      []SellInfo // was this a profitable investment?  Can be multiple if sold across multiple sales.
	RetryCount  int        // how many times was this retried
	// Delta4      int       // t4 = t3 + Delta4 - "sell" date
}

var rnderr = float64(0.01) // if we have less than this amount of C2 remaining, just assume we're done.

// GetDB returns the database associated with the investor
func (i *Investor) GetDB() *newdata.Database {
	return i.db
}

// EnsureID adds an ID if it doesn't already exist
// --------------------------------------------------------------------------------
func (i *Investor) EnsureID() {
	if len(i.ID) == 0 {
		i.DNA() // force id to be generated
	}
}

// ShortID returns the first 5 characters of the ID
// --------------------------------------------------------------------------------
func (i *Investor) ShortID() string {
	return i.ID[:5]
}

// SelectNUniqueSubclasses shuffles the indexes to the map of MInfluencerSubclasses
// then selects the first n, and returns the list
// ----------------------------------------------------------------------------------
func (i *Investor) SelectNUniqueSubclasses(n int) []newdata.MInfluencerSubclass {
	if n <= 0 || n > len(i.db.Mim.MInfluencerSubclasses) {
		fmt.Println("Invalid n; it must be in range 1 to len(m)")
		return nil
	}

	// Shuffle the keys slice
	rand.Shuffle(len(i.db.Mim.MInfluencerSubclassMetricNames), func(k, j int) {
		i.db.Mim.MInfluencerSubclassMetricNames[k], i.db.Mim.MInfluencerSubclassMetricNames[j] = i.db.Mim.MInfluencerSubclassMetricNames[j], i.db.Mim.MInfluencerSubclassMetricNames[k]
	})

	selected := make([]newdata.MInfluencerSubclass, n)
	for j, key := range i.db.Mim.MInfluencerSubclassMetricNames[:n] {
		selected[j] = i.db.Mim.MInfluencerSubclasses[key]
	}

	return selected
}

// Init is called during Generation 1 to get things started.  All settable
// fields are set to random values.
// ----------------------------------------------------------------------------
func (i *Investor) Init(cfg *util.AppConfig, f *Factory, db *newdata.Database) {
	i.cfg = cfg

	i.BalanceC1, i.BalanceC2 = f.InitialFundsSplit()
	i.StopLossThreshold = (1 - cfg.StopLoss) * i.BalanceC1
	i.FitnessCalculated = false
	i.Fitness = float64(0)
	i.factory = f
	i.db = db

	if !i.CreatedByDNA {
		i.W1 = 1
		i.W2 = 1
	}

	//--------------------------------------------------------------
	// If we're creatng by DNA, do not create the influencers here
	//--------------------------------------------------------------
	if i.CreatedByDNA {
		return
	}

	//------------------------------------------------------------------
	// Pick a strategy for this influencer to use
	//------------------------------------------------------------------
	i.Strategy = util.RandomInRange(0, len(InvestmentStrategies)-1) // 0 = Distributed Decsion, 1 = majority wins

	//------------------------------------------------------------------
	// Create a team of influencers.
	//------------------------------------------------------------------
	min := i.cfg.MinInfluencers
	max := i.cfg.MaxInfluencers
	if max > len(i.db.Mim.MInfluencerSubclasses) {
		log.Fatalf("The config file has MaxInfluencers set to %d, however there are only %d Influencers available.\n", max, len(i.db.Mim.MInfluencerSubclasses))
	}
	numInfluencers := util.RandomInRange(min, max) // create this many
	inflist := i.SelectNUniqueSubclasses(numInfluencers)
	for j := 0; j < len(inflist); j++ {
		subclass := inflist[j].Subclass
		dna := fmt.Sprintf("{%s,Metric=%s"+"}", subclass, inflist[j].Metric)
		inf, err := f.NewInfluencer(dna) // create with minimal DNA -- this causes random values to be generated where needed
		if err != nil {
			fmt.Printf("*** ERROR ***  From Influencer Factory: %s\n", err.Error())
			return
		}
		inf.Init(i, cfg) // regardless of the influencer's sell date offset is, we need to force it to this one so that all are consistent
		i.Influencers = append(i.Influencers, inf)
	}
}

// DNA returns a string containing descriptions all its influencers.
// Here is the format of a DNA string for an Investor:
//
//	Delta4=5;Influencers=[{subclass,var1=val1,var2=val2,...}|{subclass,var1=val1,var2=val2,...}|...]
//
// ----------------------------------------------------------------------------
func (i *Investor) DNA() string {
	s := fmt.Sprintf("Strategy=%s;InvW1=%6.4f;InvW2=%6.4f;Influencers=[", InvestmentStrategies[i.Strategy], i.W1, i.W2)
	//----------------------------------------------------------------------------
	// only sort them if this is the first time DNA has been asked for...
	//----------------------------------------------------------------------------
	if !i.IDGenerated {
		i.SortInfluencers()
		i.IDGenerated = true
	}
	for j := 0; j < len(i.Influencers); j++ {
		s += i.Influencers[j].DNA()
		if j+1 < len(i.Influencers) {
			s += "|"
		}
	}
	s += "]}"
	//--------------------------------------------------------------------------------------------
	// s is now the string that we want to generate a hash from. It contains all the critical
	// information that makes it unique.
	//--------------------------------------------------------------------------------------------
	i.ID = util.HashDNA(s)
	s = fmt.Sprintf("{Investor;ID=%s;", i.ID) + s
	return s
}

// GenerateInvestorID generates a unique investor id string
func (i *Investor) GenerateInvestorID() string {
	i.DNA() // call this so that the ID is generated
	return i.ID
}

// SortInfluencers sorts the Influencers of an Investor in a consistent order.
func (i *Investor) SortInfluencers() {
	sort.Slice(i.Influencers, func(k, j int) bool {
		return i.Influencers[k].GetMetric() < i.Influencers[j].GetMetric()
	})
}

// DecideCourseOfAction returns the Investor's "buy", "sell", "hold", or "abstain"
// prediction for T3
// --------------------------------------------------------------------------------
func (i *Investor) DecideCourseOfAction(T3 time.Time) (CourseOfAction, error) {
	var coa CourseOfAction
	coa.Action = "abstain" // the prediction, assume the worst for now
	var recs []Prediction

	//---------------------------------------------------------------------
	// Before doing anything, see if we have a stop-loss situation...
	// For this, we look at the PortfolioValue. If it is less than
	// the Initial Funding then we have lost money.  Then we look at the
	// StopLoss percentage.  If we have lost that percentage or more then
	// We convert everything back to C1 (we don't hold C2 any longer).
	//---------------------------------------------------------------------
	pv := i.PortfolioValue(T3)
	if pv < i.StopLossThreshold {
		if err := i.ExecuteSell(T3, 1); err != nil {
			return coa, err
		}
		i.StopLossThreshold = (1 - i.cfg.StopLoss) * i.BalanceC1
		if (i.cfg.Trace && !i.cfg.CrucibleMode) || i.cfg.PredictionMode {
			fmt.Printf("        <<<STOP LOSS>>>  %s StopLoss, PV = %8.2f, new StopLoss amount: %8.2f\n", i.ID, pv, i.StopLossThreshold)
		}
		i.StopLossCount++
	}

	//---------------------------------------------------------------------
	// No stop-loss. So carry on with determiniing the coarse of action
	//---------------------------------------------------------------------
	for j := 0; j < len(i.Influencers); j++ {
		pred, err := i.Influencers[j].GetPrediction(T3)

		if err != nil {
			// if the error is anything except nildata, then return now
			if !strings.Contains(err.Error(), "nildata") {
				fmt.Printf("nildata comparison failed. Returning now. Error = %s\n", err.Error())
				return coa, err
			}
		}
		pred.Metric = i.Influencers[j].GetMetric()
		pred.ID = i.Influencers[j].GetID()
		pred.Correct = false // don't know yet
		recs = append(recs, *pred)

	}

	//------------------------------------------------------------------------------
	// compute the stats on today's prediction
	//------------------------------------------------------------------------------
	if len(recs) < 1 {
		return coa, fmt.Errorf("no predictions found")
	}
	for j := 0; j < len(recs); j++ {
		switch recs[j].Action {
		case "buy":
			coa.BuyVotes += recs[j].Probability * recs[j].Weight
		case "hold":
			coa.HoldVotes += recs[j].Probability * recs[j].Weight
		case "sell":
			coa.SellVotes += recs[j].Probability * recs[j].Weight
		case "abstain":
			coa.Abstains++
			// abstainers don't add to the totalVotes
		}
	}

	setCourseOfAction(&coa, InvestmentStrategies[i.Strategy]) // use course of action strategy called out in the config file
	if (i.cfg.Trace && !i.cfg.CrucibleMode) || i.cfg.PredictionMode {
		for j := 0; j < len(recs); j++ {
			i.FormatPrediction(&recs[j], T3)
		}
		i.FormatCOA(&coa)
	}

	return coa, nil
}

// setCourseOfAction sets the Action and ActionPct based on influencers input
// ----------------------------------------------------------------------------
func setCourseOfAction(coa *CourseOfAction, method string) error {
	coa.TotalVotes = coa.BuyVotes + coa.HoldVotes + coa.SellVotes // even if it's already been added, this won't hurt anything
	switch method {
	case "DistributedDecision":
		return distributedDecisionCOA(coa)
	case "MajorityRules":
		return majorityRulesCOA(coa)
	}
	return fmt.Errorf("course of action method not recognized: %s", method)
}

// distributedDecisionCOA accommodates all votes in its course of action
//
// -----------------------------------------------------------------------------
func distributedDecisionCOA(coa *CourseOfAction) error {
	activeVotes := float64(coa.BuyVotes + coa.SellVotes + coa.HoldVotes) // total active votes
	if coa.BuyVotes == coa.TotalVotes {                                  // all votes to buy?
		coa.Action = "buy"
		coa.ActionPct = 1.0
	} else if coa.HoldVotes == coa.TotalVotes { // all votes to hold?
		coa.Action = "hold"
		coa.ActionPct = 1.0
	} else if coa.SellVotes == coa.TotalVotes { // all votes to sell?
		coa.Action = "sell"
		coa.ActionPct = 1.0
	} else if coa.BuyVotes > coa.SellVotes { // more buy votes than sell votes?
		coa.Action = "buy"
		coa.ActionPct = float64(coa.BuyVotes) / activeVotes
	} else if coa.SellVotes > coa.BuyVotes {
		coa.Action = "sell"
		coa.ActionPct = float64(coa.SellVotes) / activeVotes
	} else {
		coa.Action = "hold"
		coa.ActionPct = float64(coa.HoldVotes) / activeVotes
	}
	return nil
}

// majorityRulesCOA the draconian decision maker.  The ActionPct is always 100%
//
// -----------------------------------------------------------------------------
func majorityRulesCOA(coa *CourseOfAction) error {
	coa.ActionPct = 1
	if coa.BuyVotes > coa.SellVotes+coa.HoldVotes { // more buy votes than anything else?
		coa.Action = "buy"
	} else if coa.SellVotes > coa.BuyVotes+coa.HoldVotes {
		coa.Action = "sell"
	} else {
		coa.Action = "hold"
	}
	return nil
}

// DailyRun is the main function of an Investor - manage funds for today
//
// INPUTS
//
//	T3      - day to evaluate and act on
//	winddown - if true simulator has passed the simulation end date. Only execute
//	          sells until we're out of C2.  We'll consider anything less
//	          than 1.00 C2 to be "done"
//
// RETURNS
// err - any error encountered
// ------------------------------------------------------------------------------
func (i *Investor) DailyRun(T3 time.Time, winddown bool) error {
	if (i.cfg.Trace && !i.cfg.CrucibleMode) || i.cfg.PredictionMode {
		fmt.Printf("%s - Investor: %s\n", T3.Format("Jan _2, 2006"), i.ID)
	}
	coa, err := i.DecideCourseOfAction(T3)
	if err != nil {
		return err
	}
	if i.cfg.PredictionMode {
		return nil
	}
	switch coa.Action {
	case "buy":
		if winddown {
			return nil
		}
		if err = i.ExecuteBuy(T3, coa.ActionPct); err != nil {
			return err
		}

	case "sell":
		if err = i.ExecuteSell(T3, coa.ActionPct); err != nil {
			return err
		}
	}

	//----------------------------------
	// Update report info as needed...
	//----------------------------------
	if (i.cfg.Trace && !i.cfg.CrucibleMode) || i.cfg.PredictionMode {
		fmt.Printf("\t%s\n", i.PortfolioToString(T3))
		i.SaveTrace()
	}
	if i.cfg.CrucibleMode && i.cfg.DNALog {
		i.SaveCrucibleStats(T3)
	}

	return nil
}

// SaveCrucibleStats saves information needed for the cricible to create the
// dnalog report.  At the moment, this is just the portfolio value on a daily
// basis
// ------------------------------------------------------------------------------
func (i *Investor) SaveCrucibleStats(T3 time.Time) {
	pv := i.PortfolioValue(T3) // portfolio value for this day
	ar, err := util.AnnualizedReturn(i.cfg.InitFunds, pv, time.Time(i.factory.cfg.DtStart), T3)
	if err != nil {
		log.Fatal(err)
	}
	i.factory.sim.crucible.SaveInvestorPortfolioValue(ar) // save this info in the crucible object
}

// ExecuteBuy does an exchange of C1 for C2 on T3. It will purchase pct*i.cfg.StdInvestment
//
// INPUTS
// T3 - the date on which this buy is being executed
// pct - the percentage of the StdInvestment amount.
// RETURNS
// err - any error encountered
// -----------------------------------------------------------------------------
func (i *Investor) ExecuteBuy(T3 time.Time, pct float64) error {
	//-------------------------------------------------------
	// make sure we have funds before executing any buy...
	//-------------------------------------------------------
	if i.BalanceC1 < 1.00 {
		return nil
	}

	var inv Investment
	inv.id = util.GenerateRefNo()
	inv.T3C1 = i.cfg.StdInvestment * pct
	if i.BalanceC1 < i.cfg.StdInvestment {
		inv.T3C1 = i.BalanceC1
	}
	inv.T3 = T3
	s := i.factory.PrefixMetricC1C2("EXClose")
	ss := []newdata.FieldSelector{s}
	er3, err := i.db.Select(inv.T3, ss)
	if err != nil {
		return err
	}
	if er3 == nil {
		return fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found", inv.T3.Format("1/2/2006"))
	}

	inv.ERT3 = er3.Fields[s.FQMetric()].Value                // exchange rate on T3
	inv.T3C2Buy = inv.T3C1 * inv.ERT3                        // amount of C2 we purchased on T3
	inv.Fee = (inv.T3C1 * i.cfg.TxnFeeFactor) + i.cfg.TxnFee // cost of the transaction: flat fee plus percentage is here because a buy is wholly done here, not in chunks as with sells
	inv.T4C2Sold = 0                                         // just being explicit, haven't sold any of it yet
	i.BalanceC1 -= (inv.T3C1 + inv.Fee)                      // we spent this much C1...
	i.BalanceC2 += inv.T3C2Buy                               // to purchase this much more C2
	inv.T3BalanceC1 = i.BalanceC1                            // C1 balance after exchange
	inv.T3BalanceC2 = i.BalanceC2                            // C2 balance after exchange
	i.Investments = append(i.Investments, inv)               // add it to the list of investments

	if (i.cfg.Trace && !i.cfg.CrucibleMode) || i.cfg.PredictionMode {
		i.showBuy(&inv)
	}

	return nil
}

// ExecuteSell does an exchange of C2 for C1 on T4. It will purchase pct*i.cfg.StdInvestment
//
// INPUTS
// T4 - the date on which this buy is being executed
// pct - the percentage of the StdInvestment amount.
// RETURNS
// err - any error encountered
// -----------------------------------------------------------------------------
func (i *Investor) ExecuteSell(T4 time.Time, pct float64) error {
	//------------------------------------------
	// Make sure we have something to sell...
	//------------------------------------------
	if i.BalanceC2 < 1.00 {
		return nil
	}
	sellAmount := pct * i.BalanceC2 // the action was to sell pct * i.BalanceC2
	i.settleInvestment(T4, sellAmount)

	return nil
}

// PortfolioValue returns the value of the Investors portfolio at time t. The
// portfolio value is returned in terms of C1 and it is the current BalanceC1
// plus BalanceC2 converted to C1 at t.
// ------------------------------------------------------------------------------
func (i *Investor) PortfolioValue(t time.Time) float64 {
	if i.BalanceC2 == 0 {
		return i.BalanceC1
	}
	s := i.factory.PrefixMetricC1C2("EXClose")
	ss := []newdata.FieldSelector{}
	ss = append(ss, s)
	er, err := i.db.Select(t, ss) // exchange rate for C2 at time t
	if err != nil {
		log.Fatalf("Error getting exchange close rate")
	}
	if er == nil {
		fmt.Printf("Please check your database. EXClose value for %s not found\n", t.Format("1/2/2006"))
	}
	a := s.FQMetric()
	pv := float64(0.0)
	if len(a) > 0 {
		if v, ok := er.Fields[a]; ok {
			C2 := i.BalanceC2 / v.Value // amount of C1 we get for BalanceC2 at this exchange rate
			pv = i.BalanceC1 + C2
		}
	} else {
		fmt.Printf("*** WARNING **** PortfolioValue: ExchangeRate Record for %s not found\n", t.Format("1/2/2006"))
	}
	return pv
}

// settleInvestment - this code was moved to a method as it needed to be done
//
//	This function is called when we need to convert C2 from the jth Investment
//	back to C1
//
// INPUTS
//
//	   t4         - sell date
//	   sellAmount - the amount we're looking to sell.  Could be greater than, equal to,
//		               or less than the amount of C2 we gained in this Investment.
//
// RETURNS
//
//	   adjusted sellAmount after settling this investment
//		  any critical error encountered
//
// -----------------------------------------------------------------------------
func (i *Investor) settleInvestment(t4 time.Time, sellAmount float64) (float64, error) {
	var err error
	var thisSaleC2 float64
	var thisSaleC1 float64

	//-------------------------------------------------
	// Save the exchange rate on the day of sale, t4
	//-------------------------------------------------
	s := i.factory.PrefixMetricC1C2("EXClose")
	ss := []newdata.FieldSelector{}
	ss = append(ss, s)
	er4, err := i.db.Select(t4, ss)
	if err != nil {
		return sellAmount, err
	}
	if er4 == nil {
		// this should never happen.
		err = fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found; Investment marked as completed", t4.Format("1/2/2006"))
		fmt.Printf("%s\n", err.Error())
		return sellAmount, nil // it's was not a critical error, it's been reported, just keep going
	}

	//-------------------------------------------------------------------
	// Now that we have today's exchange rate... sort the investments
	// by ERT4 decending.  We do this so that we sell everything we can
	// at the greatest loss... per Joe, this gives us tax benefits. This
	// is accomplished by processing the Investments slice sorted by
	// ERT4 descending. Profitability is inversely proportional to T4 EXClose.
	//-------------------------------------------------------------------
	for j := 0; j < len(i.Investments); j++ {
		if !i.Investments[j].Completed {
			if er4.Fields[s.FQMetric()].Value < 0.0001 {
				log.Panicf("Invalid exchange rate on %s: %12.6f\n", er4.Date.Format("1/2/2006"), er4.Fields[s.Metric].Value)
			}
			i.Investments[j].ERT4 = er4.Fields[s.FQMetric()].Value // exchange rate on T4... just applies to this sale, we don't touch completed Investments
		}
	}
	i.sortInvestmentsDescending()

	//-----------------------------------------------------------------
	// Now spin through the investments selling "sellAmount" of C2
	//-----------------------------------------------------------------
	for j := 0; j < len(i.Investments) && sellAmount > rnderr; j++ {
		if i.Investments[j].Completed {
			continue // skip if already processed
		}

		//---------------------------------------------------------------------------------
		// If amount we're selling is greater than or equal to what's in this investment
		// then we'll sell it all.  Otherwise we'll sell enough to cover sellAmount
		//---------------------------------------------------------------------------------
		remaining := i.Investments[j].T3C2Buy - i.Investments[j].T4C2Sold // remaining is how much C2 we bought in this Investment minus what we've already sold
		if sellAmount >= remaining {
			thisSaleC2 = remaining // sell everything we have left
		} else {
			thisSaleC2 = sellAmount // sellAmount is < what we have. So we'll sell a portion
		}
		sellAmount -= thisSaleC2                        // this will be what's left to sell, now that we know how much to sell in this exchange
		thisSaleC1 = thisSaleC2 / i.Investments[j].ERT4 // This is the sell. The Amount of C1 we got back by selling "sellAmount"
		fee := thisSaleC1 * i.cfg.TxnFeeFactor          // for each chunk, add the fee factor
		i.Investments[j].T4C2Sold += thisSaleC2         // add what we're selling now to what's already been sold
		i.Investments[j].T4C1 += thisSaleC1             // add the C1 we got back to the cumulative total for this investment
		i.BalanceC1 += (thisSaleC1 - fee)               // we recovered this much C1...
		i.BalanceC2 -= thisSaleC2                       // by selling this C2

		chunkt3c1 := thisSaleC2 / i.Investments[j].ERT3 // amount of C1 in this transaction

		//------------------------------------------------------------------------
		// Create a new chunk for this investment to capture all relevant details
		//------------------------------------------------------------------------
		p := i.Investments[j].ERT4 < i.Investments[j].ERT3 // this is the profitability condition at its simplest
		chunk := SellInfo{
			T4:            t4,                        // date of exchange
			ERT4:          i.Investments[j].ERT4,     // exchange rate used in the exchange
			T4C2Sold:      thisSaleC2,                // how much was sold in this chunk
			T4C2Remaining: i.Investments[j].T4C2Sold, // how much C2 remains from the original exchange
			T4C1:          thisSaleC1,                // amount of C1 resulting from the exchange
			Profitable:    p,                         // was this exchange profitable
			ChunkProfit:   thisSaleC1 - chunkt3c1,    // how much profit
			Fee:           fee,                       // cost of this transaction
		}
		i.Investments[j].Chunks = append(i.Investments[j].Chunks, chunk) // was this transaction profitable?  Save it in the list

		i.Investments[j].Completed = (i.Investments[j].T4C2Sold+rnderr >= i.Investments[j].T3C2Buy) // we're completed when we've sold as much as we bought

		i.Investments[j].T4BalanceC1 = i.BalanceC1 // amount of C1 after this exchange
		i.Investments[j].T4BalanceC2 = i.BalanceC2 // amount of C2 after this exchange
		i.Investments[j].T4 = t4                   // the date on which this particular sale was done (we don't save all dates of sale)

		//-------------------------------------------------------------
		// Update each Influencer's predictions...
		//-------------------------------------------------------------
		for k := 0; k < len(i.Influencers); k++ {
			i.Influencers[k].FinalizePrediction(i.Investments[j].T3, t4, p)
		}
		if (i.cfg.Trace && !i.cfg.CrucibleMode) || i.cfg.PredictionMode {
			i.showSell(&i.Investments[j], thisSaleC1, thisSaleC2, fee)
		}
	}
	// One final fee... if there is an flat-fee for the transaction, add it here
	if i.cfg.TxnFee > 0 {
		fee := Investment{
			T3:          t4,                         // date on which exchange for C2 was made
			T4:          t4,                         // date of exchange back to C1
			T3BalanceC1: i.BalanceC1,                // C1 balance after exchange on T3
			T3BalanceC2: i.BalanceC2,                // C2 balance after exchange on T3
			T4BalanceC1: i.BalanceC1 - i.cfg.TxnFee, // C1 balance after exchange on T4
			T4BalanceC2: i.BalanceC2,                // C2 balance after exchange on T4
			T3C1:        0,                          // amount of C1 exchanged for C2 on T3
			T3C2Buy:     0,                          // the amount of currency in C2 that T3C1 purchased on T3
			T4C2Sold:    0,                          // we may need to sell it off over multiple transactions. This keeps track of how much we've sold.
			ERT3:        0,                          // the exchange rate on T3
			ERT4:        0,                          // the exchange rate on T4
			T4C1:        0,                          // amount of currency C1 we were able to purchase with C2 on T4 at exchange rate ERT4
			Fee:         i.cfg.TxnFee,               // Fee for converting C1 to C2, the "buy fee".  Sell fees are in the chunks
			Completed:   true,                       // true when the entire original buy amount of C2 has been exchanged for C1
			RetryCount:  0,                          // how many times was this retried
		}
		i.Investments = append(i.Investments, fee)
	}

	return sellAmount, nil
}

// sortInvestmentsDescending uses the E
func (i *Investor) sortInvestmentsDescending() {
	sort.Slice(i.Investments, func(j, k int) bool {
		return i.Investments[j].ERT4 > i.Investments[k].ERT4
	})
}

// CalculateFitnessScore calculates the fitness score for an Investor.
//
// The score depends  on the final amount of C1 the investor has at the end of the
// simulation. If the investor ends up with less C1 than it started with, a low or
// even zero fitness score makes sense. If it has more, then it did something right
// and should be rewarded.
//
// The approach will be to combine the amount of profit made and the correctness of
// its decisions, with the majority of the weight on the profit.  So, the formula
// would be:
//
//	fitnessScore = w1 * (finalC1 = initialC1) / maxProfit  +  w2 * correctness
//
// initialC1   - the amount of C1 the investor started with.
// finalC1     - the amount of C1 the investor has at the end of the simulation.
// maxProfit   - the maximum profit made by any investor. This is used to normalize
//
//	the profit made by the investor.
//
// correctness - the percentage of correct investment decisions made by the investor.
// w1 and w2   - weights that determine the relative importance of profit and correctness.
//
//	w1 is used to normalized profit. w2 rewards investors for making
//	correct decisions, even if those decisions didn't necessarily lead
//	to the highest profit.
//
// ------------------------------------------------------------------------------------
func (i *Investor) CalculateFitnessScore() float64 {
	if i.FitnessCalculated {
		return i.Fitness
	}

	//-------------------------------------------------------
	// Calculate correctness.  This will always be >= 0
	//-------------------------------------------------------
	correct := 0
	total := 0
	jlen := len(i.Investments)
	for j := 0; j < jlen; j++ {
		if i.Investments[j].Completed {
			pl := i.Investments[j].Chunks
			for k := 0; k < len(pl); k++ {
				total++
				if pl[k].Profitable {
					correct++
				}
			}
		}
	}
	correctness := float64(0)
	if total > 0 && correct > 0 {
		correctness = float64(float64(correct) / float64(total))
	}

	profit := i.PortfolioValueC1 - i.cfg.InitFunds
	weightedProfit := float64(0)
	if i.maxProfit > 0 {
		weightedProfit = float64(i.W1 * profit / i.maxProfit)
	}
	weightedCorrectness := float64(i.W2 * correctness)
	i.Fitness = weightedProfit + weightedCorrectness
	if i.Fitness < 0 {
		i.Fitness = 0
	}

	//-----------------------------------------------------------
	// if bonus plan is in effect, add it to the fitness score
	//-----------------------------------------------------------
	if i.cfg.InvestorBonusPlan {
		if len(i.cfg.GenDurSpec) > 0 {
			// TODO: fix fitness bonus for GenDurSpec
			// fmt.Printf("fitness bonus for GenDurSpec: %s needs to be coded\n", i.cfg.GenDurSpec)
		} else {
			dtStart := time.Time(i.cfg.DtStart)
			dtStop := time.Time(i.cfg.DtStop)
			pv := i.PortfolioValue(dtStop)
			ar, err := util.AnnualizedReturn(i.cfg.InitFunds, pv, dtStart, dtStop)
			if err != nil {
				fmt.Printf("Investor.CalculatFitnessScore: Error calculating annualized return: %s\n", err.Error())
			} else {
				bonus := fitnessBonus(ar)
				if bonus > 1.0 {
					i.Fitness *= bonus
					// fmt.Printf("Investor %s received fitness bonus = %f\n", i.ID, bonus)
				}
			}
		}
	}

	i.FitnessCalculated = true
	return i.Fitness
}

func fitnessBonus(ar float64) float64 {
	if ar >= 0.1 && ar < 0.15 {
		return 2 + ar*5
	} else if ar >= 0.15 && ar < 0.2 {
		return 3 + ar*6
	} else if ar >= 0.2 && ar < 0.25 {
		return 4 + ar*7
	} else if ar >= 0.25 {
		return 5 + ar*8
	}
	return 1
}
