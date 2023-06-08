package core

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/stmansour/psim/util"
)

// InfluencerSubclasses is an array of strings with all the subclasses of
// Influencer that the factory knows how to create.
// ---------------------------------------------------------------------------
var InfluencerSubclasses = []string{
	"DRInfluencer",
	"URInfluencer",
	"IRInfluencer",
}

// Factory contains methods to create objects based on a DNA string
type Factory struct {
	cfg         *util.AppConfig // system-wide configuration info
	MutateCalls int64           // how many calls were made to Mutate()
	Mutations   int64           // how many times did mutation happen
}

// InfluencerDNA is a struct of information used during the process of
// crossover when creating a new Investor based on the results from a
// simulation cycle.
// -----------------------------------------------------------------------
type InfluencerDNA struct {
	Subclass string // Influencer subclass
	DNA1     string // DNA from one parent
	DNA2     string // DNA from the other parent if it exists
}

// Init - initializes the factory
//
// --------------------------------------------------------------------------------
func (f *Factory) Init(cfg *util.AppConfig) {
	f.cfg = cfg
}

// NewPopulation creates a new population based on the current population
// and their fitness scores.
//
// INPUT
//
//		population - current population of Investors, it is assumed that this
//	              population has just completed a simulation cycle.
//
//		cfg        - the app configuration file
//
// RETURN
//
//	[]Investor - the population that just finished its simulation cycle
//	error any error encountered
//
// -------------------------------------------------------------------------
func (f *Factory) NewPopulation(population []Investor) ([]Investor, error) {
	if len(population) < 2 {
		return nil, errors.New("population size must be at least 2")
	}

	newPopulation := make([]Investor, f.cfg.PopulationSize)
	fitnessSum := float64(0.0)                              // used by rouletteSelect
	influencerFitnessSums := make(map[string]float64)       // stores the fitness sum of each Influencer subclass
	influencersBySubclass := make(map[string][]*Influencer) // stores pointers to each Influencer of each subclass

	for i := 0; i < len(population); i++ {
		fitnessSum += population[i].FitnessScore()
		for j := range population[i].Influencers {
			subclass := population[i].Influencers[j].Subclass()
			influencerFitnessSums[subclass] += population[i].Influencers[j].FitnessScore()
			influencersBySubclass[subclass] = append(influencersBySubclass[subclass], &population[i].Influencers[j])
		}
	}

	// Build the new population... Select parents, create a new Investor
	for i := 0; i < f.cfg.PopulationSize; i++ {
		idxParent1 := f.rouletteSelect(population, fitnessSum) // parent 1
		idxParent2 := f.rouletteSelect(population, fitnessSum) // parent 2

		// ensure idxParent2 is different from idxParent1
		dbgCounter := 0
		for idxParent2 == idxParent1 {
			idxParent2 = f.rouletteSelect(population, fitnessSum)
			dbgCounter++
			if dbgCounter > 2 {
				util.DPrintf("population len = %d\n", len(population))
				for j := 0; j < len(population); j++ {
					util.DPrintf("population[%d].DNA() = %s, FitnessCalculated = %v, Fitness = %6.2f\n", j, population[j].DNA(), population[j].FitnessCalculated, population[j].Fitness)
				}
				log.Panicf("Looks like we're stuck in the loop\n")
			}
		}

		newPopulation[i] = f.BreedNewInvestor(&population, idxParent1, idxParent2)
		if newPopulation[i].factory == nil {
			log.Panicf("BreedNewInvestor returned a new Investor with a nil factory\n")
		}
	}

	return newPopulation, nil
}

