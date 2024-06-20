package newcore

import (
	"fmt"
	"testing"

	"github.com/stmansour/psim/util"
)

func TestInfluencerSorting(t *testing.T) {
	f, db, cfg := createConfigAndFactory()

	dnalist := []string{
		"{Investor;ID=Investor_13999eaa-bf28-407a-bf0f-1e61244783e1;Strategy=MajorityRules;InvW1=0.2674;InvW2=0.7326;Influencers=[{LSMInfluencer,Delta1=-256,Delta2=-45,Metric=GCAM_C5_4}|{LSMInfluencer,Delta1=-332,Delta2=-40,Metric=GCAM_V42_2}|{LSMInfluencer,Delta1=-90,Delta2=-47,Metric=GCAM_C4_16}|{LSMInfluencer,Delta1=-255,Delta2=-39,Metric=GCAM_C15_147}|{LSMInfluencer,Delta1=-207,Delta2=-18,Metric=GCAM_V42_3}|{LSMInfluencer,Delta1=-347,Delta2=-54,Metric=GCAM_C25_6}|{LSMInfluencer,Delta1=-75,Delta2=-53,Metric=GCAM_C25_2}|{LSMInfluencer,Delta1=-190,Delta2=-31,Metric=GCAM_C15_204}|{LSMInfluencer,Delta1=-632,Delta2=-1,Metric=HeatingOil}|{LSMInfluencer,Delta1=-183,Delta2=-42,Metric=Silver}|{LSMInfluencer,Delta1=-285,Delta2=-48,Metric=GCAM_C4_24}|{LSMInfluencer,Delta1=-136,Delta2=-56,Metric=GCAM_V42_4}]}",
		"{Investor;ID=Investor_13999eaa-bf28-407a-bf0f-1e61244783e1;Strategy=MajorityRules;InvW1=0.2674;InvW2=0.7326;Influencers=[{LSMInfluencer,Delta1=-256,Delta2=-45,Metric=GCAM_C5_4}|{LSMInfluencer,Delta1=-332,Delta2=-40,Metric=GCAM_V42_2}|{LSMInfluencer,Delta1=-90,Delta2=-47,Metric=GCAM_C4_16}|{LSMInfluencer,Delta1=-255,Delta2=-39,Metric=GCAM_C15_147}|{LSMInfluencer,Delta1=-207,Delta2=-18,Metric=GCAM_V42_3}|{LSMInfluencer,Delta1=-347,Delta2=-54,Metric=GCAM_C25_6}|{LSMInfluencer,Delta1=-75,Delta2=-53,Metric=GCAM_C25_2}|{LSMInfluencer,Delta1=-190,Delta2=-31,Metric=GCAM_C15_204}|{LSMInfluencer,Delta1=-632,Delta2=-1,Metric=HeatingOil}|{LSMInfluencer,Delta1=-183,Delta2=-42,Metric=Silver}|{LSMInfluencer,Delta1=-136,Delta2=-56,Metric=GCAM_V42_4}|{LSMInfluencer,Delta1=-285,Delta2=-48,Metric=GCAM_C4_24}]}",
		"{Investor;ID=Investor_13999eaa-bf28-407a-bf0f-1e61244783e1;Strategy=MajorityRules;InvW1=0.2674;InvW2=0.7326;Influencers=[{LSMInfluencer,Delta1=-256,Delta2=-45,Metric=GCAM_C5_4}|{LSMInfluencer,Delta1=-136,Delta2=-56,Metric=GCAM_V42_4}|{LSMInfluencer,Delta1=-332,Delta2=-40,Metric=GCAM_V42_2}|{LSMInfluencer,Delta1=-90,Delta2=-47,Metric=GCAM_C4_16}|{LSMInfluencer,Delta1=-255,Delta2=-39,Metric=GCAM_C15_147}|{LSMInfluencer,Delta1=-207,Delta2=-18,Metric=GCAM_V42_3}|{LSMInfluencer,Delta1=-347,Delta2=-54,Metric=GCAM_C25_6}|{LSMInfluencer,Delta1=-75,Delta2=-53,Metric=GCAM_C25_2}|{LSMInfluencer,Delta1=-190,Delta2=-31,Metric=GCAM_C15_204}|{LSMInfluencer,Delta1=-632,Delta2=-1,Metric=HeatingOil}|{LSMInfluencer,Delta1=-183,Delta2=-42,Metric=Silver}|{LSMInfluencer,Delta1=-285,Delta2=-48,Metric=GCAM_C4_24}]}",
	}

	fmt.Printf("DNA initial hash values\n")
	for _, dna := range dnalist {
		fmt.Printf("\t%s\n", util.HashDNA(dna))
	}

	var s Simulator
	s.Init(cfg, db, nil, false, false)
	s.Investors = make([]Investor, 0)

	for _, dna := range dnalist {
		i := f.NewInvestorFromDNA(dna)
		s.Investors = append(s.Investors, i)
	}

	s.SortInvestors()

	fmt.Printf("DNA hash values after sort\n")

	x := ""
	for _, v := range s.Investors {
		fmt.Printf("%s\n", v.DNA())
		if len(x) == 0 {
			x = util.HashDNA(v.DNA())
			fmt.Printf("\t%s  length: %d\n", x, len(x))
		} else {
			x1 := util.HashDNA(v.DNA())
			fmt.Printf("\t%s  length: %d\n", x1, len(x1))
			if x != x1 {
				t.Errorf("ERROR: Hashes don't match: %s != %s\n", x, x1)
			}
		}
	}
}
