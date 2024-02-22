package newcore

// MInfluencerSubclass is the struct that defines a metric-based influencer
type MInfluencerSubclass struct {
	Name          string   // name of this type of influencer, if blank in the database it will be set to the Metric
	Metric        string   // data type of subclass - THIS IS THE TABLE NAME
	BlocType      int      // bloc type, only type LocaleBloc reads values from Blocs
	Blocs         []string // list of associated countries. If associated with C1 & C2, blocs[0] must be associated with C1, blocs[1] with C2
	LocaleType    int      // how to handle locales
	Predictor     int      // which predictor to use
	Subclass      string   // What subclass is the container for this metric-influencer
	MinDelta1     int      // furthest back from t3 that t1 can be
	MaxDelta1     int      // closest to t3 that t1 can be
	MinDelta2     int      // furthest back from t3 that t2 can be
	MaxDelta2     int      // closest to t3 that t2 can be
	FitnessW1     float64  // weight for correctness
	FitnessW2     float64  // weight for activity
	HoldWindowPos float64  // positive hold area
	HoldWindowNeg float64  // negative hold area
}

// GetName returns the Name string or the Metric if len(Name) == 0
func (p *MInfluencerSubclass) GetName() string {
	if len(p.Name) == 0 {
		return p.Metric
	}
	return p.Name
}
