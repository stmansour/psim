package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestForexRateUnmarshalJSON(t *testing.T) {
	jsonInput := `[{"Symbol":"USDJPY:CUR","Date":"16/02/2024","Open":149.92,"High":150.64,"Low":149.81,"Close":150.21}]`
	var rates []ForexRate
	err := json.Unmarshal([]byte(jsonInput), &rates)
	if err != nil {
		t.Errorf("Unmarshal failed: %s", err)
	}

	expectedDate, _ := time.Parse("02/01/2006", "16/02/2024")
	if !rates[0].Date.Equal(expectedDate) {
		t.Errorf("Expected Date '%v', got '%v'", expectedDate, rates[0].Date)
	}

	// You can add more checks here for Open, High, Low, and Close fields
	fmt.Println("Test passed, Date:", rates[0].Date)

	// acid test
	jsonInput = `[{"Symbol":"USDJPY:CUR","Date":"16/02/2024","Open":149.920000000000,"High":150.640000000000,"Low":149.810000000000,"Close":150.210000000000},{"Symbol":"AUDUSD:CUR","Date":"16/02/2024","Open":0.652200000000,"High":0.654400000000,"Low":0.649500000000,"Close":0.653100000000},{"Symbol":"AUDUSD:CUR","Date":"15/02/2024","Open":0.648800000000,"High":0.652900000000,"Low":0.647500000000,"Close":0.652400000000},{"Symbol":"USDJPY:CUR","Date":"15/02/2024","Open":150.550000000000,"High":150.580000000000,"Low":149.530000000000,"Close":149.910000000000},{"Symbol":"USDJPY:CUR","Date":"14/02/2024","Open":150.780000000000,"High":150.800000000000,"Low":150.340000000000,"Close":150.550000000000},{"Symbol":"AUDUSD:CUR","Date":"14/02/2024","Open":0.645100000000,"High":0.649500000000,"Low":0.644400000000,"Close":0.649000000000},{"Symbol":"AUDUSD:CUR","Date":"13/02/2024","Open":0.653000000000,"High":0.653500000000,"Low":0.644100000000,"Close":0.645200000000},{"Symbol":"USDJPY:CUR","Date":"13/02/2024","Open":149.330000000000,"High":150.880000000000,"Low":149.250000000000,"Close":150.790000000000},{"Symbol":"USDJPY:CUR","Date":"12/02/2024","Open":149.240000000000,"High":149.470000000000,"Low":148.910000000000,"Close":149.340000000000},{"Symbol":"AUDUSD:CUR","Date":"12/02/2024","Open":0.651100000000,"High":0.654300000000,"Low":0.650400000000,"Close":0.652900000000},{"Symbol":"AUDUSD:CUR","Date":"09/02/2024","Open":0.649300000000,"High":0.653400000000,"Low":0.648500000000,"Close":0.652300000000},{"Symbol":"USDJPY:CUR","Date":"09/02/2024","Open":149.320000000000,"High":149.570000000000,"Low":149.000000000000,"Close":149.300000000000},{"Symbol":"USDJPY:CUR","Date":"08/02/2024","Open":148.180000000000,"High":149.480000000000,"Low":147.920000000000,"Close":149.310000000000},{"Symbol":"AUDUSD:CUR","Date":"08/02/2024","Open":0.651900000000,"High":0.653200000000,"Low":0.647800000000,"Close":0.649100000000},{"Symbol":"AUDUSD:CUR","Date":"07/02/2024","Open":0.652300000000,"High":0.654000000000,"Low":0.651400000000,"Close":0.651800000000},{"Symbol":"USDJPY:CUR","Date":"07/02/2024","Open":147.940000000000,"High":148.270000000000,"Low":147.620000000000,"Close":148.180000000000},{"Symbol":"USDJPY:CUR","Date":"06/02/2024","Open":148.650000000000,"High":148.810000000000,"Low":147.810000000000,"Close":147.940000000000},{"Symbol":"AUDUSD:CUR","Date":"06/02/2024","Open":0.648100000000,"High":0.652500000000,"Low":0.647500000000,"Close":0.652300000000},{"Symbol":"AUDUSD:CUR","Date":"05/02/2024","Open":0.650600000000,"High":0.652000000000,"Low":0.646700000000,"Close":0.648200000000},{"Symbol":"USDJPY:CUR","Date":"05/02/2024","Open":148.330000000000,"High":148.890000000000,"Low":148.250000000000,"Close":148.670000000000},{"Symbol":"USDJPY:CUR","Date":"02/02/2024","Open":146.420000000000,"High":148.580000000000,"Low":146.230000000000,"Close":148.370000000000},{"Symbol":"AUDUSD:CUR","Date":"02/02/2024","Open":0.657000000000,"High":0.661000000000,"Low":0.650100000000,"Close":0.651200000000},{"Symbol":"AUDUSD:CUR","Date":"01/02/2024","Open":0.656400000000,"High":0.657800000000,"Low":0.650600000000,"Close":0.656900000000},{"Symbol":"USDJPY:CUR","Date":"01/02/2024","Open":146.880000000000,"High":147.110000000000,"Low":145.880000000000,"Close":146.420000000000},{"Symbol":"USDJPY:CUR","Date":"31/01/2024","Open":147.600000000000,"High":147.890000000000,"Low":146.000000000000,"Close":146.880000000000},{"Symbol":"AUDUSD:CUR","Date":"31/01/2024","Open":0.660100000000,"High":0.662200000000,"Low":0.655000000000,"Close":0.656500000000},{"Symbol":"AUDUSD:CUR","Date":"30/01/2024","Open":0.661000000000,"High":0.662400000000,"Low":0.657300000000,"Close":0.660100000000},{"Symbol":"USDJPY:CUR","Date":"30/01/2024","Open":147.500000000000,"High":147.920000000000,"Low":147.090000000000,"Close":147.600000000000},{"Symbol":"USDJPY:CUR","Date":"29/01/2024","Open":148.110000000000,"High":148.330000000000,"Low":147.240000000000,"Close":147.490000000000},{"Symbol":"AUDUSD:CUR","Date":"29/01/2024","Open":0.657000000000,"High":0.661500000000,"Low":0.656700000000,"Close":0.661000000000},{"Symbol":"AUDUSD:CUR","Date":"26/01/2024","Open":0.658300000000,"High":0.660900000000,"Low":0.657100000000,"Close":0.657400000000},{"Symbol":"USDJPY:CUR","Date":"26/01/2024","Open":147.610000000000,"High":148.200000000000,"Low":147.400000000000,"Close":148.160000000000},{"Symbol":"USDJPY:CUR","Date":"25/01/2024","Open":147.510000000000,"High":147.920000000000,"Low":147.070000000000,"Close":147.650000000000},{"Symbol":"AUDUSD:CUR","Date":"25/01/2024","Open":0.657400000000,"High":0.660900000000,"Low":0.656400000000,"Close":0.658100000000},{"Symbol":"AUDUSD:CUR","Date":"24/01/2024","Open":0.657900000000,"High":0.662100000000,"Low":0.656300000000,"Close":0.657600000000},{"Symbol":"USDJPY:CUR","Date":"24/01/2024","Open":148.340000000000,"High":148.390000000000,"Low":146.640000000000,"Close":147.500000000000},{"Symbol":"USDJPY:CUR","Date":"23/01/2024","Open":148.090000000000,"High":148.700000000000,"Low":146.970000000000,"Close":148.360000000000},{"Symbol":"AUDUSD:CUR","Date":"23/01/2024","Open":0.657000000000,"High":0.661200000000,"Low":0.654900000000,"Close":0.657800000000},{"Symbol":"AUDUSD:CUR","Date":"22/01/2024","Open":0.658600000000,"High":0.661300000000,"Low":0.656400000000,"Close":0.657000000000},{"Symbol":"USDJPY:CUR","Date":"22/01/2024","Open":148.270000000000,"High":148.300000000000,"Low":147.600000000000,"Close":148.090000000000},{"Symbol":"USDJPY:CUR","Date":"19/01/2024","Open":148.110000000000,"High":148.800000000000,"Low":147.830000000000,"Close":148.140000000000},{"Symbol":"AUDUSD:CUR","Date":"19/01/2024","Open":0.657100000000,"High":0.660100000000,"Low":0.656300000000,"Close":0.659700000000},{"Symbol":"AUDUSD:CUR","Date":"18/01/2024","Open":0.655200000000,"High":0.657400000000,"Low":0.652400000000,"Close":0.657000000000},{"Symbol":"USDJPY:CUR","Date":"18/01/2024","Open":148.140000000000,"High":148.300000000000,"Low":147.640000000000,"Close":148.150000000000},{"Symbol":"USDJPY:CUR","Date":"17/01/2024","Open":147.180000000000,"High":148.520000000000,"Low":147.060000000000,"Close":148.150000000000},{"Symbol":"AUDUSD:CUR","Date":"17/01/2024","Open":0.658300000000,"High":0.659400000000,"Low":0.652300000000,"Close":0.655000000000},{"Symbol":"AUDUSD:CUR","Date":"16/01/2024","Open":0.665800000000,"High":0.666400000000,"Low":0.657500000000,"Close":0.658300000000},{"Symbol":"USDJPY:CUR","Date":"16/01/2024","Open":145.720000000000,"High":147.310000000000,"Low":145.570000000000,"Close":147.180000000000},{"Symbol":"USDJPY:CUR","Date":"15/01/2024","Open":144.900000000000,"High":145.940000000000,"Low":144.850000000000,"Close":145.730000000000},{"Symbol":"AUDUSD:CUR","Date":"15/01/2024","Open":0.667900000000,"High":0.670400000000,"Low":0.664800000000,"Close":0.666000000000},{"Symbol":"AUDUSD:CUR","Date":"12/01/2024","Open":0.668300000000,"High":0.672800000000,"Low":0.667500000000,"Close":0.668500000000},{"Symbol":"USDJPY:CUR","Date":"12/01/2024","Open":145.280000000000,"High":145.560000000000,"Low":144.340000000000,"Close":144.900000000000},{"Symbol":"USDJPY:CUR","Date":"11/01/2024","Open":145.720000000000,"High":146.410000000000,"Low":145.250000000000,"Close":145.280000000000},{"Symbol":"AUDUSD:CUR","Date":"11/01/2024","Open":0.669500000000,"High":0.672500000000,"Low":0.664500000000,"Close":0.668700000000},{"Symbol":"AUDUSD:CUR","Date":"10/01/2024","Open":0.668300000000,"High":0.671300000000,"Low":0.667700000000,"Close":0.669800000000},{"Symbol":"USDJPY:CUR","Date":"10/01/2024","Open":144.460000000000,"High":145.830000000000,"Low":144.300000000000,"Close":145.730000000000},{"Symbol":"USDJPY:CUR","Date":"09/01/2024","Open":144.260000000000,"High":144.620000000000,"Low":143.410000000000,"Close":144.470000000000},{"Symbol":"AUDUSD:CUR","Date":"09/01/2024","Open":0.671500000000,"High":0.673400000000,"Low":0.667500000000,"Close":0.668200000000},{"Symbol":"AUDUSD:CUR","Date":"08/01/2024","Open":0.670800000000,"High":0.673400000000,"Low":0.667500000000,"Close":0.671800000000},{"Symbol":"USDJPY:CUR","Date":"08/01/2024","Open":144.570000000000,"High":144.920000000000,"Low":143.650000000000,"Close":144.220000000000},{"Symbol":"USDJPY:CUR","Date":"05/01/2024","Open":144.620000000000,"High":145.980000000000,"Low":143.800000000000,"Close":144.650000000000},{"Symbol":"AUDUSD:CUR","Date":"05/01/2024","Open":0.671000000000,"High":0.674800000000,"Low":0.663900000000,"Close":0.671300000000},{"Symbol":"AUDUSD:CUR","Date":"04/01/2024","Open":0.673200000000,"High":0.676000000000,"Low":0.669400000000,"Close":0.670500000000},{"Symbol":"USDJPY:CUR","Date":"04/01/2024","Open":143.280000000000,"High":144.850000000000,"Low":142.840000000000,"Close":144.620000000000},{"Symbol":"USDJPY:CUR","Date":"03/01/2024","Open":141.980000000000,"High":143.730000000000,"Low":141.850000000000,"Close":143.290000000000},{"Symbol":"AUDUSD:CUR","Date":"03/01/2024","Open":0.676300000000,"High":0.677000000000,"Low":0.670000000000,"Close":0.672900000000},{"Symbol":"AUDUSD:CUR","Date":"02/01/2024","Open":0.681400000000,"High":0.683900000000,"Low":0.675400000000,"Close":0.676000000000},{"Symbol":"USDJPY:CUR","Date":"02/01/2024","Open":140.870000000000,"High":142.210000000000,"Low":140.800000000000,"Close":141.980000000000},{"Symbol":"USDJPY:CUR","Date":"01/01/2024","Open":140.820000000000,"High":140.890000000000,"Low":140.820000000000,"Close":140.870000000000},{"Symbol":"AUDUSD:CUR","Date":"01/01/2024","Open":0.681000000000,"High":0.682000000000,"Low":0.680400000000,"Close":0.681100000000}]`
	err = json.Unmarshal([]byte(jsonInput), &rates)
	if err != nil {
		t.Errorf("Unmarshal failed: %s", err)
	}
	fmt.Printf("Total records: %d\n", len(rates))
}