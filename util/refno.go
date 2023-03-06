package util

import (
	"math/rand"
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
}

// Init is the util library's initialization functio for all the really
// low level initialization that needs to be done...
// -----------------------------------------------------------------------------
func Init() {
	now := time.Now()
	UtilData.Rand = rand.New(rand.NewSource(now.UnixNano())) // specific seed
	rand.Seed(time.Now().UnixNano() + 42)                    // general seed
}

// GenerateRefNo generate a unique identifier for a transaction. This is
// a really simple implementation. Should be rewritten if we intend to use
// it commercially.
//
// INPUTS:
//
// RETURNS:
//
//	the id string
//
// -----------------------------------------------------------------------------
func GenerateRefNo() string {
	var l []byte

	// Generate 10 random digits and 5 random letters
	for i := 0; i < 10; i++ {
		l = append(l, Alphabet[UtilData.Rand.Intn(26)])
	}
	for i := 0; i < 10; i++ {
		l = append(l, Digits[UtilData.Rand.Intn(10)])
	}
	// move them around some random number of times
	// fmt.Printf("Initial val:  %s\n", string(l))
	swaps := 5 + UtilData.Rand.Intn(10)
	for i := 0; i < swaps; i++ {
		j := UtilData.Rand.Intn(10)
		k := 10 + UtilData.Rand.Intn(10)
		l[k], l[j] = l[j], l[k]
	}
	return string(l)
}

// func main() {
// 	now := time.Now()
// 	fmt.Printf("now = %s\n", now.Format("2006-01-02 15:04:00 MST"))
// 	fmt.Printf("now.UnixNano() = %d\n", now.UnixNano())
// 	UtilData.Rand = rand.New(rand.NewSource(now.UnixNano()))
// 	fmt.Println(GenerateRefNo())
// }
