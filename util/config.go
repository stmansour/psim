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
var InfluencerSubclasses []string

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
	/*
	   InfluencerSubclasses []string  // valid Influencer subclasses for this simulation
	   CCMinDelta1          int       // negative integer, most number of days prior to T3 for Influencer research to begin
	   CCMaxDelta1          int       // negative integer, fewest number of days prior to T3 for Invfluencer research to begin
	   CCMinDelta2          int       // research boundary
	   CCMaxDelta2          int       // research boundary
	   DRMinDelta1          int       // negative integer, most number of days prior to T3 for Influencer research to begin
	   DRMaxDelta1          int       // negative integer, fewest number of days prior to T3 for Invfluencer research to begin
	   DRMinDelta2          int       // research boundary
	   DRMaxDelta2          int       // research boundary
	   GDMinDelta1          int       // negative integer, most number of days prior to T3 for Influencer research to begin
	   GDMaxDelta1          int       // negative integer, fewest number of days prior to T3 for Invfluencer research to begin
	   GDMinDelta2          int       // research boundary
	   GDMaxDelta2          int       // research boundary
	   URMinDelta1          int       // research boundary
	   URMaxDelta1          int       // research boundary
	   URMinDelta2          int       // research boundary
	   URMaxDelta2          int       // research boundary
	   IRMinDelta1          int       // research boundary
	   IRMaxDelta1          int       // research boundary
	   IRMinDelta2          int       // research boundary
	   IRMaxDelta2          int       // research boundary
	   L0MinDelta1          int       // research boundary
	   L0MaxDelta1          int       // research boundary
	   L0MinDelta2          int       // research boundary
	   L0MaxDelta2          int       // research boundary
	   L1MinDelta1          int       // research boundary
	   L1MaxDelta1          int       // research boundary
	   L1MinDelta2          int       // research boundary
	   L1MaxDelta2          int       // research boundary
	   L2MinDelta1          int       // research boundary
	   L2MaxDelta1          int       // research boundary
	   L2MinDelta2          int       // research boundary
	   L2MaxDelta2          int       // research boundary
	   L3MinDelta1          int       // research boundary
	   L3MaxDelta1          int       // research boundary
	   L3MinDelta2          int       // research boundary
	   L3MaxDelta2          int       // research boundary
	   L4MinDelta1          int       // research boundary
	   L4MaxDelta1          int       // research boundary
	   L4MinDelta2          int       // research boundary
	   L4MaxDelta2          int       // research boundary
	   L5MinDelta1          int       // research boundary
	   L5MaxDelta1          int       // research boundary
	   L5MinDelta2          int       // research boundary
	   L5MaxDelta2          int       // research boundary
	   L6MinDelta1          int       // research boundary
	   L6MinDelta2          int
	   L6MaxDelta1          int
	   L6MaxDelta2          int
	   L7MinDelta1          int // research boundary
	   L7MinDelta2          int
	   L7MaxDelta1          int
	   L7MaxDelta2          int
	   L8MinDelta1          int // research boundary
	   L8MinDelta2          int
	   L8MaxDelta1          int
	   L8MaxDelta2          int
	   L9MinDelta1          int // research boundary
	   L9MinDelta2          int
	   L9MaxDelta1          int
	   L9MaxDelta2          int
	   LAMinDelta1          int // research boundary
	   LAMinDelta2          int
	   LAMaxDelta1          int
	   LAMaxDelta2          int
	   LBMinDelta1          int // research boundary
	   LBMinDelta2          int
	   LBMaxDelta1          int
	   LBMaxDelta2          int
	   LCMinDelta1          int // research boundary
	   LCMinDelta2          int
	   LCMaxDelta1          int
	   LCMaxDelta2          int
	   LDMinDelta1          int // research boundary
	   LDMinDelta2          int
	   LDMaxDelta1          int
	   LDMaxDelta2          int
	   LEMinDelta1          int // research boundary
	   LEMinDelta2          int
	   LEMaxDelta1          int
	   LEMaxDelta2          int
	   LFMinDelta1          int // research boundary
	   LFMinDelta2          int
	   LFMaxDelta1          int
	   LFMaxDelta2          int
	   LGMinDelta1          int // research boundary
	   LGMinDelta2          int
	   LGMaxDelta1          int
	   LGMaxDelta2          int
	   LHMinDelta1          int // research boundary
	   LHMinDelta2          int
	   LHMaxDelta1          int
	   LHMaxDelta2          int
	   LIMinDelta1          int // research boundary
	   LIMinDelta2          int
	   LIMaxDelta1          int
	   LIMaxDelta2          int
	   LJMinDelta1          int // research boundary
	   LJMinDelta2          int
	   LJMaxDelta1          int
	   LJMaxDelta2          int
	   WTMinDelta1          int // research boundary
	   WTMinDelta2          int
	   WTMaxDelta1          int
	   WTMaxDelta2          int
	   M1MinDelta1          int     // research boundary
	   M1MaxDelta1          int     // research boundary
	   M1MinDelta2          int     // research boundary
	   M1MaxDelta2          int     // research boundary
	   M2MinDelta1          int     // research boundary
	   M2MaxDelta1          int     // research boundary
	   M2MinDelta2          int     // research boundary
	   M2MaxDelta2          int     // research boundary
	   SPMinDelta1          int     // research boundary
	   SPMaxDelta1          int     // research boundary
	   SPMinDelta2          int     // research boundary
	   SPMaxDelta2          int     // research boundary
	   MinDelta4            int     // closest to t3 that t4 can be
	   MaxDelta4            int     // furthest out from t3 that t4 can be
	   CCW1                 float64 // weighting in fitness calculation
	   CCW2                 float64 // weighting in fitness calculation
	   DRW1                 float64 // weighting in fitness calculation
	   DRW2                 float64 // weighting in fitness calculation
	   GDW1                 float64 // weighting in fitness calculation
	   GDW2                 float64 // weighting in fitness calculation
	   IRW1                 float64 // weighting in fitness calculation
	   IRW2                 float64 // weighting in fitness calculation
	   L0W1                 float64 // weighting in fitness calculation
	   L0W2                 float64 // weighting in fitness calculation
	   L1W1                 float64 // weighting in fitness calculation
	   L1W2                 float64 // weighting in fitness calculation
	   L2W1                 float64 // weighting in fitness calculation
	   L2W2                 float64 // weighting in fitness calculation
	   L3W1                 float64 // weighting in fitness calculation
	   L3W2                 float64 // weighting in fitness calculation
	   L4W1                 float64 // weighting in fitness calculation
	   L4W2                 float64 // weighting in fitness calculation
	   L5W1                 float64 // weighting in fitness calculation
	   L5W2                 float64 // weighting in fitness calculation
	   L6W1                 float64
	   L6W2                 float64
	   L7W1                 float64
	   L7W2                 float64
	   L8W1                 float64
	   L8W2                 float64
	   L9W1                 float64
	   L9W2                 float64
	   LAW1                 float64
	   LAW2                 float64
	   LBW1                 float64
	   LBW2                 float64
	   LCW1                 float64
	   LCW2                 float64
	   LDW1                 float64
	   LDW2                 float64
	   LEW1                 float64
	   LEW2                 float64
	   LFW1                 float64
	   LFW2                 float64
	   LGW1                 float64
	   LGW2                 float64
	   LHW1                 float64
	   LHW2                 float64
	   LIW1                 float64
	   LIW2                 float64
	   LJW1                 float64
	   LJW2                 float64
	   WTW1                 float64
	   WTW2                 float64
	   M1W1                 float64 // weighting in fitness calculation
	   M1W2                 float64 // weighting in fitness calculation
	   M2W1                 float64 // weighting in fitness calculation
	   M2W2                 float64 // weighting in fitness calculation
	   SPW1                 float64 // weighting in fitness calculation
	   SPW2                 float64 // weighting in fitness calculation
	   URW1                 float64 // weighting in fitness calculation
	   URW2                 float64 // weighting in fitness calculation
	   InvW1                float64 // weight for profit part of Investor FitnessScore
	   InvW2                float64 // weight for correctness part of Investor FitnessScore
	   MutationRate         int     // 1 - 100 indicating the % of mutation
	   DBSource             string  // {CSV | Database | OnlineService}
	   RandNano             int64   // random seed

	   //--------------------------------------------------------------------------------------------------------------------
	   // Single Investor mode...  LoopCount will be set to 1, Generations will be set to 1, PopulationSize will be set to 1
	   //--------------------------------------------------------------------------------------------------------------------
	   SingleInvestorMode bool                   // default is false, when true it means we're running a single investor... more like the production code will run
	   SingleInvestorDNA  string                 // DNA of the single investor
	*/
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
}

