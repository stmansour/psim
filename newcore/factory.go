package newcore

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/stmansour/psim/newdata"
	"github.com/stmansour/psim/util"
)

// Factory contains methods to create objects based on a DNA string
type Factory struct {
	cfg         *util.AppConfig          // system-wide configuration info
	db          *newdata.Database        // db to provide to investors
	mim         *MetricInfluencerManager //metric influencer manager describes the influencers
	MutateCalls int64                    // how many calls were made to Mutate()
	Mutations   int64                    // how many times did mutation happen
	InvCounter  int64                    // used in ID generation
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
func (f *Factory) Init(cfg *util.AppConfig, db *newdata.Database, mim *MetricInfluencerManager) {
	f.cfg = cfg
	f.db = db
	f.mim = mim
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
		fitnessSum += population[i].CalculateFitnessScore()
		for j := range population[i].Influencers {
			subclass := population[i].Influencers[j].Subclass()
			influencerFitnessSums[subclass] += population[i].Influencers[j].CalculateFitnessScore()
			influencersBySubclass[subclass] = append(influencersBySubclass[subclass], &population[i].Influencers[j])
		}
	}

	// Build the new population... Select parents, create a new Investor
	for i := 0; i < f.cfg.PopulationSize; i++ {
		idxParent1 := f.rouletteSelect(population, fitnessSum, -1) // parent 1
		var idxParent2 int
		retryLimit := 10 // Set a sensible retry limit to prevent infinite loops
		for j := 0; j < retryLimit; j++ {
			idxParent2 = f.rouletteSelect(population, fitnessSum, idxParent1) // attempt to select parent 2
			if idxParent2 != idxParent1 {
				break // We found a different parent, exit the loop
			}
			// Optional: Log or handle the case where the same index is selected
		}

		// Check if a different parent was successfully selected
		if idxParent2 == idxParent1 {
			// use desperate measures
			found := false
			for j := 0; j < len(population) && !found; j++ {
				if j != idxParent1 {
					found = true
					idxParent2 = j
				}
			}
			if !found {
				log.Panicf("Unable to select a different parent\n")
			}
		}

		newPopulation[i] = f.BreedNewInvestor(&population, idxParent1, idxParent2)
		if newPopulation[i].factory == nil {
			log.Panicf("BreedNewInvestor returned a new Investor with a nil factory\n")
		}
	}

	//-------------------------------------------
	// Check for duplicate Influencers...
	//-------------------------------------------
	// max := len(util.InfluencerSubclasses)
	// count := 0
	// for i := 0; i < len(newPopulation); i++ {
	// 	if len(newPopulation[i].Influencers) > max {
	// 		util.DPrintf("newPopulation[%d] has %d Influencers\n", i, len(newPopulation[i].Influencers))
	// 		count++
	// 	}
	// }
	// if count > 0 {
	// 	log.Panicf("Fount %d Investors with number of Influencers > %d\n", count, max)
	// }

	return newPopulation, nil
}

