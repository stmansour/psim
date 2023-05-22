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
	cfg                 *util.AppConfig
	Delta1              int
	Delta2              int
	Delta4              int
	ID                  string
	FitnessIsCalculated bool
	FitnessIsNormalized bool
	Fitness             float64
	MyPredictions       []Prediction
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
	p.Delta1 = -4 - rand.Intn(26) // -30 to -4
	l2 := 6
	if (-p.Delta1) < l2 {
		l2 = (-p.Delta1) - 2
	}
	for found := false; !found; {
		p.Delta2 = -1 - rand.Intn(l2)
		found = p.Delta2 > p.Delta1 // make sure that T1 occurs prior to T2
	}
	p.Delta4 = 1 + rand.Intn(14)
	p.ID = fmt.Sprintf("DRInfluencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, util.GenerateRefNo())
}

// Subclass - a method that returns the Influencer subclass of this object
func (p *DRInfluencer) Subclass() string {
	return "DRInfluencer"
}

// DNA - a quick description of the type of Influencer and
//
//	its key attributes.
//
// ----------------------------------------------------------------------------
func (p *DRInfluencer) DNA() string {
	return fmt.Sprintf("{%s,%d,%d,%d}", p.Subclass(), p.Delta1, p.Delta2, p.Delta4)
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
//			   The purpose of a fitness score in a genetic algorithm is to evaluate how well
//			   a potential solution, represented by an individual in the population, solves
//			   the problem at hand. The fitness score is used to guide the selection, crossover,
//			   and mutation operations of the algorithm, as individuals with higher fitness
//			   scores are more likely to be selected for reproduction and to contribute to the
//			   next generation.
//
//			   I'm thinking that the Fitness Score for an Influencer should be based on
//			   (a) the percentage of time its "buy" predictions that are "correct", and
//			   (b) the total number of predictions it made
//
//		       For example, let's consider 2 DiscountRate Influencers, DR-A
//			   and DR-B. Suppose that DR-A makes 100 "buy" predictions during the course of
//			   a simulation, and 80% of its predictions turn out to be correct.  During the
//			   same simulation, DR-B makes only 2 "buy" predictions, and both of them turn
//			   out to be correct.  In this case, I would say that 80 out of 100 predictions
//			   correct is a little more reliable than 2 out of 2.  I would have more
//			   confidence in DR-A than DR-B.  It is true that DR-B got 100% correct results,
//			   but it only made 2 predictions... it doesn't feel like there's enough history
//			   to know if it just got lucky or if it really knows better.
//
//			   On the other hand, we don't want to overcompensate if an Influencer makes
//	        many predictions, most of them are wrong, but it still scores better than
//	        DR-B.  For example, suppose that DR-C made 200 "buy" predictions during that
//	        simulation but its correctness is only 25%. We don't want its Fitness to be
//	        far better than DR-B's.  So we probably need to attenuate the number of
//	        correct predictions by the highest total of predictions made by any Influncer
//	        of the same subclass.
//
//			   So, I think the form of the Fitness Score for an Influencer is:
//
//					cp = count of correct "buy" predictions
//					tp = total number of "buy" predictions
//	             mp = maximum number of buy predictions by any Influencer in the subclass
//
//					w1 = weighting factor for correctness
//	             w2 = weighting factor for
//
//			        fitness = w1 * (cp/tp) + w2 * cp/mp
//
// RETURNS
//
//	a float64 value for the Fitness Score
//
// ------------------------------------------------------------------------------------
func (p *DRInfluencer) FitnessScore() float64 {

	//---------------------------------------------------
	// If it's already been calculated, just return it
	//---------------------------------------------------
	if p.FitnessIsCalculated {
		return p.Fitness
	}

	// MaxPredictions is the maximum number of predictions made by any influencer of the same subclass
	// MaxPredictions := getMaxPredictions(subclass)
	// FitnessScore := w1*Correctness + w2*(float64(TotalPredictions)/float64(MaxPredictions+1))

	t := float64(len(p.MyPredictions))
	if t == 0 {
		return 0
	}
	cp := 0
	for i := 0; i < len(p.MyPredictions); i++ {
		if p.MyPredictions[i].Correct {
			cp++
		}
	}
	c := float64(cp)
	p.Fitness = (c / t) * c
	p.FitnessIsCalculated = true
	return p.Fitness
}