// BreedNewInvestor creates a new Investor by going through the genetic
// algorithm. It also creates the Investor's Influencers.  Here's how
// we choose the next generation Influencers for the new Investor.
//
//			  Investor DNA is of the form:
//	          Delta4=5;Influencers=[{subclass,var1=val1,var2=val2,...}|{subclass,var1=val1,var2=val2,...}|...]
//
// INPUT
//
//	population - current population of investors
//
// RETURN
//
//	new population
//	any error encountered
//
// -------------------------------------------------------------------------
func (f *Factory) BreedNewInvestor(population *[]Investor, idxParent1, idxParent2 int) Investor {
	newInvestor := Investor{
		CreatedByDNA: true,
	}
	newInvestor.Init(f.cfg, f)
	newInvestor.FitnessCalculated = false
	newInvestor.Fitness = 0.0
	newInvestor.BalanceC1 = f.cfg.InitFunds

	parent1 := (*population)[idxParent1]
	parent2 := (*population)[idxParent2]

	parents := []Investor{parent1, parent2}

	map1, err := f.ParseInvestorDNA(parent1.DNA())
	if err != nil {
		fmt.Printf("*** ERROR *** ParseInvestorDNA returned:  %s\n", err)
		return newInvestor
	}
	map2, err := f.ParseInvestorDNA(parent2.DNA())
	if err != nil {
		fmt.Printf("*** ERROR *** ParseInvestorDNA returned:  %s\n", err)
		return newInvestor
	}

	maps := []map[string]interface{}{
		map1,
		map2,
	}

	//-----------------------------------------------------------------
	// Randomly choose one of the parents and copy its DNA value...
	//-----------------------------------------------------------------
	if val, ok := maps[util.RandomInRange(0, 1)]["Delta4"].(int); ok {
		newInvestor.Delta4 = val
	}
	if val, ok := maps[util.RandomInRange(0, 1)]["InvW1"].(float64); ok {
		newInvestor.W1 = val
	}
	if val, ok := maps[util.RandomInRange(0, 1)]["InvW2"].(float64); ok {
		newInvestor.W2 = val
	}

	//-----------------------------------------------------------------------
	// Determine the count of new influencers as a random number between
	// the counts of parent influencers
	//-----------------------------------------------------------------------
	p1Influencers := parent1.Influencers
	p2Influencers := parent2.Influencers

	//-----------------------------------------------------------------------
	// To generate new Investor's Influencers, we use a random, parental-based
	// approach:
	//
	// Influencers Count: Randomly choose the number of Influencers from
	//     either parent. For example, if Parent1 has 3 and Parent2 has 2
	//     Influencers, the offspring gets 2 or 3 Influencers.
	//
	// Influencer Types: Select Influencer types based on the frequency in
	//     both parents. The first selected type is removed from subsequent
	//     selection to avoid duplicate Influencer subclasses in the new Investor.
	//
	// DNA Configuration: Several strategies can be employed:
	//     a) Use the parents' Influencer DNA when both have the same Influencer
	//        type. Treat single occurrences as dominant.
	//     b) Apply roulette selection from the population for DNA selection,
	//        favoring more successful Influencers.
	//     c) Utilize parents' DNA when possible, and resort to roulette selection
	//        from population for single occurrences.
	//     d) Generate a random DNA string where needed.
	//
	// Strategies (a) and (c) seem most viable. (b) is simple but potentially
	//     suboptimal. (d) might be superfluous due to the mutation phase.
	//
	// For now, I'm going with option (a). We can make a configuration
	// option to flip it to (c) if we do not get the result's we're seeking
	// with (a).
	//-----------------------------------------------------------------------
	var parentInfluencers []Influencer
	for _, influencer := range append(p1Influencers, p2Influencers...) {
		parentInfluencers = append(parentInfluencers, influencer)
	}

	//-------------------------------------------------------------------
	// Select influencers based on what the parents had.
	// Build a list of InfluencerDNA structs that we'll create next
	//-------------------------------------------------------------------
	newInfluencersDNA := []InfluencerDNA{} // we're going to pick our Influencers now...
	newInfCount := len(parents[util.RandomInRange(0, 1)].Influencers)
	for i := 0; i < newInfCount && len(parentInfluencers) > 0; i++ {
		idx := rand.Intn(len(parentInfluencers)) // select a random subclass
		newInfluencerDNA := InfluencerDNA{
			Subclass: parentInfluencers[idx].Subclass(),
			DNA1:     parentInfluencers[idx].DNA(),
		}
		selectedInfluencer := parentInfluencers[idx]

		//-----------------------------------------------------------------
		// Remove all occurrences of the selected subclass from the list
		// because the Investor can only have one instance of a particular
		// subclass
		//-----------------------------------------------------------------
		var tmp []Influencer
		for _, inf := range parentInfluencers {
			if inf.Subclass() != selectedInfluencer.Subclass() {
				tmp = append(tmp, inf) // we keep it if it's not the same subclass
			} else {
				if inf.GetID() != selectedInfluencer.GetID() {
					newInfluencerDNA.DNA2 = inf.DNA() // save the second DNA if encountered it
				}
			}
		}
		parentInfluencers = tmp
		newInfluencersDNA = append(newInfluencersDNA, newInfluencerDNA)
	}

	//------------------------------------------------------------------------------------
	// The slice newInfluencersDNA now has one entry for each Influencer we must create.
	// It also has 1 or 2 DNAs for these Influences.  If both parents had the subclass
	// then the entry will have 2 DNAs otherwise it will have one.
	// Spin through the list and create the new Influencers.  If there are 2 DNAs, do a
	// crossover.  Otherwise, we'll just use the one DNA we have and assume it's dominant
	//------------------------------------------------------------------------------------
	for i := 0; i < len(newInfluencersDNA); i++ {
		dna1 := newInfluencersDNA[i].DNA1
		subclass, map1, err := f.ParseInfluencerDNA(dna1)
		if err != nil {
			log.Panicf("*** PANIC ERROR ***  BreedNewInvestor:  Error parsing Influencer DNA1 = %s : %s\n", dna1, err.Error())
		}
		dna2 := newInfluencersDNA[i].DNA2

		dna := ""
		if len(dna2) > 0 {
			//-------------------------------------
			// We have 2 DNA strands to crossover.
			//-------------------------------------
			_, map2, err := f.ParseInfluencerDNA(dna2)
			if err != nil {
				log.Panicf("BreedNewInvestor:  Error parsing Influencer DNA2 = %s : %s\n", dna2, err.Error())
			}
			m := []map[string]interface{}{map1, map2}
			//--------------------------------------------------------------------
			// build a new DNA string that is a crossover blend of dna1 and dna2
			//--------------------------------------------------------------------
			dna = "{" + subclass + "," // this will be the dna of the new Influencer
			j := 0
			for k := range map1 {
				dna += fmt.Sprintf("%s=%v,", k, m[j][k]) // first time through it gets map1[k], next time map2[k], next time map1[k]...
				j = 1 - j                                // alternates between 0 and 1
			}
			dna = dna[:len(dna)-1] // remove the trailing comma
			dna += "}"
		} else {
			//------------------------------------------------------------------------------
			// We only have 1 DNA strand. Assume it dominant and make the new Influencer...
			//------------------------------------------------------------------------------
			dna = dna1
		}
		//-----------------------------------------------------------
		// Create the new Influencer and add it to newInvestor...
		//-----------------------------------------------------------
		inf, err := f.NewInfluencer(dna)
		if err != nil {
			log.Panicf("*** PANIC ERROR ***  BreedNewInvestor:  Error from NewInfluencer(%s) : %s\n", dna, err.Error())
		}
		newInvestor.Influencers = append(newInvestor.Influencers, inf)
	}

	//----------------------------------------------------------------------------------
	// The influencer has control over its research period, however the Investor has
	// control of the "sell" point. So, no matter what T4 the Influencers need the Delta4
	// value of the Investor. Additionally, each Influencer must be initialized with
	// a pointer back to its parent, and it needs the global config data...
	//----------------------------------------------------------------------------------
	for i := 0; i < len(newInvestor.Influencers); i++ {
		newInvestor.Influencers[i].Init(&newInvestor, f.cfg, newInvestor.Delta4)
	}

	f.Mutate(&newInvestor) // mutate only after *everything* has been set

	return newInvestor
}

