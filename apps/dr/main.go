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
	"flag"
	"fmt"
	"os"
	"psim/core"
	"psim/data"
	"time"
)

type probInfo struct {
	count   int64
	correct int64
	prob    float64
}

var app struct {
	dri         data.DRInfo
	eri         data.ERInfo
	probMap     map[string]probInfo
	showInfo    bool
	showAllRecs bool
	showResults bool
}

func displayStats() (data.DRInfo, data.ERInfo) {
	drinfo := data.DRGetDataInfo()
	if app.showInfo {
		fmt.Printf("Discount Rate Info:\n")
		fmt.Printf("   Records:\t%d\n", data.DR.DRRecs.Len())
		fmt.Printf("   Beginning:\t%s\n", drinfo.DtStart.Format("Jan 2, 2006"))
		fmt.Printf("   Ending:\t%s\n", drinfo.DtStop.Format("Jan 2, 2006"))
	}

	erinfo := data.ERGetDataInfo()
	if app.showInfo {
		fmt.Printf("Exchange Rate Info:\n")
		fmt.Printf("   Records:\t%d\n", data.ER.ERRecs.Len())
		fmt.Printf("   Beginning:\t%s\n", erinfo.DtStart.Format("Jan 2, 2006"))
		fmt.Printf("   Ending:\t%s\n", erinfo.DtStop.Format("Jan 2, 2006"))
	}

	return drinfo, erinfo
}

func checkDR(t3 time.Time) {
	rec := data.DRFindRecord(t3)
	if rec == nil {
		fmt.Println("DiscountRate Record not found.")
		os.Exit(1)
	}

	if !rec.Date.Equal(t3) {
		fmt.Printf("date did not match!\n")
		os.Exit(1)
	}
	if rec.USJPDRRatio != -15.0 {
		fmt.Printf("USJPDRRatio did not match!  Read %7.4f, looking for: -15 \n", rec.USJPDRRatio)
		os.Exit(1)
	}
	if rec.USDiscountRate != 0.015 {
		fmt.Printf("USDiscountRate did not match!  Read %7.4f, looking for: 0.015 \n", rec.USDiscountRate)
		os.Exit(1)
	}
	if rec.JPDiscountRate != -0.001 {
		fmt.Printf("JPDiscountRate did not match!  Read %7.4f, looking for: -0.001 \n", rec.JPDiscountRate)
		os.Exit(1)
	}
}

func readCommandLineArgs() {
	infoPtr := flag.Bool("i", false, "show info about the probability data range")
	allRecsPtr := flag.Bool("a", false, "output all records used in the analysis")
	flag.Parse()
	app.showInfo = *infoPtr
	app.showAllRecs = *allRecsPtr
	app.showResults = !app.showAllRecs
}

func main() {
	app.probMap = map[string]probInfo{}
	readCommandLineArgs()

	data.Init()
	checkDR(time.Date(2018, 2, 14, 0, 0, 0, 0, time.UTC)) // Just make sure everything looks OK before starting...

	//-------------------------------------
	// Now set up the boundaries...
	//-------------------------------------
	drinfo, erinfo := displayStats()
	app.dri = drinfo
	app.eri = erinfo

	//--------------------------------------------------------------------------
	// We must insure that the date range for which we calculate probabilities
	// is such that data exists.  Now that we know how much data we have, make
	// any adjustments necessary.
	//--------------------------------------------------------------------------
	dtStart := erinfo.DtStart.AddDate(0, 0, -core.DR.T1min)
	dtStop := erinfo.DtStop.AddDate(0, 0, -core.DR.T4max-1)

	//--------------------------------------------------------------------------
	// Adjust these dates if the DR data does not yet exist...
	//--------------------------------------------------------------------------
	if dtStop.After(drinfo.DtStop) {
		dtStop = drinfo.DtStop
	}
	if drinfo.DtStart.After(dtStart) {
		dtStart = drinfo.DtStart
	}

	dtEnd := dtStop.AddDate(0, 0, 1)
	fmt.Printf("Probs from %s to %s\n", dtStart.Format("Jan 2, 2006"), dtEnd.Format("Jan 2, 2006"))

	// CSV column headings
	if app.showAllRecs {
		fmt.Printf("t1,t2,t3,t4, dt1, dt2, dt4, dDRR, dERR, prediction, actual\n")
	}

	for dt := dtStart; dtEnd.After(dt); dt = dt.AddDate(0, 0, 1) {
		genProbs(dt)
	}

	//----------------------------------------------------------------------------------
	// Now we need to compute the accuracy.  That is, when the prediction was "buy" how
	// often was it right.
	// TODO: how often was it correct when it said to hold?  This might be worth knowing
	//----------------------------------------------------------------------------------
	if app.showResults {
		fmt.Printf("Sig1,Sig2,Sig3,Correct Predictions,Total Predictions, Correct Pct\n")
		for i := core.DR.T1min; i <= core.DR.T1max; i++ {
			for j := core.DR.T2min; j <= core.DR.T2max; j++ {
				if i == j {
					continue
				}
				for k := core.DR.T4min; k <= core.DR.T4max; k++ {
					s := fmt.Sprintf("%d,%d,%d", i, j, k)
					v := app.probMap[s]
					fmt.Printf("%s, %d, %d, %5.1f%%\n", s, v.correct, v.count, 100.0*v.prob)
				}
			}
		}
	}
}

