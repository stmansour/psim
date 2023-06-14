package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	json5 "github.com/yosuke-furukawa/json5/encoding/json5"
)

// CustomDate is used so that unmarshaling a date will work with
// dates in the format we want to enter them.
// ---------------------------------------------------------------------------
type CustomDate time.Time

type RangeLimits struct {
	MinDelta1 int // furthest back from t3 that t1 can be
	MaxDelta1 int // closest to t3 that t1 can be
	MinDelta2 int // furthest back from t3 that t2 can be
	MaxDelta2 int // closest to t3 that t2 can be
}

// UnmarshalJSON implements an interface that allows our specially formatted
// dates to be parsed by go's json unmarshaling code.
// ----------------------------------------------------------------------------
func (t *CustomDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	parsedTime, err := StringToDate(string(data))
	if err != nil {
		return err
	}
	*t = CustomDate(parsedTime)
	return nil
}

// FileConfig contains all the configuration values for the Simulator,
// Investors, and Influencers. I had to put it here in order to be visible
// to all areas of code in this project
// ---------------------------------------------------------------------------
type FileConfig struct {
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
	DRMinDelta1    int        `json:"DRMinDelta1"`
	DRMaxDelta1    int        `json:"DRMaxDelta1"`
	DRMinDelta2    int        `json:"DRMinDelta2"`
	DRMaxDelta2    int        `json:"DRMaxDelta2"`
	URMinDelta1    int        `json:"URMinDelta1"`
	URMaxDelta1    int        `json:"URMaxDelta1"`
	URMinDelta2    int        `json:"URMinDelta2"`
	URMaxDelta2    int        `json:"URMaxDelta2"`
	IRMinDelta1    int        `json:"IRMinDelta1"`
	IRMaxDelta1    int        `json:"IRMaxDelta1"`
	IRMinDelta2    int        `json:"IRMinDelta2"`
	IRMaxDelta2    int        `json:"IRMaxDelta2"`
	MinDelta4      int        // closest to t3 that t4 can be
	MaxDelta4      int        // furthest out from t3 that t4 can be
	DRW1           float64    // weighting for correctness part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	DRW2           float64    // weighting for prediction count part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	InvW1          float64    // weight for profit part of Investor FitnessScore
	InvW2          float64    // weight for correctness part of Investor FitnessScore
	MutationRate   int        // 1 - 100 indicating the % of mutation
	DBSource       string     // {CSV | Database | OnlineService}
}

// AppConfig contains all the configuration values for the Simulator,
// Investors, and Influencers. I had to put it here in order to be visible
// to all areas of code in this project
// ---------------------------------------------------------------------------
type AppConfig struct {
	// ExchangeRate   string     // the exchange rate that controls investing for this simulation
	C1             string                 // Currency1 - the currency that we're trying to maximize
	C2             string                 // Currency2 - the currency that we invest in to sell later and make a profit (or loss)
	DtStart        CustomDate             // simulation begins on this date
	DtStop         CustomDate             // simulation ends on this date
	PopulationSize int                    // how many investors are in this population
	InitFunds      float64                // amount of funds each Investor is "staked" at the outset of the simulation
	StdInvestment  float64                // standard investment amount
	TradingDay     int                    // this needs to be completely re-thought -- it's like a recurrence rule
	TradingTime    time.Time              // time of day when buy/sell is executed
	Generations    int                    // current generation in the simulator
	MaxInf         int                    // maximum number of influencers for any Investor
	MinInf         int                    // minimum number of influencers for any Investor
	DLimits        map[string]RangeLimits // map from Influencer subtype to its research time limits
	MinDelta4      int                    // closest to t3 that t4 can be
	MaxDelta4      int                    // furthest out from t3 that t4 can be
	DRW1           float64                // weighting for correctness part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	DRW2           float64                // weighting for prediction count part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	InvW1          float64                // weight for profit part of Investor FitnessScore
	InvW2          float64                // weight for correctness part of Investor FitnessScore
	MutationRate   int                    // 1 - 100 indicating the % of mutation
	DBSource       string                 // {CSV | Database | OnlineService}
}

