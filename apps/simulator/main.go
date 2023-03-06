package main

import (
	"log"
	"psim/core"
	"psim/util"
)

func main() {
	util.Init()
	cfg, err := util.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	var sim core.Simulator
	sim.Init(&cfg)
	sim.Run()
}
