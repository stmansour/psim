package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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
	TopInvestorCount int        // how many top investors to include in financial report
	// HoldWindowPos    float64    // positive space to consider as "no difference" when subtracting two ratios
	// HoldWindowNeg    float64    // negative space to consider as "no difference" when subtracting two ratios

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
	Index    int
	DtStart  CustomDate // simulation begins on this date
	DtStop   CustomDate // simulation ends on this date
	Duration string     // how long the simulation is
	Ending   string     // typically a keyword: "today", "yesterday", "lastmonthend", or possibly a CustomDate
}

// CruciblePeriod defines the start and end time for a simulation of top investors
type CruciblePeriod struct {
	DtStart time.Time // simulation begins on this date
	DtStop  time.Time // simulation ends on this date
}

// CustomCell is for custom reporting
type CustomCell struct {
	Col string // spreadsheet column
	Val string // a formula or value
}

// AppConfig is the struct of config data used throughout the code by the Simulator,
// Investors, and Influencers. It is here in the util directory in order to be visible
// to all areas of code in this project
// ---------------------------------------------------------------------------
type AppConfig struct {
	SimulationName          string              // name for this simulation
	ConfigFilename          string              // filename of the configuration file read
	C1                      string              // Currency1 - the currency that we're trying to maximize
	C2                      string              // Currency2 - the currency that we invest in to sell later and make a profit (or loss)
	DtStart                 CustomDate          // simulation begins on this date
	DtStop                  CustomDate          // simulation ends on this date. Guaranteed that no "buys" happen after this date
	EnforceStopDate         bool                // stops on DtStop even if there is a C2 Balance, if false and C2 Balance > 0 on StopDate, simulation will continue in sell-only mode until C2 < 1.00
	COAStrategy             string              // course of action strategy used by Investors (choices are: DistributedDecision)
	LoopCount               int                 // how many times to loop over DtStart to DtStop
	TopInvestorCount        int                 // how many top investors to include in financial report
	MinInfluencers          int                 // minimum number of influencers per Investor
	MaxInfluencers          int                 // maximum number of influencers per Investor
	GenDurSpec              string              // gen dur spec
	GenDur                  *GenerationDuration // parsed gen dur spec
	DtSettle                time.Time           // later of DtStop or date on which the last sale was made
	PopulationSize          int                 // how many investors are in this population
	InitFunds               float64             // amount of funds each Investor is "staked" at the outset of the simulation
	SplitInitFunds          bool                // if true split initial funds evenly between C1 and C2 on DtStart
	StdInvestment           float64             // standard investment amount
	TradingDay              int                 // this needs to be completely re-thought -- it's like a recurrence rule
	TradingTime             time.Time           // time of day when buy/sell is executed
	Generations             int                 // current generation in the simulator
	MutationRate            int                 // 1 - 100 indicating the % of mutation
	DBSource                string              // {CSV | Database | OnlineService}
	RandNano                int64               // random number seed used for this simulation
	InfPredDebug            bool                // print debug info about every prediction
	Trace                   bool                // use this flag to cause full trace information to be printed regarding every Investor decision every day.
	SingleInvestorMode      bool                // default is false, when true it means we're running a single investor... more like the production code will run
	SingleInvestorDNA       string              // DNA of the single investor
	TopInvestors            []TopInvestor       // a list of top investors
	CrucibleSpans           []CruciblePeriod    // list of times to run the simulation
	CustomRow               []CustomCell        // a list of custom cells to add to the spreadsheet for each Investor over all the CrucibleSpans
	CrucibleMode            bool                // if true then run all TopInvestor DNA through the CrucibleSpans
	Recommendation          bool                // if true then show buy/sell/hold recommendation for DtStart
	CrucibleName            string              // name of the crucible
	CrucibleARThreshold     float64             // AR threshold... it only counts if if the annualized return is above this amount.  Use 0.15 for 15%
	ReportDirectory         string              // final directory where all reports should be
	ReportTimestamp         string              // timestamp to used for archived reports
	ReportDirSet            bool                // when false the info needs to be set, when true it's already set
	ArchiveBaseDir          string              // directory where archive will be created.  If no value is supplied, current directory will be used
	ArchiveMode             bool                // archive reports to directory when true
	PreserveElite           bool                // when true it replicates the top PreserverElitePct of DNA from gen x to gen x+1
	PreserveElitePct        float64             // floating point value representing the amount of DNA to preserve. 0.0 to 100.0
	EliteCount              int                 // calculated by the simulator
	ExecutableFilePath      string              // path to the executable
	StopLoss                float64             // Expressed as a percentage of the Portfolio Value. That is, use 0.10 for 10%.  Sell all C2 immediately if the PV has lost this much of the initial funding.
	TxnFeeFactor            float64             // cost per transaction that is multiplied by the amount. 0.0002 == 2 basis points, 0 if not set
	TxnFee                  float64             // a flat cost that is added for each transaction, 0 if not set
	InvestorBonusPlan       bool                // rewards Investors earning high ROI by giving a bonus to their FitnessScore.  PV >= 110% receive 100% bonus, PV >= 115% get 200%, PV >= 120% get 300%, and PV >= 400% get 500%
	Gen0Elites              bool                // if true, introduce TopInvestors DNA into the initial population
	WorkerPoolSize          int                 // number of cores to utilize, if < 1 then the program decides, if 1 or more then that many cores are used, it will be capped at the number of cores the hardware actually has
	HoldWindowStatsLookBack int                 // how many days make up the rolling window of data used in HoldWindow stats calculations (mean and StdDev)
	StdDevVariationFactor   float64             // how much variance from thethe standard deviation is needed for the hold window.
	AllowDuplicateInvestors bool                // if false then check for duplicate investors
	PredictionMode          bool                // true if the code is making a prediction about the next day
	DNALog                  bool                // if true, generate DNA log report
	SID                     int64               // simulation id if > 0
	GracePeriodDays         int                 // Grace period in days; max days between DtStop and the last date in the database that is allowed. Can happen when db is not updated and DtStop is a keyword like "yesterday"
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
		MutationRate:   1,       // percentage number, from 1 - 100, what percent of the time does mutation occur
		DBSource:       "CSV",   // {CSV | Database | OnlineService}
	}

	cfg.DtStart = CustomDate(dt1)
	cfg.DtStop = CustomDate(dt2)

	cfg.DtSettle = time.Time(cfg.DtStop)

	return &cfg
}

