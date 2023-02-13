package main

// This app is used to calculate the correctness probabilities for all the
// Discount Rate data.  Within the Influencer for DiscountRate there are
// 3 variables.  Based on a purchase date, t3, they are:
//
//		dt1 = number of days prior to t3 where data research begins.
//            As of this writing its range is [2,30].
//      dt2 = number of days prior to t3 where data research stops
//            As of this writing its range is [1,5].
//		dt4 = number of days after t3 when the purchased currency is sold.
//            As of this writing its range is [1,5].
//---------------------------------------------------------------------------
import (
	"fmt"
	"os"
	"psim/core"
	"psim/data"
	"time"
)

func displayStats() (data.DRInfo, data.ERInfo) {
	drinfo := data.DRGetDataInfo()
	fmt.Printf("Discount Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", data.DR.DRRecs.Len())
	fmt.Printf("   Beginning:\t%s\n", drinfo.DtStart.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", drinfo.DtStop.Format("Jan 2, 2006"))

	erinfo := data.ERGetDataInfo()
	fmt.Printf("Exchange Rate Info:\n")
	fmt.Printf("   Records:\t%d\n", data.ER.ERRecs.Len())
	fmt.Printf("   Beginning:\t%s\n", erinfo.DtStart.Format("Jan 2, 2006"))
	fmt.Printf("   Ending:\t%s\n", erinfo.DtStop.Format("Jan 2, 2006"))

	return drinfo, erinfo
}

func checkDR(t3 time.Time) {
	rec := data.DRFindRecord(t3)
	if rec == nil {
		fmt.Println("ExchangeRate Record not found.")
		os.Exit(1)
	}

	if !rec.Date.Equal(t3) {
		fmt.Printf("date did not match!\n")
		os.Exit(1)
	}
	if rec.USJPDRRatio != -2.5 {
		fmt.Printf("USJPDRRatio did not match!\n")
		os.Exit(1)
	}
	if rec.USDiscountRate != 0.0025 {
		fmt.Printf("USDiscountRate did not match!\n")
		os.Exit(1)
	}
	if rec.JPDiscountRate != -0.001 {
		fmt.Printf("JPDiscountRate did not match!\n")
		os.Exit(1)
	}
}

func main() {
	data.Init()
	_, erinfo := displayStats()

	//--------------------------------------------------------------------------
	// We must insure that the date range for which we calculate probabilities
	// is such that data exists.  Now that we know how much data we have, make
	// any adjustments necessary.
	//--------------------------------------------------------------------------
	dtStart := erinfo.DtStart.AddDate(0, 0, -core.DR.T1min)
	dtStop := erinfo.DtStop.AddDate(0, 0, -core.DR.T4max)
	dtEnd := dtStop.AddDate(0, 0, 1)

	t3 := time.Date(2022, 2, 10, 0, 0, 0, 0, time.UTC)
	checkDR(t3) // Just make sure everything looks OK before starting...

	fmt.Printf("Probs from %s to %s\n", dtStart.Format("Jan 2, 2006"), dtStop.Format("Jan 2, 2006"))

	fmt.Printf("t1,t2,t3,t4, dt1, dt2, dt4, drr, err, prediction, actual\n")
	for dt := dtStart; dtEnd.After(dt); dt = dt.AddDate(0, 0, 1) {
		fmt.Printf("%s\n", dt.Format("Jan 2, 2006"))
		generateProbabilities(dt)
	}

}

func generateProbabilities(t3 time.Time) {
	t1a := core.DR.T1min // range for t1 relative to t3
	t1b := core.DR.T1max

	t2a := core.DR.T2min // range for t2 relative to t3
	t2b := core.DR.T2max

	t4a := core.DR.T4min // range for t4 relative to t3
	t4b := core.DR.T4max

	dt2 := 0 // define the deltas for each of the moving dates
	dt4 := 0
	dt1 := t1a

	n := 0
	for t1 := t3.AddDate(0, 0, t1a); t1.Before(t3.AddDate(0, 0, t1b+1)); t1 = t1.AddDate(0, 0, 1) {
		dt2 = t2a
		for t2 := t3.AddDate(0, 0, t2a); t2.Before(t3.AddDate(0, 0, t2b+1)); t2 = t2.AddDate(0, 0, 1) {
			if t2.After(t1.AddDate(0, 0, 1)) {
				dt4 = t4a
				for t4 := t3.AddDate(0, 0, t4a); t4.Before(t3.AddDate(0, 0, t4b+1)); t4 = t4.AddDate(0, 0, 1) {
					//-------------------------------------------------------------
					// Determine DRR = (DiscountRate at t1) - (DiscountRate at t2)
					//-------------------------------------------------------------
					rec1 := data.DRFindRecord(t1)
					if rec1 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t1.Format("1/2/2006"))
						os.Exit(1)
					}
					rec2 := data.DRFindRecord(t2)
					if rec2 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t2.Format("1/2/2006"))
						os.Exit(1)
					}
					drr := rec1.USJPDRRatio - rec2.USJPDRRatio

					//-----------------------------------------------------------------
					// Determine ERR = (ExchangeRatio at t1) - (ExchangeRatio at t2)
					//-----------------------------------------------------------------
					er1 := data.ERFindRecord(t1)
					if er1 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t1.Format("1/2/2006"))
						os.Exit(1)
					}
					er2 := data.ERFindRecord(t2)
					if er2 == nil {
						fmt.Printf("ExchangeRate Record for %s not found.\n", t2.Format("1/2/2006"))
						os.Exit(1)
					}
					err := er1.Close - er2.Close

					fmt.Printf("%s,%s,%s,%s,%d,%d,%d,%6.2f,%6.2f,%s\n",
						t1.Format("01/02/2006"), t2.Format("01/02/2006"),
						t3.Format("01/02/2006"), t4.Format("01/02/2006"),
						dt1, dt2, dt4,
						drr, err,
						"n/a",
					)
					dt4++
					n++
				}
			}
			dt2++
		}
		dt1++
	}
	fmt.Printf("total possibilities: %d\n", n)
}
