package newcore

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Convenience function
func createConfigAndFactory() (*Factory, *newdata.Database, *util.AppConfig) {
	var f Factory
	util.Init(-1)
	cfg := util.CreateTestingCFG()
	if err := util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	db, err := newdata.NewDatabase("CSV", cfg, nil)
	if err != nil {
		log.Panicf("*** PANIC ERROR ***  NewDatabase returned error: %s\n", err)
	}
	if err := db.Open(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Panicf("Could not get current working directory: %s\n", err.Error())
	}
	db.SetCSVFilename(dir + "/data/platodb.csv")
	if err := db.Init(); err != nil {
		log.Panicf("*** PANIC ERROR ***  db.Init returned error: %s\n", err)
	}
	f.Init(cfg, db)
	return &f, db, cfg
}

// TestInfluencerPredictions
func TestInfluencerPredictions(t *testing.T) {
	f, db, cfg := createConfigAndFactory()
	inv := Investor{}
	inv.Init(cfg, f, db)

	// create one of each Influencer
	for _, v := range db.Mim.MInfluencerSubclasses {
		// fmt.Printf("k = %v, v = %v\n", k, v)
		dna := fmt.Sprintf("{%s,Metric=%s}", v.Subclass, v.Metric)
		inf, err := f.NewInfluencer(dna)
		if err != nil {
			t.Errorf("Error creating new influencer: %s\n", err.Error())
		}
		inf.SetMyInvestor(&inv)
		inv.Influencers = append(inv.Influencers, inf)
	}

	// Ask for a prediction on a particular date to
	// ensure that every Influencer can access the data
	// it needs.
	dt := time.Date(2023, time.July, 15, 12, 0, 0, 0, time.UTC)
	for _, inf := range inv.Influencers {
		pred, err := inf.GetPrediction(dt)
		if err != nil {
			t.Errorf("Error from influencer(%s) generating prediction: %s\n", inf.GetMetric(), err.Error())
		}
		fmt.Printf("Prediction from %s: action: %s, data(%6.2f,%6.2f), probability: %6.1f,  weight: %6.1f\n", inf.GetMetric(), pred.Action, pred.Val1, pred.Val2, pred.Probability, pred.Weight)
	}
}

// TestNewInfestorFromDNA - create Investors from DNA
func TestNewInvestorFromDNA(t *testing.T) {
	dnas := []string{
		"{Investor;Strategy=DistributedDecision;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Delta1=-77,Delta2=-14,Metric=WTOILClose}|{LSMInfluencer,Delta1=-83,Delta2=-13,Metric=GCAM_C5_4}|{LSMInfluencer,Delta1=-63,Delta2=-21,Metric=GCAM_C15_148_ECON}|{LSMInfluencer,Delta1=-77,Delta2=-10,Metric=GCAM_C15_147}|{LSMInfluencer,Delta1=-95,Delta2=-30,Metric=IR}|{LSMInfluencer,Delta1=-149,Delta2=-26,Metric=GCAM_C16_60}|{LSMInfluencer,Delta1=-132,Delta2=-9,Metric=GCAM_C3_2_ECON}|{LSMInfluencer,Delta1=-163,Delta2=-35,Metric=M2}|{LSMInfluencer,Delta1=-249,Delta2=-3,Metric=GCAM_C16_121}|{LSMInfluencer,Delta1=-104,Delta2=-42,Metric=BC}]}",
		"{Investor;Strategy=DistributedDecision;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Delta1=-63,Delta2=-23,Metric=GCAM_C16_47_ECON}]}",
		"{Investor;Strategy=DistributedDecision;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Delta1=-80,Delta2=-24,Metric=GCAM_C16_60_ECON}|{LSMInfluencer,Delta1=-242,Delta2=-5,Metric=GCAM_C16_121_ECON}]}",
		"{Investor;Strategy=DistributedDecision;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Delta1=-138,Delta2=-15,Metric=GCAM_C16_121}|{LSMInfluencer,Delta1=-71,Delta2=-5,Metric=GCAM_C15_147}|{LSMInfluencer,Delta1=-166,Delta2=-44,Metric=BC}]}",
		"{Investor;Strategy=DistributedDecision;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Delta1=-132,Delta2=-27,Metric=GCAM_C3_1}|{LSMInfluencer,Delta1=-144,Delta2=-14,Metric=GCAM_C15_148}|{LSMInfluencer,Delta1=-137,Delta2=-16,Metric=GCAM_C16_60}|{LSMInfluencer,Delta1=-147,Delta2=-6,Metric=GCAM_C15_137}|{LSMInfluencer,Delta1=-159,Delta2=-16,Metric=GCAM_C3_1}|{LSMInfluencer,Delta1=-84,Delta2=-16,Metric=GCAM_C15_204}|{LSMInfluencer,Delta1=-167,Delta2=-43,Metric=M2}|{LSMInfluencer,Delta1=-97,Delta2=-34,Metric=CC}]}",
	}
	f, db, cfg := createConfigAndFactory()
	for i := 0; i < len(dnas); i++ {
		inv := f.NewInvestorFromDNA(dnas[i])
		inv.Init(cfg, f, db)
		fmt.Printf("InvestorID = %s\n", inv.ID)
		li := len(inv.Influencers)
		for j := 0; j < li; j++ {
			metric := inv.Influencers[j].GetMetric()
			fmt.Printf("\t%d. %s - %s\n", j, inv.Influencers[j].Subclass(), metric)
		}
		if cfg.MinInfluencers > li || cfg.MaxInfluencers < li {
			t.Errorf("BreedNewInvestor returned an investor with %d Influencers. Config states MinInfluencers = %d and MaxInfluencers = %d", li, cfg.MinInfluencers, cfg.MaxInfluencers)
		}
	}
}

