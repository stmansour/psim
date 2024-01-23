package core

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/data"
	"github.com/stmansour/psim/util"
)

// L0Influencer is the Influencer that predicts based on Linguistic Sentiment associated with a country
type L0Influencer struct {
	cfg                 *util.AppConfig
	Delta1              int
	Delta2              int
	Delta4              int
	HoldMin             float64 // ratio diffs below this amount indicate sell, HoldMin - HoldMax defines the "zero" range
	HoldMax             float64 // ratio diffs above this amount indicate buy
	ID                  string
	FitnessIsCalculated bool
	FitnessIsNormalized bool
	Fitness             float64
	MyPredictions       []Prediction
	myInvestor          *Investor // my parent, the investor that holds me
	flagpos             int
	nilDataCount        int // how many times did we encounter nil data in research
}

// GetNilDataCount returns the value for nilDataCount
func (p *L0Influencer) GetNilDataCount() int {
	return p.nilDataCount
}

// IncNilDataCount the bit position of the valid data flag for this Influencer
func (p *L0Influencer) IncNilDataCount() {
	p.nilDataCount++
}

// GetFlagPos the bit position of the valid data flag for this Influencer
func (p *L0Influencer) GetFlagPos() int {
	return p.flagpos
}

// GetFitnessScore returns the current value of Fitness
func (p *L0Influencer) GetFitnessScore() float64 {
	return p.Fitness
}

// SetFitnessScore sets this objects FitnessScore to the supplied value
func (p *L0Influencer) SetFitnessScore(x float64) {
	p.Fitness = x
	p.FitnessIsCalculated = true
}

// IsFitnessCalculated returns the boolean FitnessIsCalculated indicating whether
// or not we have a valid value for Fitness.
func (p *L0Influencer) IsFitnessCalculated() bool {
	return p.FitnessIsCalculated
}

// MyInvestor returns a pointer to the investor object that holds this influencer
func (p *L0Influencer) MyInvestor() *Investor {
	return p.myInvestor
}

// SetMyInvestor returns a pointer to the investor object that holds this influencer
func (p *L0Influencer) SetMyInvestor(inv *Investor) {
	p.myInvestor = inv
}

// SetMyPredictions is used primarily for testing and sets the Prediction
// slice to the supplied value
func (p *L0Influencer) SetMyPredictions(ps []Prediction) {
	p.MyPredictions = ps
}

// GetMyPredictions is used primarily for testing and returns MyPredictions
func (p *L0Influencer) GetMyPredictions() []Prediction {
	return p.MyPredictions
}

