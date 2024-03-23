package newcore

import (
	"fmt"
	"testing"
)

func TestMutateInfluencer(t *testing.T) {
	f, _, cfg := createConfigAndFactory()

	// set limits on the number of influencers we can have...
	cfg.MaxInfluencers = 3
	cfg.MinInfluencers = 2

	// now create an investor with 2 Influencers.
	dna := "{Investor;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Metric=LSNScore,Delta1=-65,Delta2=-5}|{LSMInfluencer,Metric=LSNScore,Delta1=-65,Delta2=-5}}]}"
	inv := f.NewInvestorFromDNA(dna)

	// we should have 2 influencers now
	if len(inv.Influencers) != 2 {
		t.Errorf("Wrong number of influencers created from dna, expected 2, found %d\n", len(inv.Influencers))
	}

	// Now add an influencer
	f.doMutateInfluencer(&inv, 0)
	if len(inv.Influencers) != 3 {
		t.Errorf("doMutateInfluencer, mutation=0, did not add an influencer\n")
	}

	// Now try to add another Influencer... this would make it total 4, but
	// we set cfg.MaxInfluencers to 3, so it should stay at 3
	f.doMutateInfluencer(&inv, 0)
	if len(inv.Influencers) != 3 {
		t.Errorf("doMutateInfluencer, mutation=0, increased the influencer count above the max\n")
	}

	// Now delete an influencer
	f.doMutateInfluencer(&inv, 1)
	if len(inv.Influencers) != 2 {
		t.Errorf("doMutateInfluencer, mutation=1, did not reduce the influencer count\n")
	}

	// Now try to delete another influencer, this would put the total below the minimum
	f.doMutateInfluencer(&inv, 1)
	if len(inv.Influencers) != 2 {
		t.Errorf("doMutateInfluencer, mutation=1, reduced the influencer count below the min\n")
	}

	// make a list of the current Metric Influencer types...
	fmt.Printf("Influencers prior to mutation:")
	mm := map[string]bool{}
	for _, j := range inv.Influencers {
		mm[j.GetMetric()] = true
		fmt.Printf("  {%s,Metric=%s}", j.Subclass(), j.GetMetric())
	}
	fmt.Printf("\n")
	// Now mutate one of those influencers
	f.doMutateInfluencer(&inv, 2)
	// first, make sure we haven't increased or decreased the influencer count
	if len(inv.Influencers) != 2 {
		t.Errorf("doMutateInfluencer, mutation=2, has changed the total number of influencers rather than changed an existing one\n")
	}
	// now verify that there is exactly 1 change...
	fmt.Printf("After doMutateInfluencer(inv,2): ")
	n := 0
	for _, j := range inv.Influencers {
		if _, ok := mm[j.GetMetric()]; !ok {
			n++ // count the change
		}
		fmt.Printf("  {%s,Metric=%s}", j.Subclass(), j.GetMetric())
	}
	fmt.Printf("\n")
	if n != 1 {
		t.Errorf("doMutateInfluencer, mutation=2, changed more than %d Influencers\n", n)
	}

}
