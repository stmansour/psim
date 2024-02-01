package data

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/stmansour/psim/util"
)

// HoldLimits are used to define the delta between ratios that are to be considered as a "hold"
// The values are a percentage. The percentages are applied to the first value on or after the
// simulation date. The area between delta*mn and delta*mx is the hold area.
// -----------------------------------------------------------------------------------------------
type HoldLimits struct {
	Mn float64
	Mx float64
}

// DBInfo is meta information about the discount rate data
var DBInfo struct {
	MaxValidBitpos int

	//-----------------------------------------------------------------------------------------------------
	// note that HoldRec.Date is not necessarily the date for the values found.
	// The values are the first valid values found on or after the simulation start date.
	//-----------------------------------------------------------------------------------------------------
	HoldRec RatesAndRatiosRecord // values on first day of simulation to help set "hold" area.

	//-----------------------------------------------------------------------------------------------------
	// HoldSpace is a map that defines the hold space for the data type used as an index
	//-----------------------------------------------------------------------------------------------------
	HoldSpace map[string]HoldLimits
}

//****************************************************************************
//
//  Column naming and formatting conventions:
//
// Date,USD_DiscountRate,JPY_DiscountRate,USDJPY_DRRatio
//
//----------------------------------------------------------------------------
//  COLUMN 1  =  DATE
//----------------------------------------------------------------------------
//  Date     - MM/DD/YYYY
//
//----------------------------------------------------------------------------
//  COLUMN 2 thru n  =  statistics data
//----------------------------------------------------------------------------
//  The general format is:
//
//      [C1][C2]{DataType}(Qualifier)
//
//  For Currency, use the ISO 4217 naming conventions, 3-letter strings, the
//  first two identify the country, the last is represents the currency name.
//
//  	Examples:
//              USD = United States Dollar
//              JPY Japanese Yen
//
//  DataType - use a 2 letter identifier:
//      BC = Business Confidence
//      BP = Building Permits
//      CC = Consumer Confidence
//		CP = Corporate Profits  ***
//      CU = Capacity Utilization
//      DR = Discount Rate
//      EX = Exchange Rate -- can be appended with "Open", "Low", "High", "Close"
//      GD = Government Debt to GDP
//      HS = Housing Starts
//      IE = Inflation Expectation
//      IP = Industrial Production
//      IR = Inflation Rate
//      L0 = Linguistic sentiment positive
//      L1 = Linguistic LSNScore_ECON
//      L2 = Linguistic WHAScore_ECON
//      L3 = Linguistic WHOScore_ECON
//      L4 = Linguistic WHLScore_ECON
//      L5 = Linguistic WPAScore_ECON
//		L6 = Linguistic LALLWDECount
//		L7 = Linguistic LALLWDFCount
//		L8 = Linguistic LALLWDPCount
//		L9 = Linguistic LALLWDMCount
//		LA = Linguistic LUSDLSNScore_ECON
//		LB = Linguistic LUSDLSPScore_ECON
//		LC = Linguistic LUSDWHAScore_ECON
//		LD = Linguistic LUSDWHOScore_ECON
//		LE = Linguistic LUSDWHLScore_ECON
//		LF = Linguistic LUSDWPAScore_ECON
//		LG = Linguistic LUSDWDECount_ECON
//		LH = Linguistic LUSDWDFCount_ECON
//		LI = Linguistic LUSDWDPCount_ECON
//		LJ = Linguistic LUSDLIMCount_ECON
//      M1 = Money Supply - short term
//      M2 = Money Supply - long term
//      MR = Manufacturing ?
//      RS = Retail Sales
//      SP = Stock Prices
//      UR = Unemployment Rate
//
//  Qualifier
//      Ratio - indicates that the value is a ratio
//      Close - indicates that this is the "Close" value for the date.  Currently,
//              it applies only the Exchange Rate (EX) info
//
//      Examples:
//              USDJPYDRRatio - USD / JPY Discount Rate Ratio
//              USDJPYEXClose = USD / JPY Exchange Rate Closing value
//
//****************************************************************************

