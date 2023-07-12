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
		l = append(l, Alphabet[RandomInRange(0, 25)])
	}
	for i := 0; i < 10; i++ {
		l = append(l, Digits[RandomInRange(0, 9)])
	}
	// move them around some random number of times
	// fmt.Printf("Initial val:  %s\n", string(l))
	swaps := 5 + RandomInRange(0, 9)
	for i := 0; i < swaps; i++ {
		j := RandomInRange(0, 9)
		k := 10 + RandomInRange(0, 9)
		l[k], l[j] = l[j], l[k]
	}
	return string(l)
}
