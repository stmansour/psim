package core

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/stmansour/psim/util"

	"github.com/stmansour/psim/data"
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

// Investor is the class that manages one or more influencers to pursue an
// investment strategy in currency exchange.
// ----------------------------------------------------------------------------
type Investor struct {
	cfg               *util.AppConfig // program wide configuration values
	factory           *Factory        // used to create Influencers
	BalanceC1         float64         // total amount of currency C1
	BalanceC2         float64         // total amount of currency C2
	PortfolioValueC1  float64         // the C1 value of BalanceC1 + BalanceC2 on DtPortfolioValue
	DtPortfolioValue  time.Time       // the date for which PortfolioValueC1 was calculated
	Delta4            int             // t4 = t3 + Delta4 - must be the same Delta4 for all influencers in this investor
	Investments       []Investment    // a record of all investments made by this investor
	Influencers       []Influencer    // all the influencerst that advise this Investor
	maxProfit         float64         // maximum profit of ALL Investors during this simulation cycle, set by simulator at the end of each simulation cycle
	maxPredictions    map[string]int  // max predictions indexed by Influncer subclass, set by simulator at the end of each simulation cycle
	W1                float64         // weight for profit in Fitness Score
	W2                float64         // weight for correctness
	FitnessCalculated bool            // true after fitness score is calculated and stored in Fitness
	Fitness           float64         // Fitness score calculated at the end of a simulation cycle
	CreatedByDNA      bool            // some init steps must be skipped if it's created from DNA
	ID                string          // unique id for this investor
	// BalanceSettled    float64         // amount of C2 converted to C1 because simulation ended before T4 arrived
}

// Investment describes a full transaction when the Investor decides to buy.
// The buy-related info is filled in at the time the purchase is made.  T4
// is also set at buy time.  When T4 arrives in the simulation, the
// transaction is completed and the remaining fields are filled in. All
// Investment structures are saved during the simulation. They can be dumped
// to a CSV file for analysis.
// ----------------------------------------------------------------------------
type Investment struct {
	id          string    // investment id
	T3          time.Time // date on which purchase of C2 was made
	T4          time.Time // date on which C2 will be exchanged for C1
	T3BalanceC1 float64   // C1 balance after exchange on T3
	T3BalanceC2 float64   // C2 balance after exchange on T3
	T4BalanceC1 float64   // C1 balance after exchange on T4
	T4BalanceC2 float64   // C2 balance after exchange on T4
	T3C1        float64   // amount of C1 exchanged for C2 on T3
	T3C2Buy     float64   // the amount of currency in C2 that T3C1 purchased on T3
	T4C2Sold    float64   // we may need to sell it off over multiple transactions. This keeps track of how much we've sold.
	ERT3        float64   // the exchange rate on T3
	ERT4        float64   // the exchange rate on T4
	T4C1        float64   // amount of currency C1 we were able to purchase with C2 on T4 at exchange rate ERT4
	Delta4      int       // t4 = t3 + Delta4 - "sell" date
	Completed   bool      // true when the investmnet has been exchanged C2 for C1
	Profitable  []bool    // was this a profitable investment?  Can be multiple if sold across multiple sales.
	RetryCount  int       // how many times was this retried
}

var rnderr = float64(0.01) // if we have less than this amount of C2 remaining, just assume we're done.

