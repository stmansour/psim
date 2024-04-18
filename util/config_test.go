package util

import (
	"fmt"
	"testing"
	"time"
)

func TestCalculateEndDate(t *testing.T) {
	tests := []struct {
		ending     string
		systemTime time.Time // Allow specifying the fixed time for each test
		expected   time.Time
	}{
		{
			ending:     "2023-5-31",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2023, 5, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 1m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2023, 4, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 2m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2023, 3, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 3m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 4m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 5m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 6m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 11, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 7m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 10, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 8m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 9, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 9m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 8, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 10m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 7, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 11m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 6, 30, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-5-31 - 12m",
			systemTime: time.Date(2024, 4, 17, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 5, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2024-2-29 - 1y",
			systemTime: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "2023-12-31 - 1y",
			systemTime: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			expected:   time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "today - 4d",
			systemTime: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), // Dec 15, 2023 as the fixed time
			expected:   time.Date(2023, 12, 11, 0, 0, 0, 0, time.UTC), // Dec 11, 2023
		},
		{
			ending:     "today - 2w",
			systemTime: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), // Dec 15, 2023 as the fixed time
			expected:   time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),  // Dec 1, 2023
		},
		{
			ending:     "yesterday - 6m",
			systemTime: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC), // Dec 31, 2023 as the fixed time
			expected:   time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC),  // June 30, 2023
		},
		{
			ending:     "today - 6m",
			systemTime: time.Date(2023, 12, 15, 0, 0, 0, 0, time.UTC), // Dec 15, 2023 as the fixed time
			expected:   time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),  // June 15, 2023
		},
		{
			ending:     "yesterday - 1m",
			systemTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),  // April 1, 2024, as the fixed time
			expected:   time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), // Leap year correction
		},
		{
			ending:     "yesterday - 30d",
			systemTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), // April 1, 2024, as the fixed time
			expected:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "yesterday - 2w",
			systemTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), // April 1, 2024, as the fixed time
			expected:   time.Date(2024, 3, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "today - 1m",
			systemTime: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), // April 1, 2024, as the fixed time
			expected:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ending:     "yesterday - 3m",
			systemTime: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), // April 1, 2024, as the fixed time
			expected:   time.Date(2023, 11, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		name := fmt.Sprintf("%s on %s", tc.ending, tc.systemTime.Format("2006-01-02"))
		t.Run(name, func(t *testing.T) {
			result := calculateEndDate(tc.ending, tc.systemTime)
			if !result.Equal(tc.expected) {
				t.Errorf("Test %v failed. Expected %v, got %v", name, tc.expected, result)
			}
		})
	}
}

