package core

import (
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"time"

	"github.com/stmansour/psim/util"

	"github.com/stmansour/psim/data"
)

// Investor is the class that manages one or more influencers to pursue an
// investment strategy in currency exchange.
// =---------------------------------------------------------------------------
type Investor struct {
	cfg               *util.AppConfig // program wide configuration values
	factory           *Factory        // used to create Influencers
	BalanceC1         float64         // total amount of currency C1
	BalanceC2         float64         // total amount of currency C2
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
}

// Investment describes a full transaction when the Investor decides to buy.
// The buy-related info is filled in at the time the purchase is made.  T4
// is also set at buy time.  When T4 arrives in the simulation, the
// transaction is completed and the remaining fields are filled in. All
// Investment structures are saved during the simulation. They can be dumped
// to a CSV file for analysis.
// ----------------------------------------------------------------------------
type Investment struct {
	id         string    // investment id
	T3         time.Time // date on which purchase of C2 was made
	T4         time.Time // date on which C2 will be exchanged for C1
	T3C1       float64   // amount of C2 to purchase at T3
	BuyC2      float64   // the amount of currency in C2 that C1 purchased on T3
	SellC2     float64   // for now, this is always going to be the same as BuyC2
	ERT3       float64   // the exchange rate on T3
	ERT4       float64   // the exchange rate on T4
	T4C1       float64   // amount of currency C1 we were able to purchase with C2 on T4 at exchange rate ERT4
	Delta4     int       // t4 = t3 + Delta4 - "sell" date
	Completed  bool      // true when the investmnet has been exchanged C2 for C1
	Profitable bool      // was this a profitable investment?
}

// Init is called during Generation 1 to get things started.  All settable
// fields are set to random values.
// ----------------------------------------------------------------------------
func (i *Investor) Init(cfg *util.AppConfig, f *Factory) {
	i.cfg = cfg
	i.BalanceC1 = cfg.InitFunds
	i.FitnessCalculated = false
	i.Fitness = float64(0)

	if !i.CreatedByDNA {
		i.Delta4 = util.RandomInRange(cfg.MinDelta4, cfg.MaxDelta4) // all Influencers will be constrained to this
		i.W1 = i.cfg.InvW1
		i.W2 = i.cfg.InvW2
	}

	//------------------------------------------------------------------
	// Create a team of influencers.  For now, we're just going to add
	// one influencer to get things compiling and running.
	//------------------------------------------------------------------
	inf, err := f.NewInfluencer("{DRInfluencer}") // create with minimal DNA -- this causes random values to be generated where needed
	if err != nil {
		fmt.Printf("*** ERROR ***  From Influencer Factory: %s\n", err.Error())
		return
	}
	inf.Init(i, cfg, i.Delta4) // regardless of the influencer's sell date offset is, we need to force it to this one so that all are consistent
	i.Influencers = append(i.Influencers, inf)
}

// DNA returns a string containing descriptions all its influencers.
// Here is the format of a DNA string for an Investor:
//
//	Delta4=5;Influencers=[{subclass,var1=val1,var2=val2,...}|{subclass,var1=val1,var2=val2,...}|...]
//
// ----------------------------------------------------------------------------
func (i *Investor) DNA() string {
	s := fmt.Sprintf("{Investor;Delta4=%d;InvW1=%6.4f;InvW2=%6.4f;Influencers=[", i.Delta4, i.W1, i.W2)
	for j := 0; j < len(i.Influencers); j++ {
		s += i.Influencers[j].DNA()
		if j+1 < len(i.Influencers) {
			s += "|"
		}
	}
	s += "]}"
	return s
}

