package main

import (
	"testing"
)


// Tests that a Gossiper correctly creates a Status packet based on its rumors.
func TestInsertSortedNoValue(t *testing.T) {

	array := []uint32{}
	array = insertSorted(array, 3)

	if len(array) != 1 {
		t.Errorf("Array size should be %v, got %v instead.", 1, len(array))
	}
}

// Tests that a Gossiper correctly creates a Status packet based on its rumors.
func TestInsertSortedOneValueBigger(t *testing.T) {

	array := []uint32{0}
	array = insertSorted(array, 1)

	if len(array) != 2 {
		t.Errorf("Array size should be %v, got %v instead.", 2, len(array))
	}

	if !(array[0] == 0 && array[1] == 1) {
		t.Errorf("Expected array to be %v, got %v instead.", []uint{0, 1}, array)
	}
}

// Tests that a Gossiper correctly creates a Status packet based on its rumors.
func TestInsertSortedOneValueSmaller(t *testing.T) {

	array := []uint32{0}
	array = insertSorted(array, 1)

	if len(array) != 2 {
		t.Errorf("Array size should be %v, got %v instead.", 2, len(array))
	}

	if !(array[0] == 0 && array[1] == 1) {
		t.Errorf("Expected array to be %v, got %v instead.", []uint{0, 1}, array)
	}
}

// Tests that a Gossiper correctly creates a Status packet based on its rumors.
func TestInsertSortedMiddle(t *testing.T) {

	array := []uint32{0, 2}
	array = insertSorted(array, 1)

	if len(array) != 3 {
		t.Errorf("Array size should be %v, got %v instead.", 2, len(array))
	}

	if !(array[0] == 0 && array[1] == 1 && array[2] == 2) {
		t.Errorf("Expected array to be %v, got %v instead.", []uint{0, 1, 2}, array)
	}
}

// Tests that a Gossiper correctly creates a Status packet based on its rumors.
func TestInsertSortedThird(t *testing.T) {

	array := []uint32{0, 1}
	array = insertSorted(array, 2)

	if len(array) != 3 {
		t.Errorf("Array size should be %v, got %v instead.", 2, len(array))
	}

	if !(array[0] == 0 && array[1] == 1 && array[2] == 2) {
		t.Errorf("Expected array to be %v, got %v instead.", []uint{0, 1, 2}, array)
	}
}