func TestCalculateStartDate(t *testing.T) {
	tests := []struct {
		duration string
		endDate  string // Use string and compute endDate in test
		expected time.Time
	}{
		// Subtract months from May 31, 2023
		{"1m", "2023-5-31", time.Date(2023, 5, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 1m", time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 2m", time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 3m", time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 4m", time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 5m", time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 6m", time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 7m", time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 8m", time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 9m", time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 10m", time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)},
		{"1m", "2023-5-31 - 11m", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC)},

		// Additional cases for subtracting years, weeks and days
		{"3w", "2023-5-31", time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC)},
		{"10d", "2023-5-31", time.Date(2023, 5, 21, 0, 0, 0, 0, time.UTC)},
		{"1y", "2024-2-29", time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tc := range tests {
		endDate := calculateEndDate(tc.endDate, time.Now()) // Calculate end date first using calculateEndDate
		name := fmt.Sprintf("Start date for %s ending %s", tc.duration, endDate.Format("2006-01-02"))
		t.Run(name, func(t *testing.T) {
			result := calculateStartDate(tc.duration, endDate)
			if !result.Equal(tc.expected) {
				t.Errorf("%s: calculated start date %v, expected %v", name, result, tc.expected)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	Init(-1)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig failed: %s", err)
		return
	}

	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}

	// Print out the crucible date list...
	//--------------------------------------
	for i := 0; i < len(cfg.CrucibleSpans); i++ {
		fmt.Printf("%d: %s - %s\n", i, cfg.CrucibleSpans[i].DtStart.Format("2006-01-02"), cfg.CrucibleSpans[i].DtStop.Format("2006-01-02"))
	}
}
func TestSingleInvestorMode(t *testing.T) {
	Init(-1)
	cfg, err := LoadConfig("singleInvestor.json5")
	if err != nil {
		t.Errorf("LoadConfig failed: %s", err)
		return
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}
	if !cfg.SingleInvestorMode {
		t.Errorf("Single investor mode was expected to be true, but it was false")
	}
	if cfg.LoopCount != 1 || cfg.PopulationSize != 1 {
		t.Errorf("Expected LoopCount and PopulationSize to be 1.  Found LoopCount = %d, PopulationSize = %d", cfg.LoopCount, cfg.PopulationSize)
	}
}

func TestConfig(t *testing.T) {
	Init(-1)
	cfg := CreateTestingCFG()

	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}
}

func TestDateFunctions(t *testing.T) {
	Init(-1)
	dt1 := time.Date(2023, time.July, 12, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2023, time.October, 24, 0, 0, 0, 0, time.UTC)
	dt3 := time.Date(2023, time.November, 11, 0, 0, 0, 0, time.UTC)

	dt4 := time.Date(1957, time.October, 24, 0, 0, 0, 0, time.UTC)
	dt5 := time.Date(1964, time.November, 11, 0, 0, 0, 0, time.UTC)

	dt6 := time.Date(1964, time.November, 11, 0, 0, 0, 0, time.UTC)
	dt7 := time.Date(1966, time.October, 10, 9, 8, 7, 0, time.UTC)

	var sa []string
	var gold = []string{
		"0 years 3 months 12 days",
		"0 years 0 months 18 days",
		"7 years 0 months 18 days",
		"1 year 10 months 29 days",
	}
	sa = append(sa, DateDiffString(dt2, dt1))
	sa = append(sa, DateDiffString(dt2, dt3))
	sa = append(sa, DateDiffString(dt4, dt5))
	sa = append(sa, DateDiffString(dt6, dt7))

	for i := 0; i < len(sa); i++ {
		if sa[i] != gold[i] {
			t.Errorf("%d. Expected %s, got %s\n", i, gold[i], sa[i])
		}
	}
}

func TestGenerateRefNo(t *testing.T) {
	Init(-1)
	DisableConsole()
	r := GenerateRefNo()
	Console("r = %s\n", r)
	EnableConsole()
	DPrintf("r = %s.  THIS IS OUTPUT FROM A TEST\n", r)
	Console("len(r) = %d\n", len(r))
	if len(r) != 20 {
		t.Errorf("expecting len(r) to be 20, got %d\n", len(r))
	}
}

func TestRandomFunctions(t *testing.T) {
	Init(-1)
	x := RandomInRange(25, 0)
	if x < 0 || x > 25 {
		t.Errorf("Expected 0 <= x <= 25, got x = %d\n", x)
	}
}

type structA struct {
	Field1 int
	Field2 string
}

type structB struct {
	Field1 int
	Field2 string
	Field3 float64
}

func TestMigrateStruct(t *testing.T) {
	var a = structA{
		Field1: 10,
		Field2: "Hello",
	}

	var b = structB{
		Field1: 20,
		Field2: "World",
		Field3: 3.1416,
	}

	var c structA
	var d structB
	cExp := `c = util.structA{Field1:10, Field2:"Hello"}`
	dExp := `d = util.structB{Field1:0, Field2:"", Field3:0}`

	MigrateStructVals(&a, &c)
	x := fmt.Sprintf("c = %#v", c)
	if x != cExp {
		t.Errorf("Expected: %s\n", cExp)
		t.Errorf("Got:      %s\n", x)
	}
	MigrateStructVals(&b, &c)
	x = fmt.Sprintf("d = %#v", d)
	if x != dExp {
		t.Errorf("Expected: %s\n", dExp)
		t.Errorf("Got:      %s\n", x)
	}
}
