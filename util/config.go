package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	json5 "github.com/yosuke-furukawa/json5/encoding/json5"
)

// CustomDate is used so that unmarshaling a date will work with
// dates in the format we want to enter them.
// ---------------------------------------------------------------------------
type CustomDate time.Time

// InfluencerSubclassInfo limits for Influencers data research time
// ---------------------------------------------------------------------------
type InfluencerSubclassInfo struct {
	MinDelta1 int     // furthest back from t3 that t1 can be
	MaxDelta1 int     // closest to t3 that t1 can be
	MinDelta2 int     // furthest back from t3 that t2 can be
	MaxDelta2 int     // closest to t3 that t2 can be
	FitnessW1 float64 // weight for correctness
	FitnessW2 float64 // weight for activity
}

// ValidInfluencerSubclasses anything other than these values is an error
// ---------------------------------------------------------------------------
var ValidInfluencerSubclasses = []string{
	"BCInfluencer",
	"BPInfluencer",
	"CCInfluencer",
	"CUInfluencer",
	"DRInfluencer", // discount rate
	"GDInfluencer",
	"HSInfluencer", // housing starts
	"IEInfluencer",
	"IPInfluencer",
	"IRInfluencer", // inflation rate
	"L0Influencer", // LSPScore_ECON - linguistic sentiment positive
	"L1Influencer", //
	"L2Influencer", //
	"L3Influencer", //
	"L4Influencer", //
	"L5Influencer", //
	"L6Influencer",
	"L7Influencer",
	"L8Influencer",
	"L9Influencer",
	"LAInfluencer",
	"LBInfluencer",
	"LCInfluencer",
	"LDInfluencer",
	"LEInfluencer",
	"LFInfluencer",
	"LGInfluencer",
	"LHInfluencer",
	"LIInfluencer",
	"LJInfluencer",
	"M1Influencer", // money supply short term
	"M2Influencer", // money supply long term
	"MPInfluencer",
	"RSInfluencer",
	"SPInfluencer", // stock price
	"URInfluencer", // unemployment rate
	"WTInfluencer", // unemployment rate
}

// InfluencerSubclasses is an array of strings with all the subclasses of
// Influencer that the factory knows how to create.
// ---------------------------------------------------------------------------
// var InfluencerSubclasses []string

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
// Investors, and Influencers. It needs to be in this directory so that
// it is visible to all areas of code in this project.
// ---------------------------------------------------------------------------
type FileConfig struct {
	C1               string     // Currency1 - the currency that we're trying to maximize
	C2               string     // Currency2 - the currency that we invest in to sell later and make a profit (or loss)
	DtStart          CustomDate // simulation begins on this date
	DtStop           CustomDate // simulation ends on this date. Guaranteed that no "buys" happen after this date
	LoopCount        int        // how many times to loop over DtStart to DtStop
	COAStrategy      string     // course of action strategy used by Investors (choices are: DistributedDecision)
	TopInvestorCount int        // how many top investors to include in financial report
	HoldWindowPos    float64    // positive space to consider as "no difference" when subtracting two ratios
	HoldWindowNeg    float64    // negative space to consider as "no difference" when subtracting two ratios

	//--------------------------------------------------------------------------------
	// The format of the GenDurSpec string is one to four pairs of the the following
	// values:  an integer,  one of the following letters: YMWD. There can be 1,
	// 2, 3 or 4 pairs, but each pair must contain a different letter from the
	// YMWD.  For example, ‘1 Y’ means 1 year, ‘1 Y 2 M’ means 1 year and 2 months.
	// W is for weeks,  and D is for Days.  It is fine to have a string like ’60 D’,
	// which means 60 days.
	//--------------------------------------------------------------------------------
	GenDurSpec      string                 // described above
	Generations     int                    // current generation in the simulator
	PopulationSize  int                    // how many investors are in this population
	InitFunds       float64                // amount of funds each Investor is "staked" at the outset of the simulation
	StdInvestment   float64                // standard investment amount
	TradingDay      int                    // this needs to be completely re-thought -- it's like a recurrence rule
	TradingTime     time.Time              // time of day when buy/sell is executed
	MaxInf          int                    // maximum number of influencers for any Investor
	MinInf          int                    // minimum number of influencers for any Investor
	CruciblePeriods []CustomCruciblePeriod // we read in from the config file here -- it gets changed to time.Time vals in the AppConfig struct at the end of LoadConfig
}

