package newcore

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/stmansour/psim/util"
)

// InfluencerPrediction represents a single prediction for an Influencer
type InfluencerPrediction struct {
	Action   string
	Dt1      time.Time
	Dt2      time.Time
	Val1     float64
	Val2     float64
	AvgDelta float64
	StdDev   float64
	Factor   float64
	Trigger  float64
	Metric   string
}

// BSEvent buy or sell event
type BSEvent struct {
	T3C1                float64 // amount of c1 spent to get T3c2buy
	T3C2Buy             float64 // amount of c2 purchased for t3c1
	Fee                 float64
	Sell                bool // true if sell, false if buy
	TSC1                float64
	TSC2                float64
	Gains               int
	Losses              int
	InvestmentsAffected int
}

// TEvent represents a single event in the trace
type TEvent struct {
	Dt             time.Time
	COA            string
	C1             float64
	C2             float64
	PV             float64
	BuyVotes       float64
	HoldVotes      float64
	SellVotes      float64
	Abstains       float64
	InfPredictions []*InfluencerPrediction
	BSEvents       []*BSEvent
}

// Trace is a data struct for trace events
type Trace struct {
	Events []*TEvent
	Event  *TEvent // current tevent, set to nil after added to Events
}

// FormatInfluencer prints a readable version of the Influencers predictions
// It is used for debugging. It can be used to validate the decisions of an
// Investor.

// TraceInit initializes the trace object for this Investor
func (i *Investor) TraceInit() {
	i.COATrace.Event = nil
}

// SaveTrace adds the current Event to the trace Events list
func (i *Investor) SaveTrace() {
	i.COATrace.Events = append(i.COATrace.Events, i.COATrace.Event)
	i.COATrace.Event = nil
}