// GenerateInvestorID generates a unique investor id string
func (f *Factory) GenerateInvestorID() string {
	f.InvCounter++
	return fmt.Sprintf("Investor%d", f.InvCounter)
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
	newInvestor.Init(f.cfg, f, f.mim, f.db)
	newInvestor.FitnessCalculated = false
	newInvestor.Fitness = 0.0
	newInvestor.BalanceC1 = f.cfg.InitFunds
	newInvestor.ID = f.GenerateInvestorID()

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
	if util.RandomInRange(0, 1) == 0 {
		if val, ok := maps[util.RandomInRange(0, 1)]["InvW1"].(float64); ok {
			newInvestor.W1 = val
			newInvestor.W2 = 1 - val
		}
	} else {
		if val, ok := maps[util.RandomInRange(0, 1)]["InvW2"].(float64); ok {
			newInvestor.W2 = val
			newInvestor.W1 = 1 - val
		}
	}
	if val, ok := maps[util.RandomInRange(0, 1)]["Strategy"].(int); ok {
		newInvestor.Strategy = val
	}

	parent := parents[util.RandomInRange(0, 1)]
	newInfCount := len(parent.Influencers) // use the count from one of the parents
	if newInfCount == 0 {
		log.Panicf("newInfCount == 0, we cannot have an Investor with 0 Influencers\n")
	}
	newInfluencersDNA := f.createInfluencerDNAList(&parent1, &parent2, newInfCount)
	if newInfCount > len(util.InfluencerSubclasses) {
		log.Panicf("Factory.BreedNewInvestor len(newInvestor.Influencers) = %d\n", len(newInvestor.Influencers))
	}

	//------------------------------------------------------------------------------------
	// The slice newInfluencersDNA now has one entry for each Influencer we must create.
	// It also has 1 or 2 DNAs for these Influencers.  If both parents had the subclass
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
				j = 1 - j                                // alternates between 0 and 1, you have to think about this, it's a very efficient way to do this kind of a toggle
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
		inf.SetMyInvestor(&newInvestor)
		newInvestor.Influencers = append(newInvestor.Influencers, inf)
		if len(newInvestor.Influencers) > len(util.InfluencerSubclasses) {
			log.Panicf("Factory.BreedNewInvestor len(newInvestor.Influencers) = %d.  i = %d, newInfCount = %d\n", len(newInvestor.Influencers), i, newInfCount)
		}
	}

	f.Mutate(&newInvestor) // mutate only after *everything* has been set

	//----------------------------------------------------------------------------------
	// The influencer has control over its research period, however the Investor has
	// control of the "sell" point. So, no matter what T4 the Influencers need the Delta4
	// value of the Investor. Additionally, each Influencer must be initialized with
	// a pointer back to its parent, and it needs the global config data...
	//----------------------------------------------------------------------------------
	for i := 0; i < len(newInvestor.Influencers); i++ {
		newInvestor.Influencers[i].Init(&newInvestor, f.cfg)
	}

	return newInvestor
}

// createInfluencerDNAList returns a list of subclasses to be created for a newly bred
// Investor. It also includes the relevant DNA of the parents for each subclass.
func (f *Factory) createInfluencerDNAList(parent1, parent2 *Investor, n int) []InfluencerDNA {

	// Validate n
	minInfluencers := f.min(len(parent1.Influencers), len(parent2.Influencers))
	maxInfluencers := f.max(len(parent1.Influencers), len(parent2.Influencers))
	if n < minInfluencers || n > maxInfluencers {
		fmt.Printf("n must be between %d and %d; adjusting to %d\n", minInfluencers, maxInfluencers, minInfluencers)
		n = minInfluencers
	}

	infMetricMap := make(map[string]InfluencerDNA)
	for _, investor := range []*Investor{parent1, parent2} {
		for _, influencer := range investor.Influencers {
			metric := influencer.GetMetric()
			if _, exists := infMetricMap[metric]; !exists {
				infMetricMap[metric] = InfluencerDNA{Subclass: influencer.Subclass(), DNA1: influencer.DNA()}
			} else {
				// If already exists and DNA1 is filled, fill DNA2
				dna := infMetricMap[metric]
				if dna.DNA1 != "" && dna.DNA2 == "" {
					dna.DNA2 = influencer.DNA()
					infMetricMap[metric] = dna
				}
			}
		}
	}

	// Convert map to slice
	allInfluencersDNA := make([]InfluencerDNA, 0, len(infMetricMap))
	for _, dna := range infMetricMap {
		allInfluencersDNA = append(allInfluencersDNA, dna)
	}

	// Shuffle slice to randomize
	util.UtilData.Rand.Shuffle(len(allInfluencersDNA), func(i, j int) {
		allInfluencersDNA[i], allInfluencersDNA[j] = allInfluencersDNA[j], allInfluencersDNA[i]
	})

	// Select n Influencers, ensuring uniqueness
	if n > len(allInfluencersDNA) {
		n = len(allInfluencersDNA)
	}
	return allInfluencersDNA[:n]
}

