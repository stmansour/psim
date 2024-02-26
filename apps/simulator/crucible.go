package main

import (
	"github.com/stmansour/psim/util"
)

// Crucible is the class that implements and reports on a crucible... a
// series of simulations on a list of Investor DNA
// ---------------------------------------------------------------------------
type Crucible struct {
	cfg *util.AppConfig
	// file *os.File
}

// NewCrucible returns a pointer to a new crucible object
func NewCrucible() *Crucible {
	c := Crucible{}
	return &c
}

// Init initializes the crucible object
func (c *Crucible) Init(cfg *util.AppConfig) {
	c.cfg = cfg
}

// DumpResults sends the crucible report to a csv file
func (c *Crucible) DumpResults() {

}
