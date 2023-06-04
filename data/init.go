package data

import "github.com/stmansour/psim/util"

// DInfo maintains data needed by the data subsystem.
// The primary need is the two currencies C1 & C2
type DInfo struct {
	cfg *util.AppConfig
}

// Init calls the initialize routine for all data types
// ------------------------------------------------------------
func Init(cfg *util.AppConfig) {
	DRInit()
	ERInit()
}
