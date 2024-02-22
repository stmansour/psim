package newcore

import "testing"

func TestMutateAddInfluencer(t *testing.T) {
	f, _, _, cfg := createConfigAndFactory()

	// set limits on the number of influencers we can have...
	cfg.MaxInfluencers = 3
	cfg.MinInfluencers = 2

	// now create an investor with 2 Influencers.
	dna := "{Investor;InvW1=0.5000;InvW2=0.5000;Influencers=[{LSMInfluencer,Metric=LSNScore,Delta1=-45,Delta2=-5}|{LSMInfluencer,Metric=LSPScore,Delta1=-45,Delta2=-5}]}"
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

}
