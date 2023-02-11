package core

import "time"

type Influencer interface {
	// Returns the prediction made by the influencer for a specific date
	Predict(date time.Time) (string, float64, error)
	// Returns the influencer's identifier
	ID() string
	// Returns the influencer's settings
	Settings() map[string]interface{}
	// Updates the influencer's settings
	SetSettings(settings map[string]interface{}) error
}

func (i *Influencer) SetSetting(setting string, value float64) {
	i.settings[setting] = value
}

func (i *Influencer) Predict(date time.Time) (string, float64, error) {

}
