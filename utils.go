package main

import (
	"math/rand"
)


func flipCoin() bool {
  	return rand.Int() % 2 == 0
}

func contains(array []string, element string) bool {
	for _, o := range array {
		if o == element {
			return true
		}
	}
	return false
}
