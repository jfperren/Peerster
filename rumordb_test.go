package main

import (
	"testing"
)

func TestRumorsPutSameOrigin(t *testing.T) {

	rumors := MakeRumorDatabase()

	rumors.Put(&RumorMessage{"A", 0, "Hello"})
	rumors.Put(&RumorMessage{"A", 1, "Hi"})

	if len(rumors.IDs) != 1   {
		t.Errorf("Expected length of rumor IDs to be %v, go %v instead.", 1, len(rumors.rumors))
	}

	if len(rumors.rumors) != 1 {
		t.Errorf("Expected length of Rumors to be %v, go %v instead.", 1, len(rumors.IDs))
	}

	if !(rumors.IDs["A"][0] == 0 && rumors.IDs["A"][1] == 1) {
		t.Errorf("Expected IDs['A'] to be %v, go %v instead.", []uint{0, 1}, rumors.IDs["A"])
	}

	if !(rumors.rumors["A"][0].ID == 0 && rumors.rumors["A"][0].Origin == "A" && rumors.rumors["A"][0].Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.rumors["A"][0])
	}

	if !(rumors.rumors["A"][1].ID == 1 && rumors.rumors["A"][1].Origin == "A" && rumors.rumors["A"][1].Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.rumors["A"][1])
	}
}

func TestRumorsPutTwoOrigins(t *testing.T) {

	rumors := MakeRumorDatabase()

	rumors.Put(&RumorMessage{"A", 0, "Hello"})
	rumors.Put(&RumorMessage{"B", 1, "Hi"})

	if len(rumors.IDs) != 2   {
		t.Errorf("Expected length of rumor IDs to be %v, go %v instead.", 2, len(rumors.rumors))
	}

	if len(rumors.rumors) != 2 {
		t.Errorf("Expected length of Rumors to be %v, go %v instead.", 2, len(rumors.IDs))
	}

	if !(rumors.IDs["A"][0] == 0) {
		t.Errorf("Expected IDs['A'] to be %v, go %v instead.", []uint{0}, rumors.IDs["A"])
	}

	if !(rumors.IDs["B"][0] == 1) {
		t.Errorf("Expected IDs['B'] to be %v, go %v instead.", []uint{1}, rumors.IDs["B"])
	}

	if !(rumors.rumors["A"][0].ID == 0 && rumors.rumors["A"][0].Origin == "A" && rumors.rumors["A"][0].Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.rumors["A"][0])
	}

	if !(rumors.rumors["B"][1].ID == 1 && rumors.rumors["B"][1].Origin == "B" && rumors.rumors["B"][1].Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.rumors["A"][1])
	}
}

func TestRumorsGetSameOrigin(t *testing.T) {

	rumors := MakeRumorDatabase()

	rumors.Put(&RumorMessage{"A", 0, "Hello"})
	rumors.Put(&RumorMessage{"A", 1, "Hi"})

	firstRumor := rumors.Get("A", 0)
	secondRumor := rumors.Get("A", 1)

	if !(firstRumor.ID == 0 && firstRumor.Origin == "A" && firstRumor.Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.rumors["A"][0])
	}

	if !(secondRumor.ID == 1 && secondRumor.Origin == "A" && secondRumor.Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.rumors["A"][1])
	}
}

func TestRumorsGetTwoOrigins(t *testing.T) {

	rumors := MakeRumorDatabase()

	rumors.Put(&RumorMessage{"A", 0, "Hello"})
	rumors.Put(&RumorMessage{"B", 1, "Hi"})

	firstRumor := rumors.Get("A", 0)
	secondRumor := rumors.Get("B", 1)

	if !(firstRumor.ID == 0 && firstRumor.Origin == "A" && firstRumor.Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.rumors["A"][0])
	}

	if !(secondRumor.ID == 1 && secondRumor.Origin == "B" && secondRumor.Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.rumors["A"][1])
	}
}

func TestRumorsGetInexistant(t *testing.T) {

	rumors := MakeRumorDatabase()

	rumors.Put(&RumorMessage{"A", 0, "Hello"})
	rumors.Put(&RumorMessage{"B", 1, "Hi"})

	firstRumor := rumors.Get("A", 1)
	secondRumor := rumors.Get("B", 0)

	if firstRumor != nil {
		t.Errorf("Should not return any rumor at Rumors['A'][1], got %v.", firstRumor)
	}

	if secondRumor != nil {
		t.Errorf("Should not return any rumor at Rumors['B'][0], got %v.", secondRumor)
	}
}

func TestRumorsContains(t *testing.T) {

	rumors := MakeRumorDatabase()

	firstRumor := &RumorMessage{"A", 0, "Hello"}
	secondRumor := &RumorMessage{"B", 1, "Hi"}
	thirdRumor := &RumorMessage{"B", 2, "Hey"}
	fourthRumor := &RumorMessage{"C", 0, "Greetings"}

	rumors.Put(firstRumor)
	rumors.Put(secondRumor)

	if !rumors.Contains(firstRumor) {
		t.Errorf("Should contain ('A', 0)")
	}

	if !rumors.Contains(secondRumor) {
		t.Errorf("Should contain ('B', 1)")
	}

	if rumors.Contains(thirdRumor) {
		t.Errorf("Should not contain ('B', 2)")
	}

	if rumors.Contains(fourthRumor) {
		t.Errorf("Should not contain ('C', 0)")
	}
}