// BuyConversion spins through all the influencers and asks for recommendations
// on whether to buy or hold on T3. Then the Investor decides whether to buy
// or hold.  If a "buy" is made, then an entry is added to the Investments
// list so that it can be completed when T4 arrives.  Currency C2 is purchased
// using C1.  If there are no remaining funds, the function returns immediately.
// Balances of C1 and C2 are adjusted whenever a conversion is done.
// -----------------------------------------------------------------------------
func (i *Investor) BuyConversion(T3 time.Time) (int, error) {
	T4 := T3.AddDate(0, 0, i.Delta4) // here if we need it
	BuyCount := 0
	if i.BalanceC1 <= 0.0 {
		return BuyCount, nil // we cannot do anything else, we have no C1 left
	}

	//----------------------------------------------------------------------------
	// Have each Influencer make their prediction. Hold the predictions in recs[]
	//----------------------------------------------------------------------------
	var recs []Prediction
	for j := 0; j < len(i.Influencers); j++ {
		influencer := i.Influencers[j]
		prediction, probability, err := influencer.GetPrediction(T3)
		if err != nil {
			return BuyCount, err
		}
		recs = append(recs,
			Prediction{
				T3:          T3,
				T4:          T4,
				Action:      prediction,
				Probability: probability,
				IType:       reflect.TypeOf(influencer).String(),
				ID:          influencer.GetID(),
				Correct:     false, // don't know yet
			})
	}

	//------------------------------------------------------------------------------
	// make decision based on predictions.
	// TODO:  This code needs to be rethought. For now, I'm using a 'majority wins'
	//        strategy, which is probably not a good approach.
	//------------------------------------------------------------------------------
	if len(recs) < 1 {
		return BuyCount, fmt.Errorf("No predictions found")
	}
	buyVotes := 0
	holdVotes := 0
	buy := false // assume that we will not buy
	for j := 0; j < len(recs); j++ {
		if recs[j].Action == "buy" {
			buyVotes++
		} else {
			holdVotes++
		}
	}

	if buyVotes > holdVotes {
		buy = true
		// util.DPrintf("Buy decision on %s\n", T3.Format("1/2/2006"))
	}

	//------------------------------------------------------------------------------
	// if buy, get exchange rate and add to investments and update balances
	//------------------------------------------------------------------------------
	if buy {
		BuyCount++
		var inv Investment
		inv.id = util.GenerateRefNo()
		inv.T3C1 = i.cfg.StdInvestment
		if i.BalanceC1 < i.cfg.StdInvestment {
			inv.T3C1 = i.BalanceC1
		}
		inv.T3 = T3
		inv.T4 = T4 // we sell in Delta4 days
		er3 := data.ERFindRecord(inv.T3)
		if er3 == nil {
			return BuyCount, fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found", inv.T3.Format("1/2/2006"))
		}
		inv.ERT3 = er3.Close                       // exchange rate on T3
		inv.BuyC2 = inv.T3C1 * inv.ERT3            // amount of C2 we purchased on T3
		i.Investments = append(i.Investments, inv) // add it to the list of investments
		i.BalanceC1 -= inv.T3C1                    // we spent this much C1...
		i.BalanceC2 += inv.BuyC2                   // to purchase this much more C2

		//----------------------------------------------------------------
		// we need to update each of the Influencers predictions...
		//         *** ONLY THE BUY PREDICTIONS ARE SAVED ***
		//----------------------------------------------------------------
		for j := 0; j < len(i.Influencers); j++ {
			i.Influencers[j].AppendPrediction(recs[j])
		}
	}
	return BuyCount, nil
}

// SellConversion scans the Investment table for any Investment that concludes on
// the supplied t4.  When one is found, it converts C2 back to C1 and updates the
// Investments table withe the results of the conversion. Balances of C1 and C2
// are made after the conversion is completed.
//
// RETURNS:
//
//	The Investor
//	Number of investments sold
//	any error encountered, or nil if no errors were found
//
// ----------------------------------------------------------------------------------
func (i *Investor) SellConversion(t4 time.Time) (int, error) {
	var err error
	SellCount := 0
	err = nil

	// Look for investments to sell on t4
	jlen := len(i.Investments)
	for j := 0; j < jlen; j++ {
		//-------------------------------
		// Skip completed investments...
		//-------------------------------
		if i.Investments[j].Completed {
			continue
		}

		//------------------------------------------------
		// Check to see if the sell date has arrived...
		//------------------------------------------------
		dtSell := time.Time(i.Investments[j].T4)  // date on which we need to sell (convert) C2
		if t4.Equal(dtSell) || t4.After(dtSell) { // if the sell date has arrived...
			//-----------------------------------------------------------
			// The time has arrived. Get the exchange rate for today...
			//-----------------------------------------------------------
			er4 := data.ERFindRecord(t4) // get the exchange rate on t4
			if er4 == nil {
				err = fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found; Investment marked as completed", t4.Format("1/2/2006"))
				fmt.Printf("%s\n", err.Error())
				i.Investments[j].Completed = true
				continue
			}

			//-----------------------------------------------------------
			// Document this specific investment...
			//-----------------------------------------------------------
			i.Investments[j].ERT4 = er4.Close                                           // exchange rate on T4
			i.Investments[j].SellC2 = i.Investments[j].BuyC2                            // sell exactly what we bought on the associated T3
			i.Investments[j].T4C1 = i.Investments[j].SellC2 / i.Investments[j].ERT4     // amount of C1 we got back by selling SellC2 on T4 at the exchange rate on T4
			i.Investments[j].Profitable = i.Investments[j].T4C1 > i.Investments[j].T3C1 // did we make money on this trade?
			i.Investments[j].Completed = true                                           // this investment is now completed

			//-------------------------------------------------------------
			// Update Investor's totals having concluded the investment...
			//-------------------------------------------------------------
			i.BalanceC1 += i.Investments[j].T4C1   // we recovered this much C1...
			i.BalanceC2 -= i.Investments[j].SellC2 // by selling this C2

			//-------------------------------------------------------------
			// Update each Influencer's predictions...
			//-------------------------------------------------------------
			for k := 0; k < len(i.Influencers); k++ {
				i.Influencers[k].FinalizePrediction(i.Investments[j].T3, t4, i.Investments[j].Profitable)
			}

			SellCount++
		}
	}
	return SellCount, err
}