func (f *Factory) min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (f *Factory) max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Mutate - there's a one percent chance that something will get completely changed.
//
//	TODO - add code that may increase or decrease the number of Influencers
//
// ------------------------------------------------------------------------------------
func (f *Factory) Mutate(inv *Investor) {
	f.MutateCalls++ // this marks another call to Mutate

	if util.RandomInRange(1, 100) > f.cfg.MutationRate {
		return
	}

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
	randomKey := keys[util.UtilData.Rand.Intn(len(keys))]
	// fmt.Printf("Random key: %s, value: %v\n", randomKey, m[randomKey])

	switch randomKey {
	case "InvW1":
		w := float64(0)
		found := false
		for !found {
			w = util.UtilData.Rand.Float64()
			found = (w != inv.W1)
		}
		inv.W1 = w
		inv.W2 = 1.0 - w
	case "InvW2":
		w := float64(0)
		found := false
		for !found {
			w = util.UtilData.Rand.Float64()
			found = (w != inv.W2)
		}
		inv.W2 = w
		inv.W1 = 1.0 - w
	case "Influencers":
		f.MutateInfluencer(inv)

	case "Strategy":
		inv.Strategy = util.UtilData.Rand.Intn(len(InvestmentStrategies))

	default:
		log.Panicf("*** PANIC ERROR *** Unhandled key from DNA: %s\n", randomKey)
	}
}

// MutateInfluencer will mutate the supplied investor by adding or removing an Influencer
// where possible.
//
// INPUTS
//
//	inv - the Investor to mutate
//
// RETURNS
//
//	nothing at this time
//
// ----------------------------------------------------------------------------------------------------
func (f *Factory) MutateInfluencer(inv *Investor) {
	//---------------------------------------------
	// 50% chance that we change influencer count
	//---------------------------------------------
	if util.RandomInRange(0, 1) == 0 {
		//------------------------------------
		// ADD or REMOVE
		//------------------------------------
		if util.RandomInRange(0, 1) == 0 { // 50% chance of adding
			//-----------------------------------------------------------------------------
			// ADD, but only if influencer count is < the number of influencer subclasses
			// and also only if the total number of influencers is < the max allowed
			//-----------------------------------------------------------------------------
			if len(inv.Influencers) < len(f.mim.InfluencerSubclasses) && len(inv.Influencers) < f.cfg.MaxInfluencers {
				//------------------------------------------------------------------
				// Randomly select a new subclass until we find one that does not
				// yet exist in the investor's influencers
				//------------------------------------------------------------------
				subclass, metric := f.RandomUnusedSubclassAndMetric(inv)
				//-----------------------------------------------------------------
				// Now that we know the subclass and metric, create it with random values...
				//-----------------------------------------------------------------
				newdna := fmt.Sprintf("{%s,Metric=%s}", subclass, metric)
				inf, err := f.NewInfluencer(newdna)
				if err != nil {
					log.Panicf("NewInfluencer(%s) returned error: %s\n", subclass, err)
				}
				inf.Init(inv, inv.cfg)
				inv.Influencers = append(inv.Influencers, inf)

			}
		} else {
			//--------------------------------------------------------------------
			// REMOVE - but only if there are more than MinInfluencers in the slice
			//--------------------------------------------------------------------
			if len(inv.Influencers) > f.cfg.MinInfluencers {
				index := util.UtilData.Rand.Intn(len(inv.Influencers))
				inv.Influencers = append(inv.Influencers[:index], inv.Influencers[index+1:]...)
			}
		}
	} else {
		//--------------------------------------------------------------------------------
		// CHANGE EXISTING
		// here we pick a random position... we'll delete the Influencer in that position
		// but we will keep its subclass info, then we'll create a new one with random
		// values to replace it.
		//--------------------------------------------------------------------------------
		idx := util.RandomInRange(0, len(inv.Influencers)-1) // pick the one to mutate
		dna := "{" + util.InfluencerSubclasses[idx] + "}"    // remember its subclass
		r, err := f.NewInfluencer(dna)                       // create a new one
		if err != nil {
			log.Panicf("*** PANIC ERROR NewInfluncer(%q) returned error: %s\n", dna, err)
		}
		r.Init(inv, inv.cfg)     // intialize it
		inv.Influencers[idx] = r // and replace it in the slot we chose randomly
	}
}