// Mutate - there's a one percent chance that something will get completely changed
// ------------------------------------------------------------------------------------
func (f *Factory) Mutate(inv *Investor) {
	f.MutateCalls++ // this marks another call to Mutate

	// if util.RandomInRange(1, 100) != 1 {
	// 	return
	// }

	f.Mutations++ // if we hit this point, we're going to mutate
	dna := inv.DNA()
	m, err := f.ParseInvestorDNA(dna)
	if err != nil {
		log.Panicf("*** PANIC ERROR ***  Mutate:  Error from ParseInvestorDNA(%s) : %s\n", dna, err.Error())
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	randomKey := keys[rand.Intn(len(keys))]
	// fmt.Printf("Random key: %s, value: %v\n", randomKey, m[randomKey])
	switch randomKey {
	case "Delta4":
		d := 0
		found := false
		for !found {
			d = util.RandomInRange(f.cfg.MinDelta4, f.cfg.MaxDelta4)
			found = (d == inv.Delta4)
		}
		inv.Delta4 = d
		//-------------------------------------------------------------------------------------
		// if we mutate Delta4, we have to change ALL influencers to use the new Delta4 value
		//-------------------------------------------------------------------------------------
		for i := 0; i < len(inv.Influencers); i++ {
			inv.Influencers[i].SetDelta4(inv.Delta4)
		}
	case "InvW1":
		w := float64(0)
		found := false
		for !found {
			w = rand.Float64()
			found = (w != inv.W1)
		}
		inv.W1 = w
		inv.W2 = 1.0 - w
	case "InvW2":
		w := float64(0)
		found := false
		for !found {
			w = rand.Float64()
			found = (w != inv.W2)
		}
		inv.W2 = w
		inv.W1 = 1.0 - w
	case "Influencers":
		idx := util.RandomInRange(0, len(InfluencerSubclasses)-1)
		dna := fmt.Sprintf("{%s}", InfluencerSubclasses[idx])
		idx = 0
		if len(inv.Influencers) > 1 {
			idx = util.RandomInRange(0, len(inv.Influencers)-1)
		}
		r, err := f.NewInfluencer(dna)
		if err != nil {
			log.Panicf("*** PANIC ERROR NewInfluncer(%q) returned error: %s\n", dna, err)
		}
		r.Init(inv, inv.cfg, inv.Delta4)
		inv.Influencers[idx] = r
	default:
		log.Panicf("*** PANIC ERROR *** Unhandled key from DNA: %s\n", randomKey)
	}
}

// NewInvestor creates a new investor from supplied DNA
// -----------------------------------------------------------------------------
func (f *Factory) NewInvestor(DNA string) Investor {
	m, err := f.ParseInvestorDNA(DNA)
	if err != nil {
		log.Panicf("*** PANIC ERROR *** ParseInvestorDNA returned: %s\n", err.Error())
	}

	inv := Investor{}
	if val, ok := m["Delta4"].(int); ok {
		if val >= f.cfg.MinDelta4 && val <= f.cfg.MaxDelta4 {
			inv.Delta4 = val
		} else {
			log.Panicf("*** PANIC ERROR ***invalid Delta4 value: %d, it must be in the range %d to %d\n", val, f.cfg.MinDelta4, f.cfg.MaxDelta4)
		}
	} else {
		inv.Delta4 = util.RandomInRange(f.cfg.MinDelta4, f.cfg.MaxDelta4)
	}
	if val, ok := m["InvW1"].(float64); ok {
		inv.W1 = val
	}
	if val, ok := m["InvW2"].(float64); ok {
		inv.W2 = val
	}
	inv.cfg = f.cfg
	inv.factory = f
	inv.BalanceC1 = inv.cfg.InitFunds

	infDNA, ok := m["Influencers"].(string)
	if !ok {
		log.Panicf("*** PANIC ERROR *** no string available for Influencers from DNA\n")
	}
	s := infDNA[1 : len(infDNA)-1]
	sa := strings.Split(s, "|")
	for i := 0; i < len(sa); i++ {
		inf, err := f.NewInfluencer(sa[i])
		if err != nil {
			log.Panicf("*** PANIC ERROR *** NeInfluencer(%s) returned error: %s\n", sa[i], err.Error())
		}
		inf.Init(&inv, f.cfg, inv.Delta4)
		inv.Influencers = append(inv.Influencers, inf)
	}

	return inv
}

// ParseInvestorDNA parses an Investor DNA string. out the list of Influencers and returns
//
// The format of a DNA string:
//
//	"{invVar1=YesIDo;invVar2=34;Influencers=[{subclass1,var1=NotAtAll,var2=1.0}|{subclass2,var1=2,var2=2.0}];invVar3=3.1416}"
//
// We use commas to separate the Influencer variables, we use semicolons to
// separate the Investor variables. This just simplifies the parsing.
//
// RETURNS
//
//	a map of interface{} values indexed by variable name
//	any error found, nil if no errors
//
// --------------------------------------------------------------------------------
func (f *Factory) ParseInvestorDNA(DNA string) (map[string]interface{}, error) {
	DNA = strings.TrimSpace(DNA)
	if len(DNA) < 2 || DNA[0] != '{' || DNA[len(DNA)-1] != '}' {
		return nil, fmt.Errorf("invalid DNA format")
	}
	DNA = DNA[1 : len(DNA)-1] // Remove the braces
	investorVarMap := make(map[string]interface{})
	parts := strings.Split(DNA, ";")

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			investorVarMap[key] = int(val)
		} else if val, err := strconv.ParseFloat(value, 64); err == nil {
			investorVarMap[key] = float64(val)
		} else {
			investorVarMap[key] = value
		}
	}

	return investorVarMap, nil
}

