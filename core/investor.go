package core

import (
	"fmt"
	"os"
	"psim/data"
	"psim/util"
	"reflect"
	"time"
)

// Investor is the class definition
// =---------------------------------------------------------------------------
type Investor struct {
	cfg         *util.AppConfig // program wide configuration values
	BalanceC1   float64
	BalanceC2   float64
	Delta4      int // t4 = t3 + Delta4 - must be the same Delta4 for all influencers in this investor
	Investments []Investment
	Influencers []Influencer
	OutputIdx   int
}

// Recommendation holds the recommendations from Influencers. Based on a list
// of these recommendations, the Investor will decide whether to "buy" or
// "hold".
// ----------------------------------------------------------------------------
type Recommendation struct {
	Action      string
	Probability float64
	IType       string
}

// Investment describes a full transaction when the Investor decides to buy.
// The buy-related info is filled in at the time the purchase is made.  T4
// is also set at buy time.  When T4 arrives in the simulation, the
// transaction is completed and the remaining fields are filled in. All
// Investment structures are saved during the simulation. They can be dumped
// to a CSV file for analysis.
// ----------------------------------------------------------------------------
type Investment struct {
	id        string    // investment id
	T3        time.Time // date on which purchase of C2 was made
	T4        time.Time // date on which C2 will be exchanged for C1
	T3C1      float64   // amount of C2 to purchase at T3
	BuyC2     float64   // the amount of currency in C2 that C1 purchased on T3
	SellC2    float64   // for now, this is always going to be the same as BuyC2
	ERT3      float64   // the exchange rate on T3
	ERT4      float64   // the exchange rate on T4
	ResultC1  float64   // amount of currency C1 we were able to purchase with C2 on T4 at exchange rate ERT4
	Delta4    int       // t4 = t3 + Delta4 - "sell" date
	Completed bool      // true when the investmnet has been exchanged C2 for C1
}

// Init is called during Generation 1 to get things started.  All settable
// fields are set to random values.
// ----------------------------------------------------------------------------
func (i *Investor) Init(cfg *util.AppConfig) {
	i.cfg = cfg
	i.BalanceC1 = cfg.InitFunds
	i.Delta4 = util.RandomInRange(cfg.MinDelta4, cfg.MaxDelta4) // all Influencers will be constrained to this

	//------------------------------------------------------------------
	// Create a team of influencers.  For now, we're just going to add
	// one influencer to get things compiling and running.
	//------------------------------------------------------------------
	var inf Influencer = &DRInfluencer{}
	inf.Init(cfg, i.Delta4)
	i.Influencers = append(i.Influencers, inf)
}

// ProfileString returns a string containing descriptions all its influencers
// ----------------------------------------------------------------------------
func (i *Investor) ProfileString() string {
	s := ""
	for j := 0; j < len(i.Influencers); j++ {
		s += i.Influencers[j].ProfileString()
	}
	return s
}

// BuyConversion spins through all the influencers and asks for recommendations
// on whether to buy or hold on T3. Then the Investor decides whether to buy
// or hold.  If a "buy" is made, then an entry is added to the Investments
// list so that it can be completed when T4 arrives.  Currency C2 is purchased
// using C1.  If there are no remaining funds, the function returns immediately.
// Balances of C1 and C2 are adjusted whenever a conversion is done.
// ----------------------------------------------------------------------------
func (i *Investor) BuyConversion(T3 time.Time) (int, error) {
	BuyCount := 0
	if i.BalanceC1 <= 0.0 {
		return BuyCount, nil // we cannot do anything else, we have no C1 left
	}

	var recs []Recommendation
	for _, influencer := range i.Influencers {
		prediction, probability, err := influencer.GetPrediction(T3)
		if err != nil {
			return BuyCount, err
		}
		recs = append(recs,
			Recommendation{
				Action:      prediction,
				Probability: probability,
				IType:       reflect.TypeOf(influencer).String(),
			})
	}

	//------------------------------------------------------------------------------
	// make decision based on predictions.
	// TODO:  This code needs to be rethought. For now, I'm using a 'majority wins'
	//        strategy, which is probably not a good approach.
	//------------------------------------------------------------------------------
	if len(recs) < 1 {
		return BuyCount, fmt.Errorf("No recommendations found")
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
		inv.T3C1 = 100.00 // assume $100
		if i.BalanceC1 < 100.00 {
			inv.T3C1 = i.BalanceC1
		}
		inv.T3 = T3
		inv.T4 = T3.AddDate(0, 0, i.Delta4) // we sell in Delta4 days
		er3 := data.ERFindRecord(inv.T3)
		if er3 == nil {
			return BuyCount, fmt.Errorf("*** ERROR *** SellConversion: ExchangeRate Record for %s not found", inv.T3.Format("1/2/2006"))
		}
		inv.ERT3 = er3.Close                       // exchange rate on T3
		inv.BuyC2 = inv.T3C1 * inv.ERT3            // amount of C2 we purchased on T3
		i.Investments = append(i.Investments, inv) // add it to the list of investments
		i.BalanceC1 -= inv.T3C1                    // we spent this much C1...
		i.BalanceC2 += inv.BuyC2                   // to purchase this much more C2
		// util.DPrintf("New $100 investment, exchange date: %s, total pending = %d, C1 Bal: %8.2f %s, C2 Bal: %8.2f %s\n",
		// 	            inv.T4.Format("1/2/2006"), len(i.Investments), i.BalanceC1, i.cfg.C1, i.BalanceC2, i.cfg.C2)
	}
	return BuyCount, nil
}

