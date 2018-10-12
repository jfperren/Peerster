package main

import (
	"fmt"
	"strings"
)

func logClientMessage(message *SimpleMessage) {
	fmt.Printf("CLIENT MESSAGE %v\n", message.Contents)
}

func logSimpleMessage(message *SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
		message.OriginalName, message.RelayPeerAddr, message.Contents)
}

func logRumor(rumor *RumorMessage, relayAddress string) {
	fmt.Printf("RUMOR origin %v from %v ID %v contents %v\n", rumor.Origin, relayAddress, rumor.ID, rumor.Text)
}

func logMongering(peerAddress string) {
	fmt.Printf("MONGERING with %v\n", peerAddress)
}

func logStatus(status *StatusPacket, relayAddress string) {
	fmt.Printf("STATUS from %v ", relayAddress)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}

func logFlippedCoin(peerAddress string) {
	fmt.Printf("FLIPPED COIN sending rumor to %v\n", peerAddress)
}

func logInSyncWith(peerAddress string) {
	fmt.Printf("IN SUNC WITH %v\n", peerAddress)
}

func logPeers(peers []string) {
	fmt.Printf("PEERS %v\n", strings.Join(peers, ","))
}

// --- Debug Messages ---
// These are optional messages, not required in the assignment
// that might be used for debugging.

func debugStopMongering(rumor *RumorMessage) {
	fmt.Printf("STOP MONGERING rumor %v\n", rumor.Text)
}

func debugTimeout(peer string) {
	fmt.Printf("TIMEOUT from %v\n", peer)
}

func debugSendStatus(status *StatusPacket, to string) {
	fmt.Printf("SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}