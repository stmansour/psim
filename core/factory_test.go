package core

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stmansour/psim/util"
)

// TestParseInvestorDNA - verify that the parser can correctly parse n Investor DNA string
// ------------------------------------------------------------------------------------------
func TestParseInvestorDNA(t *testing.T) {
	var f Factory
	var tests = []struct {
		input   string
		wantMap map[string]interface{}
	}{
		{
			"{invVar1=YesIDo;invVar2=34;Influencers=[{subclass1,var1=NotAtAll,var2=1.0}|{subclass2,var1=2,var2=2.0}];invVar3=3.1416}",
			map[string]interface{}{
				"invVar1": "YesIDo",
				"invVar2": 34,
				"invVar3": float64(3.1416),
			},
		},
		// add more test cases here
	}

	for _, tt := range tests {
		gotMap, err := f.ParseInvestorDNA(tt.input)
		if err != nil {
			fmt.Printf("Error returned by ParseInvestorDNA = %s\n", err.Error())
			continue
		}

		if !reflect.DeepEqual(gotMap, tt.wantMap) {
			t.Errorf("parseInvestorDNA(%q) map = %v, want %v", tt.input, gotMap, tt.wantMap)
		}
	}
}

func TestInvestorFromDNA(t *testing.T) {
	var f Factory
	util.Init()
	f.Init(CreateTestingCFG())

	parent1 := Investor{
		Delta4: 4,
		W1:     0.5,
		W2:     0.5,
	}
	dr := DRInfluencer{
		Delta1: -30,
		Delta2: -2,
		Delta4: parent1.Delta4,
	}
	ir := IRInfluencer{
		Delta1: -17,
		Delta2: -3,
		Delta4: parent1.Delta4,
	}
	dr.SetID()
	ir.SetID()
	parent1.Influencers = append(parent1.Influencers, &dr, &ir)

	parent2 := Investor{
		Delta4: 2,
		W1:     0.5,
		W2:     0.5,
	}
	dr2 := DRInfluencer{
		Delta1: -28,
		Delta2: -3,
		Delta4: parent2.Delta4,
	}
	ur2 := URInfluencer{
		Delta1: -17,
		Delta2: -3,
		Delta4: parent2.Delta4,
	}
	dr2.SetID()
	ur2.SetID()
	parent2.Influencers = append(parent2.Influencers, &dr2, &ur2)
	population := []Investor{}
	population = append(population, parent1, parent2)

	investor := f.NewInvestor(&population, 0, 1)
	newDNA := investor.DNA()
	util.DPrintf("newDNA = %s\n", newDNA)
	t.Fail()

}

func CreateTestingCFG() *util.AppConfig {
	cfg := util.AppConfig{
		// DtStart: "2022-01-01", // simulation start date for each generation
		// DtStop: "2022-12-31",  // simulation stop date for each generation
		Generations:    1,        // how many generations should the simulator run
		PopulationSize: 10,       // Total number Investors in the population
		C1:             "USD",    // main currency  (ISO 4217 code)
		C2:             "YEN",    // currency that we will invest in (ISO 4217 code)
		ExchangeRate:   "USDJPY", // forex conventional notation for Exchange Rate
		InitFunds:      1000.00,  // how much each Investor is funded at the start of a simulation cycle
		StdInvestment:  100.00,   // the "standard" investment amount if a decision is made to invest in C2
		MinDelta1:      -30,      // greatest amount of time prior to T3 that T1 can be
		MaxDelta1:      -2,       // least amount of time prior to T3 that T1 can be
		MinDelta2:      -5,       // greatest amount of time prior to T3 that T2 can be, constraint: MaxDelta2 > MaxDelta1
		MaxDelta2:      -1,       // least amount of time prior to T3 that T2 can be, with the constraint that MinDelta1 < MaxDelta2
		MinDelta4:      1,        // shortest period of time after a "buy" on T3 that we can do a "sell"
		MaxDelta4:      14,       // greatest period of time after a "buy" on T3 that we can do a "sell"
		DRW1:           0.6,      // DRInfluencer Fitness Score weighting for "correctness" of predictions. Constraint: DRW1 + DRW2 = 1.0
		DRW2:           0.4,      // DRInfluencer Fitness Score weighting for number of predictions made. Constraint: DRW1 + DRW2 = 1.0
		InvW1:          0.5,      // Investor Fitness Score weighting for "correctness" of predictions. Constraint: InvW1 + InvW2 = 1.0
		InvW2:          0.5,      // Investor Fitness Score weighting for profit. Constraint: InvW1 + InvW2 = 1.0
	}
	return &cfg
}
