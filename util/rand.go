package util

// RandomInRange returns a random number, r, such that:
//
//	a <= r <= b
//
// It works for positive or negative values of a and b
// made threadsafe on Mar 30, 2024
// -------------------------------------------------------
func RandomInRange(a, b int) int {
	UtilData.mu.Lock()
	defer UtilData.mu.Unlock()
	if a > b {
		a, b = b, a
	}
	return UtilData.Rand.Intn(b-a+1) + a
}
