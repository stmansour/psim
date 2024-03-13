package newcore

import (
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Prediction holds the predictions from Influencers. Based on a list
// of these recommendations, the Investor will decide whether to "buy" or
// "hold". Also, each Influencer keeps a list of its predictions to assess
// its own performance.
// -----------------------------------------------------------------------------
type Prediction struct {
	Action      string                        // buy or hold
	Probability float64                       // probability that the action is correct
	Weight      float64                       // how heavily should this prediction weigh in the overall decision
	Delta1      int                           // research start offset
	Delta2      int                           // research stop offset
	T3          time.Time                     // date of buy
	Recs        []*newdata.EconometricsRecord // one, two, or more records of data, depending on the LocaleType
	Fields      []newdata.FieldSelector       // names of database fields
	Val1        float64                       // value or ratio at time T1
	Val2        float64                       // value ratio at time T2
	DeltaPct    float64                       // percent change from T1 to T2
	IType       string                        // specific influencer type
	ID          string                        // id of this influencer
	Correct     bool                          // was this profitable (correct)?
	Completed   bool                          // has this Prediction been Finalized
}

// Influencer is a base class / struct definition for the types of objects that will
//
//	use a particular type of data to make a prediction to buy or hold currency.
//
// ------------------------------------------------------------------------------------------
type Influencer interface {
	Init(i *Investor, cfg *util.AppConfig)
	GetID() string
	SetID()
	Subclass() string
	GetMetric() string
	SetDelta1(d int)
	SetDelta2(d int)
	SetAppConfig(cfg *util.AppConfig)
	GetDelta1() int
	GetDelta2() int
	GetFlagPos() int
	GetNilDataCount() int
	IncNilDataCount()
	DNA() string

	CalculateFitnessScore() float64
	IsFitnessCalculated() bool
	SetFitnessScore(x float64)
	GetFitnessScore() float64

	MyInvestor() *Investor
	SetMyInvestor(inv *Investor)

	AppendPrediction(pr Prediction)
	FinalizePrediction(t3, t4 time.Time, profitable bool)
	GetLenMyPredictions() int
	GetMyPredictions() []Prediction
	GetPrediction(t3 time.Time) (*Prediction, error)
	SetMyPredictions(ps []Prediction)
}