// TopInvestor is a struct containing the DNA for a top-performing Investor and an associated name
type TopInvestor struct {
	Name string
	DNA  string
}

// CustomCruciblePeriod is a struct containing a start and end time for the simulation of TopInvestors
// The CustomDate type is used to force our custome string to date function when it is read in through
// the csv file
type CustomCruciblePeriod struct {
	DtStart CustomDate // simulation begins on this date
	DtStop  CustomDate // simulation ends on this date
}

// CruciblePeriod defines the start and end time for a simulation of top investors
type CruciblePeriod struct {
	DtStart time.Time // simulation begins on this date
	DtStop  time.Time // simulation ends on this date
}

// AppConfig is the struct of config data used throughout the code by the Simulator,
// Investors, and Influencers. It is here in the util directory in order to be visible
// to all areas of code in this project
// ---------------------------------------------------------------------------
type AppConfig struct {
	Filename             string                            // filename of the configuration file read
	C1                   string                            // Currency1 - the currency that we're trying to maximize
	C2                   string                            // Currency2 - the currency that we invest in to sell later and make a profit (or loss)
	DtStart              CustomDate                        // simulation begins on this date
	DtStop               CustomDate                        // simulation ends on this date
	EnforceStopDate      bool                              // stops on DtStop even if there is a C2 Balance, if false and C2 Balance > 0 on StopDate, simulation will continue in sell-only mode until C2 < 1.00
	COAStrategy          string                            // course of action strategy used by Investors (choices are: DistributedDecision)
	LoopCount            int                               // how many times to loop over DtStart to DtStop
	TopInvestorCount     int                               // how many top investors to include in financial report
	MinInfluencers       int                               // minimum number of influencers per Investor
	MaxInfluencers       int                               // maximum number of influencers per Investor
	HoldWindowPos        float64                           // positive space to consider as "no difference" when subtracting two ratios
	HoldWindowNeg        float64                           // negative space to consider as "no difference" when subtracting two ratios
	GenDurSpec           string                            // gen dur spec
	GenDur               *GenerationDuration               // parsed gen dur spec
	DtSettle             time.Time                         // later of DtStop or date on which the last sale was made
	PopulationSize       int                               // how many investors are in this population
	InitFunds            float64                           // amount of funds each Investor is "staked" at the outset of the simulation
	StdInvestment        float64                           // standard investment amount
	TradingDay           int                               // this needs to be completely re-thought -- it's like a recurrence rule
	TradingTime          time.Time                         // time of day when buy/sell is executed
	Generations          int                               // current generation in the simulator
	MaxInf               int                               // maximum number of influencers for any Investor
	MinInf               int                               // minimum number of influencers for any Investor
	SCInfo               map[string]InfluencerSubclassInfo // map from Influencer subtype to its research time limits
	MinDelta4            int                               // closest to t3 that t4 can be
	MaxDelta4            int                               // furthest out from t3 that t4 can be
	DRW1                 float64                           // weighting for correctness part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	DRW2                 float64                           // weighting for prediction count part of DR Fitness Score calculation, (0 to 1), DRW1 + DRW2 must = 1
	InvW1                float64                           // weight for profit part of Investor FitnessScore
	InvW2                float64                           // weight for correctness part of Investor FitnessScore
	MutationRate         int                               // 1 - 100 indicating the % of mutation
	DBSource             string                            // {CSV | Database | OnlineService}
	InfluencerSubclasses []string                          // allowable Influencer subclasses for this run
	RandNano             int64                             // random number seed used for this simulation
	InfPredDebug         bool                              // print debug info about every prediction
	Trace                bool                              // use this flag to cause full trace information to be printed regarding every Investor decision every day.
	SingleInvestorMode   bool                              // default is false, when true it means we're running a single investor... more like the production code will run
	SingleInvestorDNA    string                            // DNA of the single investor
	TopInvestors         []TopInvestor                     // a list of top investors
	CrucibleSpans        []CruciblePeriod                  // list of times to run the simulation
	CrucibleMode         bool                              // if true then run all TopInvestor DNA through the CrucibleSpans
	ArchiveBaseDir       string                            // directory where archive will be created.  If no value is supplied, current directory will be used
	ArchiveMode          bool                              // archive reports to directory when true
	PreserveElite        bool                              // when true it replicates the top PreserverElitePct of DNA from gen x to gen x+1
	PreserveElitePct     float64                           // floating point value representing the amount of DNA to preserve. 0.0 to 100.0
	EliteCount           int                               // calculated by the simulator
}

