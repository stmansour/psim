package core

import (
	"fmt"
	"math/rand"
	"psim/data"
	"psim/util"
	"time"
)

// DRInfluencer is the Influencer that predicts based on DiscountRate
type DRInfluencer struct {
	cfg           *util.AppConfig
	Delta1        int
	Delta2        int
	Delta4        int
	ID            string
	MyPredictions []Prediction
}

// AppendPrediction - add a new buy prediction
func (p *DRInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *DRInfluencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizPrediction - finalize the results of this prediction
func (p *DRInfluencer) FinalizePrediction(t3, t4 time.Time, profitable bool) {
	for i := 0; i < len(p.MyPredictions); i++ {
		if p.MyPredictions[i].Completed {
			continue
		}
		if t3.Equal(p.MyPredictions[i].T3) && t4.Equal(p.MyPredictions[i].T4) {
			p.MyPredictions[i].Correct = profitable
			p.MyPredictions[i].Completed = true
			return
		}
	}
}

// GetID - get ID string
func (p *DRInfluencer) GetID() string {
	return p.ID
}

// SetAppConfig - set cfg
func (p *DRInfluencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *DRInfluencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *DRInfluencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *DRInfluencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *DRInfluencer) SetDelta2(d int) {
	p.Delta2 = d
}

// GetDelta4 - get Delta4
func (p *DRInfluencer) GetDelta4() int {
	return p.Delta4
}

// SetDelta4 - set Delta4
func (p *DRInfluencer) SetDelta4(x int) {
	p.Delta4 = x
}

// SetID - set ID
func (p *DRInfluencer) SetID(x string) {
	p.ID = x
}

// Init - initializes a DRInfluencer
func (p *DRInfluencer) Init(cfg *util.AppConfig, delta4 int) {
	p.cfg = cfg
	p.Delta1 = rand.Intn(30) - 30 // -30 to -1
	for found := false; !found; {
		p.Delta2 = -1 - rand.Intn(6)
		found = p.Delta2 > p.Delta1 // make sure that T1 occurs prior to T2
	}
	p.Delta4 = 1 + rand.Intn(14)
	p.ID = fmt.Sprintf("DRInfluencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, util.GenerateRefNo())
}

// ProfileString - a quick description of the type of Influencer and
//
//	its key attributes.
//
// ----------------------------------------------------------------------------
func (p *DRInfluencer) ProfileString() string {
	return p.ID + fmt.Sprintf(" (Discount Rate, T1 = %d, T2 = %d, T4 = %d)", p.Delta1, p.Delta2, p.Delta4)
}

// GetPrediction - using the supplied date, it researches data and makes
// a prediction on whther to "buy" or "hold"
//
// RETURNS
//
//	action     -  "buy" or "hold"
//	prediction - probability of correctness - most valid for "buy" action
//	error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func (p *DRInfluencer) GetPrediction(t3 time.Time) (string, float64, error) {
	t1 := t3.AddDate(0, 0, p.Delta1)
	t2 := t3.AddDate(0, 0, p.Delta2)
	//---------------------------------------------------------------------------
	// Determine dDRR = (DiscountRateRatio at t1) - (DiscountRateRatio at t2)
	//---------------------------------------------------------------------------
	rec1 := data.DRFindRecord(t1)
	if rec1 == nil {
		err := fmt.Errorf("ExchangeRate Record for %s not found", t1.Format("1/2/2006"))
		return "hold", 0, err
	}
	rec2 := data.DRFindRecord(t2)
	if rec2 == nil {
		err := fmt.Errorf("ExchangeRate Record for %s not found", t2.Format("1/2/2006"))
		return "hold", 0, err
	}
	dDRR := rec1.USJPDRRatio - rec2.USJPDRRatio

	//-------------------------------------------------------------------------------
	// Prediction formula (based on the change in DiscountRateRatios):
	//     dDRR > 0:   buy on t3, sell on t4
	//     dDRR <= 0:  take no action
	//-------------------------------------------------------------------------------
	prediction := "hold"
	if dDRR > 0 {
		prediction = "buy"
	}

	// todo - return proper probability
	return prediction, 0.5, nil
}

// FitnessScore - Discount Rate Fitness Score.
//
//		   The purpose of a fitness score in a genetic algorithm is to evaluate how well
//		   a potential solution, represented by an individual in the population, solves
//		   the problem at hand. The fitness score is used to guide the selection, crossover,
//		   and mutation operations of the algorithm, as individuals with higher fitness
//		   scores are more likely to be selected for reproduction and to contribute to the
//		   next generation.
//
//		   I'm thinking that the Fitness Score for an Influencer should be based on
//		   (a) the percentage of time its "buy" predictions are "correct" and penalized
//	    for the incorrect predictions, and (b) the total number of predictions it made
//	    For example, let's consider 2 DiscountRate
//		   Influencers, let's call them DR-A and DR-B. Suppose that DR-A makes 100 "buy"
//		   predictions during the course of a simulation, and 80% of its predictions turn
//		   out to be correct.  During the same simulation, DR-B makes only 2 "buy"
//		   predictions, and both of them turn out to be correct.  In this case, I would
//		   say that 80 out of 100 predictions correct is a
//		   little more reliable than 2 out of 2.  I would have more confidence in DR-A than
//		   DR-B.  It is true that DR-B got 100% correct results, but it only made 2
//		   predictions... it doesn't feel like there's enough history to know if it
//		   just got lucky or if it really knows better.
//
//		   So, I think the form of the Fitness Score for an Influencer is:
//
//				cp = correct "buy" predictions
//				tp = total "buy" predictions
//		        ic = incorrect "buy" predictions = tp - cp
//
//		        fitness = (cp/tp) * (cp - ic)
//		                  (cp/tp) * (cp - (tp-cp))
//				          (cp/tp) * (2cp - tp)
//
// RETURNS
//
//	any error encountered or nil if no error
//
// ------------------------------------------------------------------------------------
func (p *DRInfluencer) FitnessScore(cp, tp int) float64 {
	c := float64(cp)
	t := float64(tp)
	return (c / t) * (2*c - t)
}