func hasPrefix(tag string, prefixes []string, mod string) bool {
	for i := 0; i < len(prefixes); i++ {
		pfx := prefixes[i] + mod
		if strings.HasPrefix(tag, pfx) {
			return true
		}
	}
	return false
}

// LoadConfig reads the configuration data from config.json into an
// internal struct and returns that struct.
// ---------------------------------------------------------------------
func LoadConfig() (AppConfig, error) {
	var cfg AppConfig
	var fcfg FileConfig

	configFile, err := os.Open("config.json5")
	if err != nil {
		return cfg, fmt.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()

	byteValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %v", err)
	}

	// read into our config struct
	//-------------------------------------
	err = json5.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config data: %v", err)
	}

	// now read into fcfg to pick up the other values...
	//------------------------------------------------------
	err = json5.Unmarshal(byteValue, &fcfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config data: %v", err)
	}

	//-------------------------------------------
	// now build the map[subclass]RangeLimits
	//-------------------------------------------
	mapper := make(map[string]RangeLimits)
	t := reflect.TypeOf(fcfg)
	v := reflect.ValueOf(fcfg)

	prefixes := []string{"DR", "IR", "UR"}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		value := v.Field(i).Interface()

		if strings.HasSuffix(jsonTag, "Delta1") || strings.HasSuffix(jsonTag, "Delta2") {
			subclassName := jsonTag[:2]
			rangeLimits := mapper[subclassName]
			if strings.HasSuffix(jsonTag, "Delta1") {
				if hasPrefix(jsonTag, prefixes, "Min") {
					rangeLimits.MinDelta1 = value.(int)
				} else if hasPrefix(jsonTag, prefixes, "Max") {
					rangeLimits.MaxDelta1 = value.(int)
				}
			} else if strings.HasSuffix(jsonTag, "Delta2") {
				if hasPrefix(jsonTag, prefixes, "Min") {
					rangeLimits.MinDelta2 = value.(int)
				} else if hasPrefix(jsonTag, prefixes, "Max") {
					rangeLimits.MaxDelta2 = value.(int)
				}
			}
			mapper[subclassName] = rangeLimits
		}
	}
	cfg.DLimits = mapper

	return cfg, nil
}

// CreateTestingCFG is a function that creates a test cfg file with no secrets for us in testing
func CreateTestingCFG() *AppConfig {
	dt1 := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)

	cfg := AppConfig{
		// DtStart: "2022-01-01", // simulation start date for each generation
		// DtStop: "2022-12-31",  // simulation stop date for each generation
		DtStart:        CustomDate(dt1),
		DtStop:         CustomDate(dt2),
		Generations:    1,       // how many generations should the simulator run
		PopulationSize: 10,      // Total number Investors in the population
		C1:             "USD",   // main currency  (ISO 4217 code)
		C2:             "JPY",   // currency that we will invest in (ISO 4217 code)
		InitFunds:      1000.00, // how much each Investor is funded at the start of a simulation cycle
		StdInvestment:  100.00,  // the "standard" investment amount if a decision is made to invest in C2
		MinDelta4:      1,       // shortest period of time after a "buy" on T3 that we can do a "sell"
		MaxDelta4:      14,      // greatest period of time after a "buy" on T3 that we can do a "sell"
		DRW1:           0.6,     // DRInfluencer Fitness Score weighting for "correctness" of predictions. Constraint: DRW1 + DRW2 = 1.0
		DRW2:           0.4,     // DRInfluencer Fitness Score weighting for number of predictions made. Constraint: DRW1 + DRW2 = 1.0
		InvW1:          0.5,     // Investor Fitness Score weighting for "correctness" of predictions. Constraint: InvW1 + InvW2 = 1.0
		InvW2:          0.5,     // Investor Fitness Score weighting for profit. Constraint: InvW1 + InvW2 = 1.0
		MutationRate:   1,       // percentage number, from 1 - 100, what percent of the time does mutation occur
		DBSource:       "CSV",   // {CSV | Database | OnlineService}
	}

	mapper := map[string]RangeLimits{
		"DR": {
			MinDelta1: -30,
			MaxDelta1: -3,
			MinDelta2: -7,
			MaxDelta2: -1,
		},
		"UR": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"IR": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -60,
			MaxDelta2: -30,
		},
	}
	cfg.DLimits = mapper

	return &cfg
}
