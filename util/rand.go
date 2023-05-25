package util

import (
	"math/rand"
)

// RandomInRange returns a random number, r, such that:
//
//	a <= r <= b
//
// It works for positive or negative values of a and b
// -------------------------------------------------------
func RandomInRange(a, b int) int {
	//  rand.Seed(time.Now().UnixNano())
	if a > b {
		a, b = b, a
	}
	return rand.Intn(b-a+1) + a
}
