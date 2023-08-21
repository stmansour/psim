package util

import (
	"errors"
	"testing"
)

func TestParseGenerationDuration(t *testing.T) {
	tests := []struct {
		input  string
		result *GenerationDuration
		err    error
	}{
		{
			input:  "1 Y",
			result: &GenerationDuration{Years: 1},
		},
		{
			input:  "2 M 1 Y",
			result: &GenerationDuration{Years: 1, Months: 2},
		},
		{
			input:  "1 Y 2 M 3 W 4 D",
			result: &GenerationDuration{Years: 1, Months: 2, Weeks: 3, Days: 4},
		},
		{
			input: "1 Y 2 M 2 M",
			err:   errors.New("unit M is repeated"),
		},
		{
			input: "1 Z",
			err:   errors.New("unknown unit: Z"),
		},
		{
			input: "A Y",
			err:   errors.New("error parsing number: strconv.Atoi: parsing \"A\": invalid syntax"),
		},
		{
			input: "1 Y 2",
			err:   errors.New("invalid input format"),
		},
	}

	for _, test := range tests {
		res, err := ParseGenerationDuration(test.input)
		if (err != nil && test.err == nil) || (err == nil && test.err != nil) || (err != nil && test.err != nil && err.Error() != test.err.Error()) {
			t.Errorf("For input '%s', expected %v (error: %v), but got %v (error: %v)", test.input, test.result, test.err, res, err)
		}
	}
}

func TestReadGenerationDur(t *testing.T) {
	Init(-1)
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %s", err)
		return
	}
	if len(cfg.GenDurSpec) > 0 {
		if cfg.GenDur.Years != 1 {
			t.Errorf("Expecting GenDur to show 1 year duration, found: years = %d, months = %d, weeks = %d, days = %d", cfg.GenDur.Years, cfg.GenDur.Months, cfg.GenDur.Weeks, cfg.GenDur.Days)
		}
	}

}