// NewInfluencer creates and returns a new influencer of the specified subclass.
// It can initialize the objects either randomly or from supplied DNA
//
// INPUTS
//
//	 DNA - a string with configuration information.
//			  It will be parsed for its subclass, then it is used as the DNA.
//			  Investor DNA is of the form:
//	          Delta4=5;Influencers=[{subclass,var1=val1,var2=val2,...}|{subclass,var1=val1,var2=val2,...}|...]
//	       The subclass is required. Other values will be randomly generated
//	       if they are not present.
//
// RETURNS
//
//	a pointer to the Influencer object (*Influencer)
//	any error encountered
//
// --------------------------------------------------------------------------------
func (f *Factory) NewInfluencer(DNA string) (Influencer, error) {
	subclassName, DNAmap, err := f.ParseInfluencerDNA(DNA)
	if err != nil {
		return nil, err
	}

	Delta1, Delta2, Delta4, err := f.GenerateDeltas(DNAmap)
	if err != nil {
		return nil, err
	}

	switch subclassName {
	case "DRInfluencer":
		dri := DRInfluencer{
			Delta1: Delta1,
			Delta2: Delta2,
			Delta4: Delta4,
			cfg:    f.cfg,
		}
		return &dri, nil
	case "IRInfluencer":
		iri := IRInfluencer{
			Delta1: Delta1,
			Delta2: Delta2,
			Delta4: Delta4,
			cfg:    f.cfg,
		}
		return &iri, nil
	case "URInfluencer":
		uri := URInfluencer{
			Delta1: Delta1,
			Delta2: Delta2,
			Delta4: Delta4,
			cfg:    f.cfg,
		}
		return &uri, nil
	default:
		return nil, errors.New("unknown subclass")
	}
}