// Helper function to check if a file exists
func checkFileExists(fname string) bool {
	_, err := os.Stat(fname)
	return err == nil
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
func LoadConfig(cfname string) (*AppConfig, error) {
	var fcfg FileConfig
	var cfg AppConfig
	fname := "config.json5"
	fnameSet := false

	// Check if a config file name was provided
	if len(cfname) > 0 {
		fname = cfname
		fnameSet = true
	}

	// Check if the config file exists in the current directory
	if !fnameSet && checkFileExists(fname) {
		fnameSet = true
	}

	// If the config file was not found in the current directory, try the executable directory
	if !fnameSet {
		exePath, err := GetExecutableDir()
		if err != nil {
			return &cfg, err
		}
		fname = exePath + "/" + fname
		cfg.ConfigFilename = fname
	}

	// If the config file still wasn't found, return an error
	if !checkFileExists(fname) {
		if len(cfname) > 0 {
			return &cfg, fmt.Errorf("no configuration file was found")
		}
		return &cfg, fmt.Errorf("no configuration file was found in the current directory or the executable directory")
	}

	configFile, err := os.Open(fname)
	if err != nil {
		return &cfg, fmt.Errorf("failed to open config file: %v", err)
	}
	defer configFile.Close()
	cfg.ConfigFilename = fname // This was the file ultimately opened
	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		return &cfg, fmt.Errorf("failed to read config file: %v", err)
	}

	//-------------------------------------
	// read into our config struct
	//-------------------------------------
	err = json5.Unmarshal(byteValue, &cfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to unmarshal config data into cfg: %v", err)
	}

	//------------------------------------------------------
	// now read into fcfg to pick up the other values...
	//------------------------------------------------------
	err = json5.Unmarshal(byteValue, &fcfg)
	if err != nil {
		return &cfg, fmt.Errorf("failed to unmarshal config data into fcfg: %v", err)
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
	if cfg.HoldWindowStatsLookBack < 1 {
		cfg.HoldWindowStatsLookBack = 365 // guarantee a reasonable number
	}
	if cfg.StdDevVariationFactor == 0 {
		cfg.StdDevVariationFactor = 0.1 // guarantee a reasonable number
	}

	if cfg.GracePeriodDays == 0 {
		cfg.GracePeriodDays = 5 // 5 days of grace period
	}
	//-------------------------------------------------------------------
	// CRUCIBLE processing...
	//-------------------------------------------------------------------
	if err = ProcessCrucibleSettings(&cfg, &fcfg); err != nil {
		return &cfg, err
	}

	cfg.DtStart = fcfg.DtStart
	cfg.DtStop = fcfg.DtStop
	return &cfg, nil
}

// parseCustomCruciblePeriod takes a CustomCruciblePeriod and calculates the start and stop times.
func parseCustomCruciblePeriod(ccp *CustomCruciblePeriod) (CruciblePeriod, error) {
	var cp CruciblePeriod
	now := time.Now()

	// Handling the Ending field
	if ccp.Ending != "" {
		cp.DtStop = time.Time(calculateEndDate(ccp.Ending, now))
	} else {
		cp.DtStop = time.Time(ccp.DtStop)
	}

	// Calculate start date based on duration if available
	if ccp.Duration != "" {
		cp.DtStart = time.Time(calculateStartDate(ccp.Duration, cp.DtStop))
	} else {
		cp.DtStart = time.Time(ccp.DtStart)
	}

	return cp, nil
}

// ProcessCrucibleSettings handles the date ranges for the Crucible mode
func ProcessCrucibleSettings(cfg *AppConfig, fcfg *FileConfig) error {
	if len(fcfg.CruciblePeriods) == 0 {
		return nil
	}

	// for i := 0; i < len(cfg.TopInvestors); i++ {
	// 	if len(cfg.TopInvestors[i].Name) == 0 {
	// 		cfg.TopInvestors[i].Name = fmt.Sprintf("TopInvestor%d", i)
	// 	}
	// }

	for i := 0; i < len(fcfg.CruciblePeriods); i++ {
		cp, err := parseCustomCruciblePeriod(&fcfg.CruciblePeriods[i])
		if err != nil {
			return err
		}
		cfg.CrucibleSpans = append(cfg.CrucibleSpans, cp)
	}
	return nil
}

// calculateStartDate calculates the start date by subtracting the specified 'duration' from 'endDate'.
func calculateStartDate(duration string, endDate time.Time) time.Time {
	parts := strings.Fields(duration)
	if len(parts) == 0 {
		return endDate
	}

	// Parse duration and unit
	unit := parts[0][len(parts[0])-1] // 'm', 'y', 'd', 'w'
	amountStr := parts[0][:len(parts[0])-1]
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		fmt.Println("Error converting duration amount to integer:", err)
		return time.Time{}
	}

	switch unit {
	case 'y':
		return endDate.AddDate(-amount, 0, 0)
	case 'm':
		if endDate.Day() >= 28 { // Handle end-of-month scenario
			newMonth := int(endDate.Month()) - amount
			newYear := endDate.Year()
			if newMonth <= 0 {
				newYear -= (-newMonth / 12) + 1
				newMonth = 12 + (newMonth % 12)
			}
			startDate := time.Date(newYear, time.Month(newMonth), endDate.Day(), 0, 0, 0, 0, time.UTC)
			lastDayOfMonth := time.Date(newYear, time.Month(newMonth+1), 0, 0, 0, 0, 0, time.UTC)
			if !startDate.Equal(lastDayOfMonth) {
				return lastDayOfMonth.AddDate(0, 0, 1)
			}
			return startDate.AddDate(0, 0, 1)
		}
		return endDate.AddDate(0, -amount, 0)
	case 'w':
		return endDate.AddDate(0, 0, 1-(7*amount))
	case 'd':
		return endDate.AddDate(0, 0, -amount)
	default:
		fmt.Println("Unknown duration unit:", string(unit))
		return time.Time{}
	}
}