// LoadConfig reads the configuration data from config.json into an
// internal struct and returns that struct.
//
// INPUT
//
//	overrideConfig - name of override config file.  use nil or "" to
//	                 indicate no override
//
// RETURNS
//
//	the AppConfig struct
//	any error encountered
//
// ---------------------------------------------------------------------
func LoadConfig(cfname string) (AppConfig, error) {
	var cfg AppConfig
	var fcfg FileConfig

	fname := "config.json5"
	if len(cfname) > 0 {
		fname = cfname
	}
	configFile, err := os.Open(fname)
	if err != nil {
		return cfg, fmt.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()
	cfg.Filename = fname
	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config file: %v", err)
	}

	//-------------------------------------
	// read into our config struct
	//-------------------------------------
	err = json5.Unmarshal(byteValue, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config data into cfg: %v", err)
	}

	//------------------------------------------------------
	// now read into fcfg to pick up the other values...
	//------------------------------------------------------
	err = json5.Unmarshal(byteValue, &fcfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config data into fcfg: %v", err)
	}

	cfg.DtSettle = time.Time(cfg.DtStop) // start it here... it will be updated later if needed

	if len(cfg.GenDurSpec) != 0 {
		cfg.GenDur, err = ParseGenerationDuration(cfg.GenDurSpec)
		if err != nil {
			log.Panicf("Invalid GenDurSpec specification: %s\n", cfg.GenDurSpec)
		}
	}

	if cfg.SingleInvestorMode || cfg.CrucibleMode {
		cfg.LoopCount = 1
		cfg.Generations = 1
		cfg.PopulationSize = 1
	}
	if cfg.TopInvestorCount < 1 {
		cfg.TopInvestorCount = 10 // guarantee a reasonable number
	}

	//-------------------------------------------------------------------
	// CRUCIBLE processing...
	//-------------------------------------------------------------------
	for i := 0; i < len(cfg.TopInvestors); i++ {
		if len(cfg.TopInvestors[i].Name) == 0 {
			cfg.TopInvestors[i].Name = fmt.Sprintf("TopInvestor%d", i)
		}
	}
	//-------------------------------------------------------------------
	// Convert the dates now, it's much easier to deal with time.Time
	// values after they've be read in.
	//-------------------------------------------------------------------
	for i := 0; i < len(fcfg.CruciblePeriods); i++ {
		var cp CruciblePeriod
		cp.DtStart = time.Time(fcfg.CruciblePeriods[i].DtStart)
		cp.DtStop = time.Time(fcfg.CruciblePeriods[i].DtStop)
		cfg.CrucibleSpans = append(cfg.CrucibleSpans, cp)
	}
	//-------------------------------------------------------------------
	// If the weights were not specified, or they do not add up to 1
	// then set default values here...
	//-------------------------------------------------------------------
	if cfg.InvW1+cfg.InvW2 != 1.0 {
		cfg.InvW1 = 0.5
		cfg.InvW2 = 0.5
	}
	return cfg, nil
}