func genProbs(t3 time.Time) {
	l := 0
	for i := core.DR.T1min; i <= core.DR.T1max; i++ {
		t1 := t3.AddDate(0, 0, i)
		for j := core.DR.T2min; j <= core.DR.T2max; j++ {
			if i == j {
				continue
			}
			t2 := t3.AddDate(0, 0, j)
			for k := core.DR.T4min; k <= core.DR.T4max; k++ {
				t4 := t3.AddDate(0, 0, k)
				l++
				// fmt.Printf("%d:  %d,%d,%d:   t1: %s, t2: %s, t4: %s\n", l, i, j, k, t1.Format("Jan 2, 2006"), t2.Format("Jan 2, 2006"), t4.Format("Jan 2, 2006"))
				computeDRProbability(t1, t2, t3, t4, i, j, k)

			}
		}
	}
}

// computeDRProbability
//
// INPUTS
//
//			t1 - first date of DRR analysis
//		    t2 - last date of DRR analysis
//		    t3 - check date, that is the date of buy or hold
//		    t4 - sell date if the prediction is to buy
//		    dt1, dt2, dt4 - these 3 numbers form the unique signature of a
//		         DiscountRate Influencer.  Their meanings are:
//	             * # days prior to check date to start analysis
//	             * # days prior to check date to stop analysis
//	             * # days after check date to sell (if the prediction is to buy)
//
// --------------------------------------------------------------------------------
func computeDRProbability(t1, t2, t3, t4 time.Time, dt1, dt2, dt4 int) {
	//---------------------------------------------------------------------------
	// Determine dDRR = (DiscountRateRatio at t1) - (DiscountRateRatio at t2)
	//---------------------------------------------------------------------------
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
	dDRR := rec1.USJPDRRatio - rec2.USJPDRRatio

	//-------------------------------------------------------------------------------
	// Prediction formula (based on the change in DiscountRateRatios):
	//     dDRR > 0:   buy on t3, sell on t4
	//     dDRR <= 0:  take no action
	//-------------------------------------------------------------------------------
	prediction := "hold"
	if dDRR > 0 {
		prediction = "buy"
	}

	//-------------------------------------------------------------------------------
	// Determine deltaER (dER) =
	//         (ExchangeRateRatio at t3) - (ExchangeRateRatio at t4)
	//-------------------------------------------------------------------------------
	er1 := data.ERFindRecord(t3)
	if er1 == nil {
		fmt.Printf("ExchangeRate Record for %s not found.\n", t3.Format("1/2/2006"))
		os.Exit(1)
	}
	er2 := data.ERFindRecord(t4)
	if er2 == nil {
		fmt.Printf("ExchangeRate Record for %s not found.\n", t4.Format("1/2/2006"))
		os.Exit(1)
	}
	dER := er1.Close - er2.Close

	//-------------------------------------------------------------------------------
	// Check to see if the prediction is correct. If dDRR > 0 AND dER > 0 then
	// then the prediction was correct.
	//-------------------------------------------------------------------------------
	predictionResult := false
	if dER > 0 && dDRR > 0 {
		predictionResult = true
	}

	//-------------------------------------------------------------------------------
	// Record the correctness.  A unique DiscountRateInfluencer type is defined by its
	// t1, t2, and t4 values.  Record the correctness (predictionResult) for each
	// DiscountInfluencer type
	//-------------------------------------------------------------------------------
	addToProbabilities(dt1, dt2, dt4, prediction, predictionResult)

	//-------------------------------------------------------------------------------
	// Print out for manual checking...
	//-------------------------------------------------------------------------------
	if app.showAllRecs {
		fmt.Printf("%s,%s,%s,%s,%d,%d,%d,%6.2f,%6.2f,%s,%t\n",
			t1.Format("01/02/2006"), t2.Format("01/02/2006"),
			t3.Format("01/02/2006"), t4.Format("01/02/2006"),
			dt1, dt2, dt4,
			dDRR, dER,
			prediction,
			predictionResult,
		)
	}

}

func addToProbabilities(i, j, k int, prediction string, predWasCorrect bool) {
	s := fmt.Sprintf("%d,%d,%d", i, j, k)
	v, ok := app.probMap[s]
	if !ok {
		v = probInfo{
			count:   0,
			correct: 0,
			prob:    0.0,
		}
	}
	if prediction == "buy" {
		v.count++
		if predWasCorrect {
			v.correct++
		}
		v.prob = float64(v.correct) / float64(v.count)
	}
	app.probMap[s] = v
}
