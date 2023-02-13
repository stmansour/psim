package core

import "time"

// DiscountRate is the DR Influencer
type DiscountRate struct {
	T1 time.Time
	T2 time.Time
	T4 time.Time
}

// DRSettings defines the variables for the simulation of a DR Simulator
type DRSettings struct {
	T1max int
	T1min int
	T2max int
	T2min int
	T4min int
	T4max int
}

// DR is the struct of data used for the simulator
var DR = DRSettings{
	T1min: -30,
	T1max: -5,
	T2min: -5,
	T2max: -2,
	T4min: 1,
	T4max: 5,
}
