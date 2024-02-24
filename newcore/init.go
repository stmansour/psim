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
	SingleValGT   = iota // 0. use a single value at T1 and T2, predict "buy" for val@T1 > val@T2
	SingleValLT          // 1. use a single value at T1 and T2, preduct "buy" for val@T1 < val@T2
	C1C2RatioGT          // 2. use ratio, predict "buy" for valT1C1/valT1C2 > valT2C1/valT2C2
	C1C2RatioLT          // 3. use ratio, predict "buy" for valT1C1/valT1C2 < valT2C1/valT2C2
	CustomPredict        // 4. values for metric must be handled by a custom handler
)

// PredictorTypeMap is used in unmarshalling
var PredictorTypeMap = map[string]int{
	"SingleValGT":   SingleValGT,
	"SingleValLT":   SingleValLT,
	"C1C2RatioGT":   C1C2RatioGT,
	"C1C2RatioLT":   C1C2RatioLT,
	"CustomPredict": CustomPredict,
}

// MetricInfluencerManager maintains the types and contextual info associated with
// all metric influencers
type MetricInfluencerManager struct {
	MInfluencerSubclasses        map[string]MInfluencerSubclass
	MInfluencerSubclassesIndexer []string
	InfluencerSubclasses         []string
}

// NewInfluencerManager is the constructor for MetricInfluencerManager
func NewInfluencerManager() *MetricInfluencerManager {
	m := MetricInfluencerManager{}
	return &m
}

// Init initializes the MetricInfluencerManager
func (m *MetricInfluencerManager) Init() error {
	m.MInfluencerSubclasses = map[string]MInfluencerSubclass{}
	m.LoadMInfluencerSubclasses()
	m.InfluencerSubclasses = append(m.InfluencerSubclasses, "LSMInfluencer")
	return nil
}

// LoadMInfluencerSubclasses reads the definitions of Metric Influencer subclasses
// from in a table (or a CSV file) so that we don't have to create a Go
// file for every one. It loads them into the MSInfluencer
// (Metric Specific Influencer)
func (m *MetricInfluencerManager) LoadMInfluencerSubclasses() error {
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
				inf.MinDelta1 = m.parseAndCheckInt(record[index], filename, line)
			case "MaxDelta1":
				inf.MaxDelta1 = m.parseAndCheckInt(record[index], filename, line)
			case "MinDelta2":
				inf.MinDelta2 = m.parseAndCheckInt(record[index], filename, line)
			case "MaxDelta2":
				inf.MaxDelta2 = m.parseAndCheckInt(record[index], filename, line)
			case "FitnessW1":
				inf.FitnessW1 = m.parseAndCheckFloat64(record[index], filename, line)
			case "FitnessW2":
				inf.FitnessW2 = m.parseAndCheckFloat64(record[index], filename, line)
			case "HoldWindowPos":
				inf.HoldWindowPos = m.parseAndCheckFloat64(record[index], filename, line)
			case "HoldWindowNeg":
				inf.HoldWindowNeg = m.parseAndCheckFloat64(record[index], filename, line)
			}
		}
		m.MInfluencerSubclasses[inf.Metric] = inf
	}

	//--------------------------------------------
	// Create an indexer for random selection
	//--------------------------------------------
	for k := range m.MInfluencerSubclasses {
		m.MInfluencerSubclassesIndexer = append(m.MInfluencerSubclassesIndexer, k)
	}

	return nil
}

func (m *MetricInfluencerManager) parseAndCheckInt(s, filename string, line int) int {
	var intValue int
	var err error
	if intValue, err = strconv.Atoi(s); err != nil {
		fmt.Printf("Error in %s, line %d, bad integer: %q, error: %s\n", filename, line, s, err.Error())
		os.Exit(1)
	}
	return intValue
}

func (m *MetricInfluencerManager) parseAndCheckFloat64(s, filename string, line int) float64 {
	var x float64
	var err error
	if x, err = strconv.ParseFloat(s, 64); err != nil {
		fmt.Printf("Error in %s, line %d, bad floating point number: %q, error: %s\n", filename, line, s, err.Error())
		os.Exit(1)
	}
	return x
}
