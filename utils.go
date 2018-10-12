package main

import (
  "math/rand"
)


func selectRandom(list []string) string {
	if len(list) == 0 {
		panic("Cannot select randomly in an empty list")
	}

  	return list[rand.Int() % len(list)]
}

func flipCoin() bool {
  	return rand.Int() % 2 == 0
}

func containsString(array []string, element string) bool {
	for _, o := range array {
		if o == element {
			return true
		}
	}
	return false
}

func containsUInt32(array []uint32, element uint32) bool {
	for _, o := range array {
		if o == element {
			return true
		}
	}
	return false
}