// func hasPrefix(tag string, prefixes []string, mod string) bool {
// 	for i := 0; i < len(prefixes); i++ {
// 		pfx := prefixes[i] + mod
// 		if strings.HasPrefix(tag, pfx) {
// 			return true
// 		}
// 	}
// 	return false
// }

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

	//--------------------------------------------------------
	// First thing to do is set the InfluencerSubclasses...
	//--------------------------------------------------------
	// InfluencerSubclasses = fcfg.InfluencerSubclasses
	// var prefixes []string
	// for i := 0; i < len(InfluencerSubclasses); i++ {
	// 	s := InfluencerSubclasses[i][:2]
	// 	prefixes = append(prefixes, s)
	// 	// DPrintf("LOAD - InfluencerSubclasses[%d] = %s", i, InfluencerSubclasses[i])
	// }

	//----------------------------------------------------
	// now build the map[subclass]InfluencerSubclassInfo
	//----------------------------------------------------
	// mapper := make(map[string]InfluencerSubclassInfo)
	// t := reflect.TypeOf(fcfg)
	// v := reflect.ValueOf(fcfg)
	// isi := []string{"W1", "W2", "Delta1", "Delta2"}

	// for i := 0; i < t.NumField(); i++ {
	// 	field := t.Field(i)
	// 	jsonTag := field.Name
	// 	value := v.Field(i).Interface()

	// 	//--------------------------------------------------------
	// 	// Check for values belonging to InfluencerSubclassInfo
	// 	//--------------------------------------------------------
	// 	isInfo := -1
	// 	for j := 0; j < len(isi) && isInfo < 0; j++ {
	// 		if strings.HasSuffix(jsonTag, isi[j]) {
	// 			isInfo = j
	// 		}
	// 	}

	// 	if isInfo >= 0 {
	// 		//--------------------------------------------------------
	// 		// make sure this subclass is included in the simulation
	// 		//--------------------------------------------------------
	// 		subclassName := jsonTag[:2]
	// 		found := false
	// 		for j := 0; j < len(InfluencerSubclasses) && !found; j++ {
	// 			if subclassName == InfluencerSubclasses[j][0:2] {
	// 				found = true
	// 			}
	// 		}
	// 		if !found {
	// 			continue // skip the limits if we're not even using this subclass
	// 		}

	// 		info := mapper[subclassName] // grab the current version
	// 		switch isInfo {
	// 		case 0: // W1
	// 			info.FitnessW1 = value.(float64)
	// 		case 1: // W2
	// 			info.FitnessW2 = value.(float64)
	// 		case 2: // Delta1
	// 			if hasPrefix(jsonTag, prefixes, "Min") {
	// 				info.MinDelta1 = value.(int)
	// 			} else if hasPrefix(jsonTag, prefixes, "Max") {
	// 				info.MaxDelta1 = value.(int)
	// 			}
	// 		case 3: // Delta2
	// 			if hasPrefix(jsonTag, prefixes, "Min") {
	// 				info.MinDelta2 = value.(int)
	// 			} else if hasPrefix(jsonTag, prefixes, "Max") {
	// 				info.MaxDelta2 = value.(int)
	// 			}
	// 		}
	// 		mapper[subclassName] = info // save the updated version
	// 	}
	// }
	// cfg.SCInfo = mapper
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

	InfluencerSubclasses = []string{
		"CCInfluencer",
		"DRInfluencer",
		// "GDInfluencer",
		"IRInfluencer",
		"L0Influencer",
		"L1Influencer",
		"L2Influencer",
		"L3Influencer",
		"L4Influencer",
		"L5Influencer",
		"WTInfluencer",
		// "M1Influencer",
		// "M2Influencer",
		"URInfluencer",
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
