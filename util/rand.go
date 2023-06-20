package util

// RandomInRange returns a random number, r, such that:
//
//	a <= r <= b
//
// It works for positive or negative values of a and b
// -------------------------------------------------------
func RandomInRange(a, b int) int {
	if a > b {
		a, b = b, a
	}
	return UtilData.Rand.Intn(b-a+1) + a
}
