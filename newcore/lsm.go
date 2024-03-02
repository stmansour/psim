package newcore

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// LSMInfluencer is the Influencer that predicts based on Money Supply
type LSMInfluencer struct {
	Metric              string   // data type of subclass
	Blocs               []string // list of associated countries. If associated with C1 & C2, blocs[0] must be associated with C1, blocs[1] with C2
	LocaleType          int      // how to handle locales
	Predictor           int      // which predictor to use
	cfg                 *util.AppConfig
	Delta1              int
	Delta2              int
	HoldWindowNeg       float64 // positive number defining negative hold space:  from 0 to -HoldWindowNeg should be treated as 0
	HoldWindowPos       float64 // defines positive hold space: from 0 to HoldWindowPos should be treated as 0
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
func (p *LSMInfluencer) GetNilDataCount() int {
	return p.nilDataCount
}

// IncNilDataCount the bit position of the valid data flag for this Influencer
func (p *LSMInfluencer) IncNilDataCount() {
	p.nilDataCount++
}

// GetFlagPos the bit position of the valid data flag for this Influencer
func (p *LSMInfluencer) GetFlagPos() int {
	return p.flagpos
}

// GetFitnessScore returns the current value of Fitness
func (p *LSMInfluencer) GetFitnessScore() float64 {
	return p.Fitness
}

// SetFitnessScore sets this objects FitnessScore to the supplied value
func (p *LSMInfluencer) SetFitnessScore(x float64) {
	p.Fitness = x
	p.FitnessIsCalculated = true
}

// IsFitnessCalculated returns the boolean FitnessIsCalculated indicating whether
// or not we have a valid value for Fitness.
func (p *LSMInfluencer) IsFitnessCalculated() bool {
	return p.FitnessIsCalculated
}

// MyInvestor returns a pointer to the investor object that holds this influencer
func (p *LSMInfluencer) MyInvestor() *Investor {
	return p.myInvestor
}

// SetMyInvestor returns a pointer to the investor object that holds this influencer
func (p *LSMInfluencer) SetMyInvestor(inv *Investor) {
	p.myInvestor = inv
}

// SetMyPredictions is used primarily for testing and sets the Prediction
// slice to the supplied value
func (p *LSMInfluencer) SetMyPredictions(ps []Prediction) {
	p.MyPredictions = ps
}

// GetMyPredictions is used primarily for testing and returns MyPredictions
func (p *LSMInfluencer) GetMyPredictions() []Prediction {
	return p.MyPredictions
}

// GetAppConfig - return cfg struct
func (p *LSMInfluencer) GetAppConfig() *util.AppConfig {
	return p.cfg
}

// GetLenMyPredictions - add a new buy prediction
func (p *LSMInfluencer) GetLenMyPredictions() int {
	return len(p.MyPredictions)
}

// AppendPrediction - append a new prediction to the list of buy predictions
func (p *LSMInfluencer) AppendPrediction(pr Prediction) {
	p.MyPredictions = append(p.MyPredictions, pr)
}

// FinalizePrediction - finalize the results of this prediction
func (p *LSMInfluencer) FinalizePrediction(t3, t4 time.Time, profitable bool) {
	for i := 0; i < len(p.MyPredictions); i++ {
		if p.MyPredictions[i].Completed {
			continue
		}
		if t3.Equal(p.MyPredictions[i].T3) {
			p.MyPredictions[i].Correct = profitable
			p.MyPredictions[i].Completed = true
			return
		}
	}
}

// GetID - get ID string
func (p *LSMInfluencer) GetID() string {
	return p.ID
}

// SetAppConfig - set cfg
func (p *LSMInfluencer) SetAppConfig(cfg *util.AppConfig) {
	p.cfg = cfg
}

// GetDelta1 - get Delta1
func (p *LSMInfluencer) GetDelta1() int {
	return p.Delta1
}

// SetDelta1 - set Delta1
func (p *LSMInfluencer) SetDelta1(d int) {
	p.Delta1 = d
}

// GetDelta2 - get Delta2
func (p *LSMInfluencer) GetDelta2() int {
	return p.Delta2
}

// SetDelta2 - set Delta2
func (p *LSMInfluencer) SetDelta2(d int) {
	p.Delta2 = d
}

// SetID - set ID
func (p *LSMInfluencer) SetID() {
	p.ID = util.GenerateRefNo()
}

// Init - initializes a LSMInfluencer
func (p *LSMInfluencer) Init(i *Investor, cfg *util.AppConfig) {
	p.myInvestor = i
	p.cfg = cfg
	p.SetID()
	// p.flagpos = 5
}

// Subclass - a method that returns the Influencer subclass of this object
func (p *LSMInfluencer) Subclass() string {
	return "LSMInfluencer"
}

// GetMetric - a method that returns the Influencer subclass of this object
func (p *LSMInfluencer) GetMetric() string {
	return p.Metric
}

// DNA - returns the DNA of this influencer.
// A quick description of the type of Influencer and its key attributes.
// ----------------------------------------------------------------------------
func (p *LSMInfluencer) DNA() string {
	return fmt.Sprintf("{%s,Delta1=%d,Delta2=%d,Metric=%s}", p.Subclass(), p.Delta1, p.Delta2, p.Metric)
}

// GetPrediction - using the supplied date, it researches data and makes
// a prediction on whther to "buy" or "hold"
//
// RETURNS
//
//	action     -  "buy" or "hold" or "sell" or "abstain"
//	prediction - probability of correctness - most valid for "buy" action
//	error      - nil on success, error encountered otherwise
//
// ---------------------------------------------------------------------------
func (p *LSMInfluencer) GetPrediction(t3 time.Time) (*Prediction, error) {
	var pred Prediction
	pred.Action = "abstain" // assume no data
	pred.T3 = t3
	pred.Delta1 = p.Delta1
	pred.Delta2 = p.Delta2

	if err := p.SetDataFields(&pred); err != nil {
		return &pred, err
	}

	var res float64
	rec1 := pred.Recs[0]
	rec2 := pred.Recs[1]

	switch p.Predictor {
	case newdata.SingleValGT, newdata.SingleValLT:
		if len(pred.Recs[0].Fields) != 1 {
			return &pred, nil // need to abstain, the data was not available
		}
		pred.Val1 = rec1.Fields[pred.Fields[0]]
		pred.Val2 = rec2.Fields[pred.Fields[0]]
		res = pred.Val2 - pred.Val1
	case newdata.C1C2RatioGT, newdata.C1C2RatioLT:
		if len(pred.Recs[0].Fields) != 2 || len(pred.Recs[1].Fields) != 2 {
			return &pred, nil // need to abstain, the data was not available
		}
		pred.Val1 = rec1.Fields[pred.Fields[0]] / rec1.Fields[pred.Fields[1]]
		pred.Val2 = rec2.Fields[pred.Fields[0]] / rec2.Fields[pred.Fields[1]]
		res = pred.Val2 - pred.Val1
	default:
		log.Fatalf("Need to handle this case\n")
	}

	sc := p.myInvestor.db.Mim.MInfluencerSubclasses[p.Metric]
	pred.Action = "hold" // we have the data and made the calculation.  Assume "hold"

	switch p.Predictor {
	case newdata.SingleValGT, newdata.C1C2RatioGT:
		if res > sc.HoldWindowPos {
			pred.Action = "buy"
		} else if res < sc.HoldWindowNeg {
			pred.Action = "sell"
		}
	case newdata.SingleValLT, newdata.C1C2RatioLT:
		if res < sc.HoldWindowNeg {
			pred.Action = "buy" // check buy condition
		} else if res > sc.HoldWindowPos {
			pred.Action = "sell" // check sell condition
		}
	default:
		log.Fatalf("Need to handle this case\n")
	}

	// todo - return proper probability and weight
	pred.Probability = 1.0
	pred.Weight = 1.0
	return &pred, nil
}

// SetDataFields files in the db record metric info for the Prediction
func (p *LSMInfluencer) SetDataFields(pred *Prediction) error {
	db := p.myInvestor.db
	sc := p.myInvestor.db.Mim.MInfluencerSubclasses[p.Metric]

	// the dates for the
	t1 := pred.T3.AddDate(0, 0, pred.Delta1)
	t2 := pred.T3.AddDate(0, 0, pred.Delta2)

	// the fields for the Select
	switch sc.LocaleType {
	case newdata.LocaleNone:
		pred.Fields = []string{p.Metric} // just the metric as-is

	case newdata.LocaleC1C2:
		f1 := p.MyInvestor().cfg.C1 + p.Metric
		f2 := p.MyInvestor().cfg.C2 + p.Metric
		pred.Fields = []string{f1, f2}

	case newdata.LocaleBloc:
		log.Fatalf("Need to implement this!")
	}

	// Do the Select(s)
	var rec1, rec2 *newdata.EconometricsRecord
	var err error
	if sc.LocaleType == newdata.LocaleNone || sc.LocaleType == newdata.LocaleC1C2 {
		rec1, err = db.Select(t1, pred.Fields)
		if err != nil {
			return err
		}
		if rec1 == nil {
			err := fmt.Errorf("nildata: newdata.EconometricsRecord for %s not found", t1.Format("1/2/2006"))
			return err
		}
		pred.Recs = append(pred.Recs, rec1)
	}

	rec2, err = db.Select(t2, pred.Fields)
	if err != nil {
		return err
	}
	if rec2 == nil {
		err := fmt.Errorf("nildata: newdata.EconometricsRecord for %s not found", t2.Format("1/2/2006"))
		return err
	}
	pred.Recs = append(pred.Recs, rec2)
	return nil
}

// CalculateFitnessScore - See explanation in common.go calculateFitnessScore
//
// RETURNS - the fitness score
// ------------------------------------------------------------------------------------
func (p *LSMInfluencer) CalculateFitnessScore() float64 {
	/*
	**  I don't think this makes sense here. I think it needs to be calculated
	**  outside the context of the simulator.  The simulator does not determine
	**  if the influencer's predictions are correct. We only really determine
	**  if the Investor is correct.  And the CreateInfluencerFromDNA function
	**  does not base its creation of Influencers on their fitness values
	 */

	// //---------------------------------------------------
	// // If it's already been calculated, just return it
	// //---------------------------------------------------
	// if p.IsFitnessCalculated() {
	// 	return p.GetFitnessScore()
	// }
	// myPredictions := p.GetMyPredictions()

	// t := float64(len(myPredictions))
	// if t == 0 {
	// 	return 0
	// }
	// cp := 0
	// for i := 0; i < len(myPredictions); i++ {
	// 	if myPredictions[i].Correct {
	// 		cp++
	// 	}
	// }
	// c := float64(cp)

	// subclassKey := p.Subclass()[:2] // Extract the first two characters of the subclass name

	// // FitnessScore := W1 * Correctness  +  W2 * TotalPredictions/(MaxPredictions+1)    --- NOTE: we add 1 to MaxPredictions to prevent division by 0
	// x := cfg.SCInfo[subclassKey].FitnessW1*(c/t) + cfg.SCInfo[subclassKey].FitnessW2*(t/float64(1+p.MyInvestor().maxPredictions[subclassKey]))
	// p.SetFitnessScore(x)

	// return x
	return 1
}
