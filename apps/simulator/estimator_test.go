package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stmansour/psim/util"
)

func TestEstimator(t *testing.T) {
	app.randNano = -1
	app.cfg = &util.AppConfig{}
	app.cfg.LoopCount = 1
	app.cfg.Generations = 500
	app.sim.GensCompleted = 220
	app.sim.TrackingGenStop = time.Date(2024, 1, 1, 0, 0, 22, 0, time.UTC)
	app.sim.TrackingGenStart = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	totalGenerations, timePerGen, remainingDuration, estimatedCompletionTime := endTimeEstimator()

	fmt.Printf("Total generations: %d\n", totalGenerations)
	fmt.Printf("Time for last generation: %v\n", timePerGen)
	fmt.Printf("Remaining duration: %v\n", remainingDuration)
	fmt.Printf("Estimated completion time: %v\n", estimatedCompletionTime)
	if !estimatedCompletionTime.After(time.Now()) {
		t.Errorf("Expected estimated completion time to be in the future")
	}
}
