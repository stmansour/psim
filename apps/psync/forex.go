package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
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
	// Set the time part to 00:00:00 and use UTC timezone
	fr.Date = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

// FetchForexRates fetches forex rates from Trading Economics.
// -----------------------------------------------------------------------------
func FetchForexRates(startDate, endDate time.Time) ([]ForexRate, error) {
	var currencies []string
	for _, currency := range app.currencies {
		currencies = append(currencies, currency+":CUR")
	}
	currencyPairs := strings.Join(currencies, ",")

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
func UpdateForex(fxrs []ForexRate) error {
	for i := 0; i < len(fxrs); i++ {
		fields := []newdata.FieldSelector{}
		var f newdata.FieldSelector
		f.Metric = fxrs[i].Symbol[0:6] + "EXClose"
		fields = append(fields, f)

		//---------------------------------------------------------
		// Try to read this value and let's see what we may already
		// have for it
		//---------------------------------------------------------
		rec, err := app.SQLDB.Select(fxrs[i].Date, fields)
		if err != nil {
			return err
		}
		metric := util.Stripchars(f.Metric, " ")
		if len(rec.Fields) == 0 {
			// currently we do not have data for this exchange rate.  Add it...
			fld := newdata.MetricInfo{
				Value: fxrs[i].Close,
				MSID:  app.MSID,
			}
			flds := make(map[string]newdata.MetricInfo, 1)
			flds[metric] = fld
			rec.Fields = flds
			rec.Date = fxrs[i].Date
			if err = app.SQLDB.Insert(rec); err != nil {
				return err
			}
		} else if rec.Fields[metric].Value != fxrs[i].Close {
			fmt.Printf("*** MISCOMPARE ***\n")
			fmt.Printf("    Rec:  Date = %s, metric = %s, rec.Fields[metric] = %v\n", rec.Date.Format("2006-01-02"), metric, rec.Fields[metric].Value)
			fmt.Printf("    API:  Date = %s, i = %d, fxrs[i].Close = %.6f\n", fxrs[i].Date.Format("2006-01-02"), i, fxrs[i].Close)
		}
	}

	return nil
}
