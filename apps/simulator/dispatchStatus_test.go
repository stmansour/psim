package main

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/stmansour/psim/util"
	"github.com/stretchr/testify/assert"
)

func isProcessRunning(name string) (bool, error) {
	cmd := exec.Command("pgrep", name)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Check if the error is due to no processes found
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				// No matching process found, which is not an error in our case
				return false, nil
			}
		}
		// For any other error, return it
		return false, fmt.Errorf("error running pgrep: %w", err)
	}

	// If we get here, pgrep found at least one matching process
	return len(output) > 0, nil
}
func TestSendDispatchStatus(t *testing.T) {
	isRunning, err := isProcessRunning("dispatcher")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	assert.NoError(t, err)
	if !isRunning {
		t.Log("dispatcher not running")
		return
	}
	var cfg util.AppConfig
	app.randNano = -1
	app.DispatcherURL = "http://localhost:8250/"
	app.SID = 2
	app.MachineID, err = util.GetMachineUUID()
	assert.NoError(t, err)
	app.cfg = &cfg
	app.cfg.LoopCount = 10
	app.cfg.Generations = 100
	app.cfg.PopulationSize = 200
	app.sim.LoopsCompleted = 1
	app.sim.TrackingGenStart = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // made up date
	app.sim.TrackingGenStop = time.Date(2024, 1, 1, 0, 42, 0, 0, time.UTC) // made up date + 42 mins

	err = SendStatusUpdate(nil)
	assert.NoError(t, err)

}

func TestSendCompletion(t *testing.T) {
	isRunning, err := isProcessRunning("dispatcher")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	assert.NoError(t, err)
	if !isRunning {
		t.Log("dispatcher not running")
		return
	}
	var cfg util.AppConfig
	app.randNano = -1
	app.DispatcherURL = "http://localhost:8250/"
	app.SID = 2
	app.MachineID, err = util.GetMachineUUID()
	assert.NoError(t, err)
	app.cfg = &cfg
	app.cfg.LoopCount = 10
	app.cfg.Generations = 100
	app.cfg.PopulationSize = 200
	app.sim.LoopsCompleted = 1000
	app.sim.StopTimeSet = true
	app.sim.SimStop = time.Date(2024, 1, 2, 10, 24, 0, 0, time.UTC)

	SendStatusUpdate(&app.sim.SimStop)

}
