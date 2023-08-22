package core

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stmansour/psim/data"
	"github.com/stmansour/psim/util"
)

func TestInvestorDNA(t *testing.T) {
	var f Factory
	util.Init(-1)
	cfg := util.CreateTestingCFG()
	if err := util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	data.Init(cfg)
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
	InvestorDNA := v.DNA()

	expected := "{Investor;Delta4=4;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-30,Delta2=-2}|{IRInfluencer,Delta1=-17,Delta2=-3}]}"
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
	dr2.Init(&v2, dr.GetAppConfig(), v2.Delta4)
	ur2.Init(&v2, nil, v2.Delta4)
	v2.Influencers = []Influencer{&dr2, &ur2}
	Investor2DNA := v2.DNA()
	expected2 := "{Investor;Delta4=2;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-28,Delta2=-3}|{URInfluencer,Delta1=-17,Delta2=-3}]}"
	if Investor2DNA != expected2 {
		fmt.Printf("MISMATCHED DNA\n")
		fmt.Printf("Investor2DNA = %s\n", Investor2DNA)
		fmt.Printf("expected2    = %s\n", expected2)
		t.Fail()
	}

	var x Influencer
	var xa = []Influencer{&ur2, &dr}
	dt := time.Date(2018, time.July, 15, 0, 0, 0, 0, time.UTC)

	for i := 0; i < len(xa); i++ {
		x = xa[i]
		x.SetAppConfig(cfg)
		x.SetDelta1(-18)
		x.SetDelta2(-4)
		x.SetDelta4(5)
		if x.GetDelta1() != -18 {
			t.Errorf("x.Delta1: should be -18, but it is %d", x.GetDelta1())
		}
		if x.GetDelta2() != -4 {
			t.Errorf("x.Delta1: should be -4, but it is %d", x.GetDelta2())
		}
		if x.GetDelta4() != 5 {
			t.Errorf("x.Delta1: should be 5, but it is %d", x.GetDelta4())
		}

		action, prob, err := x.GetPrediction(dt)
		if err != nil {
			t.Errorf("GetPrediction returned error: %s", err)
		}
		fmt.Printf("action = %s, prob = %4.2f\n", action, prob)

		buy, err := x.MyInvestor().BuyConversion(dt)
		if err != nil {
			t.Errorf("BuyConversion returned error: %s", err)
		}
		fmt.Printf("buy = %d\n", buy)

		t4 := dt.AddDate(0, 0, x.GetDelta4())
		sellcount, err := x.MyInvestor().SellConversion(t4)
		if err != nil {
			t.Errorf("SellConversion returned error: %s", err)
		}
		fmt.Printf("sellcount = %d\n", sellcount)

	}

}
