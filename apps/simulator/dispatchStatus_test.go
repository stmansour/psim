package main

import (
	"testing"
	"time"

	"github.com/stmansour/psim/util"
	"github.com/stretchr/testify/assert"
)

func TestSendDispatchStatus(t *testing.T) {
	var err error
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

	SendStatusUpdate(nil)

}
