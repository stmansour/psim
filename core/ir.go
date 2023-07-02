package core

import (
	"fmt"
	"time"

	"github.com/stmansour/psim/util"
)

// IRInfluencer is the Influencer that predicts based on DiscountRate
type IRInfluencer struct {
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
}

// GetFitnessScore returns the current value of Fitness
func (p *IRInfluencer) GetFitnessScore() float64 {
	return p.Fitness
}

// SetFitnessScore sets this objects FitnessScore to the supplied value
// ------------------------------------------------------------------------
func (p *IRInfluencer) SetFitnessScore(x float64) {
	p.Fitness = x
	p.FitnessIsCalculated = true
}

// IsFitnessCalculated returns the boolean FitnessIsCalculated indicating whether
// or not we have a valid value for Fitness.
func (p *IRInfluencer) IsFitnessCalculated() bool {
	return p.FitnessIsCalculated
}

// MyInvestor returns a pointer to the investor object that holds this influencer
func (p *IRInfluencer) MyInvestor() *Investor {
	return p.myInvestor
}

// SetMyInvestor returns a pointer to the investor object that holds this influencer
func (p *IRInfluencer) SetMyInvestor(inv *Investor) {
	p.myInvestor = inv
}

// SetMyPredictions is used primarily for testing and sets the Prediction
// slice to the supplied value
func (p *IRInfluencer) SetMyPredictions(ps []Prediction) {
	p.MyPredictions = ps
}

// GetMyPredictions is used primarily for testing and returns MyPredictions
func (p *IRInfluencer) GetMyPredictions() []Prediction {
	return p.MyPredictions
}

// GetAppConfig - return cfg struct
func (p *IRInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// GetLenMyPredictions - add a new buy prediction
func (p *IRInfluencer) GetLenMyPredictions() int {
	return len(p.MyPredictions)
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *IRInfluencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizePrediction - finalize the results of this prediction
func (p *IRInfluencer) FinalizePrediction(t3, t4 time.Time, profitable bool) {
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
func (p *IRInfluencer) GetID() string {
	return p.ID
}

// SetAppConfig - set cfg
func (p *IRInfluencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *IRInfluencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *IRInfluencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *IRInfluencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *IRInfluencer) SetDelta2(d int) {
	p.Delta2 = d
}

// GetDelta4 - get Delta4
func (p *IRInfluencer) GetDelta4() int {
	return p.Delta4
}

// SetDelta4 - set Delta4
func (p *IRInfluencer) SetDelta4(x int) {
	p.Delta4 = x
}

// SetID - set ID
func (p *IRInfluencer) SetID() {
	p.ID = fmt.Sprintf("IRInfluencer|%d|%d|%d|%s", p.Delta1, p.Delta2, p.Delta4, util.GenerateRefNo())
}

// Init - initializes a IRInfluencer
func (p *IRInfluencer) Init(i *Investor, cfg *util.AppConfig, delta4 int) {
	p.myInvestor = i
	p.cfg = cfg
	p.SetID()
	p.Delta4 = delta4
}

// Subclass - a method that returns the Influencer subclass of this object
func (p *IRInfluencer) Subclass() string {
	return "IRInfluencer"
}

// DNA - a quick description of the type of Influencer and
//
//	its key attributes.
//
// ----------------------------------------------------------------------------
func (p *IRInfluencer) DNA() string {
	return fmt.Sprintf("{%s,Delta1=%d,Delta2=%d,Delta4=%d}", p.Subclass(), p.Delta1, p.Delta2, p.Delta4)
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
//
//	func (p *IRInfluencer) GetPrediction(t3 time.Time) (string, float64, error) {
//		return getPrediction(t3, p.Delta1, p.Delta2, func(rec1, rec2 *data.RatesAndRatiosRecord) float64 {
//			return rec1.IRRatio - rec2.IRRatio
//		})
//	}
func (p *IRInfluencer) GetPrediction(t3 time.Time) (string, float64, error) {
	// t1 := t3.AddDate(0, 0, p.Delta1)
	// t2 := t3.AddDate(0, 0, p.Delta2)
	// //---------------------------------------------------------------------------
	// // Determine dDRR = (DiscountRateRatio at t1) - (DiscountRateRatio at t2)
	// //---------------------------------------------------------------------------
	// rec1 := data.CSVDBFindRecord(t1)
	// if rec1 == nil {
	// 	err := fmt.Errorf("ExchangeRate Record for %s not found", t1.Format("1/2/2006"))
	// 	return "hold", 0, err
	// }
	// rec2 := data.CSVDBFindRecord(t2)
	// if rec2 == nil {
	// 	err := fmt.Errorf("ExchangeRate Record for %s not found", t2.Format("1/2/2006"))
	// 	return "hold", 0, err
	// }
	// dDRR := rec1.Ratio - rec2.Ratio

	//-------------------------------------------------------------------------------
	// Prediction formula (based on the change in DiscountRateRatios):
	//     dDRR > 0:   buy on t3, sell on t4
	//     dDRR <= 0:  take no action
	//-------------------------------------------------------------------------------
	prediction := "hold"
	// if dDRR > 0 {
	// 	prediction = "buy"
	// }

	// todo - return proper probability
	return prediction, 0.5, nil
}

// CalculateFitnessScore - See explanation in common.go calculateFitnessScore
//
// RETURNS - the fitness score
// ------------------------------------------------------------------------------------
func (p *IRInfluencer) CalculateFitnessScore() float64 {
	return calculateFitnessScore(p, p.cfg)
}
