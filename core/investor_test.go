package core

import (
	"testing"

	"github.com/stmansour/psim/util"
)

func TestInvestorDNA(t *testing.T) {
	// var f Factory
	v := Investor{
		Delta4: 4,
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

	drDNA := dr.DNA()
	util.DPrintf("drDNA = %s\n", drDNA)
	irDNA := ir.DNA()
	util.DPrintf("irDNA = %s\n", irDNA)

	InvestorDNA := v.DNA()
	util.DPrintf("InvestorDNA = %s\n", InvestorDNA)

	expected := "{Investor,Delta4=4;Influencers=[{DRInfluencer,Delta1=-30,Delta2=-2,Delta4=4}|{IRInfluencer,Delta1=-17,Delta2=-3,Delta4=4}]}"
	if InvestorDNA != expected {
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
	util.DPrintf("Investor2DNA = %s\n", Investor2DNA)
	expected2 := "{Investor,Delta4=2;Influencers=[{DRInfluencer,Delta1=-28,Delta2=-3,Delta4=2}|{URInfluencer,Delta1=-17,Delta2=-3,Delta4=2}]}"
	if Investor2DNA != expected2 {
		t.Fail()
	}

}
