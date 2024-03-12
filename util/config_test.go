package util

import (
	"fmt"
	"testing"
	"time"
)

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

	for k, v := range cfg.SCInfo {
		fmt.Printf("Key: %s, Value: %#v\n", k, v)
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}

	cfg.InfluencerSubclasses = append(cfg.InfluencerSubclasses, "URInfluencer,")
	if err := ValidateConfig(cfg); err == nil {
		t.Errorf("ValidateConfig failed: %q is not a valid Influencer subclass", "URInfluencer,")
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

	for i := 0; i < len(cfg.InfluencerSubclasses); i++ {
		fmt.Printf("%d: %s\n", i, cfg.InfluencerSubclasses[i])
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
	DPrintf("r = %s\n", r)
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
