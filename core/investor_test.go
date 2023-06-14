package core

import (
	"fmt"
	"log"
	"testing"

	"github.com/stmansour/psim/util"
)

func TestInvestorDNA(t *testing.T) {
	var f Factory
	util.Init()
	cfg := util.CreateTestingCFG()
	if err := util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	f.Init(cfg)

	v := Investor{}
	v.Init(cfg, &f)
	v.Delta4 = 4        // Since we're not creating from DNA we need to force this, otherwise it will be random
	v.W1 = float64(0.5) // same as above
	v.W2 = float64(0.5) // same as above

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
	dr.Init(&v, cfg, v.Delta4)
	ir.Init(&v, cfg, v.Delta4)
	v.Influencers = []Influencer{&dr, &ir}

	// drDNA := dr.DNA()
	// util.DPrintf("drDNA = %s\n", drDNA)
	// irDNA := ir.DNA()
	// util.DPrintf("irDNA = %s\n", irDNA)
	InvestorDNA := v.DNA()
	// util.DPrintf("InvestorDNA = %s\n", InvestorDNA)

	expected := "{Investor;Delta4=4;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-30,Delta2=-2,Delta4=4}|{IRInfluencer,Delta1=-17,Delta2=-3,Delta4=4}]}"
	if InvestorDNA != expected {
		fmt.Printf("-----------------------------------------------------------------------\n")
		fmt.Printf("MISMATCHED DNA\n")
		fmt.Printf("InvestorDNA = %s\n", InvestorDNA)
		fmt.Printf("expected    = %s\n\n", expected)
		t.Fail()
	}

	v2 := Investor{
		Delta4: 2,
	}
	v2.Init(cfg, &f)
	v2.Delta4 = 2        // Since we're not creating from DNA we need to force this, otherwise it will be random
	v2.W1 = float64(0.5) // same as above
	v2.W2 = float64(0.5) // same as above
	dr2 := DRInfluencer{
		Delta1: -28,
		Delta2: -3,
	}
	ur2 := URInfluencer{
		Delta1: -17,
		Delta2: -3,
	}
	dr2.Init(&v2, cfg, v2.Delta4)
	ur2.Init(&v2, cfg, v2.Delta4)
	v2.Influencers = []Influencer{&dr2, &ur2}
	Investor2DNA := v2.DNA()
	expected2 := "{Investor;Delta4=2;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-28,Delta2=-3,Delta4=2}|{URInfluencer,Delta1=-17,Delta2=-3,Delta4=2}]}"
	if Investor2DNA != expected2 {
		fmt.Printf("MISMATCHED DNA\n")
		fmt.Printf("Investor2DNA = %s\n", Investor2DNA)
		fmt.Printf("expected2    = %s\n", expected2)
		t.Fail()
	}

}
