package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	json5 "github.com/yosuke-furukawa/json5/encoding/json5"
)

// CustomDate is used so that unmarshaling a date will work with
// dates in the format we want to enter them.
// ---------------------------------------------------------------------------
type CustomDate time.Time

// UnmarshalJSON implements an interface that allows our specially formatted
// dates to be parsed by go's json unmarshaling code.
// ----------------------------------------------------------------------------
func (t *CustomDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	parsedTime, err := time.Parse(`"2006-01-02"`, string(data))
	if err != nil {
		return err
	}
	*t = CustomDate(parsedTime)
	return nil
}

// AppConfig contains all the configuration values for the Simulator,
// Investors, and Influencers. I had to put it here in order to be visible
// to all areas of code in this project
// ---------------------------------------------------------------------------
type AppConfig struct {
	ExchangeRate   string     // the exchange rate that controls investing for this simulation
	C1             string     // Currency1 - the currency that we're trying to maximize
	C2             string     // Currency2 - the currency that we invest in to sell later and make a profit (or loss)
	DtStart        CustomDate // simulation begins on this date
	DtStop         CustomDate // simulation ends on this date
	PopulationSize int        // how many investors are in this population
	InitFunds      float64    // amount of funds each Investor is "staked" at the outset of the simulation
	StdInvestment  float64    // standard investment amount
	TradingDay     int        // this needs to be completely re-thought -- it's like a recurrence rule
	TradingTime    time.Time  // time of day when buy/sell is executed
	Generations    int        // current generation in the simulator
	MaxInf         int        // maximum number of influencers for any Investor
	MinInf         int        // minimum number of influencers for any Investor
	MinDelta1      int        // furthest back from t3 that t1 can be
	MaxDelta1      int        // closest to t3 that t1 can be
	MinDelta2      int        // furthest back from t3 that t2 can be
	MaxDelta2      int        // closest to t3 that t2 can be
	MinDelta4      int        // closest to t3 that t4 can be
	MaxDelta4      int        // furthest out from t3 that t4 can be
	DRW1           float64    // weighting for correctness part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	DRW2           float64    // weighting for prediction count part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	InvW1          float64    // weight for profit part of Investor FitnessScore
	InvW2          float64    // weight for correctness part of Investor FitnessScore
}

// LoadConfig reads the configuration data from config.json into an
// internal struct and returns that struct.
// ---------------------------------------------------------------------
func LoadConfig() (AppConfig, error) {
	var cfg AppConfig

	configFile, err := os.Open("config.json5")
	if err != nil {
		return cfg, fmt.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()

	byteValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %v", err)
	}

	err = json5.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config data: %v", err)
	}

	return cfg, nil
}