// OutputInvestments dumps the Investments table to a .csv file
// named investments.csv
//
// RETURNS
//
//	any error encountered or nil if no error
//
// ------------------------------------------------------------------------------------
func (i *Investor) OutputInvestments(j int) error {
	fname := fmt.Sprintf("Investments%03d.csv", j)
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	// the header row
	fmt.Fprintf(file, "id,T3,T4,T3C1,ERT3,BuyC2,SellC2,ERT4,T4C1,Completed,Profitable\n")

	// investment rows
	for _, inv := range i.Investments {
		//                  1  2  3      4     5      6      7     8      9 10 11
		fmt.Fprintf(file, "%s,%s,%s,%10.2f,%6.2f,%10.2f,%10.2f,%6.2f,%10.2f,%v,%v\n",
			inv.id,                      //1 s
			inv.T3.Format("01/02/2006"), //2 s
			inv.T4.Format("01/02/2006"), //3 s
			inv.T3C1,                    //4 f
			inv.ERT3,                    //5 f
			inv.BuyC2,                   //6 f
			inv.SellC2,                  //7 f
			inv.ERT4,                    //8 f
			inv.T4C1,                    //9 f
			inv.Completed,               //10 b
			inv.Profitable)              //11 b
	}
	return nil
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
	fmt.Fprintf(file, "\nInvluencers:\n")

	for j := 0; j < len(i.Influencers); j++ {
		fmt.Fprintf(file, "%d. %s\n", j+1, i.Influencers[j].DNA())
	}

	return nil
}

// ToString simply returns a printable version of the Investment struct as a string.
// ------------------------------------------------------------------------------------
func (i *Investment) ToString() string {
	s := fmt.Sprintf("    id		= %s\n", i.id)
	s += fmt.Sprintf("    T3		= %s\n", i.T3)
	s += fmt.Sprintf("    T4		= %s\n", i.T4)
	s += fmt.Sprintf("    T3C1		= %8.2f\n", i.T3C1)
	s += fmt.Sprintf("    BuyC2		= %8.2f\n", i.BuyC2)
	s += fmt.Sprintf("    SellC2	= %8.2f\n", i.SellC2)
	s += fmt.Sprintf("    ERT3		= %8.2f\n", i.ERT3)
	s += fmt.Sprintf("    ERT4		= %8.2f\n", i.ERT4)
	s += fmt.Sprintf("    T4C1	= %8.2f\n", i.T4C1)
	s += fmt.Sprintf("    Delta4	= %d\n", i.Delta4)
	return s
}

// FitnessScore calculates the fitness score for an Investor.
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
func (i *Investor) FitnessScore() float64 {
	if i.FitnessCalculated {
		return i.Fitness
	}

	// Calculate correctness...
	correct := 0
	total := 0
	jlen := len(i.Investments)
	for j := 0; j < jlen; j++ {
		total++
		if i.Investments[j].Completed && i.Investments[j].Profitable {
			correct++
		}
	}
	correctness := float64(0)
	if total > 0 {
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
