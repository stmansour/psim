package util

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	Init(-1)
	cfg, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig failed: %s", err)
	}
	for k, v := range cfg.DLimits {
		fmt.Println("Key:", k, "Value:", v)
	}

}
