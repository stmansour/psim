package core

import (
	"time"

	"github.com/stmansour/psim/util"
)

// Steps to create a new Influencer:
// 1. Update the Influencer name in data/init.go
// 2. Update mapper in util/config.go
// 3. Update DInfo.Types in LoadCsvDB
// 4. Update LoadCsvDB in data/csvdb.go with a value for the new influencer ratio
// 5. Create an influencer class file in core/  copy and modify a file like dr.go
// 6

// Prediction holds the predictions from Influencers. Based on a list
// of these recommendations, the Investor will decide whether to "buy" or
// "hold". Also, each Influencer keeps a list of its predictions to assess
// its own performance.
// ----------------------------------------------------------------------------
type Prediction struct {
	T3          time.Time // date of buy
	T4          time.Time // date of sell
	Action      string    // buy or hold
	Probability float64   // probability that the action is correct
	IType       string    // specific influencer type
	ID          string    // id of this influencer
	Correct     bool      // was this profitable (correct)?
	Completed   bool      // has this Prediction been Finalized
}

// Influencer is a base class / struct definition for the types of objects that will
//
//	use a particular type of data to make a prediction to buy or hold currency.
//
// ------------------------------------------------------------------------------------------
type Influencer interface {
	Init(i *Investor, cfg *util.AppConfig, delta4 int)
	GetID() string
	SetID()
	Subclass() string
	SetDelta1(d int)
	SetDelta2(d int)
	SetDelta4(d int)
	SetAppConfig(cfg *util.AppConfig)
	GetDelta1() int
	GetDelta2() int
	GetDelta4() int
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
	GetPrediction(t3 time.Time) (string, float64, error)
	SetMyPredictions(ps []Prediction)
}
