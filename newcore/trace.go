package newcore

import (
	"fmt"
	"math"
	"time"
)

// FormatInfluencer prints a readable version of the Influencers predictions
// It is used for debugging. It can be used to validate the decisions of an
// Investor.

// FormatPrediction prints a readable version of the Influencers predictions
// ----------------------------------------------------------------------------
func (i *Investor) FormatPrediction(p *Prediction, T3 time.Time) {
	stdDev := math.Sqrt(p.StdDevSquared)
	// name := i.db.Mim.MInfluencerSubclasses[p.IType].Name
	fmt.Printf("\t%s:  %s   (T1 %s [%4.2f] -  T2 %s [%4.2f]   AvgDelta: %.4f  StdDev: %.4f,  Factor: %.4f, Trigger: %.4f)\n",
		p.IType,
		p.Action,
		T3.AddDate(0, 0, int(p.Delta1)).Format("Jan _2, 2006"),
		p.Val1,
		T3.AddDate(0, 0, int(p.Delta2)).Format("Jan _2, 2006"),
		p.Val2,
		p.AvgDelta,
		stdDev,
		i.cfg.StdDevVariationFactor,
		i.cfg.StdDevVariationFactor*stdDev,
		// p.DeltaPct*100,  // TODO - remove DeltaPct
	)
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

// showBuy prints the key information about a buy
// ----------------------------------------------------------------------------
func (i *Investor) showBuy(inv *Investment) {
	fmt.Printf("        *** BUY ***   %8.2f %s (%8.2f %s, fee = %6.2f)\n", inv.T4C1, i.cfg.C1, inv.T3C2Buy, i.cfg.C2, inv.Fee)
}

// showSell prints the key information about a sell
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
	fmt.Printf("        *** SELL ***  %8.2f %s (fee: %6.2f), [%8.2f %s], investments affected: %d -->  %d profited, %d lost\n", tsc1, i.cfg.C1, fee, tsc2, i.cfg.C2, n, gains, losses)
}