// TraceWriteFile writes the events to the trace file.
// ----------------------------------------------------------------------------
func (i *Investor) TraceWriteFile() error {
	var f *os.File
	var err error
	filename := fmt.Sprintf("trace-%s", i.ID)
	filename = i.cfg.GenerateFName(filename)
	f, err = os.Create(filename)
	if err != nil {
		log.Panicf("Error creating %s: %s\n", filename, err.Error())
	}

	defer f.Close()

	tsc1 := "Sell Amount in " + i.cfg.C1
	tsc2 := "Sell Amount in " + i.cfg.C2
	t3c1 := "Buy Amount in " + i.cfg.C1
	t3c2 := "Buy Amount in " + i.cfg.C2

	cvals := []util.FormatField{
		{Label: "Date", Fmt: "%q"},
		{Label: "COA", Fmt: "%q"},
		{Label: "BuySell", Fmt: "%q"},
		{Label: "BuyVotes", Fmt: "%.4f"},  //
		{Label: "HoldVotes", Fmt: "%.4f"}, //
		{Label: "SellVotes", Fmt: "%.4f"}, //
		{Label: "Abstains", Fmt: "%.4f"},  //
		{Label: "Metric", Fmt: "%q"},
		{Label: "Action", Fmt: "%q"},
		{Label: "T1", Fmt: "%q"},
		{Label: "Val1", Fmt: "%.4f"},
		{Label: "T2", Fmt: "%q"},
		{Label: "Val2", Fmt: "%.4f"},
		{Label: "AvgDelta", Fmt: "%.4f"},
		{Label: "StdDev", Fmt: "%.4f"},
		{Label: "Factor", Fmt: "%.4f"},
		{Label: "Trigger", Fmt: "%.4f"},
		{Label: "Sell", Fmt: "%.4f"}, //
		{Label: "T3C1", Fmt: "%.4f", ColumnHeader: t3c1},
		{Label: "T3C2Buy", Fmt: "%.4f", ColumnHeader: t3c2},
		{Label: "Fee", Fmt: "%.4f"},
		{Label: "TSC1", Fmt: "%.4f", ColumnHeader: tsc1},
		{Label: "TSC2", Fmt: "%.4f", ColumnHeader: tsc2},
		{Label: "InvestmentsAffected", Fmt: "%d"},
		{Label: "Gains", Fmt: "%d"},
		{Label: "Losses", Fmt: "%d"},
		{Label: "Investor ID", Fmt: "%q"},
		{Label: "C1bal", Fmt: "%.4f"},
		{Label: "C2bal", Fmt: "%.4f"},
		{Label: "PV", Fmt: "%.4f"},
		{Label: "ID", Fmt: "%s"},
	}

	formater := util.NewFormater(cvals)

	// header
	sf := formater.Header()
	fmt.Fprintf(f, "%s", sf+"\n")

	for _, e := range i.COATrace.Events {

		//------------------
		// DATE and ID
		//------------------
		id := []util.ColData{
			{Label: "Date", Val: e.Dt.Format("Jan _2, 2006")},
			{Label: "ID", Val: i.ID},
		}
		p := formater.Row(id)
		fmt.Fprintf(f, "%s", p+"\n")

		//------------------
		// INFLUENCER VOTES
		//------------------
		for _, ip := range e.InfPredictions {
			cd := []util.ColData{
				{Label: "Metric", Val: ip.Metric},
				{Label: "Action", Val: ip.Action},
				{Label: "T1", Val: ip.Dt1.Format("Jan _2, 2006")},
				{Label: "Val1", Val: ip.Val1},
				{Label: "T2", Val: ip.Dt2.Format("Jan _2, 2006")},
				{Label: "Val2", Val: ip.Val2},
				{Label: "AvgDelta", Val: ip.AvgDelta},
				{Label: "StdDev", Val: ip.StdDev},
				{Label: "Factor", Val: ip.Factor},
				{Label: "Trigger", Val: ip.Trigger},
			}
			p = formater.Row(cd)
			fmt.Fprintf(f, "%s", p+"\n")
		}

		//------------------------
		// COA - Course Of Action
		//------------------------
		coa := []util.ColData{
			{Label: "COA", Val: e.COA},
			{Label: "BuyVotes", Val: e.BuyVotes},
			{Label: "HoldVotes", Val: e.HoldVotes},
			{Label: "SellVotes", Val: e.SellVotes},
			{Label: "Abstains", Val: e.Abstains},
		}
		p = formater.Row(coa)
		fmt.Fprintf(f, "%s", p+"\n")

		//------------------
		// BUYS  /  SELLS
		//------------------
		if len(e.BSEvents) == 0 {
			cd := []util.ColData{
				{Label: "BuySell", Val: "Hold"},
			}
			fmt.Fprintf(f, "%s", formater.Row(cd)+"\n")
		}
		for _, bs := range e.BSEvents {
			b := "Buy"
			if bs.Sell {
				b = "Sell"
			}

			// BUY
			if !bs.Sell {
				cd := []util.ColData{
					{Label: "BuySell", Val: b},
					{Label: "T3C1", Val: bs.T3C1},
					{Label: "T3C2Buy", Val: bs.T3C2Buy},
					{Label: "Fee", Val: bs.Fee},
				}
				fmt.Fprintf(f, "%s", formater.Row(cd)+"\n")
			} else {
				// SELL
				cd := []util.ColData{
					{Label: "BuySell", Val: b},
					{Label: "TSC1", Val: bs.TSC1},
					{Label: "TSC2", Val: bs.TSC2},
					{Label: "Fee", Val: bs.Fee},
					{Label: "InvestmentsAffected", Val: bs.InvestmentsAffected},
					{Label: "Gains", Val: bs.Gains},
					{Label: "Losses", Val: bs.Losses},
				}
				p = formater.Row(cd)
				fmt.Fprintf(f, "%s", p+"\n")
			}
		}

		//------------------
		// BALANCES
		//------------------
		cd := []util.ColData{
			{Label: "C1bal", Val: e.C1},
			{Label: "C2bal", Val: e.C2},
			{Label: "PV", Val: e.PV},
		}
		p = formater.Row(cd)
		fmt.Fprintf(f, "%s", p+"\n\n")
	}

	return nil
}

