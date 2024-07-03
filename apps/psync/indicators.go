package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// BC = Business Climate Indicator	Business ???
//      Business Conditions Index	Business ???
//      Business Confidence			Business ???
// CU = Currency?   Currency
// DR = Discount Rate  ???
// GD = Government Debt ??
// HS = Housing Starts  -->   Housing Starts MoM
// IE = Inflation Expectations ???
// IP = Industrial Production ????
// IR = Inflation Rate  ---->   Inflation Rate MoM
// M1 = Money Supply M1
// M2 = Money Supply M2
// MP = Manufacturing Production???
// RS = Retail Sales???? there are 5 versions of this
// SP = Steel Production ???  NOTE: there is no "Stock Price" indicator -- there is a Stock Market indicator
// UR = Unemployment Rate

// Indicator struct as previously discussed.
type Indicator struct {
	Country              string    `json:"Country"`
	Category             string    `json:"Category"`
	DateTime             time.Time `json:"DateTime"`
	Value                float64   `json:"Value"`
	Frequency            string    `json:"Frequency"`
	HistoricalDataSymbol string    `json:"HistoricalDataSymbol"`
	LastUpdate           time.Time `json:"LastUpdate"`
}

// UnmarshalJSON implements an interface that allows our specially formatted
// dates to be parsed by go's json unmarshaling code.
// ---------------------------------------------------------------------------
func (i *Indicator) UnmarshalJSON(data []byte) error {
	type Alias Indicator
	aux := &struct {
		DateTime   string `json:"DateTime"`
		LastUpdate string `json:"LastUpdate"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var err error
	i.DateTime, err = time.Parse("2006-01-02T15:04:05", aux.DateTime)
	if err != nil {
		return err
	}
	i.LastUpdate, err = time.Parse("2006-01-02T15:04:05", aux.LastUpdate)
	if err != nil {
		return err
	}

	return nil
}

// Construct the URL.
//              https://api.tradingeconomics.com/historical/country/mexico/indicator/gold%20reserves/2019-01-01/2019-12-31?f=json;c=guest:guest
//              https://api.tradingeconomics.com/historical/country/United%20States/indicator/Unemployment%20Rate/2024-04-01/2024-04-16?f=json
//              https://api.tradingeconomics.com/historical/country/United%20States,Japan/indicator/Housing%20Starts%20MoM,Stock%20Market,Unemployment%20Rate/2024-01-01/2024-04-16&f=json

// FetchIndicators fetches economic indicators from Trading Economics.
func FetchIndicators(startDate, endDate time.Time, ind []PML) ([]Indicator, error) {
	const maxIndicators = 13
	allcountries := strings.Join(app.countries, ",")
	allind := []Indicator{}

	for i := 0; i < len(ind); i += maxIndicators {
		n := maxIndicators
		if len(ind)-i < maxIndicators {
			n = len(ind) - i
		}
		si := []string{}
		for j := i; j < i+n; j++ {
			si = append(si, ind[j].Handle)
		}
		nd, err := doFetch(allcountries, strings.Join(si, ","), startDate, endDate)
		if err != nil {
			return nil, err
		}
		allind = append(allind, nd...)
		time.Sleep(1 * time.Second)
	}
	return allind, nil
}

// doFetch does the API work of fetching the data
// "https://api.tradingeconomics.com/historical/country/Australia,Japan,United%20States/indicator/Government%20Debt%20to%20GDP/2024-04-26/2024-05-03?f=json&c=8dba8cc926874fd:7fgqrflgveh7wfi"
//
//		https://api.tradingeconomics.com/historical/country/Australia,Japan,United%20States/indicator/Government%20Debt%20to%20GDP/2015-03-01/2024-05-03?f=json&c=8dba8cc926874fd:7fgqrflgveh7wfi
//	    https://api.tradingeconomics.com/historical/country/Australia,Japan,United%20States/indicator/Government%20Debt%20to%20GDP/2015-01-01/2024-05-03?f=json&c=8dba8cc926874fd:7fgqrflgveh7wfi
func doFetch(allcountries, allindicators string, startDate, endDate time.Time) ([]Indicator, error) {
	const maxRetries = 5
	const retryDelay = 5 * time.Second

	prefix := "https://api.tradingeconomics.com/historical/country/"
	teurl := fmt.Sprintf("%s/indicator/%s/%s/%s?f=json&c=%s", allcountries, allindicators, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), app.APIKey)
	teurl = prefix + strings.ReplaceAll(teurl, " ", "%20")

	if len(teurl) > 2000 {
		fmt.Printf("The URL is too long now, it is %d characters long.\n", len(teurl))
		fmt.Printf("we need to break it up\n")
	}

	for i := 0; i < maxRetries; i++ {
		// Send the HTTP request.
		resp, err := http.Get(teurl)
		app.HTTPGetCalls++
		if err != nil {
			app.HTTPGetErrs++
			if i < maxRetries-1 {
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("error making request: %v", err)
		}
		defer resp.Body.Close()

		// Read the response body.
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		// Parse the JSON data.
		var indicators []Indicator
		if err := json.Unmarshal(body, &indicators); err != nil {
			return nil, fmt.Errorf("error unmarshaling response: %v", err)
		}

		return indicators, nil
	}

	return nil, fmt.Errorf("exceeded maximum number of retries")
}