// RandomUnusedSubclassAndMetric selects a random subclass not yet present in the given Investor's Influencers.
func (f *Factory) RandomUnusedSubclassAndMetric(inv *Investor) (string, string) {
	subclass := "LSMInfluencer"
	// Map to track existing subclasses
	existingMetrics := make(map[string]bool)
	for _, influencer := range inv.Influencers {
		existingMetrics[influencer.GetMetric()] = true
	}

	// Filter the util.InfluencerSubclasses to find those not in existingMetrics
	var availableMetrics []string
	for _, subclass := range util.InfluencerSubclasses {
		if !existingMetrics[subclass] {
			availableMetrics = append(availableMetrics, subclass)
		}
	}

	// If there are no available subclasses, return an empty string
	if len(availableMetrics) == 0 {
		return subclass, ""
	}

	// Randomly select a new subclass from the available ones
	return subclass, availableMetrics[util.UtilData.Rand.Intn(len(availableMetrics))]
}

// // RandomUnusedSubclassAndMetric looks at the subclasses in the Investor's Influencers
// // and returns a randomly selected subclass that is NOT in the Investor's Influencers
// // ---------------------------------------------------------------------------------------
// func (*Factory) RandomUnusedSubclassAndMetric(inv *Investor) string {
// 	found := false
// 	index := -1
// 	for !found {
// 		index = util.UtilData.Rand.Intn(len(util.InfluencerSubclasses))
// 		for i := 0; i < len(inv.Influencers) && !found; i++ {
// 			found = (inv.Influencers[i].Subclass() == util.InfluencerSubclasses[index])
// 		}
// 	}
// 	return util.InfluencerSubclasses[index]
// }

