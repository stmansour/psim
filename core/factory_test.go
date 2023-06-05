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
		"{Investor;Delta4=14;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-20,Delta2=-5,Delta4=2}|{IRInfluencer,Delta1=-17,Delta2=-3,Delta4=4}]}",
		"{Investor;Delta4=12;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-27,Delta2=-5,Delta4=6}]}",
		"{Investor;Delta4=9;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-21,Delta2=-1,Delta4=5}]}",
		"{Investor;Delta4=4;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-7,Delta2=-4,Delta4=7}]}",
		"{Investor;Delta4=10;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-11,Delta2=-5,Delta4=3}]}",
		"{Investor;Delta4=10;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-30,Delta2=-1,Delta4=3}]}",
		"{Investor;Delta4=2;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-29,Delta2=-3,Delta4=9}]}",
		"{Investor;Delta4=13;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-23,Delta2=-5,Delta4=8}]}",
		"{Investor;Delta4=9;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-4,Delta2=-3,Delta4=10}]}",
		"{Investor;Delta4=2;InvW1=0.5000;InvW2=0.5000;Influencers=[{DRInfluencer,Delta1=-21,Delta2=-5,Delta4=3}]}",
	}
	util.Init()
	var f Factory
	cfg := util.CreateTestingCFG()
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
	nc := 0
	nf := 0
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
		// USDJPRDRRatio%d].FitnessCalculated = %v, .Fitness = %6.2f\n", i, sim.Investors[i].FitnessCalculated, sim.Investors[i].Fitness)
	}
}

func TestInvestorFromDNA(t *testing.T) {
	var f Factory
	util.Init()
	f.Init(util.CreateTestingCFG())

	// t.Fail()

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
