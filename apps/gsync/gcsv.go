package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
	"gonum.org/v1/gonum/stat"
)

// LocInfo contains information about a country
type LocInfo struct {
	Country      string
	CurrencyCode string
	Mentioned    bool // true if this country has been mentioned in the article we're  processing
}

// toleranceMiscompare checks if the two values are within application's
// allowable tolerance. Returns true if they are not.
// -------------------------------------------------------------------------
func toleranceMiscompare(v1, v2 float64) bool {
	x := v1 - v2
	if x < 0 {
		x = -x
	}
	return x > app.Tolerance
}

// ProcessGDELTCSV reads a tab-delimited file line by line
// and processes it.
// -----------------------------------------------------------------
func ProcessGDELTCSV(filename string) error {
	//---------------------------------
	// Open the file
	//---------------------------------
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	//----------------------------------------------------------------------
	// we read the values for the variables into an array of float64 values
	//----------------------------------------------------------------------
	gvals := make(map[string][]float64)
	app.locs = []LocInfo{
		{"United States", "USD", false},
		{"Japan", "JPY", false},
	}

	//-------------------------------------------------
	// Create a Scanner to read the file line by line
	//-------------------------------------------------
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)   // Start with a 1 MB buffer
	scanner.Buffer(buf, 10*1024*1024) // Allow the buffer to grow up to 16 times the default size 1024*1024/65536

	if app.Verbose {
		fmt.Printf("Beginning processing of: %s\n", filename)
	}
	t0 := time.Now()

	row := 0
	for scanner.Scan() {
		line := scanner.Text()
		row++
		if row == 1 {
			continue
		}

		fields := strings.Split(line, "\t") // Split the line by tab delimiter
		if len(fields) < 18 {
			fmt.Printf("*** WARNING *** skipping row %d: invalid number of fields: %d\n", row, len(fields))
			if app.Verbose {
				for i := 0; i < len(fields); i++ {
					fmt.Printf("Field %d: %s\n", i, fields[i])
				}
			}
			continue
		}
		gcam := strings.Split(fields[17], ",") // this is the GCAM column

		//-------------------------------------------------------------------
		// Check for country "mentions" before we process the GCAM column so
		// we only loop through it once.
		//-------------------------------------------------------------------
		for j := 0; j < len(app.locs); j++ {
			target := "1#" + app.locs[j].Country
			app.locs[j].Mentioned = false                                                     // assume it's not mentioned
			if strings.Contains(fields[10], target) && strings.Contains(fields[8], "ECON_") { // TODO - check this line
				app.locs[j].Mentioned = true // it's mentioned
			}
		}

		for i := 0; i < len(gcam); i++ {
			//--------------------------------------------------------------
			// First pass... we parse each variable in the GCAM column
			//--------------------------------------------------------------
			tuple := strings.Split(gcam[i], ":")                        // this is the metric/value pair, example: m1:val1
			metric := "GCAM_" + strings.Replace(tuple[0], ".", "_", -1) // the metric name is "GCAM_" + tuple[0] where all dots are underscores
			metric = strings.ToUpper(metric)                            // GCAM metrics are upper case

			//-----------------------------
			// Is this a known metric?
			//-----------------------------
			if _, ok := app.SQLDB.Mim.MInfluencerSubclasses[metric]; !ok {
				continue // we're not tracking this metric
			}
			val, err := strconv.ParseFloat(tuple[1], 64) // the value is a float64
			if err != nil {
				return fmt.Errorf("error row %d, col 17, tuple = %s, error parsing value: %s", row, gcam[i], err)
			}
			if _, ok := gvals[metric]; !ok {
				gvals[metric] = make([]float64, 0)
				if app.Verbose {
					fmt.Printf("Found metric: %s\n", metric)
				}
			}

		//	if app.Verbose {
		//		fmt.Printf("%s : %f\n", metric, val) // DEBUG
		//	}
			gvals[metric] = append(gvals[metric], val)

			//-----------------------------------------------------------------------
			// Next, we do the ECON pass, where we append ECON to the variable name
			// and process it only when the string "1#" + countryName appears in
			// column J, Locations
			//-----------------------------------------------------------------------
			for j := 0; j < len(app.locs); j++ {
				if app.locs[j].Mentioned {
					lclmetric := metric + "_ECON" // don't add the locale here
					//-----------------------------
					// Is this a known metric?
					//-----------------------------
					if _, ok := app.SQLDB.Mim.MInfluencerSubclasses[lclmetric]; !ok {
						continue // we're not tracking this metric
					}
					lclmetric = app.locs[j].CurrencyCode + lclmetric // add the locale here
					if _, ok := gvals[lclmetric]; !ok {
						gvals[lclmetric] = make([]float64, 0)
						if app.Verbose {
							fmt.Printf("Found metric: %s\n", lclmetric)
						}
					}
					gvals[lclmetric] = append(gvals[lclmetric], val)
				}
			}
		}
	}
	t1 := time.Now()

	if app.Verbose {
		fmt.Printf("Finished in %s. Total variables found = %d\n", util.ElapsedTime(t0, t1), len(gvals))
		fmt.Printf("Will now write to database\n")
	}

	grandtotal := int64(0)
	//------------------------------------------------------------------------------------------
	// Now we've got all the numbers... compute the mean and write each metric to the database
	//------------------------------------------------------------------------------------------
	for k, v := range gvals {
		//-----------------------------------
		// Prepare the record to be written
		//-----------------------------------
		mean := stat.Mean(v, nil)
		if app.Verbose {
			l := len(v)
			grandtotal += int64(l)
			fmt.Printf("Number of values for %s: %d,  mean = %f\n", k, l, mean)
		}
		mi := newdata.MetricInfo{
			Value: mean,
			MSID:  app.MSID,
		}
		rec := newdata.EconometricsRecord{
			Date: app.dt,
		}
		rec.Fields = make(map[string]newdata.MetricInfo, 1)
		rec.Fields[k] = mi

		//-------------------------------------------------
		// Read the value first to see if we have this data
		//-------------------------------------------------
		ss := []newdata.FieldSelector{}
		fld := newdata.FieldSelector{
			Metric: k,
		}
		ss = append(ss, fld)
		existingRec, err := app.SQLDB.Select(app.dt, ss)
		if err != nil {
			return err
		}

		//------------------------------------------------------------------------------
		// one of 3 things happens now:
		//    1. if this data was not found in the database, we insert it
		//    2. if the value in the database is different, we report the difference
		//    3. if it matches, we mark a successful validation of the value
		//------------------------------------------------------------------------------
		if app.Verbose {
			fmt.Printf("%s - %s: %.4f", app.dt.Format("2006-01-02"), k, mean)
		}

		if len(existingRec.Fields) > 0 {
			if erf, ok := existingRec.Fields[k]; ok {
				//---------------------------------------------------------------------------------
				// Found it in the database. Check to see if the value matches what we calculated.
				//---------------------------------------------------------------------------------
				if toleranceMiscompare(erf.Value, mean) {
					//------------------------------------------------
					// This is case 2. We have it, but it miscompares.
					//------------------------------------------------
					if app.Verbose {
						fmt.Printf(" || SQL Record miscompare\n")
					}
					fmt.Printf("*** MISCOMPARE - GCAM VALUE ***\n")
					fmt.Printf("      Rec:  Date = %s, metric = %s, rec.Fields[metric] = %v\n", rec.Date.Format("2006-01-02"), k, existingRec.Fields[k].Value)
					fmt.Printf("    GDELT:  Date = %s, mean(gvals[%q]) = %.2f\n", app.dt.Format("2006-01-02"), k, mean)
					app.Miscompared++

					//----------------------------------------------------------------------------
					// If we're fixing miscompares, we need to update the value in the database
					//----------------------------------------------------------------------------
					if app.FixMiscompares {
						mi.ID = existingRec.Fields[k].ID // this is the MEID of the value we're replacing
						rec.Fields = make(map[string]newdata.MetricInfo, 1)
						rec.Fields[k] = mi

						if err := app.SQLDB.Update(&rec); err != nil {
							return err
						}
						app.Corrected++
						if app.Verbose {
							fmt.Printf("corrected\n")
						}
					}
				} else {
					//------------------------------------------------
					// This is case 3. We have it, and it matches.
					//------------------------------------------------
					if app.Verbose {
						fmt.Printf(" || SQL Record values matched, verified\n")
					}
					app.Verified++
					continue
				}
			} else {
				//------------------------------------------------------------------------------------------
				// This is Case 1. This veriable at this date is not in the database. We need to insert it.
				//------------------------------------------------------------------------------------------
				if app.Verbose {
					fmt.Printf(" || SQL Record not found, adding\n")
				}
				if err := app.SQLDB.Insert(&rec); err != nil {
					return err
				}
			}
		} else {
			//------------------------------------------------------------------------------------------
			// This is Case 1. This veriable at this date is not in the database. We need to insert it.
			//------------------------------------------------------------------------------------------
			if err := app.SQLDB.Insert(&rec); err != nil {
				return err
			}
			if app.Verbose {
				fmt.Printf("  || SQL Record not found, adding\n")
			}
		}
	}

	if app.Verbose {
		fmt.Printf("Total number of values processed: %d\n", grandtotal)
	}

	//-------------------------------------------------
	// Save the max number of metrics we processed...
	//-------------------------------------------------
	if len(gvals) > app.MetricsTotal {
		app.MetricsTotal = len(gvals)
	}

	//------------------------------------
	// Check for errors during scanning
	//------------------------------------
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from file: %w", err)
	}

	return nil
}
