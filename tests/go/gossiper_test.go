package tests

//import (
//	"github.com/jfperren/Peerster/common"
//	"github.com/jfperren/Peerster/gossiper"
//	"testing"
//	"time"
//)
//
//func MockGossiper() *gossiper.Gossiper {
//
//	gossipAddress := ":5050"
//	clientAddress := ":8080"
//
//	return &gossiper.Gossiper{
//		Simple:        true,
//		GossipSocket:  common.NewUDPSocket(gossipAddress),
//		ClientSocket:  common.NewUDPSocket(clientAddress),
//		GossipAddress: gossipAddress,
//		ClientAddress: clientAddress,
//		Name:          "Tester",
//		Peers:         []string{"127.0.0.1:5051", "127.0.0.1:5052"},
//		Handlers:      make(map[string]chan*common.StatusPacket),
//		Rumors:        gossiper.MakeRumorDatabase(),
//		NextID:        common.InitialId,
//	}
//}
//
//// Tests that a Gossiper correctly creates a Status packet based on its Rumors.
//func TestStatusPacket(t *testing.T) {
//
//	gossiper := MockGossiper()
//	defer gossiper.Stop()
//
//	gossiper.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//
//	time.Sleep(100 * time.Millisecond)
//
//	gossiper.Rumors.Put(&common.RumorMessage{"B", 2, "Hi"})
//
//	statusPacket := gossiper.GenerateStatusPacket()
//
//	if len(statusPacket.Want) != 1 {
//		t.Errorf("Length of statusPacket is %v, expected %v", len(statusPacket.Want), 2)
//	}
//
//	if statusPacket.Want[0].Identifier != "A" {
//		t.Errorf("Expect Origin %v at position %v, got %v", "A", 1, statusPacket.Want[0].Identifier)
//	}
//
//	if gossiper.Rumors.Get("A", 1) == nil {
//		t.Errorf("Should return (A, 1)")
//	}
//
//	if gossiper.Rumors.Get("B", 2) != nil {
//		t.Errorf("Should return nil for (B, 2)")
//	}
//
//	if gossiper.Rumors.NextIDFor("A") != 2 {
//		t.Errorf("Next ID for A should be 2")
//	}
//
//	if gossiper.Rumors.NextIDFor("B") != 1 {
//		t.Errorf("Next ID for B should be 1")
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when they are identical.
//func TestCompareStatusPacketWithSame(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//	g.Rumors.Put(&common.RumorMessage{"B", 2, "Hi"})
//
//	otherStatusPacket := []common.PeerStatus{
//		{"A", 2},
//		{"B", 1},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor != nil {
//		t.Errorf("Found rumor to send when statusPackets are the same.")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet != nil {
//		t.Errorf("Found packet to send when statusPackets are the same.")
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when the other is missing a message
//func TestCompareStatusPacketWithMissingMessage(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//	g.Rumors.Put(&common.RumorMessage{"B", 2, "Hi"})
//
//	// The other node has not seen (A, 0) yet
//	otherStatusPacket := []common.PeerStatus{
//		{"A", 1},
//		{"B", 2},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor == nil {
//		t.Fatalf("Should send rumor if statusPacket is missing one.")
//	}
//
//	if packet != nil {
//		t.Errorf("Found packet to send when statusPackets are the same.")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if rumor.ID != 1 || rumor.Origin != "A" || rumor.Text != "Hello" {
//		t.Errorf("Sent wrong rumor %v, expected %v", g.Rumors.Get("A", 1), rumor)
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when the other is missing a node
//func TestCompareStatusPacketWithMissingNode(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//	g.Rumors.Put(&common.RumorMessage{"B", 2, "Hi"})
//
//	// The other node has not seen (A, 0) yet
//	otherStatusPacket := []common.PeerStatus{
//		{"B", 1},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor == nil {
//		t.Errorf("Should send rumor if statusPacket is missing one.")
//	}
//
//	if packet != nil {
//		t.Errorf("Found packet to send when statusPackets are the same.")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if rumor.ID != 1 || rumor.Origin != "A" || rumor.Text != "Hello" {
//		t.Errorf("Sent wrong rumor %v, expected %v", g.Rumors.Get("A", 1), rumor)
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when the other has more messages
//func TestCompareStatusPacketWithAdditionalMessages(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//	g.Rumors.Put(&common.RumorMessage{"B", 1, "Hi"})
//
//	// The other node has (B, 1) which we don't have
//	otherStatusPacket := []common.PeerStatus{
//		{"A", 2},
//		{"B", 3},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor != nil {
//		t.Errorf("It should not send rumor if the other node has all our messages")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet == nil {
//		t.Errorf("It should send status packet if we are missing messages")
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when the other has more messages
//func TestCompareStatusPacketWithAdditionalNodes(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//
//	// The other node has (C, 1) which we don't have and don't know about
//	otherStatusPacket := []common.PeerStatus{
//		{"A", 2},
//		{"B", 1},
//		{"C", 2},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor != nil {
//		t.Errorf("It should not send rumor if the other node has all our messages")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet == nil {
//		t.Errorf("It should send status packet if we are missing messages")
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when the other is missing a node
//// but we don't have any message for this node.
//func TestCompareStatusPacketMissingNodeWithInitialID(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"B", 1, "Hello"})
//	g.Rumors.Put(&common.RumorMessage{"C", 2, "Hi"})
//
//	// The other node does not know about C but we have no message
//	otherStatusPacket := []common.PeerStatus{
//		{"B", 2},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor != nil {
//		t.Errorf("Should not send rumor if we don't have any new")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet != nil {
//		t.Errorf("Should not send packet if they don't have any new rumor")
//	}
//}
//
//// Test that a gossiper compares statusPackets correctly when the other has a unknown node
//// but they don't have any message from this node.
//func TestCompareStatusPacketAdditionalNodeWithInitialID(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//
//	// We haven't seen B yet but they don't have any message
//	otherStatusPacket := []common.PeerStatus{
//		{"A", 2},
//		{"B", 1},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor != nil {
//		t.Errorf("Should not send rumor if we don't have any new")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet != nil {
//		t.Errorf("Should not send packet if they don't have any new rumor")
//	}
//}
//
//// Test that we prioritize our Rumors when both status are missing
//func TestCompareStatusPacketPrioritizeRumors(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//
//	// They haven't seen (A, 0) but we haven't seen (B, 0)
//	otherStatusPacket := []common.PeerStatus{
//		{"B", 2},
//	}
//
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor == nil {
//		t.Errorf("Should send a rumor if both statuses are missing Rumors")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet != nil {
//		t.Errorf("Should not send a status if both statuses are missing Rumors")
//	}
//}
//
//// Test that we send the first rumor when multiple are missing
//func TestCompareStatusPacketFirstRumor(t *testing.T) {
//
//	g := MockGossiper()
//	defer g.Stop()
//
//	g.Rumors.Put(&common.RumorMessage{"A", 1, "Hello"})
//	g.Rumors.Put(&common.RumorMessage{"A", 2, "Hello Again"})
//	g.Rumors.Put(&common.RumorMessage{"A", 3, "Is anyone here?"})
//
//	// They haven't seen anything
//	otherStatusPacket := make([]common.PeerStatus, 0)
//	rumor, allRumors, packet := g.CompareStatus(otherStatusPacket, gossiper.ComparisonModeMissingOrNew)
//
//	if rumor == nil {
//		t.Fatalf("Should send a rumor if they are missing Rumors")
//	}
//
//	if allRumors != nil {
//		t.Fatalf("Should not return Rumors in MissingOrNew mode")
//	}
//
//	if packet != nil {
//		t.Errorf("Should not send a status if they are missing Rumors")
//	}
//
//	if rumor.ID != 1 || rumor.Origin != "A" || rumor.Text != "Hello" {
//		t.Errorf("Should send the first rumor, instead sent rumor %v", rumor.ID)
//	}
//}