// TestNewPopulation - test generating a new population
func TestNewPopulation(t *testing.T) {
	f, db, cfg := createConfigAndFactory()

	Investors := make([]Investor, 0)
	for i := 0; i < 50; i++ {
		var v Investor
		v.ID = f.GenerateInvestorID()
		Investors = append(Investors, v)
	}
	// Now initialize them all
	for i := 0; i < len(Investors); i++ {
		Investors[i].Init(cfg, f, db)
		fmt.Printf("%s\n", Investors[i].DNA())
	}
}

// TestInvestorFromParents - test generating a new investor from two parent investors
func TestInvestorFromParents(t *testing.T) {
	f, _, cfg := createConfigAndFactory()
	// t.Fail()

	parent1 := Investor{Strategy: 1, W1: 0.5, W2: 0.5}
	dr := LSMInfluencer{Metric: "WTOILClose", Delta1: -145, Delta2: -2}
	ir := LSMInfluencer{Metric: "SP", Delta1: -89, Delta2: -5}
	dr.Init(&parent1, cfg)
	dr.SetID()
	ir.Init(&parent1, cfg)
	ir.SetID()
	parent1.Influencers = append(parent1.Influencers, &dr, &ir)

	parent2 := Investor{Strategy: 0, W1: 0.5, W2: 0.5}
	dr2 := LSMInfluencer{Metric: "CC", Delta1: -95, Delta2: -30}
	ur2 := LSMInfluencer{Metric: "DR", Delta1: -89, Delta2: -1}
	dr2.Init(&parent2, cfg)
	dr2.SetID()
	ur2.Init(&parent2, cfg)
	ur2.SetID()
	parent2.Influencers = append(parent2.Influencers, &dr2, &ur2)
	population := []Investor{}
	population = append(population, parent1, parent2)

	investor := f.BreedNewInvestor(&population, 0, 1)
	if len(investor.Influencers) == 0 {
		t.Errorf("BreedNewInvestor returned an investor with no Influencers")
	}
}

// TestParseInvestorDNA - verify that the parser can correctly parse n Investor DNA string
// ------------------------------------------------------------------------------------------
func TestParseInvestorDNA(t *testing.T) {
	var f Factory
	var tests = []struct {
		input   string
		wantMap map[string]interface{}
	}{
		{
			"{Investor;invVar1=YesIDo;invVar2=34;Influencers=[{subclass1,metric=\"WTOILClose\",var1=NotAtAll,var2=1.0}|{subclass2,var1=2,var2=2.0}];invVar3=3.1416}",
			map[string]interface{}{
				"invVar1":     "YesIDo",
				"invVar2":     34,
				"Influencers": "[{subclass1,metric=\"WTOILClose\",var1=NotAtAll,var2=1.0}|{subclass2,var1=2,var2=2.0}]",
				"invVar3":     float64(3.1416),
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
			for k, v := range tt.wantMap {
				fmt.Printf("want: k = %s, v = %v    found: %v\n", k, v, gotMap[k])
			}
			t.Errorf("parseInvestorDNA(%q) map = %v, want %v", tt.input, gotMap, tt.wantMap)
		}
	}
}
