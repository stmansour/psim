package util

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Alphabet contains caps of the alphabet
var Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Digits contains characters for 0 - 9
var Digits = "0123456789"

// UtilData is util's data struct that is available to other modules if
// they know what they're doing.
// -----------------------------------------------------------------------------
var UtilData struct {
	Rand *rand.Rand
	mu   sync.Mutex // Protects Rand
}

// Init is the util library's initialization functio for all the really
// low level initialization that needs to be done...
//
//	INPUTS
//	randNano - the seed to use. If -1 then time.Now().UnixNano() will be used
//
//	RETURNS
//	  the randNano used to seed rand.NewSource()
//
// -----------------------------------------------------------------------------
func Init(randNano int64) int64 {
	if randNano == -1 {
		now := time.Now()
		randNano = now.UnixNano()
	}
	fmt.Printf("Random number seed:  %d\n", randNano)
	UtilData.Rand = rand.New(rand.NewSource(randNano)) // specific seed
	return randNano
}