// Init is called during Generation 1 to get things started.  All settable
// fields are set to random values.
// ----------------------------------------------------------------------------
func (i *Investor) Init(cfg *util.AppConfig, f *Factory) {
	i.cfg = cfg
	i.BalanceC1 = cfg.InitFunds
	i.FitnessCalculated = false
	i.Fitness = float64(0)
	i.factory = f

	if !i.CreatedByDNA {
		i.Delta4 = util.RandomInRange(cfg.MinDelta4, cfg.MaxDelta4) // all Influencers will be constrained to this
		i.W1 = i.cfg.InvW1
		i.W2 = i.cfg.InvW2
	}

	//--------------------------------------------------------------
	// If we're creatng by DNA, do not create the influencers here
	//--------------------------------------------------------------
	if i.CreatedByDNA {
		return
	}

	//------------------------------------------------------------------
	// Create a team of influencers.
	//------------------------------------------------------------------
	numInfluencers := util.RandomInRange(1, len(util.InfluencerSubclasses))
	for j := 0; (j < numInfluencers) && (len(i.Influencers) < numInfluencers); j++ {
		subclassOK := false
		subclass := ""
		for !subclassOK {
			subclass = util.InfluencerSubclasses[util.RandomInRange(0, len(util.InfluencerSubclasses)-1)]
			found := false
			for k := 0; k < len(i.Influencers) && !found; k++ {
				s := i.Influencers[k].Subclass()
				if s == subclass {
					found = true
				}
			}
			if !found {
				subclassOK = true
			}
		}
		dna := "{" + subclass + "}"

		inf, err := f.NewInfluencer(dna) // create with minimal DNA -- this causes random values to be generated where needed
		if err != nil {
			fmt.Printf("*** ERROR ***  From Influencer Factory: %s\n", err.Error())
			return
		}
		inf.Init(i, cfg, i.Delta4) // regardless of the influencer's sell date offset is, we need to force it to this one so that all are consistent
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
	s := fmt.Sprintf("{Investor;InvW1=%6.4f;InvW2=%6.4f;Influencers=[", i.W1, i.W2)
	for j := 0; j < len(i.Influencers); j++ {
		s += i.Influencers[j].DNA()
		if j+1 < len(i.Influencers) {
			s += "|"
		}
	}
	s += "]}"
	return s
}

// GetCourseOfAction returns the Investor's "buy", "sell", "hold", or "abstain"
// prediction for T3
// --------------------------------------------------------------------------------
func (i *Investor) GetCourseOfAction(T3 time.Time) (CourseOfAction, error) {
	// T4 := T3.AddDate(0, 0, i.Delta4) // here if we need it
	var coa CourseOfAction
	coa.Action = "abstain" // the prediction, assume the worst for now
	var recs []Prediction

	for j := 0; j < len(i.Influencers); j++ {
		prediction, r1, r2, probability, weight, err := i.Influencers[j].GetPrediction(T3)
		if err != nil {
			// if the error is anything except nildata, then return now
			if !strings.Contains(err.Error(), "nildata") {
				fmt.Printf("nildata comparison failed. Returning now. Error = %s\n", err.Error())
				return coa, err
			}
		}
		recs = append(recs,
			Prediction{
				Delta1:      int64(i.Influencers[j].GetDelta1()),
				Delta2:      int64(i.Influencers[j].GetDelta2()),
				T3:          T3,
				RT1:         r1,
				RT2:         r2,
				Action:      prediction,
				Probability: probability,
				Weight:      weight,
				IType:       reflect.TypeOf(i.Influencers[j]).String(),
				ID:          i.Influencers[j].GetID(),
				Correct:     false, // don't know yet
			})

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

	setCourseOfAction(&coa, i.cfg.COAStrategy) // use course of action strategy called out in the config file
	if i.cfg.Trace {
		for j := 0; j < len(recs); j++ {
			i.FormatPrediction(&recs[j], T3)
		}
		i.FormatCOA(&coa)
	}

	return coa, nil
}

// FormatPrediction prints a readable version of the Influencers predictions
// ----------------------------------------------------------------------------
func (i *Investor) FormatPrediction(p *Prediction, T3 time.Time) {
	fmt.Printf("\t%s: %s   (T1 %s [%4.2f] -  T2 %s [%4.2f])\n",
		p.IType[6:8],
		p.Action,
		T3.AddDate(0, 0, int(p.Delta1)).Format("Jan _2, 2006"),
		p.RT1,
		T3.AddDate(0, 0, int(p.Delta2)).Format("Jan _2, 2006"),
		p.RT2)
}

// FormatCOA prints a readable version of the Influencers predictions
// ----------------------------------------------------------------------------
func (i *Investor) FormatCOA(c *CourseOfAction) {
	fmt.Printf("\tCOA:  Action: %s  %3.0f%%  (buy: %3.2f, hold: %3.2f, sell: %3.2f, abs: %3.2f)\n", c.Action, c.ActionPct*100, c.BuyVotes, c.HoldVotes, c.SellVotes, c.Abstains)
}

// PortfolioToString returns a string with the portfolio balance at time t
// ----------------------------------------------------------------------------
func (i *Investor) PortfolioToString(t time.Time) string {
	pv := i.PortfolioValue(t)
	return fmt.Sprintf("C1bal = %6.2f %s, C2bal = %6.2f %s, PV = %6.2f %s\n", i.BalanceC1, i.cfg.C1, i.BalanceC2, i.cfg.C2, pv, i.cfg.C1)
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
	if i.cfg.Trace {
		fmt.Printf("%s - %s\n", T3.Format("Jan _2, 2006"), i.ID)
	}
	coa, err := i.GetCourseOfAction(T3)
	if err != nil {
		return err
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
	if i.cfg.Trace {
		fmt.Printf("\t%s\n", i.PortfolioToString(T3))
	}

	return nil
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
	er3 := data.CSVDBFindRecord(inv.T3)
	if er3 == nil {
		return fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found", inv.T3.Format("1/2/2006"))
	}

	inv.ERT3 = er3.EXClose                     // exchange rate on T3
	inv.T3C2Buy = inv.T3C1 * inv.ERT3          // amount of C2 we purchased on T3
	inv.T4C2Sold = 0                           // just being explicit, haven't sold any of it yet
	i.BalanceC1 -= inv.T3C1                    // we spent this much C1...
	i.BalanceC2 += inv.T3C2Buy                 // to purchase this much more C2
	inv.T3BalanceC1 = i.BalanceC1              // C1 balance after exchange
	inv.T3BalanceC2 = i.BalanceC2              // C2 balance after exchange
	i.Investments = append(i.Investments, inv) // add it to the list of investments

	if i.cfg.Trace {
		i.showBuy(&inv)
	}

	return nil
}

func (i *Investor) showBuy(inv *Investment) {
	fmt.Printf("        *** BUY ***   %8.2f %s (%8.2f %s)\n", inv.T4C1, i.cfg.C1, inv.T3C2Buy, i.cfg.C2)
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
	er := data.CSVDBFindRecord(t)  // exchange rate for C2 at time t
	C2 := i.BalanceC2 / er.EXClose // amount of C1 we get for BalanceC2 at this exchange rate
	pv := i.BalanceC1 + C2
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
	er4 := data.CSVDBFindRecord(t4)
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
			i.Investments[j].ERT4 = er4.EXClose // exchange rate on T4... just applies to this sale, we don't touch completed Investments
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
		i.Investments[j].T4C2Sold += thisSaleC2         // add what we're selling now to what's already been sold
		i.Investments[j].T4C1 += thisSaleC1             // add the C1 we got back to the cumulative total for this investment

		p := i.Investments[j].ERT4 < i.Investments[j].ERT3                   // this is the profitability condition at its simplest
		i.Investments[j].Profitable = append(i.Investments[j].Profitable, p) // was this transaction profitable?  Save it in the list

		i.Investments[j].Completed = (i.Investments[j].T4C2Sold+rnderr >= i.Investments[j].T3C2Buy) // we're completed when we've sold as much as we bought

		i.BalanceC1 += thisSaleC1                  // we recovered this much C1...
		i.BalanceC2 -= thisSaleC2                  // by selling this C2
		i.Investments[j].T4BalanceC1 = i.BalanceC1 // amount of C1 after this exchange
		i.Investments[j].T4BalanceC2 = i.BalanceC2 // amount of C2 after this exchange
		i.Investments[j].T4 = t4                   // the date on which this particular sale was done (we don't save all dates of sale)

		//-------------------------------------------------------------
		// Update each Influencer's predictions...
		//-------------------------------------------------------------
		for k := 0; k < len(i.Influencers); k++ {
			i.Influencers[k].FinalizePrediction(i.Investments[j].T3, t4, p)
		}
		if i.cfg.Trace {
			i.showSell(&i.Investments[j], thisSaleC1, thisSaleC2)
		}
	}

	return sellAmount, nil
}

func (i *Investor) showSell(inv *Investment, tsc1, tsc2 float64) {
	gains := 0
	losses := 0
	n := len(inv.Profitable)
	for j := 0; j < n; j++ {
		if inv.Profitable[j] {
			gains++
		} else {
			losses++
		}
	}
	fmt.Printf("        *** SELL ***  %8.2f %s, [%8.2f %s], investments affected: %d -->  %d profited, %d lost\n", tsc1, i.cfg.C1, tsc2, i.cfg.C2, n, gains, losses)
}

// sortInvestmentsDescending uses the E
func (i *Investor) sortInvestmentsDescending() {
	sort.Slice(i.Investments, func(j, k int) bool {
		return i.Investments[j].ERT4 > i.Investments[k].ERT4
	})
}

// InvestorProfile outputs information about this investor and its influencers
// to a file named "investorProfile.txt"
//
// RETURNS
//
//	any error encountered or nil if no error
//
// ------------------------------------------------------------------------------------
func (i *Investor) InvestorProfile() error {
	file, err := os.Create("investorProfile.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "INVESTOR PROFILE\n")
	fmt.Fprintf(file, "Initial cash: %14.2f %s\n", i.cfg.InitFunds, i.cfg.C1)
	fmt.Fprintf(file, "              %14.2f %s\n", 0.0, i.cfg.C2)
	fmt.Fprintf(file, "Ending cash:  %14.2f %s\n", i.BalanceC1, i.cfg.C1)
	fmt.Fprintf(file, "              %14.2f %s\n", i.BalanceC2, i.cfg.C2)
	fmt.Fprintf(file, "\nInfluencers:\n")

	for j := 0; j < len(i.Influencers); j++ {
		fmt.Fprintf(file, "%d. %s\n", j+1, i.Influencers[j].DNA())
	}

	return nil
}

// // ToString simply returns a printable version of the Investment struct as a string.
// // ------------------------------------------------------------------------------------
// func (i *Investment) ToString() string {
// 	s := fmt.Sprintf("    id		= %s\n", i.id)
// 	s += fmt.Sprintf("    T3		= %s\n", i.T3)
// 	s += fmt.Sprintf("    T4		= %s\n", i.T4)
// 	s += fmt.Sprintf("    T3C1		= %8.2f\n", i.T3C1)
// 	s += fmt.Sprintf("    T3C2Buy		= %8.2f\n", i.T3C2Buy)
// 	s += fmt.Sprintf("    T4C2Sold	= %8.2f\n", i.T4C2Sold)
// 	s += fmt.Sprintf("    ERT3		= %8.2f\n", i.ERT3)
// 	s += fmt.Sprintf("    ERT4		= %8.2f\n", i.ERT4)
// 	s += fmt.Sprintf("    T4C1	= %8.2f\n", i.T4C1)
// 	s += fmt.Sprintf("    Delta4	= %d\n", i.Delta4)
// 	return s
// }

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
			pl := i.Investments[j].Profitable
			for k := 0; k < len(pl); k++ {
				total++
				if pl[k] {
					correct++
				}
			}
		}
	}
	correctness := float64(0)
	if total > 0 && correct > 0 {
		correctness = float64(float64(correct) / float64(total))
	}

	// And now the fitness score
	// util.DPrintf("FitnessScore:  Investor dna is %s\n", i.DNA())
	// util.DPrintf("i.Balance: %6.2f\n", i.BalanceC1)
	// util.DPrintf("i = %#v\n", *i)
	// util.DPrintf("i.cfg.InitFunds: %8.2f\n", i.cfg.InitFunds)

	dda := i.BalanceC1 - i.cfg.InitFunds
	if math.IsNaN(dda) || math.IsInf(dda, 0) {
		log.Panicf("Investor.FitnessSocre() is dda is invalid\n")
	}

	ddb := float64(0)
	if i.maxProfit > 0 {
		ddb = float64(i.W1 * dda / i.maxProfit)
	}
	if math.IsNaN(ddb) || math.IsInf(ddb, 0) {
		log.Panicf("Investor.FitnessSocre() is ddb is invalid.  i.W1 = %4.2f, dda = %8.2f, i.maxProfit = %8.2f\n", i.W1, dda, i.maxProfit)
	}
	ddc := float64(i.W2 * correctness)

	if math.IsNaN(ddc) || math.IsInf(ddc, 0) {
		log.Panicf("Investor.FitnessSocre() is ddc is invalid\n")
	}

	i.Fitness = ddb + ddc
	if i.Fitness < 0 {
		i.Fitness = 0
	}
	// i.Fitness = float64(i.W1*(i.BalanceC1-i.cfg.InitFunds)/i.maxProfit) + float64(i.W2*correctness)
	i.FitnessCalculated = true

	// util.DPrintf("W1 = %3.1f, BalanceC1 = %6.2f, InitFunds = %6.2f, maxProfit = %6.2f, W2 = %3.1f, correctness = %d / %d = %6.2f  ",
	// 	i.W1, i.BalanceC1, i.cfg.InitFunds, i.maxProfit, i.W2, correct, total, correctness)
	// util.DPrintf("Fitness = %6.3f\n", i.Fitness)

	if math.IsNaN(i.Fitness) || math.IsInf(i.Fitness, 0) {
		log.Panicf("Investor.FitnessSocre() is STORING AN INVALID FITNESS!!!!\n")
	}

	return i.Fitness
}