// CreateTestingCFG is a function that creates a test cfg file with no secrets
// for use in testing
// -----------------------------------------------------------------------------
func CreateTestingCFG() *AppConfig {
	dt1 := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)

	cfg := AppConfig{
		Generations:    1,       // how many generations should the simulator run
		PopulationSize: 10,      // Total number Investors in the population
		LoopCount:      10,      // How many times to loop over DtStart to DtSTop
		C1:             "USD",   // main currency  (ISO 4217 code)
		C2:             "JPY",   // currency that we will invest in (ISO 4217 code)
		InitFunds:      1000.00, // how much each Investor is funded at the start of a simulation cycle
		StdInvestment:  100.00,  // the "standard" investment amount if a decision is made to invest in C2
		MinInfluencers: 1,       // at least this many per Investor
		MaxInfluencers: 10,      // no more than this many
		MinDelta4:      1,       // shortest period of time after a "buy" on T3 that we can do a "sell"
		MaxDelta4:      14,      // greatest period of time after a "buy" on T3 that we can do a "sell"
		DRW1:           0.6,     // DRInfluencer Fitness Score weighting for "correctness" of predictions. Constraint: DRW1 + DRW2 = 1.0
		DRW2:           0.4,     // DRInfluencer Fitness Score weighting for number of predictions made. Constraint: DRW1 + DRW2 = 1.0
		InvW1:          0.5,     // Investor Fitness Score weighting for "correctness" of predictions. Constraint: InvW1 + InvW2 = 1.0
		InvW2:          0.5,     // Investor Fitness Score weighting for profit. Constraint: InvW1 + InvW2 = 1.0
		MutationRate:   1,       // percentage number, from 1 - 100, what percent of the time does mutation occur
		DBSource:       "CSV",   // {CSV | Database | OnlineService}
		COAStrategy:    "DistributedDecision",
		InfluencerSubclasses: []string{ // default case is to enable all Influencer subclasses
			"CCInfluencer",
			"DRInfluencer",
			// "GDInfluencer",
			// "IRInfluencer",
			// "M1Influencer",
			// "M2Influencer",
			"URInfluencer",
		},
	}

	cfg.DtStart = CustomDate(dt1)
	cfg.DtStop = CustomDate(dt2)

	mapper := map[string]InfluencerSubclassInfo{
		"BC": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"BP": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"CC": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"CU": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"DR": {
			MinDelta1: -30,
			MaxDelta1: -7,
			MinDelta2: -6,
			MaxDelta2: -1,
		},
		"GD": {
			MinDelta1: -730,
			MaxDelta1: -630,
			MinDelta2: -180,
			MaxDelta2: -120,
		},
		"HS": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -60,
			MaxDelta2: -30,
		},
		"IE": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -60,
			MaxDelta2: -30,
		},
		"IP": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -60,
			MaxDelta2: -30,
		},
		"IR": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -60,
			MaxDelta2: -30,
		},
		"L0": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L1": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L2": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L3": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L4": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L5": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L6": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L7": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L8": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"L9": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LA": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LB": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LC": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LD": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LE": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LF": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LG": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LH": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LI": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"LJ": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
		"M1": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"M2": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"MR": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"RS": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"SP": {
			MinDelta1: -30,
			MaxDelta1: -6,
			MinDelta2: -5,
			MaxDelta2: -1,
		},
		"UR": {
			MinDelta1: -180,
			MaxDelta1: -90,
			MinDelta2: -50,
			MaxDelta2: -20,
		},
		"WT": {
			MinDelta1: -90,
			MaxDelta1: -30,
			MinDelta2: -29,
			MaxDelta2: -1,
		},
	}
	cfg.SCInfo = mapper
	cfg.DtSettle = time.Time(cfg.DtStop)

	return &cfg
}
