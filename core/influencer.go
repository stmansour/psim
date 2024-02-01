package core

import (
	"time"

	"github.com/stmansour/psim/util"
)

// LINGUISTIC INFLUENCER LIST
//  L0 - uses LxxxLSPScore_ECON   -   Lexicoder Sentiment Positivity - where xxx is the ISO currency code:  USD, AUD, JPY, ...
// 	L1 - uses LxxxLSNScore_ECON   -   Lexicoder Sentiment Negativity
// 	L2 - uses LxxxWHAScore_ECON
// 	L3 - uses LxxxWHOScore_ECON
// 	L4 - uses LxxxWHLScore_ECON
// 	L5 - uses LxxxWPAScore_ECON
//  L6 - uses LxxxWDECount_ECON
// 	L7 - uses LxxxWDFCount_ECON
// 	L8 - uses LxxxWDPCount_ECON
// 	L9 - uses LxxxLIMCount_ECON
// 	LA - LALLLSNScore
//	LB - LALLLSPScore
//	LC - LALLWHAScore
//	LD - LALLWHOScore
//	LE - LALLWHLScore
//	LF - LALLWPAScore
//	LG - LALLWDECount
//	LH - LALLWDFCount
//	LI - LALLWDPCount
//	LJ - LALLWDMCount

// Steps to create a new Influencer:
//
// 1. Update the Influencer name in data/init.go (this may be obsolete now)
// 2. Update mapper in util/config.go
//    Update: ValidInfluencerSubclasses in util/config.go
//    Update: 6 lines in type FileConfig:  xxW1, xxW2, and xxMinDelta1, xxMinDelta2, xxMaxDelta1, xxMaxDelta2 in config.go
// 3. Update DInfo.Types in LoadCsvDB (data/csvdb.go)
// 4. Update LoadCsvDB in data/csvdb.go with a data flag position for the new influencer ratio
//    This will probably require renumbering all of them.
// 6. Update core/factory.c - NewInfluencer to create one
// 5. Create an influencer class file in core/  copy and modify a file like dr.go
//
//  TODO: this needs to be simplied in the next redesign
//-----------------------------------------------------------------------------

// Prediction holds the predictions from Influencers. Based on a list
// of these recommendations, the Investor will decide whether to "buy" or
// "hold". Also, each Influencer keeps a list of its predictions to assess
// its own performance.
// ----------------------------------------------------------------------------
type Prediction struct {
	Delta1      int64     // research start offset
	Delta2      int64     // research stop offset
	T3          time.Time // date of buy
	T4          time.Time // date of sell
	RT1         float64   // ratio at time T1
	RT2         float64   // ratio at time T2
	Action      string    // buy or hold
	Probability float64   // probability that the action is correct
	Weight      float64   // how heavily should this prediction weigh in the overall decision
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
	GetPrediction(t3 time.Time) (string, float64, float64, float64, float64, error)
	SetMyPredictions(ps []Prediction)
}