// ParseInfluencerDNA does what you think
//
// The format of a DNA string:
//
//	{subclass,var1=val1,var2=val2,...}
//
// RETURNS
//
//	a string representing the subclass
//	a map of interface{} values indexed by variable name
//	any error found, nil if no errors
//
// --------------------------------------------------------------------------------
func (f *Factory) ParseInfluencerDNA(DNA string) (string, map[string]interface{}, error) {
	values := make(map[string]interface{})
	DNA = strings.Trim(DNA, "{}")
	tokens := strings.Split(DNA, ",")
	for i, token := range tokens {
		if i == 0 {
			found := false
			for _, v := range InfluencerSubclasses {
				if v == token {
					found = true
					break
				}
			}
			if !found {
				return "", nil, fmt.Errorf("unknown subclass: %s", token)
			}
		} else {
			pair := strings.Split(token, "=")
			if len(pair) != 2 {
				return "", nil, fmt.Errorf("invalid variable assignment: %s", token)
			}
			if val, err := strconv.ParseInt(pair[1], 10, 64); err == nil {
				values[pair[0]] = int(val)
			} else if val, err := strconv.ParseFloat(pair[1], 64); err == nil {
				values[pair[0]] = float64(val)
			} else {
				values[pair[0]] = pair[1]
			}
		}
	}
	return tokens[0], values, nil
}

