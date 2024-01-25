package core

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/data"
	"github.com/stmansour/psim/util"
)

// RatioFunc is a type used by the GetPrediction method of Influencer subclasses.
// It returns the metric ratio for the subclass's specific metrics.
// -----------------------------------------------------------------------------------
type RatioFunc func(*data.RatesAndRatiosRecord, *data.RatesAndRatiosRecord) (float64, float64, float64)

// getPrediction using the supplied date, it researches data and makes
//
//		 a prediction on whether to "buy", "sell", "hold", or "abstain".
//	  This function consolidates the same code used by all Influencers.
//
//		 Current method:  subtract the C2 data ratio from the C1 data ratio.
//		 The prediction is based on where the result lands in this space:
//
//		                    |<------------------HOLD------------------->|
//		           <-- SELL |<--      mn       -->|<--       mx      -->| BUY -->
//		 -------------------|---------------------|---------------------|------------------------
//		                                         0.0
//
// INPUTS
//
//	    t3 = date of the transaction
//	delta1 = # days prior to t3 to begin research
//	delta2 = # days prior to t3 to end research
//	bitpos = bit position of valid data flag for this Influencer
//	dbg    - print date, numbers, dRR, and prediction
//	mn,mx  - the min and max values less than and greater than 0
//	         that is considered the "hold space".
//
// RETURNS
//
//				action     - "buy", "sell", "hold", or "abstain"
//			                 "abstain" means remove it from the decision making process
//			                 because there was an error in doing its research (most likely
//			                 the data it needed was missing)
//				confidence - probability that the prediction is correct.  TEMPORARY IMPL
//		                     (it always returns 1.0 for confidence at this point)
//	            weight     - how much to weight this decision (always 1.0 for now)
//				error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func getPrediction(t3 time.Time, p Influencer, f RatioFunc, dbg bool, mn, mx float64) (string, float64, float64, float64, float64, error) {
	prediction := "abstain" // assume no data

	t1 := t3.AddDate(0, 0, p.GetDelta1())
	t2 := t3.AddDate(0, 0, p.GetDelta2())

	rec1 := data.CSVDBFindRecord(t1)
	if rec1 == nil {
		err := fmt.Errorf("nildata: data.RatesAndRatiosRecord for %s not found", t1.Format("1/2/2006"))
		return prediction, 0, 0, 0, 0, err
	}
	rec2 := data.CSVDBFindRecord(t2)
	if rec2 == nil {
		err := fmt.Errorf("nildata: data.RatesAndRatiosRecord for %s not found", t2.Format("1/2/2006"))
		return prediction, 0, 0, 0, 0, err
	}
	flagpos := p.GetFlagPos()
	flagslot := uint64(1 << flagpos)

	if ((rec1.FLAGS & flagslot) | (rec2.FLAGS & flagslot)) == 0 {
		if rec1.FLAGS&flagslot == 0 {
			p.IncNilDataCount()
		}
		if rec2.FLAGS&flagslot == 0 {
			p.IncNilDataCount()
		}
		err := fmt.Errorf("nildata")
		return prediction, 0, 0, 0, 0, err
	}

	d1, d2, dRR := f(rec1, rec2)

	prediction = "hold" // we have the data and made the calculation.  Assume "hold"
	if dRR > mx {
		prediction = "buy" // check buy condition
	} else if dRR < mn {
		prediction = "sell" // check sell condition
	}

	if dbg {
		fmt.Printf("%s: ratio(t1) = %5.2f, ratio(t2) = %5.2f, dRR = %5.2f, prediction = %s\n", t3.Format("01/02/2006"), d1, d2, dRR, prediction)
	}
	// todo - return proper probability and weight
	return prediction, d1, d2, 1.0, 1.0, nil
}

