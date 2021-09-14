package mathutil

// Max returns the larger of a and b.
func Max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