// calculateEndDate calculates based on the user's specification
func calculateEndDate(ending string, currentDate time.Time) time.Time {
	parts := strings.Fields(ending)
	baseDate := time.Time{}

	// Determine the initial date
	switch strings.ToLower(parts[0]) {
	case "yesterday":
		baseDate = currentDate.AddDate(0, 0, -1)
	case "today":
		baseDate = currentDate
	case "lastmonthend":
		firstOfCurrentMonth := time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
		baseDate = firstOfCurrentMonth.AddDate(0, 0, -1)
	default:
		var err error
		baseDate, err = StringToDate(parts[0])
		if err != nil {
			fmt.Println("Error parsing base date:", err)
			return time.Time{}
		}
	}

	if len(parts) > 1 {
		offsetUnit := parts[2][len(parts[2])-1]
		offsetAmountStr := parts[2][:len(parts[2])-1]
		offsetAmount, err := strconv.Atoi(offsetAmountStr)
		if err != nil {
			fmt.Println("Error converting offset amount to integer:", err)
			return time.Time{}
		}

		if parts[1] == "-" {
			offsetAmount = -offsetAmount
		}

		// Adjust month and find the last day of the adjusted month if the day of the baseDate is 28 or higher
		switch offsetUnit {
		case 'm', 'y':
			if offsetUnit == 'y' {
				offsetAmount = offsetAmount * 12
			}
			if baseDate.Day() >= 28 {
				// Calculate new month and year considering the offset
				newMonth := int(baseDate.Month()) + offsetAmount
				newYear := baseDate.Year()

				if newMonth <= 0 {
					newYear -= 1 + (-newMonth / 12)
					newMonth = 12 + (newMonth % 12)
				} else if newMonth > 12 {
					newYear += (newMonth - 1) / 12
					newMonth = ((newMonth - 1) % 12) + 1
				}

				// Find the last day of the new month
				firstOfNextMonth := time.Date(newYear, time.Month(newMonth+1), 1, 0, 0, 0, 0, baseDate.Location())
				lastDayOfNewMonth := firstOfNextMonth.AddDate(0, 0, -1)
				return lastDayOfNewMonth
			}
			endDate := baseDate.AddDate(0, offsetAmount, 0)
			return endDate
		case 'w':
			return baseDate.AddDate(0, 0, offsetAmount*7)
		case 'd':
			return baseDate.AddDate(0, 0, offsetAmount)
		default:
			fmt.Println("Unknown offset unit:", string(offsetUnit))
			return time.Time{}
		}
	}

	return baseDate
}

// GenerateFName returns the filename for the specified basename
// in the appropriate directory.
// --------------------------------------------------------------------------
func (cfg *AppConfig) GenerateFName(basename string) string {
	fname := ""
	if len(cfg.ReportDirectory) > 0 {
		fname = cfg.ReportDirectory + "/"
	}
	fname += basename
	if cfg.ArchiveMode {
		fname += cfg.ReportTimestamp
	}
	fname += ".csv"
	return fname
}
