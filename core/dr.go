package core

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/util"

	"github.com/stmansour/psim/data"
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
	myInvestor          *Investor // my parent, the investor that holds me
	flagpos             int       // bit position in the data flags to indicate whether or not the value exists
	nilDataCount        int       // how many times did we encounter nil data in research
}

// GetNilDataCount returns the value for nilDataCount
func (p *DRInfluencer) GetNilDataCount() int {
	return p.nilDataCount
}

// IncNilDataCount the bit position of the valid data flag for this Influencer
func (p *DRInfluencer) IncNilDataCount() {
	p.nilDataCount++
}

// GetFlagPos the bit position of the valid data flag for this Influencer
func (p *DRInfluencer) GetFlagPos() int {
	return p.flagpos
}

// GetFitnessScore returns the current value of Fitness
func (p *DRInfluencer) GetFitnessScore() float64 {
	return p.Fitness
}

// SetFitnessScore sets this objects FitnessScore to the supplied value
// ------------------------------------------------------------------------
func (p *DRInfluencer) SetFitnessScore(x float64) {
	p.Fitness = x
	p.FitnessIsCalculated = true
}

// IsFitnessCalculated returns the boolean FitnessIsCalculated indicating whether
// or not we have a valid value for Fitness.
func (p *DRInfluencer) IsFitnessCalculated() bool {
	return p.FitnessIsCalculated
}

// MyInvestor returns a pointer to the investor object that holds this influencer
func (p *DRInfluencer) MyInvestor() *Investor {
	return p.myInvestor
}

// SetMyInvestor returns a pointer to the investor object that holds this influencer
func (p *DRInfluencer) SetMyInvestor(inv *Investor) {
	p.myInvestor = inv
}

// SetMyPredictions is used primarily for testing and sets the Prediction
// slice to the supplied value
func (p *DRInfluencer) SetMyPredictions(ps []Prediction) {
	p.MyPredictions = ps
}

// GetMyPredictions is used primarily for testing and returns MyPredictions
func (p *DRInfluencer) GetMyPredictions() []Prediction {
	return p.MyPredictions
}

// GetAppConfig - return the config struct
func (p *DRInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// GetLenMyPredictions - how many buy predictions
func (p *DRInfluencer) GetLenMyPredictions() int {
	return len(p.MyPredictions)
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *DRInfluencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizePrediction - finalize the results of this prediction
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
func (p *DRInfluencer) SetID() {
	rn := util.GenerateRefNo()
	p.ID = fmt.Sprintf("DRInfluencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, rn)
}

// Init - initializes a DRInfluencer
func (p *DRInfluencer) Init(i *Investor, cfg *util.AppConfig, delta4 int) {
	p.myInvestor = i
	p.cfg = cfg
	p.SetID()
	p.Delta4 = delta4
	p.flagpos = 2
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
	inv := p.MyInvestor()

	if inv == nil {
		log.Panicf("YIPES!  Influencer's MyInvestor is nil!\n")
	}
	if p.Delta4 != inv.Delta4 {
		util.DPrintf("YIPES!  Influencer Delta4 (%d) is not the same as Investor.Delta4 (%d)\n", p.Delta4, inv.Delta4)
	}
	return fmt.Sprintf("{%s,Delta1=%d,Delta2=%d}", p.Subclass(), p.Delta1, p.Delta2)
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
	return getPrediction(t3, p,
		func(rec1, rec2 *data.RatesAndRatiosRecord) (float64, float64, float64) {
			return rec1.DRRatio, rec2.DRRatio, rec1.DRRatio - rec2.DRRatio
		},
		p.cfg.InfPredDebug)
}

// CalculateFitnessScore - See explanation in common.go calculateFitnessScore
//
// RETURNS - the fitness score
// ------------------------------------------------------------------------------------
func (p *DRInfluencer) CalculateFitnessScore() float64 {
	return calculateFitnessScore(p, p.cfg)
}
