package core

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/util"

	"github.com/stmansour/psim/data"
)

// BCInfluencer is the Influencer that predicts based on DiscountRate
type BCInfluencer struct {
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
func (p *BCInfluencer) GetNilDataCount() int {
	return p.nilDataCount
}

// IncNilDataCount the bit position of the valid data flag for this Influencer
func (p *BCInfluencer) IncNilDataCount() {
	p.nilDataCount++
}

// GetFlagPos the bit position of the valid data flag for this Influencer
func (p *BCInfluencer) GetFlagPos() int {
	return p.flagpos
}

// GetFitnessScore returns the current value of Fitness
func (p *BCInfluencer) GetFitnessScore() float64 {
	return p.Fitness
}

// SetFitnessScore sets this objects FitnessScore to the supplied value
// ------------------------------------------------------------------------
func (p *BCInfluencer) SetFitnessScore(x float64) {
	p.Fitness = x
	p.FitnessIsCalculated = true
}

// IsFitnessCalculated returns the boolean FitnessIsCalculated indicating whether
// or not we have a valid value for Fitness.
func (p *BCInfluencer) IsFitnessCalculated() bool {
	return p.FitnessIsCalculated
}

// MyInvestor returns a pointer to the investor object that holds this influencer
func (p *BCInfluencer) MyInvestor() *Investor {
	return p.myInvestor
}

// SetMyInvestor returns a pointer to the investor object that holds this influencer
func (p *BCInfluencer) SetMyInvestor(inv *Investor) {
	p.myInvestor = inv
}

// SetMyPredictions is used primarily for testing and sets the Prediction
// slice to the supplied value
func (p *BCInfluencer) SetMyPredictions(ps []Prediction) {
	p.MyPredictions = ps
}

// GetMyPredictions is used primarily for testing and returns MyPredictions
func (p *BCInfluencer) GetMyPredictions() []Prediction {
	return p.MyPredictions
}

// GetAppConfig - return the config struct
func (p *BCInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// GetLenMyPredictions - how many buy predictions
func (p *BCInfluencer) GetLenMyPredictions() int {
	return len(p.MyPredictions)
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *BCInfluencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizePrediction - finalize the results of this prediction
func (p *BCInfluencer) FinalizePrediction(t3, t4 time.Time, profitable bool) {
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
func (p *BCInfluencer) GetID() string {
	return p.ID
}

// SetAppConfig - set cfg
func (p *BCInfluencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *BCInfluencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *BCInfluencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *BCInfluencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *BCInfluencer) SetDelta2(d int) {
	p.Delta2 = d
}

// GetDelta4 - get Delta4
func (p *BCInfluencer) GetDelta4() int {
	return p.Delta4
}

// SetDelta4 - set Delta4
func (p *BCInfluencer) SetDelta4(x int) {
	p.Delta4 = x
}

// SetID - set ID
func (p *BCInfluencer) SetID() {
	rn := util.GenerateRefNo()
	p.ID = fmt.Sprintf("BCInfluencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, rn)
}

// Init - initializes a BCInfluencer
func (p *BCInfluencer) Init(i *Investor, cfg *util.AppConfig, delta4 int) {
	p.myInvestor = i
	p.cfg = cfg
	p.SetID()
	p.Delta4 = delta4
	p.flagpos = 2
}

// Subclass - a method that returns the Influencer subclass of this object
func (p *BCInfluencer) Subclass() string {
	return "BCInfluencer"
}

// DNA - a quick description of the type of Influencer and
//
//	its key attributes.
//
// ----------------------------------------------------------------------------
func (p *BCInfluencer) DNA() string {
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
func (p *BCInfluencer) GetPrediction(t3 time.Time) (string, float64, error) {
	return getPrediction(t3, p,
		func(rec1, rec2 *data.RatesAndRatiosRecord) (float64, float64, float64) {
			return rec1.BCRatio, rec2.BCRatio, rec1.BCRatio - rec2.BCRatio
		},
		p.cfg.InfPredDebug)
}

// CalculateFitnessScore - See explanation in common.go calculateFitnessScore
//
// RETURNS - the fitness score
// ------------------------------------------------------------------------------------
func (p *BCInfluencer) CalculateFitnessScore() float64 {
	return calculateFitnessScore(p, p.cfg)
}
