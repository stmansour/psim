package util

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	Init(-1)
	cfg := CreateTestingCFG()

	for k, v := range cfg.SCInfo {
		fmt.Printf("Key: %s, Value: %#v\n", k, v)
	}
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}

}
func TestLoadConfig(t *testing.T) {
	Init(-1)
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %s", err)
		return
	}

	if err := ValidateConfig(&cfg); err != nil {
		t.Errorf("ValidateConfig failed: %s", err)
	}

	for i := 0; i < len(cfg.InfluencerSubclasses); i++ {
		fmt.Printf("%d: %s\n", i, cfg.InfluencerSubclasses[i])
	}

}
