package util

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	Init(-1)
	cfg := CreateTestingCFG()

	for k, v := range cfg.DLimits {
		fmt.Printf("Key: %s, Value: %#v\n", k, v)
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}

}
