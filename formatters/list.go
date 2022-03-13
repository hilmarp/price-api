package formatters

import "strings"

// IsInStringList checks if a string is in list, has to match exactly
func IsInStringList(str string, lst []string) bool {
	for _, s := range lst {
		if s == str {
			return true
		}
	}

	return false
}

// ListItemContainsString checks if an item in list contains a string, not exact match
func ListItemContainsString(str string, lst []string) bool {
	for _, s := range lst {
		if strings.Contains(s, str) {
			return true
		}
	}

	return false
}

// ChunkWorkers splits worker list up
func ChunkWorkers(workers []func(), maxRunning int) [][]func() {
	var chunked [][]func()

	var j int
	for i := 0; i < len(workers); i += maxRunning {
		j += maxRunning
		if j > len(workers) {
			j = len(workers)
		}
		chunked = append(chunked, workers[i:j])
	}

	return chunked
}
