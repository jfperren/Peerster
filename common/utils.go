package common

import (
	"math/rand"
)

// Return 0 or 1 with 50% probability
func FlipCoin() bool {
	return rand.Int() % 2 == 0
}

// Check if an array contains an element
func Contains(array []string, element string) bool {
	for _, o := range array {
		if o == element {
			return true
		}
	}
	return false
}

// Return 1 if bool is true
func boolCount(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

// Split a budget (uint64) equally among a given number of peers and return the assignment.
func SplitBudget(budget uint64, into int) []uint64 {

	splits := make([]uint64, 0)

	if into == 0 {
		return splits
	}

	split := budget / uint64(into)

	if split < 0 {
		split = 0
	}

	remainder := budget - uint64(into) * split

	for i := 0; i < into; i++ {

		if uint64(i) < remainder {
			splits = append(splits, split + 1)
		} else {
			splits = append(splits, split)
		}
	}

	// Random permutation
	randomSplits := make([]uint64, len(splits))
	randomIndices := rand.Perm(len(splits))

	for i, v := range randomIndices {
		randomSplits[v] = splits[i]
	}

	return randomSplits
}
