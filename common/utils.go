package common

import (
	"math/rand"
)

func FlipCoin() bool {
	return rand.Int()%2 == 0
}

func Contains(array []string, element string) bool {
	for _, o := range array {
		if o == element {
			return true
		}
	}
	return false
}

func SplitBudget(budget int64, into int) []int64 {

	splits := make([]int64, 0)

	if into == 0 {
		return splits
	}

	split := budget / int64(into)

	if split < 0 {
		split = 0
	}

	remainder := budget - int64(into) * split

	for i := 0; i < into; i++ {

		if int64(i) < remainder {
			splits = append(splits, split + 1)
		} else {
			splits = append(splits, split)
		}
	}

	// Random permutation
	randomSplits := make([]int64, len(splits))
	randomIndices := rand.Perm(len(splits))

	for i, v := range randomIndices {
		randomSplits[v] = splits[i]
	}

	return randomSplits
}