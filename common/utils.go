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