// SellConversion scans the Investment table for any Investment that concludes on
// the supplied t4.  When one is found, it converts C2 back to C1 and updates the
// Investments table withe the results of the conversion. Balances of C1 and C2
// are made after the conversion is completed.
// ----------------------------------------------------------------------------------
func (i *Investor) SellConversion(t4 time.Time) (Investor, int, error) {
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

			// util.DPrintf("SellConversion -- BEFORE Txn:  C1 Balance: %8.2f %s,  C2 Balance: %8.2f %s\n", i.BalanceC1, i.cfg.C1, i.BalanceC2, i.cfg.C2)
			// util.DPrintf("SellConversion -- Exchange Rate on %s :  %8.2f\n", t4.Format("Jan 2, 2006"), er4.Close)

			i.Investments[j].ERT4 = er4.Close                                           // exchange rate on T4
			i.Investments[j].SellC2 = i.Investments[j].BuyC2                            // sell exactly what we bought on the associated T3
			i.Investments[j].ResultC1 = i.Investments[j].SellC2 / i.Investments[j].ERT4 // amount of C1 we got back by selling SellC2 on T4 at the exchange rate on T4

			i.BalanceC1 += i.Investments[j].ResultC1 // we recovered this much C1...
			i.BalanceC2 -= i.Investments[j].SellC2   // by selling this C2
			i.Investments[j].Completed = true
			SellCount += 1

			// util.DPrintf("SellConversion -- AFTER Txn: C1 Balance: %8.2f %s,  C2 Balance: %8.2f %s\n",
			// 	            i.BalanceC1, i.cfg.C1, i.BalanceC2, i.cfg.C2)
		}
	}
	// util.DPrintf("SellConversion -- @RETURN: C1 Balance: %8.2f %s,  C2 Balance: %8.2f %s\n",
	// 	            i.BalanceC1, i.cfg.C1, i.BalanceC2, i.cfg.C2)
	return (*i), SellCount, err
}

// OutputInvestments dumps the Investments table to a .csv file
// named investments.csv
// ------------------------------------------------------------------------------------
func (i *Investor) OutputInvestments(j int) error {
	fname := fmt.Sprintf("Investments%03d.csv", j)
	i.OutputIdx++
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	// write header row
	fmt.Fprintf(file, "id,T3,T4,T3C1,ERT3,BuyC2,SellC2,ERT4,ResultC1,Completed\n")

	// write investment rows
	for _, inv := range i.Investments {
		//                  1  2  3      4     5      6      7     8      9 10
		fmt.Fprintf(file, "%s,%s,%s,%10.2f,%6.2f,%10.2f,%10.2f,%6.2f,%10.2f,%v\n",
			inv.id,                      //1 s
			inv.T3.Format("01/02/2006"), //2 s
			inv.T4.Format("01/02/2006"), //3 s
			inv.T3C1,                    //4 f
			inv.ERT3,                    //5 f
			inv.BuyC2,                   //6 f
			inv.SellC2,                  //7 f
			inv.ERT4,                    //8 f
			inv.ResultC1,                //9 f
			inv.Completed)               //10 b
	}
	return nil
}

// InvestorProfile outputs information about this investor and its influencers
// to a file named "investorProfile.txt"
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
		fmt.Fprintf(file, "%d. %s\n", j+1, i.Influencers[j].ProfileString())
	}

	return nil
}

func (i *Investment) ToString() string {
	s := fmt.Sprintf("    id		= %s\n", i.id)
	s += fmt.Sprintf("    T3		= %s\n", i.T3)
	s += fmt.Sprintf("    T4		= %s\n", i.T4)
	s += fmt.Sprintf("    T3C1		= %8.2f\n", i.T3C1)
	s += fmt.Sprintf("    BuyC2		= %8.2f\n", i.BuyC2)
	s += fmt.Sprintf("    SellC2	= %8.2f\n", i.SellC2)
	s += fmt.Sprintf("    ERT3		= %8.2f\n", i.ERT3)
	s += fmt.Sprintf("    ERT4		= %8.2f\n", i.ERT4)
	s += fmt.Sprintf("    ResultC1	= %8.2f\n", i.ResultC1)
	s += fmt.Sprintf("    Delta4	= %d\n", i.Delta4)
	return s
}
