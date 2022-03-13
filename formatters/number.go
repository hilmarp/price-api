package formatters

// GetNumbersDiff returns the absolute diff between two numbers
func GetNumbersDiff(a, b int) int {
	if a < b {
		return b - a
	}
	return a - b
}
