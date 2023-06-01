package core

import (
	"time"

	"github.com/stmansour/psim/util"
)

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
	Correct     bool      // was this profitable?
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
	SetDelta4(d4 int)
	GetPrediction(t3 time.Time) (string, float64, error)
	DNA() string
	AppendPrediction(pr Prediction)
	FinalizePrediction(t3, t4 time.Time, profitable bool)
	FitnessScore() float64
	Subclass() string
	MyInvestor() *Investor
	GetLenMyPredictions() int
}