// NewInvestorFromDNA creates a new investor from supplied DNA.
// -----------------------------------------------------------------------------
func (f *Factory) NewInvestorFromDNA(DNA string) Investor {
	m, err := f.ParseInvestorDNA(DNA)
	if err != nil {
		log.Panicf("*** PANIC ERROR *** ParseInvestorDNA returned: %s\n", err.Error())
	}

	inv := Investor{}
	if val, ok := m["InvW1"].(float64); ok {
		inv.W1 = val
	}
	if val, ok := m["InvW2"].(float64); ok {
		inv.W2 = val
	}
	if inv.W1+inv.W2 > 1.0 {
		log.Panicf("Investor Weights > 0\n")
	}
	inv.cfg = f.cfg
	inv.factory = f
	inv.db = f.db
	inv.mim = f.mim
	inv.BalanceC1 = inv.cfg.InitFunds
	inv.CreatedByDNA = true

	infDNA, ok := m["Influencers"].(string)
	if !ok {
		log.Panicf("*** PANIC ERROR *** no string available for Influencers from DNA\n")
	}
	s := infDNA[1 : len(infDNA)-1]
	sa := strings.Split(s, "|")
	for i := 0; i < len(sa); i++ {
		inf, err := f.NewInfluencer(sa[i])
		if err != nil {
			log.Panicf("*** PANIC ERROR *** NewInfluencer(%s) returned error: %s\n", sa[i], err.Error())
		}
		inf.Init(&inv, f.cfg)
		inv.Influencers = append(inv.Influencers, inf)
	}

	if inv.W1+inv.W2 > 1.0 {
		log.Panicf("Investor Weights > 0\n")
	}
	inv.ID = f.GenerateInvestorID()
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
		if len(kv) != 2 { // this should be the name of the class, "Investor"
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			investorVarMap[key] = int(val)
		} else if val, err := strconv.ParseFloat(value, 64); err == nil {
			investorVarMap[key] = float64(val)
		} else {
			if value[0] == '"' {
				value = strings.ReplaceAll(value, "\"", "")
			}
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
	metric, ok := DNAmap["Metric"].(string)
	if !ok {
		log.Panicf("Could not get a string value for Metric!\n")
	}
	if _, ok := f.mim.MInfluencerSubclasses[metric]; !ok {
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}
	Delta1, Delta2, err := f.GenerateDeltas(metric, DNAmap)
	if err != nil {
		fmt.Printf("Error generating Delta1, Delta2: %s\n", err.Error())
		return nil, err
	}

	//============================================
	// TODO:  Add mn,mx to the influencer data
	//============================================
	switch subclassName {

	case "LSMInfluencer":
		x := LSMInfluencer{
			Delta1:        Delta1,
			Delta2:        Delta2,
			HoldWindowNeg: f.mim.MInfluencerSubclasses[metric].HoldWindowNeg,
			HoldWindowPos: f.mim.MInfluencerSubclasses[metric].HoldWindowPos,
			Metric:        metric,
			cfg:           f.cfg,
		}
		minf := f.mim.MInfluencerSubclasses[metric]
		x.Blocs = minf.Blocs
		x.LocaleType = minf.LocaleType
		x.Predictor = minf.Predictor
		return &x, nil

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
		//--------------------------------------
		// ensure it's a valid subclass
		//--------------------------------------
		if i == 0 {
			found := false
			for _, v := range f.mim.InfluencerSubclasses {
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
				values[pair[0]] = strings.Trim(pair[1], "\"")
			}
		}
	}
	return tokens[0], values, nil
}

// GenerateDeltas creates values needed for Delta1 and Delta2 based
// on what was supplied in the DNA string.
//
// # The ranges for Delta1, Delta2 are read from the config information
//
// INPUTS
//
//	metric   - we need to know Influencer's metric  so that we can
//	           determine the proper bounds for Delta1 & Delta2
//
//	DNA      - the basic dna mapped into [attribute]value
//
// RETURNS
//
//	Delta1, Delta2, and Delta4
//
// --------------------------------------------------------------------------------
func (f *Factory) GenerateDeltas(metric string, DNA map[string]interface{}) (Delta1 int, Delta2 int, err error) {
	// Generate or validate Delta1
	if val, ok := DNA["Delta1"].(int); ok {
		if f.mim.MInfluencerSubclasses[metric].MinDelta1 <= val && val <= f.mim.MInfluencerSubclasses[metric].MaxDelta1 {
			Delta1 = val
		} else {
			util.DPrintf("f.mim.MInfluencerSubclasses[%s] = %#v\n", metric, f.mim.MInfluencerSubclasses[metric])
			return 0, 0, fmt.Errorf("invalid Delta1 value: %d, it must be in the range %d to %d", val, f.mim.MInfluencerSubclasses[metric].MinDelta1, f.mim.MInfluencerSubclasses[metric].MaxDelta1)
		}
	} else {
		// if no value found, generate based on configuration limits
		Delta1 = util.RandomInRange(f.mim.MInfluencerSubclasses[metric].MinDelta1, f.mim.MInfluencerSubclasses[metric].MaxDelta1)
	}

	// Generate or validate Delta2
	if val, ok := DNA["Delta2"].(int); ok {
		if f.mim.MInfluencerSubclasses[metric].MinDelta2 <= val && val <= f.mim.MInfluencerSubclasses[metric].MaxDelta2 {
			Delta2 = val
		} else {
			util.DPrintf("f.mim.MInfluencerSubclasses[%s] = %#v\n", metric, f.mim.MInfluencerSubclasses[metric])
			return 0, 0, fmt.Errorf("invalid Delta2 value: %d, it must be in the range %d to %d", val, f.mim.MInfluencerSubclasses[metric].MinDelta2, f.mim.MInfluencerSubclasses[metric].MaxDelta2)
		}
	} else {
		// if no value found, generate based on configuration limits
		Delta2 = util.RandomInRange(f.mim.MInfluencerSubclasses[metric].MinDelta2, f.mim.MInfluencerSubclasses[metric].MaxDelta2)
	}

	return Delta1, Delta2, nil
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
func (f *Factory) rouletteSelect(population []Investor, fitnessSum float64, used int) int {
	spin := util.UtilData.Rand.Float64() * fitnessSum
	runningSum := 0.0
	zeros := 0 // count the number of Investors in the population with a 0 fitness score

	for i, investor := range population {
		if i == used {
			continue
		}
		score := investor.CalculateFitnessScore()
		if score == 0 {
			zeros++
		}
		runningSum += score
		if runningSum >= spin {
			return i
		}
	}

	return len(population) - 1 // In case of rounding errors or zero fitnessSum, return the last investor
}
