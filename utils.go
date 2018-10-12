package main

import (
	"math"
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

func insertAt(array []uint32, index int, elem uint32) []uint32 {
	array = append(array, 0)
	copy(array[index+1:], array[index:])
	array[index] = elem
	return array
}

func insertSorted(array []uint32, elem uint32) []uint32 {

	if len(array) == 0 {
		return append(array, elem)
	}

	low := 0
	high := len(array)

	for low != high {

		mid := int(math.Floor((float64(high - low) / 2)))
		midElem := array[mid]

		switch {
		case elem < midElem:
			high = mid
		case elem > midElem:
			low = mid + 1
		case elem == midElem:
			panic("ID should not already be in the table")
		}
	}

	return insertAt(array, low, elem)
}