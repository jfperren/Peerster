package tests

import (
	"github.com/jfperren/Peerster/common"
	"testing"
)

func TestSplitsDeterministic(t *testing.T) {

    splits := common.SplitBudget(0, 1)

    if len(splits) != 1 {
		t.Errorf("SplitBudget (0, 1) has wrong length %v, should be 1", len(splits))
	}

	if splits[0] != 0 {
		t.Errorf("SplitBudget (0, 1) has wrong value %v, should be 0", splits[0])
	}

	splits = common.SplitBudget(0, 0)

	if len(splits) != 0 {
		t.Errorf("SplitBudget (0, 0) has wrong length %v, should be 0", len(splits))
	}

	splits = common.SplitBudget(6, 3)

	if len(splits) != 3 {
		t.Errorf("SplitBudget (6, 3) has wrong length %v, should be 3", len(splits))
	}

	if splits[0] != 2 {
		t.Errorf("SplitBudget (6, 3) has wrong value %v, should be 2", splits[0])
	}

	if splits[1] != 2 {
		t.Errorf("SplitBudget (6, 3) has wrong value %v, should be 2", splits[1])
	}

	if splits[2] != 2 {
		t.Errorf("SplitBudget (6, 3) has wrong value %v, should be 2", splits[2])
	}
}

func TestSplitsRandom(t *testing.T) {

	splits := common.SplitBudget(2, 4)
	results := make(map[int]map[uint64]bool)

	if len(splits) != 4 {
		t.Errorf("SplitBudget (2, 4) has wrong length %v, should be 4", len(splits))
	}

	for i, _ := range (splits) {
		results[i] = make(map[uint64]bool)
	}

	for i := 0; i < 10; i++ {
		for i, v := range(splits) {
			results[i][v] = true
		}

		splits = common.SplitBudget(2, 4)
	}

	for i, _ := range(splits) {

		if !results[i][0] {
			t.Errorf("SplitBudget (2, 4) never has value 0 at index %v", i)
		}

		if !results[i][1] {
			t.Errorf("SplitBudget (2, 4) never has value 1 at index %v", i)
		}
	}

	splits = common.SplitBudget(13, 3)
	results = make(map[int]map[uint64]bool)

	if len(splits) != 3 {
		t.Errorf("SplitBudget (13, 3) has wrong length %v, should be 3", len(splits))
	}

	for i, _ := range (splits) {
		results[i] = make(map[uint64]bool)
	}

	for i := 0; i < 10; i++ {
		for i, v := range(splits) {
			results[i][v] = true
		}

		splits = common.SplitBudget(13, 3)
	}

	for i, _ := range(splits) {

		if !results[i][4] {
			t.Errorf("SplitBudget (13, 3) never has value 4 at index %v", i)
		}

		if !results[i][5] {
			t.Errorf("SplitBudget (13, 3) never has value 5 at index %v", i)
		}
	}
}
