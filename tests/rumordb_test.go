package tests

import (
	"testing"
	"github.com/jfperren/Peerster/gossiper"
	"github.com/jfperren/Peerster/common"
)

func TestRumorsPutSameOrigin(t *testing.T) {

	rumors := gossiper.MakeRumorDatabase()

	rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
	rumors.Put(&common.RumorMessage{"A", 2, "Hi"})

	if len(rumors.Rumors) != 1 {
		t.Errorf("Expected length of Rumors to be %v, go %v instead.", 1, len(rumors.Rumors))
	}

	if !(rumors.Rumors["A"][0].ID == 1 && rumors.Rumors["A"][0].Origin == "A" && rumors.Rumors["A"][0].Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.Rumors["A"][0])
	}

	if !(rumors.Rumors["A"][1].ID == 2 && rumors.Rumors["A"][1].Origin == "A" && rumors.Rumors["A"][1].Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.Rumors["A"][1])
	}
}

func TestRumorsPutTwoOrigins(t *testing.T) {

	rumors := gossiper.MakeRumorDatabase()

	rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
	rumors.Put(&common.RumorMessage{"B", 2, "Hi"})

	if len(rumors.Rumors) != 2 {
		t.Errorf("Expected length of Rumors to be %v, go %v instead.", 2, len(rumors.Rumors))
	}

	if !(rumors.Rumors["A"][0].ID == 0 && rumors.Rumors["A"][0].Origin == "A" && rumors.Rumors["A"][0].Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.Rumors["A"][0])
	}

	if !(rumors.Rumors["B"][1].ID == 1 && rumors.Rumors["B"][1].Origin == "B" && rumors.Rumors["B"][1].Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.Rumors["A"][1])
	}
}

func TestRumorsGetSameOrigin(t *testing.T) {

	rumors := gossiper.MakeRumorDatabase()

	rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
	rumors.Put(&common.RumorMessage{"A", 2, "Hi"})

	firstRumor := rumors.Get("A", 1)
	secondRumor := rumors.Get("A", 2)

	if !(firstRumor.ID == 1 && firstRumor.Origin == "A" && firstRumor.Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.Rumors["A"][0])
	}

	if !(secondRumor.ID == 2 && secondRumor.Origin == "A" && secondRumor.Text == "Hi") {
		t.Errorf("Wrong rumor at Rumors['A'][1] -> %v.", rumors.Rumors["A"][1])
	}
}

func TestRumorsGetTwoOrigins(t *testing.T) {

	rumors := gossiper.MakeRumorDatabase()

	rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
	rumors.Put(&common.RumorMessage{"B", 2, "Hi"})

	firstRumor := rumors.Get("A", 1)
	secondRumor := rumors.Get("B", 2)

	if !(firstRumor.ID == 1 && firstRumor.Origin == "A" && firstRumor.Text == "Hello") {
		t.Errorf("Wrong rumor at Rumors['A'][0] -> %v.", rumors.Rumors["A"][1])
	}

	if secondRumor != nil {
		t.Errorf("Should not store rumor (B, 0)")
	}
}

func TestRumorsGetInexistant(t *testing.T) {

	rumors := gossiper.MakeRumorDatabase()

	rumors.Put(&common.RumorMessage{"A", 0, "Hello"})
	rumors.Put(&common.RumorMessage{"B", 1, "Hi"})

	firstRumor := rumors.Get("A", 1)
	secondRumor := rumors.Get("B", 0)

	if firstRumor != nil {
		t.Errorf("Should not return any rumor at Rumors['A'][1], got %v.", firstRumor)
	}

	if secondRumor != nil {
		t.Errorf("Should not return any rumor at Rumors['B'][0], got %v.", secondRumor)
	}
}

func TestRumorsExpects(t *testing.T) {

	rumors := gossiper.MakeRumorDatabase()

	firstRumor := &common.RumorMessage{"A", 1, "Hello"}
	secondRumor := &common.RumorMessage{"B", 2, "Hi"}
	thirdRumor := &common.RumorMessage{"A", 2, "Hey"}
	fourthRumor := &common.RumorMessage{"C", 1, "Greetings"}

	rumors.Put(firstRumor)
	rumors.Put(secondRumor)

	if rumors.Expects(firstRumor) {
		t.Errorf("Should not expect ('A', 1)")
	}

	if rumors.Expects(secondRumor) {
		t.Errorf("Should no expect ('B', 2)")
	}

	if !rumors.Expects(thirdRumor) {
		t.Errorf("Should expect ('A', 2)")
	}

	if !rumors.Expects(fourthRumor) {
		t.Errorf("Should expect ('C', 1)")
	}
}