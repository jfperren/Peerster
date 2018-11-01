package common

import (
	"fmt"
	"net/http"
	"strings"
)

const LogDebug = true;

// --
// -- INFO MESSAGES
// --
//
// These are log messages to be used in the assignment.

func LogClientMessage(message *SimpleMessage) {
	fmt.Printf("CLIENT MESSAGE %v\n", message.Contents)
}

func LogSimpleMessage(message *SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
		message.OriginalName, message.RelayPeerAddr, message.Contents)
}

func LogRumor(rumor *RumorMessage, relayAddress string) {
	fmt.Printf("RUMOR origin %v from %v ID %v contents %v\n", rumor.Origin, relayAddress, rumor.ID, rumor.Text)
}

func LogMongering(peerAddress string) {
	fmt.Printf("MONGERING with %v\n", peerAddress)
}

func LogStatus(status *StatusPacket, relayAddress string) {
	fmt.Printf("STATUS from %v ", relayAddress)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}

func LogFlippedCoin(peerAddress string) {
	fmt.Printf("FLIPPED COIN sending rumor to %v\n", peerAddress)
}

func LogInSyncWith(peerAddress string) {
	fmt.Printf("IN SYNC WITH %v\n", peerAddress)
}

func LogPeers(peers []string) {
	fmt.Printf("PEERS %v\n", strings.Join(peers, ","))
}

func LogUpdateRoutingTable(origin, address string) {
	fmt.Printf("DSDV %v %v\n", origin, address)
}

func LogPrivate(private *PrivateMessage) {
	fmt.Printf("PRIVATE origin %v hop-limit %v contents %v\n", private.Origin, private.HopLimit, private.Text)
}

// --
// -- DEBUG MESSAGES
// --
//
// These are optional messages, not required in the assignment
// that might be used for debugging.

func DebugStopMongering(rumor *RumorMessage) {
	if !LogDebug { return }
	fmt.Printf("STOP MONGERING rumor %v\n", rumor.Text)
}

func DebugTimeout(peer string) {
	if !LogDebug { return }
	fmt.Printf("TIMEOUT from %v\n", peer)
}

func DebugSendStatus(status *StatusPacket, to string) {
	if !LogDebug { return }
	fmt.Printf("SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}

func DebugForwardRumor(rumor *RumorMessage) {
	if !LogDebug { return }
	fmt.Printf("FORWARD rumor %v\n", rumor.Text)
}

func DebugAskAndSendStatus(status *StatusPacket, to string) {
	if !LogDebug { return }
	fmt.Printf("ASK AND SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}

func DebugServerRequest(req *http.Request) {
	if !LogDebug { return }
	fmt.Printf("%v %v\n", req.Method, req.URL)
}

func DebugSendRouteRumor(address string) {
	if !LogDebug { return }
	fmt.Printf("SEND ROUTE RUMOR to %v\n", address)
}

func DebugReceiveRouteRumor(origin, address string) {
	if !LogDebug { return }
	fmt.Printf("RECEIVE ROUTE RUMOR from %v at %v\n", origin, address)
}

func DebugUnknownDestination(destination string) {
	if !LogDebug { return }
	fmt.Printf("UNKNOWN DESTINATION %v\n", destination)
}