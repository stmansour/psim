package data

import (
	"bytes"
	"fmt"
	"time"

	"github.com/stmansour/psim/util"
)

// RatesAndRatiosRecords is a type for an array of DR records
type RatesAndRatiosRecords []RatesAndRatiosRecord

// DInfo maintains data needed by the data subsystem.
// The primary need is the two currencies C1 & C2
var DInfo struct {
	cfg     *util.AppConfig
	DBRecs  RatesAndRatiosRecords  // all records... temporary, until we have database
	LRecs   []LinguisticDataRecord // all lingustic records
	DtStart time.Time              // earliest date with data
	DtStop  time.Time              // latest date with data
	DTypes  []string               // the list of Influencers, each has their own data type
	CSVMap  map[string]int         // which columns are where? Map the data type to a CSV column
}

// LinguisticDataRecord is a temporary structure of data for linguistic metrics
type LinguisticDataRecord struct {
	Date              time.Time
	LALLLSNScore      float64
	LALLLSPScore      float64
	LALLWHAScore      float64
	LALLWHOScore      float64
	LALLWHLScore      float64
	LALLWPAScore      float64
	LALLWDECount      float64
	LALLWDFCount      float64
	LALLWDPCount      float64
	LALLWDMCount      float64
	LUSDLSNScore_ECON float64
	LUSDLSPScore_ECON float64
	LUSDWHAScore_ECON float64
	LUSDWHOScore_ECON float64
	LUSDWHLScore_ECON float64
	LUSDWPAScore_ECON float64
	LUSDWDECount_ECON float64
	LUSDWDFCount_ECON float64
	LUSDWDPCount_ECON float64
	LUSDLIMCount_ECON float64
	LJPYLSNScore_ECON float64
	LJPYLSPScore_ECON float64
	LJPYWHAScore_ECON float64
	LJPYWHOScore_ECON float64
	LJPYWHLScore_ECON float64
	LJPYWPAScore_ECON float64
	LJPYWDECount_ECON float64
	LJPYWDFCount_ECON float64
	LJPYWDPCount_ECON float64
	LJPYLIMCount_ECON float64
}

// RatesAndRatiosRecord is the basic structure of discount rate data
type RatesAndRatiosRecord struct {
	Date time.Time

	BCRatio float64 // Check FLAGS for validity
	BPRatio float64 // Check FLAGS for validity
	CCRatio float64 // Check FLAGS for validity
	CURatio float64 // Check FLAGS for validity
	DRRatio float64 // Check FLAGS for validity
	EXClose float64 //
	GDRatio float64 // Check FLAGS for validity
	HSRatio float64 // Check FLAGS for validity
	IERatio float64 // Check FLAGS for validity
	IPRatio float64 // Check FLAGS for validity
	IRRatio float64 // Check FLAGS for validity
	MPRatio float64 // Check FLAGS for validity
	M1Ratio float64 // Check FLAGS for validity
	M2Ratio float64 // Check FLAGS for validity
	RSRatio float64 // Check FLAGS for validity
	SPRatio float64 // Check FLAGS for validity
	URRatio float64 // Check FLAGS for validity

	FLAGS uint64 // can hold flags for the first 64 values associated with the Date, see DataFlags
}

// DataFlags indicate which bit of the flag fields must be set in order for the
// associated value to be valid.
var DataFlags struct {
	BCRatioValid uint64
	BPRatioValid uint64
	CCRatioValid uint64
	CURatioValid uint64
	DRRatioValid uint64
	EXCloseValid uint64
	GDRatioValid uint64
	HSRatioValid uint64
	IERatioValid uint64
	IPRatioValid uint64
	IRRatioValid uint64
	MPRatioValid uint64
	M1RatioValid uint64
	M2RatioValid uint64
	RSRatioValid uint64
	SPRatioValid uint64
	URRatioValid uint64
}

// PLATODB is the csv data file that is used for Discount Rate information
var PLATODB = string("data/platodb.csv")

