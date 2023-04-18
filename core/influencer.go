package core

import (
	"psim/util"
	"time"
)

// InfluencerPrediction describes a prediction from an influencer
type InfluencerPrediction struct {
	Prediction   string
	Probability  float64
	InfluencerID string
}

// Influencer is a base class / struct definition for the types of objects that will
//
//	use a particular type of data to make a prediction to buy or hold currency.
//
// ------------------------------------------------------------------------------------------
type Influencer interface {
	Init(cfg *util.AppConfig, delta4 int)
	GetPrediction(t3 time.Time) (string, float64, error)
	ProfileString() string
}
