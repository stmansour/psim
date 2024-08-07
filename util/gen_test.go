package util

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseGenerationDuration(t *testing.T) {
	tests := []struct {
		input  string
		result *GenerationDuration
		fmt    string
		err    error
	}{
		{
			input:  "1 Y",
			result: &GenerationDuration{Years: 1},
			fmt:    "1 year",
		},
		{
			input:  "2 M 1 Y",
			result: &GenerationDuration{Years: 1, Months: 2},
			fmt:    "1 year 2 months",
		},
		{
			input:  "1 Y 2 M 3 W 4 D",
			result: &GenerationDuration{Years: 1, Months: 2, Weeks: 3, Days: 4},
			fmt:    "1 year 2 months 3 weeks 4 days",
		},
		{
			input: "1 Y 2 M 2 M",
			err:   errors.New("unit M is repeated"),
			fmt:   "",
		},
		{
			input: "1 Z",
			err:   errors.New("unknown unit: Z"),
			fmt:   "",
		},
		{
			input: "A Y",
			err:   errors.New("error parsing number: strconv.Atoi: parsing \"A\": invalid syntax"),
			fmt:   "",
		},
		{
			input: "1 Y 2",
			err:   errors.New("invalid input format"),
			fmt:   "",
		},
	}

	for i, test := range tests {
		res, err := ParseGenerationDuration(test.input)
		if (err != nil && test.err == nil) || (err == nil && test.err != nil) || (err != nil && test.err != nil && err.Error() != test.err.Error()) {
			t.Errorf("For input '%s', expected %v (error: %v), but got %v (error: %v)", test.input, test.result, test.err, res, err)
		}
		s := FormatGenDur(res)
		if s != tests[i].fmt {
			t.Errorf("For input '%s', expected %q,  but got %q", test.input, test.fmt, s)
		}

	}
}

func TestReadGenerationDur(t *testing.T) {
	Init(-1)
	cfg, err := LoadConfig("")
	if err != nil {
		t.Errorf("LoadConfig failed: %s", err)
		return
	}
	if len(cfg.GenDurSpec) > 0 {
		if cfg.GenDur.Years != 1 {
			t.Errorf("Expecting GenDur to show 1 year duration, found: years = %d, months = %d, weeks = %d, days = %d", cfg.GenDur.Years, cfg.GenDur.Months, cfg.GenDur.Weeks, cfg.GenDur.Days)
		}
	}

	extres, err := ReadExternalResources()
	if err != nil {
		t.Errorf("ReadExternalResources failed: %s", err)
	}
	fmt.Printf("username = %s\n", extres.DbUser)

}
