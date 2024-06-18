package newcore

import (
	"fmt"
	"log"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// LSMInfluencer is the base class for all influencers. It implements the
// Influencer interface. LSMInfluencer is derived from Locale Specific Influencer
// where one of those locales is "none".
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

// calculateAndSetValues calculates and sets values based on the specified metric name (field name).
// The field name is a string that uniquely identifies the metric in a record.
// ----------------------------------------------------------------------------------------------------
func (p *LSMInfluencer) calculateAndSetValues(pred *Prediction, fieldName string) (float64, float64, float64, bool) {
	val1, ok1 := pred.Recs[0].Fields[fieldName] // value at T1
	val2, ok2 := pred.Recs[1].Fields[fieldName] // value at T2
	if !ok1 || !ok2 {
		return 0, 0, 0, false
	}

	// TODO: explain this thoroughly
	stdDevSquared := val2.StdDevSquared      // used by trace
	pred.StdDevSquared = stdDevSquared       // pass it back in the prediction for trace
	delta := val2.Value - val1.Value         // change over T1 to T2
	da := delta / float64(p.Delta2-p.Delta1) // mean change between T1 and T2
	pred.AvgDelta = da                       // Average change between T2 and T1, used for trace
	x := p.cfg.StdDevVariationFactor         // notational simplification, use the factor from the config file
	res := da*da - x*x*stdDevSquared         // deltaAvg^2 - (x*stdDev)^2   if the result is positive, then we transact
	pred.Val1 = val1.Value                   // used in trace
	pred.Val2 = val2.Value                   // used in trace

	return val1.Value, val2.Value, res, true
}

// determineAction determines the action based on delta and the predictor type.
// ----------------------------------------------------------------------------------------------------
func (p *LSMInfluencer) determineAction(delta float64, predictor int) string {
	if delta < 0 {
		if predictor == newdata.SingleValLT || predictor == newdata.C1C2RatioLT {
			return "sell"
		}
		return "buy"
	}
	if predictor == newdata.SingleValLT || predictor == newdata.C1C2RatioLT {
		return "buy"
	}
	return "sell"
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
	pred.Action = "abstain" // Default action
	pred.T3 = t3
	pred.Delta1 = p.Delta1
	pred.Delta2 = p.Delta2

	if err := p.SetDataFields(&pred); err != nil {
		return &pred, err
	}

	// Setting default probabilities and weights for simplicity.
	pred.Probability = 1.0
	pred.Weight = 1.0
	switch p.Predictor {
	case newdata.SingleValGT, newdata.SingleValLT:
		val1, val2, res, ok := p.calculateAndSetValues(&pred, p.Metric)
		if !ok {
			// Data for the given fieldName is not available.
			return &pred, nil // Return immediately with pred.Action as "abstain".
		}
		if res > 0 {
			pred.Action = p.determineAction(val2-val1, p.Predictor)
		} else {
			pred.Action = "hold"
		}
		return &pred, nil

	case newdata.C1C2RatioGT, newdata.C1C2RatioLT:
		pred.Probability = 1.0
		pred.Weight = 1.0
		if len(pred.Recs[0].Fields) != 2 || len(pred.Recs[1].Fields) != 2 {
			return &pred, nil // need to abstain, the data was not available
		}

		// // There are two keys in pred.Recs[0].Fields. One is p.cfg.C1 + metric, the other is p.cfg.C2 + metric
		// // With Go maps, the order of the keys is not guaranteed. So we'll grab both and swap if we need to.
		// c := []string{}
		// for k := range pred.Recs {

		// 	if strings.HasPrefix(k, p.cfg.C1) {
		// 		c = append(c, k)
		// 	}
		// }
		// if c[0][:3] != p.cfg.C1 {
		// 	c[0], c[1] = c[1], c[0]
		// }

		// // Now the metric for C1 is in c[0] and the metric for C2 is in c[1]

		// vala -- is Fields[0] withing the std dev?
		valaT1, valaT2, res0, ok := p.calculateAndSetValues(&pred, p.cfg.C1+p.Metric)
		if !ok {
			// Data for the given fieldName is not available.
			return &pred, nil // Return immediately with pred.Action as "abstain".
		}
		// valb -- is Fields[1] withing the std dev?
		valbT1, valbT2, res1, ok := p.calculateAndSetValues(&pred, p.cfg.C2+p.Metric)
		if !ok {
			// Data for the given fieldName is not available.
			return &pred, nil // Return immediately with pred.Action as "abstain".
		}
		// pred.Val1 = rec1.Fields[pred.Fields[0].FQMetric()].Value / rec1.Fields[pred.Fields[1].FQMetric()].Value
		// pred.Val2 = rec2.Fields[pred.Fields[1].FQMetric()].Value / rec2.Fields[pred.Fields[1].FQMetric()].Value
		if valbT1 == 0 {
			valbT1 = 0.0000001
		}
		pred.Val1 = valaT1 / valbT1
		if valbT2 == 0 {
			valbT2 = 0.0000001
		}
		pred.Val2 = valaT2 / valbT2
		delta := pred.Val2 - pred.Val1

		if res0 > 0 && res1 > 0 { // both results need to be positive if we transact
			if delta < 0 {
				pred.Action = "buy"
				if p.Predictor == newdata.C1C2RatioLT {
					pred.Action = "sell"
				}
			} else {
				pred.Action = "sell"
				if p.Predictor == newdata.C1C2RatioLT {
					pred.Action = "buy"
				}
			}
		} else {
			pred.Action = "hold"
		}

		return &pred, nil

	default:
		log.Fatalf("Need to handle this case\n")
	}

	return &pred, nil
}

// SetDataFields files in the db record metric info for the Prediction
func (p *LSMInfluencer) SetDataFields(pred *Prediction) error {
	db := p.myInvestor.db
	sc := p.myInvestor.db.Mim.MInfluencerSubclasses[p.Metric]

	// the dates for data selection
	t1 := pred.T3.AddDate(0, 0, pred.Delta1)
	t2 := pred.T3.AddDate(0, 0, pred.Delta2)

	// the fields for the Select
	switch sc.LocaleType {
	case newdata.LocaleNone:
		pred.Fields = []newdata.FieldSelector{} // just the metric as-is
		pred.Fields = append(pred.Fields, newdata.FieldSelector{Metric: p.Metric})

	case newdata.LocaleC1C2:
		f1 := newdata.FieldSelector{Locale: p.MyInvestor().cfg.C1, Metric: p.Metric}
		f2 := newdata.FieldSelector{Locale: p.MyInvestor().cfg.C2, Metric: p.Metric}
		pred.Fields = []newdata.FieldSelector{f1, f2}

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
	return 1
}
