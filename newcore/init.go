package newcore

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// LocaleType defines how the Influencer uses locales in its prediction
const (
	LocaleNone = iota // influencer is not associated with any locale
	LocaleC1C2        // associated with locales of currencies C1 and C2 (no need to define Blocs)
	LocaleBloc        // associated with 1 or more locales listed in Blocs, prediction routines must know how to utilize
)

// LocaleTypeMap is used in unmarshalling
var LocaleTypeMap = map[string]int{
	"LocaleNone": LocaleNone,
	"LocaleC1C2": LocaleC1C2,
	"LocaleBloc": LocaleBloc,
}

// Predictor type defines how the influencer will do predictions with the metric
const (
	SingleValGT   = iota // use a single value at T1 and T2, predict "buy" for val@T1 > val@T2
	SingleValLT          // use a single value at T1 and T2, preduct "buy" for val@T1 < val@T2
	C1C2RatioGT          // use ratio, predict "buy" for valT1C1/valT1C2 > valT2C1/valT2C2
	C1C2RatioLT          // use ratio, predict "buy" for valT1C1/valT1C2 < valT2C1/valT2C2
	CustomPredict        // values for metric must be handled by a custom handler
)

// PredictorTypeMap is used in unmarshalling
var PredictorTypeMap = map[string]int{
	"SingleValGT":   SingleValGT,
	"SingleValLT":   SingleValLT,
	"C1C2RatioGT":   C1C2RatioGT,
	"C1C2RatioLT":   C1C2RatioLT,
	"CustomPredict": CustomPredict,
}

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

// MInfluencerSubclasses is a map of all recognized influencers indexed by name
var MInfluencerSubclasses = map[string]MInfluencerSubclass{}

// MInfluencerSubclassesIndexer is a list of names that can be used as keys into the map
var MInfluencerSubclassesIndexer = []string{}

// InfluencerSubclasses is a list of allowable subclasses of Influencer, the metric-specific subclasses are quasi subclasses of LSMInfluencer
var InfluencerSubclasses = []string{"LSMInfluencer"}

// Init initializes the core package
func Init() error {
	MInfluencerSubclasses = map[string]MInfluencerSubclass{}
	LoadMInfluencerSubclasses()
	return nil
}

// LoadMInfluencerSubclasses reads the definitions of Metric Influencer subclasses
// from in a table (or a CSV file) so that we don't have to create a Go
// file for every one. It loads them into the MSInfluencer
// (Metric Specific Influencer)
func LoadMInfluencerSubclasses() error {
	filename := "data/misubclasses.csv"
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return err
	}

	// Identify column indices
	colIndices := make(map[string]int)
	for i, col := range header {
		colIndices[col] = i
	}

	line := 1 // we've already read line 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line++

		inf := MInfluencerSubclass{}
		for colName, index := range colIndices {
			switch colName {
			case "Blocs":
				if len(record[index]) > 0 {
					inf.Blocs = strings.Split(record[index], ";")
				}
			case "LocaleType":
				inf.LocaleType = LocaleTypeMap[record[index]]
			case "Metric":
				inf.Metric = record[index]
			case "Name":
				inf.Name = record[index]
			case "Predictor":
				inf.Predictor = PredictorTypeMap[record[index]]
			case "Subclass":
				inf.Subclass = record[index]
			case "MinDelta1":
				inf.MinDelta1 = parseAndCheckInt(record[index], filename, line)
			case "MaxDelta1":
				inf.MaxDelta1 = parseAndCheckInt(record[index], filename, line)
			case "MinDelta2":
				inf.MinDelta2 = parseAndCheckInt(record[index], filename, line)
			case "MaxDelta2":
				inf.MaxDelta2 = parseAndCheckInt(record[index], filename, line)
			case "FitnessW1":
				inf.FitnessW1 = parseAndCheckFloat64(record[index], filename, line)
			case "FitnessW2":
				inf.FitnessW2 = parseAndCheckFloat64(record[index], filename, line)
			case "HoldWindowPos":
				inf.HoldWindowPos = parseAndCheckFloat64(record[index], filename, line)
			case "HoldWindowNeg":
				inf.HoldWindowNeg = parseAndCheckFloat64(record[index], filename, line)
			}
		}
		MInfluencerSubclasses[inf.Metric] = inf
	}

	//--------------------------------------------
	// Create an indexer for random selection
	//--------------------------------------------
	for k := range MInfluencerSubclasses {
		MInfluencerSubclassesIndexer = append(MInfluencerSubclassesIndexer, k)
	}

	return nil
}

// GetName returns the Name string or the Metric if len(Name) == 0
func (p *MInfluencerSubclass) GetName() string {
	if len(p.Name) == 0 {
		return p.Metric
	}
	return p.Name
}

func parseAndCheckInt(s, filename string, line int) int {
	var intValue int
	var err error
	if intValue, err = strconv.Atoi(s); err != nil {
		fmt.Printf("Error in %s, line %d, bad integer: %q, error: %s\n", filename, line, s, err.Error())
		os.Exit(1)
	}
	return intValue
}

func parseAndCheckFloat64(s, filename string, line int) float64 {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(s, 64); err != nil {
		fmt.Printf("Error in %s, line %d, bad floating point number: %q, error: %s\n", filename, line, s, err.Error())
		os.Exit(1)
	}
	return x
}