// LoadCsvDB - Read in the data from the CSV file
//  1. Determine whether the data will come from a CSV file, a SQL
//     database, or an online service.  As of this writing we only have
//     CSV file data implemented.
//  2. If data source is CSV read it in and validate that we have the
//     correct information.
//
// ---------------------------------------------------------------------------
func LoadCsvDB() error {
	err := LoadCsvData()
	if err != nil {
		return err
	}
	DBInfo.HoldSpace = make(map[string]HoldLimits)
	err = InitHoldSpace()
	if err != nil {
		return err
	}
	return nil
}

// LoadCsvData does the bulk of the work for LoadCsvDB
func LoadCsvData() error {
	file, err := os.Open(PLATODB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//-------------------------------------------------------
	// Here are the types of data the influencers support...
	//-------------------------------------------------------
	DInfo.DTypes = []string{
		"BCRatio", //  0 - Business Confidence Ratio
		"BPRatio", //  1 – Building Permits Ratio
		"CCRatio", //  2 - Consumer Confidence Ratio
		"CPRatio", //  3 - -Corporate Profits Ratio
		"CURatio", //  4 – Capacity Utilization Ratio
		"DRRatio", //  5 – Discount Rate Ratio
		"EXClose", //  6 - Exchange Rate Close
		"GDRatio", //  7 – Debt to GDP Ratio
		"HSRatio", //  8 - Housing Starts Ratio
		"IERatio", //  9 - Inflation Expectations Ratio
		"IPRatio", // 10 – Industrial Production Ratio
		"IRRatio", // 11 - Inflation Rate Ratio
		"M1Ratio", // 12 – M1 Money Supply Ratio
		"M2Ratio", // 13 – M2 Money Supply Ratio
		"MPRatio", // 14 – Manufacturing Production Ratio
		"RSRatio", // 15 – Retail Sales Ratio
		"SPRatio", // 16 – Stock Price Ratio
		"URRatio", // 17 – Unemployment Rate Ratio
		// "L0Ratio", // not needed for linguistics
		// "L1Ratio", // not needed for linguistics
		// "L2Ratio", // not needed for linguistics
		// "L3Ratio", // not needed for linguistics
		// "L4Ratio", // not needed for linguistics
		// "L5Ratio", // not needed for linguistics
	}
	//----------------------------------------------------------------------
	// Keep track of the column with the data needed for each ratio.  This
	// is based on the two currencies in the simulation.
	//----------------------------------------------------------------------
	DInfo.CSVMap = map[string]int{}
	for k := 0; k < len(DInfo.DTypes); k++ {
		DInfo.CSVMap[DInfo.DTypes[k]] = -1 // haven't located this column yet
	}
	records := RatesAndRatiosRecords{}

	for i, line := range lines {
		if i == 0 {
			// handle the unicode case...
			line[0] = HandleUTF8FileChars(line[0])

			if line[0] != "Date" {
				log.Panicf("Problem with %s, column 1 is labelled %q, it should be %q\n", PLATODB, line[0], "Date")
			}
			//----------------------------------------------------------------------------
			// Search for the columns of interest. Record the column numbers in the map.
			// We're looking for EXClose, URRatio, DRRatio, etc.
			//----------------------------------------------------------------------------
			for j := 1; j < len(line); j++ {
				validcpair := validCurrencyPair(line[j]) // do the first 6 chars make a currency pair that matches with the simulation configuation?
				l := len(line[j])
				for k := 0; k < len(DInfo.DTypes); k++ {
					if l == 13 && validcpair && strings.HasSuffix(line[j], DInfo.DTypes[k]) {
						DInfo.CSVMap[DInfo.DTypes[k]] = j // column located.  ex: DInfo.CSVMap["DRRatio"] = j
					}
				}
			}

			//--------------------------------------------------------------
			// Make sure we have the data we need for the simulation...
			//--------------------------------------------------------------
			for k := 0; k < len(DInfo.DTypes); k++ {
				if DInfo.DTypes[k][0] == 'L' {
					continue // at this point, we're just going to assume the linguistics are there``
				}
				if subclassIsUsedInSimulation(DInfo.DTypes[k]) && DInfo.CSVMap[DInfo.DTypes[k]] == -1 {
					s := fmt.Sprintf("no column in %s had label  %s%s%s, which is required for the current simulation configuration",
						PLATODB, DInfo.cfg.C1, DInfo.cfg.C2, DInfo.DTypes[k])
					util.DPrintf(s)
					return fmt.Errorf(s)
				}
			}

			continue // remaining rows are data, code below handles data, continue to the next line now
		}

		date, err := util.StringToDate(line[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		FLAGS := uint64(0) // assume no info exists

		BCRatio, exists := getNamedFloat("BCRatio", line, 0)
		FLAGS |= exists
		DataFlags.BCRatioValid = exists

		BPRatio, exists := getNamedFloat("BPRatio", line, 1)
		FLAGS |= exists
		DataFlags.BPRatioValid = exists

		CCRatio, exists := getNamedFloat("CCRatio", line, 2)
		FLAGS |= exists
		DataFlags.CCRatioValid = exists

		CURatio, exists := getNamedFloat("CURatio", line, 3)
		FLAGS |= exists
		DataFlags.CURatioValid = exists

		DRRatio, exists := getNamedFloat("DRRatio", line, 4)
		FLAGS |= exists
		DataFlags.DRRatioValid = exists

		EXClose, exists := getNamedFloat("EXClose", line, 5)
		FLAGS |= exists
		DataFlags.EXCloseValid = exists

		GDRatio, exists := getNamedFloat("GDRatio", line, 6)
		FLAGS |= exists
		DataFlags.GDRatioValid = exists
		// util.DPrintf("GDRatio = %9.2f, exists = %b\n", GDRatio, exists)

		HSRatio, exists := getNamedFloat("HSRatio", line, 7)
		FLAGS |= exists
		DataFlags.HSRatioValid = exists

		IERatio, exists := getNamedFloat("IERatio", line, 8)
		FLAGS |= exists
		DataFlags.IERatioValid = exists

		IPRatio, exists := getNamedFloat("IPRatio", line, 9)
		FLAGS |= exists
		DataFlags.IPRatioValid = exists

		IRRatio, exists := getNamedFloat("IRRatio", line, 10)
		FLAGS |= exists
		DataFlags.IRRatioValid = exists

		MPRatio, exists := getNamedFloat("MPRatio", line, 11)
		FLAGS |= exists
		DataFlags.MPRatioValid = exists

		M1Ratio, exists := getNamedFloat("M1Ratio", line, 12)
		FLAGS |= exists
		DataFlags.M1RatioValid = exists

		M2Ratio, exists := getNamedFloat("M2Ratio", line, 13)
		FLAGS |= exists
		DataFlags.M2RatioValid = exists

		RSRatio, exists := getNamedFloat("RSRatio", line, 14)
		FLAGS |= exists
		DataFlags.RSRatioValid = exists

		SPRatio, exists := getNamedFloat("SPRatio", line, 15)
		FLAGS |= exists
		DataFlags.SPRatioValid = exists

		URRatio, exists := getNamedFloat("URRatio", line, 16)
		FLAGS |= exists
		DataFlags.URRatioValid = exists

		records = append(records, RatesAndRatiosRecord{
			Date:    date,
			BCRatio: BCRatio,
			BPRatio: BPRatio,
			CCRatio: CCRatio,
			CURatio: CURatio,
			DRRatio: DRRatio,
			EXClose: EXClose,
			GDRatio: GDRatio,
			HSRatio: HSRatio,
			IERatio: IERatio,
			IPRatio: IPRatio,
			IRRatio: IRRatio,
			MPRatio: MPRatio,
			M1Ratio: M1Ratio,
			M2Ratio: M2Ratio,
			RSRatio: RSRatio,
			SPRatio: SPRatio,
			URRatio: URRatio,
			FLAGS:   FLAGS,
		})
	}

	DInfo.DBRecs = records
	sort.Sort(DInfo.DBRecs)
	l := DInfo.DBRecs.Len()
	DInfo.DtStart = DInfo.DBRecs[0].Date
	DInfo.DtStop = DInfo.DBRecs[l-1].Date

	if err = LoadLinguistics(lines); err != nil {
		fmt.Printf("error from LoadLingustics: %s\n", err.Error())
	}
	// util.DPrintf("Loaded %d records.   %s - %s\n", l, DInfo.DtStart.Format("jan 2, 2006"), DInfo.DtStop.Format("jan 2, 2006"))
	return nil
}

// LoadLinguistics loads the linguistic stats from the CSV file
//
// RETURNS
//
//	any error encountered
func LoadLinguistics(lines [][]string) error {
	var records []LinguisticDataRecord
	var err error
	cols := make(map[string]int, 100)
	for i, line := range lines {
		if i == 0 {
			for j := 0; j < len(line); j++ {
				if line[j][0] == 'L' {
					cols[line[j]] = j
					// fmt.Printf("col %d = %s\n", j, lines[0][j])
				}
			}
			continue // we've done all we need to do with lines[0]
		}

		var rec LinguisticDataRecord
		rec.Date, err = util.StringToDate(line[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for columnName, index := range cols {
			if len(line[index]) == 0 {
				continue
			}
			// Assuming all fields are float64 as per your struct
			if value, err := strconv.ParseFloat(line[index], 64); err == nil {
				switch columnName {
				case "LALLLSNScore":
					rec.LALLLSNScore = value
				case "LALLLSPScore":
					rec.LALLLSPScore = value
				case "LALLWHAScore":
					rec.LALLWHAScore = value
				case "LALLWHOScore":
					rec.LALLWHOScore = value
				case "LALLWHLScore":
					rec.LALLWHLScore = value
				case "LALLWPAScore":
					rec.LALLWPAScore = value
				case "LALLWDECount":
					rec.LALLWDECount = value
				case "LALLWDFCount":
					rec.LALLWDFCount = value
				case "LALLWDPCount":
					rec.LALLWDPCount = value
				case "LALLWDMCount":
					rec.LALLWDMCount = value
				case "LUSDLSNScore_ECON":
					rec.LUSDLSNScore_ECON = value
				case "LUSDLSPScore_ECON":
					rec.LUSDLSPScore_ECON = value
				case "LUSDWHAScore_ECON":
					rec.LUSDWHAScore_ECON = value
				case "LUSDWHOScore_ECON":
					rec.LUSDWHOScore_ECON = value
				case "LUSDWHLScore_ECON":
					rec.LUSDWHLScore_ECON = value
				case "LUSDWPAScore_ECON":
					rec.LUSDWPAScore_ECON = value
				case "LUSDWDECount_ECON":
					rec.LUSDWDECount_ECON = value
				case "LUSDWDFCount_ECON":
					rec.LUSDWDFCount_ECON = value
				case "LUSDWDPCount_ECON":
					rec.LUSDWDPCount_ECON = value
				case "LUSDLIMCount_ECON":
					rec.LUSDLIMCount_ECON = value
				case "LJPYLSNScore_ECON":
					rec.LJPYLSNScore_ECON = value
				case "LJPYLSPScore_ECON":
					rec.LJPYLSPScore_ECON = value
				case "LJPYWHAScore_ECON":
					rec.LJPYWHAScore_ECON = value
				case "LJPYWHOScore_ECON":
					rec.LJPYWHOScore_ECON = value
				case "LJPYWHLScore_ECON":
					rec.LJPYWHLScore_ECON = value
				case "LJPYWPAScore_ECON":
					rec.LJPYWPAScore_ECON = value
				case "LJPYWDECount_ECON":
					rec.LJPYWDECount_ECON = value
				case "LJPYWDFCount_ECON":
					rec.LJPYWDFCount_ECON = value
				case "LJPYWDPCount_ECON":
					rec.LJPYWDPCount_ECON = value
				case "LJPYLIMCount_ECON":
					rec.LJPYLIMCount_ECON = value
				default:
					// Optionally handle unknown column names
				}
			} else {
				// Handle error in conversion
				fmt.Printf("Error converting value for %s: %v", columnName, err)
			}
		}
		records = append(records, rec)
	}

	DInfo.LRecs = records
	return nil
}

// subclassIsUsedInSimulation - returns true if the supplied Influencer Subtype
// string is used in this simulation.  Any string will work as long as the first
// two letters indicate the influencer subtype. This means that strings such as
// "IPRatio" will work to indicat the IPInfluencer.
// ---------------------------------------------------------------------------------
func subclassIsUsedInSimulation(ss string) bool {
	s := ss[:2] // we only need the first 2 chars
	for i := 0; i < len(DInfo.cfg.InfluencerSubclasses); i++ {
		if s == DInfo.cfg.InfluencerSubclasses[i][:2] {

			return true
		}
	}
	return false
}

// getNamedFloat - centralize a bunch of lines that would need to be
//
//	repeated for every column of data without this func.
//
// INPUTS
//
//	val = name of data column excluding C1C2
//	line = array of strings -- parsed csv input line
//	bitpos = bit position in FLAGS for this particular column
//
// RETURNS
// float64 = the ratio if it exists, value is only valid if bool is true
// uint64  = a flag in bitpos - if 1 it means that the value is valid, 0
//
//	means the value was not supplied.
//
// --------------------------------------------------------------------------
func getNamedFloat(val string, line []string, bitpos int) (float64, uint64) {
	var flags uint64

	// util.DPrintf("bitpos = %d, find %s val... ", bitpos, val)

	key, exists := DInfo.CSVMap[val]
	if !exists || key < 0 {
		// util.DPrintf("failed! A\n")
		return 0, 0
	}
	s := line[key]
	if s == "" {
		// util.DPrintf("failed! B\n")
		return 0, 0
	}
	ratio, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Panicf("getNamedFloat: invalid value: %q, err = %s\n", val, err)
	}
	flags |= 1 << bitpos
	// util.DPrintf("success!\n")

	// keep track of the maximum bit position
	if bitpos > DBInfo.MaxValidBitpos {
		DBInfo.MaxValidBitpos = bitpos
	}

	return ratio, flags
}

func validCurrencyPair(line string) bool {
	if len(line) < 6 {
		return false
	}
	myC1 := line[0:3]
	myC2 := line[3:6]
	return myC1 == DInfo.cfg.C1 && myC2 == DInfo.cfg.C2
}

// Len returns the length of the supplied RatesAndRatiosRecords array
func (r RatesAndRatiosRecords) Len() int {
	return len(r)
}

// Less is used to sort the records
func (r RatesAndRatiosRecords) Less(i, j int) bool {
	return r[i].Date.Before(r[j].Date)
}

// Swap is used to do exactly what you think it does
func (r RatesAndRatiosRecords) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// CSVDBFindRecord returns the record associated with the input date
//
// INPUTS
//
//	dt = date of record to return
//
// RETURNS
//
//	pointer to the record on the supplied date
//	nil - record was not found
//
// ---------------------------------------------------------------------------
func CSVDBFindRecord(dt time.Time) *RatesAndRatiosRecord {
	left := 0
	right := len(DInfo.DBRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if DInfo.DBRecs[mid].Date.Year() == dt.Year() && DInfo.DBRecs[mid].Date.Month() == dt.Month() && DInfo.DBRecs[mid].Date.Day() == dt.Day() {
			return &DInfo.DBRecs[mid]
		} else if DInfo.DBRecs[mid].Date.Before(dt) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return nil
}

// CSVDBFindLRecord returns the lingustics record associated with the input date
//
// INPUTS
//
//	dt = date of record to return
//
// RETURNS
//
//	pointer to the record on the supplied date
//	nil - record was not found
//
// ---------------------------------------------------------------------------
func CSVDBFindLRecord(dt time.Time) *LinguisticDataRecord {
	left := 0
	right := len(DInfo.DBRecs) - 1

	for left <= right {
		mid := left + (right-left)/2
		if DInfo.DBRecs[mid].Date.Year() == dt.Year() && DInfo.DBRecs[mid].Date.Month() == dt.Month() && DInfo.DBRecs[mid].Date.Day() == dt.Day() {
			return &DInfo.LRecs[mid]
		} else if DInfo.DBRecs[mid].Date.Before(dt) {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return nil
}
