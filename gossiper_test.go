package main

import (
	"testing"
)

func MockGossiper() *Gossiper {

	gossipAddress := ":5050"
	clientAddress := ":8080"

	return &Gossiper{
		simple:         true,
		gossipSocket:   NewUDPSocket(gossipAddress),
		clientSocket:   NewUDPSocket(clientAddress),
		gossipAddress:  gossipAddress,
		clientAddress:  clientAddress,
		Name:           "Tester",
		peers:          []string{"127.0.0.1:5051", "127.0.0.1:5052"},
		handlers:       make(map[string]chan*StatusPacket),
		rumors:         makeRumors(),
		NextID:			0,
	}
}

// Tests that a Gossiper correctly creates a Status packet based on its rumors.
func TestStatusPacket(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})
	gossiper.rumors.put(&RumorMessage{"B", 1, "Hi"})

	statusPacket := gossiper.generateStatusPacket()

	if (len(statusPacket.Want) != 2) {
		t.Errorf("Length of statusPacket is %v, expected %v", len(statusPacket.Want), 2)
	}

	if (statusPacket.Want[0].Identifier != "A") {
		t.Errorf("Expect Origin %v at position %v, got %v", "A", 0, statusPacket.Want[0].Identifier)
	}

	if (statusPacket.Want[1].Identifier != "B") {
		t.Errorf("Expect Origin %v at position %v, got %v", "B", 1, statusPacket.Want[1].Identifier)
	}

	if (statusPacket.Want[0].NextID != 1) {
		t.Errorf("Expect NextID for Origin %v to be %v, got %v", "A", 1, statusPacket.Want[0].NextID)
	}

	if (statusPacket.Want[1].NextID != 0) {
		t.Errorf("Expect NextID for Origin %v to be %v, got %v", "B", 0, statusPacket.Want[1].NextID)
	}
}

// Test that a gossiper compares statusPackets correctly when they are identical.
func TestCompareStatusPacketWithSame(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})
	gossiper.rumors.put(&RumorMessage{"B", 1, "Hi"})

	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"A", 1},
		PeerStatus{"B", 0},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor != nil {
		logStatus(gossiper.generateStatusPacket(), "none")
		t.Errorf("Found rumor to send when statusPackets are the same.")
	}

	if packet != nil {
		t.Errorf("Found packet to send when statusPackets are the same.")
	}
}

// Test that a gossiper compares statusPackets correctly when the other is missing a message
func TestCompareStatusPacketWithMissingMessage(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})
	gossiper.rumors.put(&RumorMessage{"B", 1, "Hi"})

	// The other node has not seen (A, 0) yet
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"A", 0},
		PeerStatus{"B", 1},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor == nil {
		t.Fatalf("Should send rumor if statusPacket is missing one.")
	}

	if packet != nil {
		t.Errorf("Found packet to send when statusPackets are the same.")
	}

	if rumor.ID != 0 || rumor.Origin != "A" || rumor.Text != "Hello" {
		t.Errorf("Sent wrong rumor %v, expected %v", gossiper.rumors.get("A", 0), rumor)
	}
}

// Test that a gossiper compares statusPackets correctly when the other is missing a node
func TestCompareStatusPacketWithMissingNode(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})
	gossiper.rumors.put(&RumorMessage{"B", 1, "Hi"})

	// The other node has not seen (A, 0) yet
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"B", 0},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor == nil {
		t.Errorf("Should send rumor if statusPacket is missing one.")
	}

	if packet != nil {
		t.Errorf("Found packet to send when statusPackets are the same.")
	}

	if rumor.ID != 0 || rumor.Origin != "A" || rumor.Text != "Hello" {
		t.Errorf("Sent wrong rumor %v, expected %v", gossiper.rumors.get("A", 0), rumor)
	}
}

// Test that a gossiper compares statusPackets correctly when the other has more messages
func TestCompareStatusPacketWithAdditionalMessages(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})
	gossiper.rumors.put(&RumorMessage{"B", 0, "Hi"})

	// The other node has (B, 1) which we don't have
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"A", 1},
		PeerStatus{"B", 2},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor != nil {
		t.Errorf("It should not send rumor if the other node has all our messages")
	}

	if packet == nil {
		t.Errorf("It should send status packet if we are missing messages")
	}
}

// Test that a gossiper compares statusPackets correctly when the other has more messages
func TestCompareStatusPacketWithAdditionalNodes(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})

	// The other node has (C, 1) which we don't have and don't know about
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"A", 1},
		PeerStatus{"B", 0},
		PeerStatus{"C", 1},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor != nil {
		t.Errorf("It should not send rumor if the other node has all our messages")
	}

	if packet == nil {
		t.Errorf("It should send status packet if we are missing messages")
	}
}

// Test that a gossiper compares statusPackets correctly when the other is missing a node
// but we don't have any message for this node.
func TestCompareStatusPacketMissingNodeWithZeroNextID(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"B", 0, "Hi"})

	// The other node does not know about A but we have no message
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"B", 1},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor != nil {
		t.Errorf("Should not send rumor if we don't have any new")
	}

	if packet != nil {
		t.Errorf("Should not send packet if they don't have any new rumor")
	}
}

// Test that a gossiper compares statusPackets correctly when the other has a unknown node
// but they don't have any message from this node.
func TestCompareStatusPacketAdditionalNodeWithZeroID(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})

	// We haven't seen B yet but they don't have any message
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"A", 1},
		PeerStatus{"B", 0},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor != nil {
		t.Errorf("Should not send rumor if we don't have any new")
	}

	if packet != nil {
		t.Errorf("Should not send packet if they don't have any new rumor")
	}
}

// Test that we prioritize our rumors when both status are missing
func TestCompareStatusPacketPrioritizeRumors(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})

	// They haven't seen (A, 0) but we haven't seen (B, 0)
	otherStatusPacket := &StatusPacket{[]PeerStatus{
		PeerStatus{"B", 1},
	}}

	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor == nil {
		t.Errorf("Should send a rumor if both statuses are missing rumors")
	}

	if packet != nil {
		t.Errorf("Should not send a status if both statuses are missing rumors")
	}
}

// Test that we send the first rumor when multiple are missing
func TestCompareStatusPacketFirstRumor(t *testing.T) {

	gossiper := MockGossiper()
	defer gossiper.stop()

	gossiper.rumors.put(&RumorMessage{"A", 0, "Hello"})
	gossiper.rumors.put(&RumorMessage{"A", 1, "Hello Again"})
	gossiper.rumors.put(&RumorMessage{"A", 2, "Is anyone here?"})

	// They haven't seen anything
	otherStatusPacket := &StatusPacket{[]PeerStatus{}}
	rumor, packet := gossiper.compareStatus(otherStatusPacket)

	if rumor == nil {
		t.Fatalf("Should send a rumor if they are missing rumors")
	}

	if packet != nil {
		t.Errorf("Should not send a status if they are missing rumors")
	}

	if rumor.ID != 0 || rumor.Origin != "A" || rumor.Text != "Hello" {
		t.Errorf("Should send the first rumor, instead sent rumor %v", rumor.ID)
	}
}