// FormatPrediction prints a readable version of the Influencers predictions
// ----------------------------------------------------------------------------
func (i *Investor) FormatPrediction(p *Prediction, T3 time.Time) {
	//-----------------------------------------------
	// create the current Event if it doesn't exist
	//-----------------------------------------------
	if i.COATrace.Event == nil {
		i.COATrace.Event = &TEvent{}
		i.COATrace.Event.Dt = T3
	}

	T1 := T3.AddDate(0, 0, int(p.Delta1))
	T2 := T3.AddDate(0, 0, int(p.Delta2))
	stdDev := math.Sqrt(p.StdDevSquared)

	fmt.Printf("\t%s:  %s   (T1 %s [%4.2f] -  T2 %s [%4.2f]   AvgDelta: %.4f  StdDev: %.4f,  Factor: %.4f, Trigger: %.4f)\n",
		p.Metric,
		p.Action,
		T1.Format("Jan _2, 2006"),
		p.Val1,
		T2.Format("Jan _2, 2006"),
		p.Val2,
		p.AvgDelta,
		stdDev,
		i.cfg.StdDevVariationFactor,
		i.cfg.StdDevVariationFactor*stdDev,
	)

	// add this Influencer's prediction
	ip := InfluencerPrediction{
		Action:   p.Action,
		Dt1:      T1,
		Val1:     p.Val1,
		Dt2:      T2,
		Val2:     p.Val2,
		AvgDelta: p.AvgDelta,
		StdDev:   stdDev,
		Factor:   i.cfg.StdDevVariationFactor,
		Trigger:  i.cfg.StdDevVariationFactor * stdDev,
		Metric:   p.Metric,
	}

	i.COATrace.Event.InfPredictions = append(i.COATrace.Event.InfPredictions, &ip)

}

// FormatCOA prints a readable version of the Influencers predictions
// ----------------------------------------------------------------------------
func (i *Investor) FormatCOA(c *CourseOfAction) {
	i.COATrace.Event.Abstains = c.Abstains
	i.COATrace.Event.BuyVotes = c.BuyVotes
	i.COATrace.Event.HoldVotes = c.HoldVotes
	i.COATrace.Event.SellVotes = c.SellVotes
	fmt.Printf("\tCOA:  Action: %s  %3.0f%%  (buy: %3.2f, hold: %3.2f, sell: %3.2f, abs: %3.2f)\n", c.Action, c.ActionPct*100, c.BuyVotes, c.HoldVotes, c.SellVotes, c.Abstains)
}

// PortfolioToString returns a string with the portfolio balance at time t
// ----------------------------------------------------------------------------
func (i *Investor) PortfolioToString(t time.Time) string {
	pv := i.PortfolioValue(t)
	i.COATrace.Event.C1 = i.BalanceC1
	i.COATrace.Event.C2 = i.BalanceC2
	i.COATrace.Event.PV = pv

	return fmt.Sprintf("C1bal = %6.2f %s, C2bal = %6.2f %s, PV = %6.2f %s\n", i.BalanceC1, i.cfg.C1, i.BalanceC2, i.cfg.C2, pv, i.cfg.C1)

}

// showBuy prints the key information about a buy
// ----------------------------------------------------------------------------
func (i *Investor) showBuy(inv *Investment) {

	bse := &BSEvent{}
	bse.T3C1 = inv.T3C1       // amount of C1 used to purchase C2
	bse.T3C2Buy = inv.T3C2Buy // amount of C2 purchased for T3C1
	bse.Sell = false          // false - this is a buy
	bse.Fee = inv.Fee         // calculated fee for this purchase
	i.COATrace.Event.BSEvents = append(i.COATrace.Event.BSEvents, bse)
	fmt.Printf("        *** BUY ***   %8.2f %s (%8.2f %s, fee = %6.2f)\n", inv.T3C1, i.cfg.C1, inv.T3C2Buy, i.cfg.C2, inv.Fee)

}

// showSell prints the key information about a sell
// inv the investment struct describing the sell
// tsc1 The amount of the sale in C1
// tsc2 The amount of the sale in C2
// fee The fee associated with the sale
// ----------------------------------------------------------------------------
func (i *Investor) showSell(inv *Investment, tsc1, tsc2, fee float64) {
	gains := 0
	losses := 0
	n := len(inv.Chunks)
	for j := 0; j < n; j++ {
		if inv.Chunks[j].Profitable {
			gains++
		} else {
			losses++
		}
	}

	bse := &BSEvent{}
	bse.Sell = true
	bse.TSC1 = tsc1
	bse.TSC2 = tsc2
	bse.Fee = fee
	bse.InvestmentsAffected = n
	bse.Gains = gains
	bse.Losses = losses

	i.COATrace.Event.BSEvents = append(i.COATrace.Event.BSEvents, bse)
	fmt.Printf("        *** SELL ***  %8.2f %s (fee: %6.2f), [%8.2f %s], investments affected: %d -->  %d profited, %d lost\n", tsc1, i.cfg.C1, fee, tsc2, i.cfg.C2, n, gains, losses)
}
