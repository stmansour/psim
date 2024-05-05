package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
)

// ForexRate struct that captures the data supplied by the Trading Economics API.
type ForexRate struct {
	Symbol string    `json:"Symbol"`
	Date   time.Time `json:"Date"` // Custom parsing to handle time in UTC
	Open   float64   `json:"Open"`
	High   float64   `json:"High"`
	Low    float64   `json:"Low"`
	Close  float64   `json:"Close"`
}

// UnmarshalJSON implements an interface that allows our specially formatted
// dates to be parsed by go's json unmarshaling code.
// -----------------------------------------------------------------------------
func (fr *ForexRate) UnmarshalJSON(data []byte) error {
	type Alias ForexRate
	aux := &struct {
		Date string `json:"Date"`
		*Alias
	}{
		Alias: (*Alias)(fr),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	parsedDate, err := time.Parse("02/01/2006", aux.Date)
	if err != nil {
		return err
	}
	fr.Date = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

// FetchForexRates fetches forex rates from Trading Economics. The calls
// are broken up into 13 currencies/commodities at a time.
// -----------------------------------------------------------------------------
func FetchForexRates(startDate, endDate time.Time, forex []PML) ([]ForexRate, error) {
	const maxcurrencies = 13
	allforex := []ForexRate{}

	for i := 0; i < len(forex); i += maxcurrencies {
		n := maxcurrencies
		if len(forex)-i < maxcurrencies {
			n = len(forex) - i
		}
		//---------------------------------------------------------------
		// pull out the Handles we'll be using for this call...
		//---------------------------------------------------------------
		fx := []string{}
		for j := i; j < i+n; j++ {
			fx = append(fx, forex[j].Handle)
		}
		//--------------------------------
		// fetch the values in fx
		//--------------------------------
		ind, err := doFetchForexRates(strings.Join(fx, ","), startDate, endDate)
		if err != nil {
			return nil, err
		}
		allforex = append(allforex, ind...)
		time.Sleep(1 * time.Second)
	}
	return allforex, nil
}

// doFetchForexRates does the API work of fetching the data
func doFetchForexRates(currencyPairs string, startDate, endDate time.Time) ([]ForexRate, error) {
	url := fmt.Sprintf("https://api.tradingeconomics.com/markets/historical/%s?c=%s&d1=%s&d2=%s&f=json",
		currencyPairs, app.APIKey, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	resp, err := http.Get(url)
	app.HTTPGetCalls++
	if err != nil {
		app.HTTPGetErrs++
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	var rates []ForexRate
	if err := json.Unmarshal(body, &rates); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return rates, nil
}

// UpdateForex updates the foreign exchange rates in the SQL database
func UpdateForex(forex []ForexRate, fxrs []PML) error {
	for i := 0; i < len(forex); i++ {
		fields := []newdata.FieldSelector{}
		var f newdata.FieldSelector

		//---------------------------------------------------------
		// Exchange rates are handled a little differently from
		// the commodities. If it is an exchange rate, the first
		// 6 characters will the the 3-letter currency codes of
		// the two countries involved.
		//---------------------------------------------------------
		isExchRate := false
		if len(forex[i].Symbol) > 6 && forex[i].Symbol[6] == ':' {
			if isCurrency(forex[i].Symbol[0:3]) && isCurrency(forex[i].Symbol[3:6]) {
				isExchRate = true
			}
		}

		//-----------------------------------------------------------------------------
		// Set the metric name to all characters up to the first colon ":" if it is
		// an exchange rate.
		//-----------------------------------------------------------------------------
		if isExchRate {
			colonIndex := strings.Index(forex[i].Symbol, ":")
			if colonIndex == -1 {
				fmt.Printf("No colon in symbol: %s\n", forex[i].Symbol)
				return fmt.Errorf("no colon in symbol: %s", forex[i].Symbol)
			}
			f.Metric = forex[i].Symbol[0:colonIndex]
			f.Metric += "EXClose"
		} else {
			f.Metric = ""
			for j := 0; j < len(fxrs); j++ {
				if fxrs[j].Handle == forex[i].Symbol {
					f.Metric = fxrs[j].Metric
					break
				}
			}
			if len(f.Metric) == 0 {
				return fmt.Errorf("could not find metric for %s", forex[i].Symbol)
			}
		}
		fields = append(fields, f)

		//---------------------------------------------------------
		// Try to read this value and let's see what we may already
		// have for it
		//---------------------------------------------------------
		rec, err := app.SQLDB.Select(forex[i].Date, fields)
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
			fmt.Printf("%s - %s: %.4f", forex[i].Date.Format("2006-01-02"), forex[i].Symbol, forex[i].Close)
		}
		if len(rec.Fields) == 0 {
			//------------------------------------------------------
			// This is case 1. We do not have this data. Add it...
			//------------------------------------------------------
			if app.Verbose {
				fmt.Printf(" || SQL Record not found, adding\n")
			}
			fld := newdata.MetricInfo{
				Value: forex[i].Close,
				MSID:  app.MSID,
			}
			flds := make(map[string]newdata.MetricInfo, 1)
			flds[f.Metric] = fld
			rec.Fields = flds
			rec.Date = forex[i].Date
			if err = app.SQLDB.Insert(rec); err != nil {
				return err
			}
		} else if toleranceMiscompare(rec.Fields[f.Metric].Value, forex[i].Close) {
			//------------------------------------------------------------
			// This is case 2. We have it, but it miscompares.  Flag it!
			//------------------------------------------------------------
			if app.Verbose {
				fmt.Printf(" || SQL Record miscompare\n")
			}
			fmt.Printf("*** MISCOMPARE - FOREX VALUE ***\n")
			fmt.Printf("    Rec:  Date = %s, f.Metric = %s, rec.Fields[f.Metric] = %v\n", rec.Date.Format("2006-01-02"), f.Metric, rec.Fields[f.Metric].Value)
			fmt.Printf("    API:  Date = %s, i = %d, forex[i].Close = %.2f\n", forex[i].Date.Format("2006-01-02"), i, forex[i].Close)
			//------------------------------------------------------------------
			// If the flag is set, we can fix miscompares by using the API data
			//------------------------------------------------------------------
			if app.APIFixMiscompares {
				var newrec newdata.EconometricsRecord
				newrec.Date = rec.Date
				newfld := rec.Fields[f.Metric] // start with the original information
				newfld.Value = forex[i].Close  // here's the new value
				newfld.MSID = app.MSID         // we're now updating the value from the Metric Source
				newrec.Fields = make(map[string]newdata.MetricInfo, 1)
				newrec.Fields[f.Metric] = newfld
				if err = app.SQLDB.Update(&newrec); err != nil {
					return err
				}
				app.Corrected++
			}
			app.Miscompared++
		} else {
			//-----------------------------------------------------------------------
			// This is case 3. We have it, and it compares. Update verified count...
			//-----------------------------------------------------------------------
			if app.Verbose {
				fmt.Printf(" || SQL Record values matched, verified\n")
			}
			app.Verified++
		}
	}

	return nil
}

func isCurrency(s string) bool {
	for _, v := range app.SQLDB.SQLDB.LocaleCache {
		if v.Currency == s {
			return true
		}
	}
	return false
}
