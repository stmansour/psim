package data

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// InitHoldSpace finds the "hold" area for each data type based on
//
//	the configuration percentages for HoldWindowPos and HoldWindowNeg
//
// Look for the values on the first day. If not found, search for the first
// date past the simulation start date where the value is present and base
// if on that.
//
// INPUTS
//
//	nothing at this time
//
// RETURNS
//
//	any error found, nil if no errors
//
// --------------------------------------------------------------------------------
func InitHoldSpace() error {
	if DBInfo.HoldRec.FLAGS != 0 {
		return nil
	}

	//------------------------------------------------------
	// Here's the starting place for the HoldRec
	//------------------------------------------------------
	dbrec := CSVDBFindRecord(time.Time(DInfo.cfg.DtStart))
	if dbrec == nil {
		log.Panicf("No data record for simulation start date %s", time.Time(DInfo.cfg.DtStart).Format("Jan 2, 2006"))
	}

	DBInfo.HoldRec = *dbrec

	//--------------------------------------------
	// For each data type, set the hold area.
	//--------------------------------------------
	for i := 0; i <= DBInfo.MaxValidBitpos; i++ {
		if !strings.Contains(DInfo.DTypes[i], "Ratio") {
			continue // if it's something other than a ratio, just continue
		}
		//-----------------------------------------------------------
		// Make sure this data type was called out in config.json5
		//-----------------------------------------------------------
		found := false
		if len(DInfo.DTypes[i]) < 2 {
			continue // Skip if dType is too short to have two characters
		}
		firstTwoChars := DInfo.DTypes[i][:2] // Get first two characters of dType
		for _, influencer := range DInfo.cfg.InfluencerSubclasses {
			if strings.HasPrefix(influencer, firstTwoChars) {
				found = true
				break
			}
		}
		if !found {
			continue // this simulation doesn't use the specified data type
		}

		var x float64
		var err error
		var h HoldLimits
		if x, err = FirstValidRecord(uint64(i)); err != nil {
			log.Panicf("%s: %s", DInfo.DTypes[i], err.Error())
		}
		h.Mn = x * DInfo.cfg.HoldWindowNeg
		h.Mx = x * DInfo.cfg.HoldWindowPos
		DBInfo.HoldSpace[DInfo.DTypes[i]] = h
	}
	return nil
}

// FirstValidRecord searches for the first valid value in the specified bit position
//
//	It sets the value of DBInfo.HoldRec and sets the flag to
//	indicate that the value at that bitpos is correct.
//
// This implementation is not optimal. It should be reimplemented with
// a faster algorithm
//
// RETURNS
//
//	any error found, nil if no errors
//
// --------------------------------------------------------------------------------
func FirstValidRecord(bitpos uint64) (float64, error) {
	var x float64
	if DBInfo.HoldRec.FLAGS&(1<<bitpos) != 0 {
		x = GetHoldRecRatio(bitpos, false, nil)
		return x, nil
	}
	//----------------------------------------------
	// We need to search for the first valid entry
	//----------------------------------------------
	flg := uint64(1 << bitpos)
	dtStart := time.Time(DBInfo.HoldRec.Date).AddDate(0, 0, 1)
	dtStop := time.Time(DInfo.cfg.DtStop).AddDate(0, 0, 1)
	for dt := dtStart; dt.Before(dtStop); {
		rec := CSVDBFindRecord(dt)
		if rec.FLAGS&flg != 0 {
			DBInfo.HoldRec.FLAGS |= flg
			x = GetHoldRecRatio(bitpos, true, rec)
			return x, nil
		}
		dt = dt.AddDate(0, 0, 1)
	}
	return 0, fmt.Errorf("no valid data found between %s and %s", DBInfo.HoldRec.Date.Format("Jan 2, 2006"), dtStop.AddDate(0, 0, -1).Format("Jan 2, 2006"))
}

// GetHoldRecRatio returns the ratio indicated by the bit position
// --------------------------------------------------------------------
func GetHoldRecRatio(bitpos uint64, setval bool, rec *RatesAndRatiosRecord) float64 {
	var x float64
	switch DInfo.DTypes[bitpos] {
	case "BCRatio":
		if setval {
			DBInfo.HoldRec.BCRatio = rec.BCRatio
		}
		x = DBInfo.HoldRec.BCRatio
	case "BPRatio":
		if setval {
			DBInfo.HoldRec.BPRatio = rec.BPRatio
		}
		x = DBInfo.HoldRec.BPRatio
	case "CCRatio":
		if setval {
			DBInfo.HoldRec.CCRatio = rec.CCRatio
		}
		x = DBInfo.HoldRec.CCRatio
	case "CURatio":
		if setval {
			DBInfo.HoldRec.CURatio = rec.CURatio
		}
		x = DBInfo.HoldRec.CURatio
	case "DRRatio":
		if setval {
			DBInfo.HoldRec.DRRatio = rec.DRRatio
		}
		x = DBInfo.HoldRec.DRRatio
	case "EXClose":
		if setval {
			DBInfo.HoldRec.EXClose = rec.EXClose
		}
		x = DBInfo.HoldRec.EXClose
	case "GDRatio":
		if setval {
			DBInfo.HoldRec.GDRatio = rec.GDRatio
		}
		x = DBInfo.HoldRec.GDRatio
	case "HSRatio":
		if setval {
			DBInfo.HoldRec.HSRatio = rec.HSRatio
		}
		x = DBInfo.HoldRec.HSRatio
	case "IERatio":
		if setval {
			DBInfo.HoldRec.IERatio = rec.IERatio
		}
		x = DBInfo.HoldRec.IERatio
	case "IPRatio":
		if setval {
			DBInfo.HoldRec.IPRatio = rec.IPRatio
		}
		x = DBInfo.HoldRec.IPRatio
	case "IRRatio":
		if setval {
			DBInfo.HoldRec.IRRatio = rec.IRRatio
		}
		x = DBInfo.HoldRec.IRRatio
	case "MPRatio":
		if setval {
			DBInfo.HoldRec.MPRatio = rec.MPRatio
		}
		x = DBInfo.HoldRec.MPRatio
	case "M1Ratio":
		if setval {
			DBInfo.HoldRec.M1Ratio = rec.M1Ratio
		}
		x = DBInfo.HoldRec.M1Ratio
	case "M2Ratio":
		if setval {
			DBInfo.HoldRec.M2Ratio = rec.M2Ratio
		}
		x = DBInfo.HoldRec.M2Ratio
	case "RSRatio":
		if setval {
			DBInfo.HoldRec.RSRatio = rec.RSRatio
		}
		x = DBInfo.HoldRec.RSRatio
	case "SPRatio":
		if setval {
			DBInfo.HoldRec.SPRatio = rec.SPRatio
		}
		x = DBInfo.HoldRec.SPRatio
	case "URRatio":
		if setval {
			DBInfo.HoldRec.URRatio = rec.URRatio
		}
		x = DBInfo.HoldRec.URRatio
	}
	return x
}