// GetAppConfig - return cfg struct
func (p *L0Influencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// GetLenMyPredictions - add a new buy prediction
func (p *L0Influencer) GetLenMyPredictions() int {
	return len(p.MyPredictions)
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *L0Influencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizePrediction - finalize the results of this prediction
func (p *L0Influencer) FinalizePrediction(t3, t4 time.Time, profitable bool) {
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
func (p *L0Influencer) GetID() string {
	return p.ID
}

// SetAppConfig - set cfg
func (p *L0Influencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *L0Influencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *L0Influencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *L0Influencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *L0Influencer) SetDelta2(d int) {
	p.Delta2 = d
}

// GetDelta4 - get Delta4
func (p *L0Influencer) GetDelta4() int {
	return p.Delta4
}

// SetDelta4 - set Delta4
func (p *L0Influencer) SetDelta4(x int) {
	p.Delta4 = x
}

// SetID - set ID
func (p *L0Influencer) SetID() {
	p.ID = fmt.Sprintf("L0Influencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, util.GenerateRefNo())
}

// Init - initializes a L0Influencer
func (p *L0Influencer) Init(i *Investor, cfg *util.AppConfig, delta4 int) {
	p.myInvestor = i
	p.cfg = cfg
	p.SetID()
	p.Delta4 = delta4
	p.flagpos = 5
}

// Subclass - a method that returns the Influencer subclass of this object
func (p *L0Influencer) Subclass() string {
	return "L0Influencer"
}

// DNA - returns the DNA of this influencer.
// A quick description of the type of Influencer and its key attributes.
// ----------------------------------------------------------------------------
func (p *L0Influencer) DNA() string {
	return fmt.Sprintf("{%s,Delta1=%d,Delta2=%d}", p.Subclass(), p.Delta1, p.Delta2)
}

// GetPrediction - using the supplied date, it researches data and makes
// a prediction on whther to "buy" or "hold"
//
//	t1 = T3-p.Delta1
//	t2 = T3-p.Delta2
//
// val1 = L0_C1metric(t1)/L0_C2metric(t1)
// val2 = L0_C1metric(t2)/L0_C2metric(t2)
// delta = val2 - val1
// if ABS(delta) < HOLD then hold
// if delta > HOLD, then buy, else sell
//
// RETURNS   prediction, r1, r2, probability, weight, err
//
//	val1       - an indicator value at T3 - p.Delta1 days
//	val2       - an indicator value at T3 - p.Delta2 days
//	confidence - probability that the prediction is correct.  TEMPORARY IMPL
//	             (it always returns 1.0 for confidence at this point)
//	weight     - how much to weight this decision (always 1.0 for now)
//	error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func (p *L0Influencer) GetPrediction(t3 time.Time) (string, float64, float64, float64, float64, error) {
	prediction := "abstain" // assume no data
	C1 := p.cfg.C1
	C2 := p.cfg.C2

	t1 := t3.AddDate(0, 0, p.GetDelta1())
	t2 := t3.AddDate(0, 0, p.GetDelta2())

	// LxxxLSPScore_ECON
	// where xxx = JPY or USD
	val1, err := p.computeRatio(t1, C1, C2)
	if err != nil {
		log.Panicf("error getting %s Influencer value: %s", p.Subclass(), err.Error())
	}

	val2, err := p.computeRatio(t2, C1, C2)
	if err != nil {
		log.Panicf("error getting %s Influencer value: %s", p.Subclass(), err.Error())
	}
	delta := val2 - val1

	holdneg := p.cfg.HoldWindowNeg
	holdpos := p.cfg.HoldWindowPos

	prediction = "hold" // we have the data and made the calculation.  Assume "hold"
	if delta > holdpos {
		prediction = "buy" // check buy condition
	} else if delta < holdneg {
		prediction = "sell" // check sell condition
	}

	if p.cfg.InfPredDebug {
		fmt.Printf("%s: ratio(t1) = %6.2f, ratio(t2) = %6.2f, delta = %6.2f, prediction = %s\n", t3.Format("01/02/2006"), val1, val2, delta, prediction)
	}
	// todo - return proper probability and weight
	return prediction, val1, val2, 1.0, 1.0, nil
}

// computeRatio is the code to produce an Influencer's analysis for prediction at time t
//
// INPUTS
//
//	t - date used to
//
// RETURNS
//
//	val at time t
//	err - any error encountered
//
// ---------------------------------------------------------------------------------------------
func (p *L0Influencer) computeRatio(t time.Time, C1, C2 string) (float64, error) {
	rec := data.CSVDBFindLRecord(t)
	if rec == nil {
		err := fmt.Errorf("nildata: data.LinguisticDataRecord for %s not found", t.Format("1/2/2006"))
		return 0, err
	}
	c1valt, err := data.GetLValue(rec, C1, "LSPScore_ECON")
	if err != nil {
		log.Panicf("error getting Linguistic value: %s", err.Error())
	}
	c2valt, err := data.GetLValue(rec, C2, "LSPScore_ECON")
	if err != nil {
		log.Panicf("error getting Linguistic value: %s", err.Error())
	}
	return c1valt / c2valt, nil
}

// CalculateFitnessScore - See explanation in common.go calculateFitnessScore
//
// RETURNS - the fitness score
// ------------------------------------------------------------------------------------
func (p *L0Influencer) CalculateFitnessScore() float64 {
	return calculateFitnessScore(p, p.cfg)
}
