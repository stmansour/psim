package util

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
