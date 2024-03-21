package newdata

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/stmansour/psim/util"
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

// MInfluencerSubclass is the struct that defines a metric-based influencer
type MInfluencerSubclass struct {
	MID           int     // Metric ID in the case of SQL db
	Name          string  // name of this type of influencer, if blank in the database it will be set to the Metric
	Metric        string  // data type of subclass - THIS IS THE TABLE NAME
	BlocType      int     // bloc type, only type LocaleBloc reads values from Blocs
	LocaleType    int     // how to handle locales
	Predictor     int     // which predictor to use
	Subclass      string  // What subclass is the container for this metric-influencer
	MinDelta1     int     // furthest back from t3 that t1 can be
	MaxDelta1     int     // closest to t3 that t1 can be
	MinDelta2     int     // furthest back from t3 that t2 can be
	MaxDelta2     int     // closest to t3 that t2 can be
	FitnessW1     float64 // weight for correctness
	FitnessW2     float64 // weight for activity
	HoldWindowPos float64 // positive hold area
	HoldWindowNeg float64 // negative hold area
	// Blocs         []string // list of associated countries. If associated with C1 & C2, blocs[0] must be associated with C1, blocs[1] with C2
}

// MetricInfluencerManager maintains the types and contextual info associated with
// all metric influencers
type MetricInfluencerManager struct {
	ParentDB                       *Database                      // the parent database that holds me
	MInfluencerSubclasses          map[string]MInfluencerSubclass // enables access to MInfluencer records by metric name
	MInfluencerSubclassMetricNames []string                       // a list of metric names only
	InfluencerSubclasses           []string                       // subclasses of the Influencer interface, just LSMIntluencer at this time
	initialized                    bool                           // prevents reinit
}

// NewInfluencerManager is the constructor for MetricInfluencerManager
func NewInfluencerManager() *MetricInfluencerManager {
	m := MetricInfluencerManager{}
	return &m
}

// Init initializes the MetricInfluencerManager
func (m *MetricInfluencerManager) Init(db *Database) error {
	if m.initialized {
		return nil
	}
	m.ParentDB = db
	m.MInfluencerSubclasses = map[string]MInfluencerSubclass{}
	m.LoadMInfluencerSubclasses()
	m.InfluencerSubclasses = append(m.InfluencerSubclasses, "LSMInfluencer")
	m.initialized = true
	return nil
}

// InsertMInfluencer inserts a new metric influencer into the sql table
// --------------------------------------------------------------------------------
func (p *Database) InsertMInfluencer(m *MInfluencerSubclass) error {
	switch p.Datatype {
	case "CSV":
		return fmt.Errorf("this operation is net yet supported for CSV databases")
	case "SQL":
		return p.SQLDB.InsertMInfluencerSubclass(m)
	default:
		return fmt.Errorf("unknown database type: %s", p.Datatype)
	}
}

// GetName returns the Name string or the Metric if len(Name) == 0
func (p *MInfluencerSubclass) GetName() string {
	if len(p.Name) == 0 {
		return p.Metric
	}
	return p.Name
}

// LoadMInfluencerSubclasses reads the definitions of Metric Influencer subclasses
// from in a table (or a CSV file) so that we don't have to create a Go
// file for every one. It loads them into the MSInfluencer
// (Metric Specific Influencer)
func (m *MetricInfluencerManager) LoadMInfluencerSubclasses() error {
	switch m.ParentDB.Datatype {
	case "CSV":
		return m.loadMInfluencerSubclassesCSV()
	case "SQL":
		return m.loadMInfluencerSubclassesSQL()
	default:
		return fmt.Errorf("unrecognized database type: %s", m.ParentDB.Datatype)
	}
}
func (m *MetricInfluencerManager) loadMInfluencerSubclassesSQL() error {
	// Query to select all rows from MISubclasses table
	query :=
		`SELECT MID, Name, Metric, Subclass, LocaleType, Predictor,
        MinDelta1, MaxDelta1, MinDelta2, MaxDelta2,
		FitnessW1, FitnessW2, HoldWindowPos, HoldWindowNeg FROM MISubclasses`
	rows, err := m.ParentDB.SQLDB.DB.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	subclasses := map[string]MInfluencerSubclass{}

	for rows.Next() {
		var mi MInfluencerSubclass
		var name sql.NullString // Use sql.NullString for nullable strings
		var loc, pred uint

		// Scan each row's columns into the struct
		if err := rows.Scan(&mi.MID, &name, &mi.Metric, &mi.Subclass, &loc, &pred,
			&mi.MinDelta1, &mi.MaxDelta1, &mi.MinDelta2, &mi.MaxDelta2,
			&mi.FitnessW1, &mi.FitnessW2, &mi.HoldWindowPos, &mi.HoldWindowNeg); err != nil {
			return err
		}
		mi.LocaleType = int(loc)
		mi.Predictor = int(pred)

		// Check if the Name field is NULL and set accordingly
		if name.Valid {
			mi.Name = name.String
		} else {
			mi.Name = ""
		}

		subclasses[mi.Metric] = mi
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return err
	}
	m.MInfluencerSubclasses = subclasses
	for k := range m.MInfluencerSubclasses {
		m.MInfluencerSubclassMetricNames = append(m.MInfluencerSubclassMetricNames, k)
	}

	return nil
}

func (m *MetricInfluencerManager) loadMInfluencerSubclassesCSV() error {
	var dir string
	var err error
	filename := ""
	if len(m.ParentDB.CSVDB.DBFname) > 0 {
		dir = filepath.Dir(m.ParentDB.CSVDB.DBFname)
		filename = dir + "/misubclasses.csv"
	} else {
		dir, err = util.GetExecutableDir()
		if err != nil {
			return fmt.Errorf("error getting executable directory: %s", err.Error())
		}
		filename = dir + "/data/misubclasses.csv"
	}
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
			// case "Blocs":
			// 	if len(record[index]) > 0 {
			// 		inf.Blocs = strings.Split(record[index], ";")
			// 	}
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
		m.MInfluencerSubclassMetricNames = append(m.MInfluencerSubclassMetricNames, k)
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

// InsertMInfluencerSubclass inserts a new MInfluencerSubclass into the database
func (p *DatabaseSQL) InsertMInfluencerSubclass(m *MInfluencerSubclass) error {
	query := `
INSERT INTO MISubclasses (Name, Metric, LocaleType, Predictor, Subclass, MinDelta1, MaxDelta1, MinDelta2, MaxDelta2, FitnessW1, FitnessW2, HoldWindowPos, HoldWindowNeg) 
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := p.DB.Exec(query, m.Name, m.Metric, m.LocaleType, m.Predictor, m.Subclass, m.MinDelta1, m.MaxDelta1, m.MinDelta2, m.MaxDelta2, m.FitnessW1, m.FitnessW2, m.HoldWindowPos, m.HoldWindowNeg)
	if err != nil {
		return err
	}

	return nil
}