// CurrencyInfo contains information about currencies used in this program
type CurrencyInfo struct {
	Country      string // name of the issuing country
	CountryCode  string // two-letter designator for country
	Currency     string // name of the currency
	CurrencyCode string // typically the first char of the currency name
}

// Currencies is a list a CurrencyInfo for all the currencies supported by this program
var Currencies = []CurrencyInfo{
	{
		Country:      "United States",
		CountryCode:  "US",
		Currency:     "Dollar",
		CurrencyCode: "D",
	},
	{
		Country:      "Japan",
		CountryCode:  "JP",
		Currency:     "Yen",
		CurrencyCode: "Y",
	},
	{
		Country:      "Great Britain",
		CountryCode:  "GB",
		Currency:     "Pound",
		CurrencyCode: "P",
	},
	{
		Country:      "Australia",
		CountryCode:  "AU",
		Currency:     "Dollar",
		CurrencyCode: "D",
	},
}

// Init calls the initialize routine for all data types
// ------------------------------------------------------------
func Init(cfg *util.AppConfig) error {
	DInfo.cfg = cfg
	switch DInfo.cfg.DBSource {
	case "CSV":
		if err := LoadCsvDB(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unimplemented DBSource %s", DInfo.cfg.DBSource)
	}
	return nil
}

// DRec2String provides a visual representation of what data is
// available in the data record.
// ------------------------------------------------------------------
func DRec2String(drec *RatesAndRatiosRecord) string {
	s := fmt.Sprintf(`    Date = %s
	BCRatio = %9.3f  %s
	BPRatio = %9.3f  %s
	CCRatio = %9.3f  %s
	CURatio = %9.3f  %s
	DRRatio = %9.3f  %s
	EXClose = %9.3f  %s
	GDRatio = %9.3f  %s
	HSRatio = %9.3f  %s
	IERatio = %9.3f  %s
	IPRatio = %9.3f  %s
	IRRatio = %9.3f  %s
	MPRatio = %9.3f  %s
	M1Ratio = %9.3f  %s
	M2Ratio = %9.3f  %s
	RSRatio = %9.3f  %s
	SPRatio = %9.3f  %s
	URRatio = %9.3f  %s
	FLAGS   = %b
	`,
		drec.Date.Format("Jan 2, 2006"),
		drec.BCRatio, isValidCheck(drec.FLAGS, DataFlags.BCRatioValid),
		drec.BPRatio, isValidCheck(drec.FLAGS, DataFlags.BPRatioValid),
		drec.CCRatio, isValidCheck(drec.FLAGS, DataFlags.CCRatioValid),
		drec.CURatio, isValidCheck(drec.FLAGS, DataFlags.CURatioValid),
		drec.DRRatio, isValidCheck(drec.FLAGS, DataFlags.DRRatioValid),
		drec.EXClose, isValidCheck(drec.FLAGS, DataFlags.EXCloseValid),
		drec.GDRatio, isValidCheck(drec.FLAGS, DataFlags.GDRatioValid),
		drec.HSRatio, isValidCheck(drec.FLAGS, DataFlags.HSRatioValid),
		drec.IERatio, isValidCheck(drec.FLAGS, DataFlags.IERatioValid),
		drec.IPRatio, isValidCheck(drec.FLAGS, DataFlags.IPRatioValid),
		drec.IRRatio, isValidCheck(drec.FLAGS, DataFlags.IRRatioValid),
		drec.MPRatio, isValidCheck(drec.FLAGS, DataFlags.MPRatioValid),
		drec.M1Ratio, isValidCheck(drec.FLAGS, DataFlags.M1RatioValid),
		drec.M2Ratio, isValidCheck(drec.FLAGS, DataFlags.M2RatioValid),
		drec.RSRatio, isValidCheck(drec.FLAGS, DataFlags.RSRatioValid),
		drec.SPRatio, isValidCheck(drec.FLAGS, DataFlags.SPRatioValid),
		drec.URRatio, isValidCheck(drec.FLAGS, DataFlags.URRatioValid),
		drec.FLAGS,
	)
	return s
}

// LRecToString returns a human-readable string for the supplied lrec
func LRecToString(lrec *LinguisticDataRecord) string {
	s := fmt.Sprintf(`    Date = %s
	LALLLSNScore      = %9.3f
	LALLLSPScore      = %9.3f
	LALLWHAScore      = %9.3f
	LALLWHOScore      = %9.3f
	LALLWHLScore      = %9.3f
	LALLWPAScore      = %9.3f
	LALLWDECount      = %9.3f
	LALLWDFCount      = %9.3f
	LALLWDPCount      = %9.3f
	LALLWDMCount      = %9.3f
	LUSDLSNScore_ECON = %9.3f
	LUSDLSPScore_ECON = %9.3f
	LUSDWHAScore_ECON = %9.3f
	LUSDWHOScore_ECON = %9.3f
	LUSDWHLScore_ECON = %9.3f
	LUSDWPAScore_ECON = %9.3f
	LUSDWDECount_ECON = %9.3f
	LUSDWDFCount_ECON = %9.3f
	LUSDWDPCount_ECON = %9.3f
	LUSDLIMCount_ECON = %9.3f
	LJPYLSNScore_ECON = %9.3f
	LJPYLSPScore_ECON = %9.3f
	LJPYWHAScore_ECON = %9.3f
	LJPYWHOScore_ECON = %9.3f
	LJPYWHLScore_ECON = %9.3f
	LJPYWPAScore_ECON = %9.3f
	LJPYWDECount_ECON = %9.3f
	LJPYWDFCount_ECON = %9.3f
	LJPYWDPCount_ECON = %9.3f
	LJPYLIMCount_ECON = %9.3f
	`,
		lrec.Date.Format("Jan 2, 2006"),
		lrec.LALLLSNScore,
		lrec.LALLLSPScore,
		lrec.LALLWHAScore,
		lrec.LALLWHOScore,
		lrec.LALLWHLScore,
		lrec.LALLWPAScore,
		lrec.LALLWDECount,
		lrec.LALLWDFCount,
		lrec.LALLWDPCount,
		lrec.LALLWDMCount,
		lrec.LUSDLSNScore_ECON,
		lrec.LUSDLSPScore_ECON,
		lrec.LUSDWHAScore_ECON,
		lrec.LUSDWHOScore_ECON,
		lrec.LUSDWHLScore_ECON,
		lrec.LUSDWPAScore_ECON,
		lrec.LUSDWDECount_ECON,
		lrec.LUSDWDFCount_ECON,
		lrec.LUSDWDPCount_ECON,
		lrec.LUSDLIMCount_ECON,
		lrec.LJPYLSNScore_ECON,
		lrec.LJPYLSPScore_ECON,
		lrec.LJPYWHAScore_ECON,
		lrec.LJPYWHOScore_ECON,
		lrec.LJPYWHLScore_ECON,
		lrec.LJPYWPAScore_ECON,
		lrec.LJPYWDECount_ECON,
		lrec.LJPYWDFCount_ECON,
		lrec.LJPYWDPCount_ECON,
		lrec.LJPYLIMCount_ECON,
	)
	return s
}

func isValidCheck(u1, u2 uint64) string {
	if u1&u2 != 0 {
		return "√"
	}
	return ""
}

// HandleUTF8FileChars returns the first line of the file with
// utf8 markers removed if they were there. Otherwise, it just
// returns the input string.
// ----------------------------------------------------------------
func HandleUTF8FileChars(line string) string {
	bom := []byte{0xEF, 0xBB, 0xBF}
	strBytes := []byte(line)

	if len(strBytes) >= len(bom) && bytes.Equal(strBytes[:len(bom)], bom) {
		// If the line starts with BOM, remove it.
		line = string(strBytes[len(bom):])
	}
	return line
}
