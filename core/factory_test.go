package core

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/stmansour/psim/util"
)

func TestNewPopulation(t *testing.T) {
	var oldPopulationDNA = []string{
		"{Investor;Delta4=14;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-20,Delta2=-5}|{IRInfluencer,Delta1=-150,Delta2=-45}]}",
		"{Investor;Delta4=12;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-27,Delta2=-5}]}",
		"{Investor;Delta4=9;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-21,Delta2=-1}]}",
		"{Investor;Delta4=4;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-7,Delta2=-4}]}",
		"{Investor;Delta4=10;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-11,Delta2=-5}]}",
		"{Investor;Delta4=10;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-30,Delta2=-1}]}",
		"{Investor;Delta4=2;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-29,Delta2=-3}]}",
		"{Investor;Delta4=13;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-23,Delta2=-5}]}",
		"{Investor;Delta4=9;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-7,Delta2=-3}]}",
		"{Investor;Delta4=2;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-21,Delta2=-5}]}",
	}
	util.Init(-1)
	var f Factory
	cfg := util.CreateTestingCFG()
	// util.DPrintf("cfg.SCInfo[DR] = %#v\n", cfg.SCInfo["DR"])
	if err := util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	var sim Simulator
	f.Init(cfg)

	// t.Fail()

	//----------------------------
	// Build a population...
	//----------------------------
	pop := []Investor{}
	for i := 0; i < len(oldPopulationDNA); i++ {
		inv := f.NewInvestor(oldPopulationDNA[i])
		dna := inv.DNA()
		if dna != oldPopulationDNA[i] {
			t.Errorf("DNA and newDNA differ:\n\tcreated: %s\n\texpected: %s", dna, oldPopulationDNA[i])
		}
		inv.Fitness = float64(i)*1.35 - 0.35 // a random fitness score
		inv.FitnessCalculated = true
		pop = append(pop, inv)
	}

	//-----------------------------------------------------------------
	// now let's create a new population from our test population...
	//-----------------------------------------------------------------
	sim.Init(cfg, false, false)
	sim.Investors = pop   // put our population into the simulator
	sim.GensCompleted = 1 // make it appear that a simulation cycle just completed
	var err error
	if err = sim.NewPopulation(); err != nil {
		log.Panicf("*** PANIC ERROR ***  NewPopulation returned error: %s\n", err.Error())
	}

	// make sure they have Influencers
	nc := 0  // bad cfg coung
	nf := 0  // bad factory count
	ni := 0  // no MyInvestor pointer
	bd4 := 0 // bad Delta4 count
	for i := 0; i < len(sim.Investors); i++ {
		if len(sim.Investors[i].Influencers) == 0 {
			t.Errorf("No influencers for Investor[%d]\n", i)
		}
		if sim.Investors[i].cfg == nil {
			nc++
		}
		if sim.Investors[i].factory == nil {
			nf++
		}
		delta4 := sim.Investors[i].Delta4
		bd4 = 0
		for j := 0; j < len(sim.Investors[i].Influencers); j++ {
			if sim.Investors[i].Influencers[j].MyInvestor() == nil {
				ni++
			} else if sim.Investors[i].Influencers[j].MyInvestor().Delta4 != delta4 {
				bd4++
			}
		}
		// Ratio%d].FitnessCalculated = %v, .Fitness = %6.2f\n", i, sim.Investors[i].FitnessCalculated, sim.Investors[i].Fitness)
	}
	if ni > 0 {
		t.Errorf("NewPopulation return %d Influencers with nil pointer for MyInvestor", ni)
	}
	if bd4 > 0 {
		t.Errorf("NewPopulation return %d Influencers with Delta4 that did not match their Investor delta4", bd4)
	}
	if nc > 0 {
		t.Errorf("NewPopulation returned %d Investors with a nil cfg", nc)
	}
	if nf > 0 {
		t.Errorf("NewPopulation returned T%d investors with a nil factory", nf)
	}
}

