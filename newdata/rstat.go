package newdata

// RollingStats maintains the last WindowSize values in an array that
// will be used to calculate the rolling mean and standard deviation
// ------------------------------------------------------------------------
type RollingStats struct {
	Window     []float64
	WindowSize int
}

// NewRollingStats returns an initialized RollingStats object
func NewRollingStats(windowSize int) *RollingStats {
	return &RollingStats{
		Window:     make([]float64, 0, windowSize),
		WindowSize: windowSize,
	}
}

// AddValue adds a new value to the rolling stats and calculates the mean and
// std dev squared if it has enough data in its rolling window.
// ----------------------------------------------------------------------------
func (rs *RollingStats) AddValue(value float64) (mean, stdDevSquared float64, statsValid bool) {
	rs.Window = append(rs.Window, value)
	if len(rs.Window) > rs.WindowSize {
		rs.Window = rs.Window[1:] // Remove the oldest value. This maintains a rolling window of size rs.WindowSize
	}

	// if we have enough data, calculate the stats
	if len(rs.Window) == rs.WindowSize {
		// First pass: calculate the sum to find the mean.
		var sum float64
		for _, v := range rs.Window {
			sum += v
		}
		mean = sum / float64(len(rs.Window))

		// Second pass: now that we have the mean, calculate the variance.
		var varianceSum float64
		for _, v := range rs.Window {
			varianceSum += (v - mean) * (v - mean)
		}
		stdDevSquared = varianceSum / float64(len(rs.Window)-1) // Use N-1 for sample standard deviation if it's a sample
		statsValid = true
	}

	return mean, stdDevSquared, statsValid
}
