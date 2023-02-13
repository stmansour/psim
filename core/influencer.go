package core

import "time"

// Influencer is the interface for all Influencers
type Influencer interface {
	// Returns the prediction made by the influencer for a specific date
	Predict(date time.Time) (string, float64, error)
	// Returns the influencer's identifier
	ID() string
	// Updates the influencer's settings
	SetSettings(settings map[string]interface{}) error
}