// calculatFitness - the generic Fitness Score calculator for many of the Influencer subclasses.
//
//	The purpose of a fitness score in a genetic algorithm is to evaluate how well
//	a potential solution, represented by an individual in the population, solves
//	the problem at hand. The fitness score is used to guide the selection, crossover,
//	and mutation operations of the algorithm, as individuals with higher fitness
//	scores are more likely to be selected for reproduction and to contribute to the
//	next generation.
//
//	I'm thinking that the Fitness Score for an Influencer should be based on
//	a) the percentage of time its "buy" predictions that are "correct", and
//	b) the total number of predictions it made
//
//	For example, let's consider 2 DiscountRate Influencers, DR-A
//	and DR-B. Suppose that DR-A makes 100 "buy" predictions during the course of
//	a simulation, and 80% of its predictions turn out to be correct.  During the
//	same simulation, DR-B makes only 2 "buy" predictions, and both of them turn
//	out to be correct.  In this case, I would say that 80 out of 100 predictions
//	correct is a little more reliable than 2 out of 2.  I would have more
//	confidence in DR-A than DR-B.  It is true that DR-B got 100% correct results,
//	but it only made 2 predictions... it doesn't feel like there's enough history
//	to know if it just got lucky or if it really knows better.
//
//	On the other hand, we don't want to overcompensate if an Influencer makes
//	many predictions, most of them are wrong, but it still scores better than
//	DR-B.  For example, suppose that DR-C made 200 "buy" predictions during that
//	simulation but its correctness is only 25%. We don't want its Fitness to be
//	far better than DR-B's.  So we probably need to attenuate the number of
//	correct predictions by the highest total of predictions made by any Influencer
//	of the same subclass.
//
//	So, I think the form of the Fitness Score for an Influencer is:
//
//		cp = count of correct "buy" predictions
//		tp = total number of "buy" predictions
//		mp = maximum number of buy predictions by any Influencer in the subclass
//		w1 = weighting factor for correctness
//		w2 = weighting factor for number of prediction
//
//		fitness = w1 * (cp/tp) + w2 * cp/mp
//
// RETURNS
//
//	a float64 value for the Fitness Score
//
// ------------------------------------------------------------------------------------
func calculateFitnessScore(p Influencer, cfg *util.AppConfig) float64 {
	//---------------------------------------------------
	// If it's already been calculated, just return it
	//---------------------------------------------------
	if p.IsFitnessCalculated() {
		return p.GetFitnessScore()
	}
	myPredictions := p.GetMyPredictions()

	t := float64(len(myPredictions))
	if t == 0 {
		return 0
	}
	cp := 0
	for i := 0; i < len(myPredictions); i++ {
		if myPredictions[i].Correct {
			cp++
		}
	}
	c := float64(cp)

	subclassKey := p.Subclass()[:2] // Extract the first two characters of the subclass name

	// FitnessScore := W1 * Correctness  +  W2 * TotalPredictions/(MaxPredictions+1)    --- NOTE: we add 1 to MaxPredictions to prevent division by 0
	x := cfg.SCInfo[subclassKey].FitnessW1*(c/t) + cfg.SCInfo[subclassKey].FitnessW2*(t/float64(1+p.MyInvestor().maxPredictions[subclassKey]))
	p.SetFitnessScore(x)

	return x
}

// computeRatio is the code to produce the ratio of an Influencer's data at time t
//
// INPUTS
//
//		    t - where in the time series
//	       C1 - used to determine the locale of the datatype
//	       C2 - used to determine the locale of the datatype
//	 datatype - LSPScore_ECON, LSNScore_ECON, WHAScore_ECON, ...
//
// RETURNS
//
//	val at time t
//	err - any error encountered
//
// ---------------------------------------------------------------------------------------------
func computeRatio(t time.Time, C1, C2, datatype string) (float64, error) {
	rec := data.CSVDBFindLRecord(t)
	if rec == nil {
		err := fmt.Errorf("nildata: data.LinguisticDataRecord for %s not found", t.Format("1/2/2006"))
		return 0, err
	}
	c1val, err := data.GetLValue(rec, C1, datatype)
	if err != nil {
		log.Panicf("error getting Linguistic value: %s", err.Error())
	}
	c2val, err := data.GetLValue(rec, C2, datatype)
	if err != nil {
		log.Panicf("error getting Linguistic value: %s", err.Error())
	}
	return c1val / c2val, nil
}