func TestInvestorFromDNA(t *testing.T) {
	var f Factory
	util.Init(-1)
	cfg := util.CreateTestingCFG()
	if err := util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	f.Init(cfg)

	// t.Fail()

	parent1 := Investor{
		Delta4: 4,
		W1:     0.5,
		W2:     0.5,
	}
	dr := DRInfluencer{
		Delta1: -30,
		Delta2: -2,
	}
	ir := IRInfluencer{
		Delta1: -145,
		Delta2: -50,
	}
	dr.Init(&parent1, cfg, parent1.Delta4)
	dr.SetID()
	ir.Init(&parent1, cfg, parent1.Delta4)
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
	}
	ur2 := URInfluencer{
		Delta1: -150,
		Delta2: -45,
	}
	dr2.Init(&parent2, cfg, parent2.Delta4)
	dr2.SetID()
	ur2.Init(&parent2, cfg, parent2.Delta4)
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
			"{invVar1=YesIDo;invVar2=34;Influencers=[{subclass1,var1=NotAtAll,var2=1.0}|{subclass2,var1=2,var2=2.0}];invVar3=3.1416}",
			map[string]interface{}{
				"invVar1":     "YesIDo",
				"invVar2":     34,
				"Influencers": "[{subclass1,var1=NotAtAll,var2=1.0}|{subclass2,var1=2,var2=2.0}]",
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
			t.Errorf("parseInvestorDNA(%q) map = %v, want %v", tt.input, gotMap, tt.wantMap)
		}
	}
}

func TestMutation(t *testing.T) {
	var err error
	util.Init(-1)
	var f Factory
	cfg := util.CreateTestingCFG()
	cfg.PopulationSize = 1000
	if err = util.ValidateConfig(cfg); err != nil {
		log.Panicf("*** PANIC ERROR ***  ValidateConfig returned error: %s\n", err)
	}
	var sim Simulator
	f.Init(cfg)
	sim.Init(cfg, false, false)

	//-----------------------------------------------------------------
	// Give them some transactions and stats to affect next gen
	//-----------------------------------------------------------------
	for i := 0; i < len(sim.Investors); i++ {
		sim.Investors[i].BalanceC1 += float64(util.RandomInRange(1, 1000))/1000.00 - 500.00 // random result
		if util.UtilData.Rand.Float64() > 0.7 {                                             // add an investment 30% of the time
			cm := util.RandomInRange(1, 100) > 25 // true 75% of the time
			pr := util.RandomInRange(1, 100) > 30 // true 70% of the time
			inv := Investment{
				Completed:  cm,
				Profitable: pr,
			}
			sim.Investors[i].Investments = append(sim.Investors[i].Investments, inv)
			//----------------------------------------------
			// add info for the Influencer's Predictions...
			//----------------------------------------------
			for j := 0; j < len(sim.Investors[i].Influencers); j++ {
				ps := sim.Investors[i].Influencers[j].GetMyPredictions()
				cm = util.RandomInRange(1, 100) > 25 // true 75% of the time
				pr = util.RandomInRange(1, 100) > 30 // true 70% of the time
				prd := Prediction{
					Correct:   pr,
					Completed: cm,
				}
				ps = append(ps, prd)
				sim.Investors[i].Influencers[j].SetMyPredictions(ps)
			}
		}
	}

	//-----------------------------------------------------------------
	// now let's create a new population from our test population...
	//-----------------------------------------------------------------
	sim.GensCompleted = 2 // force it to think this
	sim.CalculateMaxVals()
	sim.CalculateAllFitnessScores()
	if err = sim.NewPopulation(); err != nil {
		log.Panicf("*** PANIC ERROR *** NewPopulation returned error: %s\n", err)
	}

	//-------------------------------------------
	// Check for too many Influencers...
	//-------------------------------------------
	max := len(util.InfluencerSubclasses)
	for i := 0; i < len(sim.Investors); i++ {
		if len(sim.Investors[i].Influencers) > max {
			t.Errorf("sim.Investor[%d] has %d Influencers\n", i, len(sim.Investors[i].Influencers))
		}
	}

}
