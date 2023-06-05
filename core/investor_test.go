package core

import (
	"fmt"
	"testing"

	"github.com/stmansour/psim/util"
)

func TestInvestorDNA(t *testing.T) {
	var f Factory
	util.Init()
	f.Init(util.CreateTestingCFG())

	v := Investor{
		Delta4: 4,
		W1:     float64(0.5),
		W2:     float64(0.5),
	}
	dr := DRInfluencer{
		Delta1: -30,
		Delta2: -2,
		Delta4: v.Delta4,
	}
	ir := IRInfluencer{
		Delta1: -17,
		Delta2: -3,
		Delta4: v.Delta4,
	}
	v.Influencers = append(v.Influencers, &dr)
	v.Influencers = append(v.Influencers, &ir)

	// drDNA := dr.DNA()
	// util.DPrintf("drDNA = %s\n", drDNA)
	// irDNA := ir.DNA()
	// util.DPrintf("irDNA = %s\n", irDNA)

	InvestorDNA := v.DNA()
	// util.DPrintf("InvestorDNA = %s\n", InvestorDNA)

	expected := "{Investor;Delta4=4;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-30,Delta2=-2,Delta4=4}|{IRInfluencer,Delta1=-17,Delta2=-3,Delta4=4}]}"
	if InvestorDNA != expected {
		fmt.Printf("MISMATCHED DNA\n")
		fmt.Printf("Investor2DNA = %s\n", InvestorDNA)
		fmt.Printf("expected     = %s\n", expected)
		t.Fail()
	}

	v2 := Investor{
		Delta4: 2,
	}
	dr2 := DRInfluencer{
		Delta1: -28,
		Delta2: -3,
		Delta4: v2.Delta4,
	}
	ur2 := URInfluencer{
		Delta1: -17,
		Delta2: -3,
		Delta4: v2.Delta4,
	}
	v2.Influencers = append(v2.Influencers, &dr2)
	v2.Influencers = append(v2.Influencers, &ur2)
	Investor2DNA := v2.DNA()
	expected2 := "{Investor;Delta4=2;InvW1=0.0000;InvW2=0.0000;Influencers=[{DRInfluencer,Delta1=-28,Delta2=-3,Delta4=2}|{URInfluencer,Delta1=-17,Delta2=-3,Delta4=2}]}"
	if Investor2DNA != expected2 {
		fmt.Printf("MISMATCHED DNA\n")
		fmt.Printf("Investor2DNA = %s\n", Investor2DNA)
		fmt.Printf("expected2    = %s\n", expected2)
		t.Fail()
	}

}
