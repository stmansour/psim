package core

import (
	"fmt"
	"time"

	"github.com/stmansour/psim/util"
)

// WTInfluencer is the Influencer that predicts based on the price of West Texas OIL
type WTInfluencer struct {
	cfg                 *util.AppConfig
	Delta1              int
	Delta2              int
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
func (p *WTInfluencer) GetNilDataCount() int {
	return p.nilDataCount
}

// IncNilDataCount the bit position of the valid data flag for this Influencer
func (p *WTInfluencer) IncNilDataCount() {
	p.nilDataCount++
}

// GetFlagPos the bit position of the valid data flag for this Influencer
func (p *WTInfluencer) GetFlagPos() int {
	return p.flagpos
}

// GetFitnessScore returns the current value of Fitness
func (p *WTInfluencer) GetFitnessScore() float64 {
	return p.Fitness
}

// SetFitnessScore sets this objects FitnessScore to the supplied value
func (p *WTInfluencer) SetFitnessScore(x float64) {
	p.Fitness = x
	p.FitnessIsCalculated = true
}

// IsFitnessCalculated returns the boolean FitnessIsCalculated indicating whether
// or not we have a valid value for Fitness.
func (p *WTInfluencer) IsFitnessCalculated() bool {
	return p.FitnessIsCalculated
}

// MyInvestor returns a pointer to the investor object that holds this influencer
func (p *WTInfluencer) MyInvestor() *Investor {
	return p.myInvestor
}

// SetMyInvestor returns a pointer to the investor object that holds this influencer
func (p *WTInfluencer) SetMyInvestor(inv *Investor) {
	p.myInvestor = inv
}

// SetMyPredictions is used primarily for testing and sets the Prediction
// slice to the supplied value
func (p *WTInfluencer) SetMyPredictions(ps []Prediction) {
	p.MyPredictions = ps
}

// GetMyPredictions is used primarily for testing and returns MyPredictions
func (p *WTInfluencer) GetMyPredictions() []Prediction {
	return p.MyPredictions
}

// GetAppConfig - return cfg struct
func (p *WTInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// GetLenMyPredictions - add a new buy prediction
func (p *WTInfluencer) GetLenMyPredictions() int {
	return len(p.MyPredictions)
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *WTInfluencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizePrediction - finalize the results of this prediction
func (p *WTInfluencer) FinalizePrediction(t3, t4 time.Time, profitable bool) {
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
func (p *WTInfluencer) GetID() string {
	return p.ID
}

// SetAppConfig - set cfg
func (p *WTInfluencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *WTInfluencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *WTInfluencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *WTInfluencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *WTInfluencer) SetDelta2(d int) {
	p.Delta2 = d
}

// SetID - set ID
func (p *WTInfluencer) SetID() {
	p.ID = fmt.Sprintf("WTInfluencer|%d|%d|%s", p.Delta1, p.Delta2, util.GenerateRefNo())
}

// Init - initializes a WTInfluencer
func (p *WTInfluencer) Init(i *Investor, cfg *util.AppConfig, delta4 int) {
	p.myInvestor = i
	p.cfg = cfg
	p.SetID()
	// p.flagpos = 5
}

// Subclass - a method that returns the Influencer subclass of this object
func (p *WTInfluencer) Subclass() string {
	return "WTInfluencer"
}

// DNA - returns the DNA of this influencer.
// A quick description of the type of Influencer and its key attributes.
// ----------------------------------------------------------------------------
func (p *WTInfluencer) DNA() string {
	return fmt.Sprintf("{%s,Delta1=%d,Delta2=%d}", p.Subclass(), p.Delta1, p.Delta2)
}

// GetPrediction - using the supplied date, it researches data and makes
// a prediction on whther to "buy" or "hold"
//
//	t1 = T3-p.Delta1
//	t2 = T3-p.Delta2
//
// # Uses LSN data
//
// RETURNS   prediction, r1, r2, probability, weight, err
//
//	val1       - an indicator value at T3 - p.Delta1 days
//	vaWT       - an indicator value at T3 - p.Delta2 days
//	confidence - probability that the prediction is correct.  TEMPORARY IMPL
//	             (it always returns 1.0 for confidence at this point)
//	weight     - how much to weight this decision (always 1.0 for now)
//	error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func (p *WTInfluencer) GetPrediction(t3 time.Time) (string, float64, float64, float64, float64, error) {
	prediction := "abstain" // assume no data
	colname := "WTOILClose"

	t1 := t3.AddDate(0, 0, p.GetDelta1())
	t2 := t3.AddDate(0, 0, p.GetDelta2())

	val1, err := getDataRecByColName(t1, colname)
	if err != nil {
		util.DPrintf("error getting %s Influencer value: %s\n", p.Subclass(), err.Error())
		return prediction, 0, 0, 1, 1, nil
	}
	val2, err := getDataRecByColName(t2, colname)
	if err != nil {
		util.DPrintf("error getting %s Influencer value: %s\n", p.Subclass(), err.Error())
		return prediction, 0, 0, 1, 1, nil
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
		fmt.Printf("%s: value(t1) = %6.2f, value(t2) = %6.2f, delta = %6.2f, prediction = %s\n", t3.Format("01/02/2006"), val1, val2, delta, prediction)
	}
	// todo - return proper probability and weight
	return prediction, val1, val2, 1.0, 1.0, nil
}

// CalculateFitnessScore - See explanation in common.go calculateFitnessScore
//
// RETURNS - the fitness score
// ------------------------------------------------------------------------------------
func (p *WTInfluencer) CalculateFitnessScore() float64 {
	return calculateFitnessScore(p, p.cfg)
}