// GenerateDeltas creates values needed for Delta1, Delta2, and Delta4 based
// on what was supplied in the DNA string.
//
// # The ranges for Delta1, Delta2, and Delta4 are read from the config information
//
// RETURNS
//
//	Delta1, Delta2, and Delta4
//
// --------------------------------------------------------------------------------
func (f *Factory) GenerateDeltas(DNA map[string]interface{}) (Delta1 int, Delta2 int, Delta4 int, err error) {
	// Generate or validate Delta1
	if val, ok := DNA["Delta1"].(int); ok {

		if val >= f.cfg.MinDelta1 && val <= f.cfg.MaxDelta1 {
			Delta1 = val
		} else {
			return 0, 0, 0, fmt.Errorf("invalid Delta1 value: %d, it must be in the range %d to %d", val, f.cfg.MinDelta1, f.cfg.MaxDelta2)
		}
	} else {
		Delta1 = rand.Intn(27) - 30 // -30 to -4
	}

	// Generate or validate Delta2
	if val, ok := DNA["Delta2"].(int); ok {
		if val > Delta1 && val <= f.cfg.MaxDelta2 && val >= f.cfg.MinDelta2 {
			Delta2 = val
		} else {
			mn := f.cfg.MinDelta2 // assume the min value is the configured lower limit
			if Delta1 >= mn {     // if this is true, then Delta1's range can overlap Delta2
				//  MinDelta2       mn              MaxDelta2
				//     |---------|--+-------------------|
				//             Delta1
				mn = Delta1 + 1            // see if we can move the minDelta2 value for this object to 1 greater than Delta1
				if mn <= f.cfg.MaxDelta2 { // as long as mn is <= MaxDelta2, we're OK
					Delta2 = mn
				} else {
				}
			} else {
				log.Panicf("Not exactly sure what to do here, val = %d\n", val)
			}
		}
	} else {
		for {
			Delta2 = util.RandomInRange(f.cfg.MinDelta2, f.cfg.MaxDelta2)
			if Delta2 > Delta1 {
				break // if Delta2 is after Delta1, we're done. Otherwise we just keep trying
			}
		}
	}

	// Generate or validate Delta4
	if val, ok := DNA["Delta4"].(int); ok {
		if val >= 1 && val <= 14 {
			Delta4 = val
		} else {
			return 0, 0, 0, fmt.Errorf("invalid Delta4 value: %d, it must be in the range 1 to 14", val)
		}
	} else {
		Delta4 = util.RandomInRange(f.cfg.MinDelta4, f.cfg.MaxDelta4)
	}

	return Delta1, Delta2, Delta4, nil
}

// rouletteSelect selects a single parent from the population, using
// "roulette wheel selection" method.
//
// Here is the detailed process:
// Calculate the sum of all fitness values in the population. This sum
// represents the entire "wheel" in the roulette analogy. Each individual
// in the population is then a "section" of the wheel whose section size
// is proportional to its fitness.
//
// Generate a random number in the range [0, SumOfFitnessScores]
// This number represents the "spin" of the roulette wheel.
//
// Iterate over the population and sum the fitness values from the start
// to each individual (the cumulative sum). Whenever this
// sum is >= to the random number generated in step 2,
// return that individual.
//
// INPUTS
//
//	population = the old population
//
// RETURN
//
//	index of the investor selected
//
// -----------------------------------------------------------------------------
func (f *Factory) rouletteSelect(population []Investor, fitnessSum float64) int {
	spin := rand.Float64() * fitnessSum
	runningSum := 0.0

	for i, investor := range population {
		runningSum += investor.FitnessScore()
		if runningSum >= spin {
			return i
		}
	}

	return len(population) - 1 // In case of rounding errors or zero fitnessSum, return the last investor
}
