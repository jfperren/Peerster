package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var Verbose = true

// --
// -- INFO MESSAGES
// --
//
// These are log messages to be used in the assignment.

func LogClientMessage(message string) {
	fmt.Printf("CLIENT MESSAGE %v\n", message)
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

func LogDownloadingMetafile(filename string, seed string) {
	fmt.Printf("DOWNLOADING metafile of %v from %v\n", filename, seed)
}

func LogDownloadingChunk(filename string, n int, seed string) {
	fmt.Printf("DOWNLOADING %v chunk %v from %v\n", filename, n, seed)
}

func LogReconstructed(filename string) {
	fmt.Printf("RECONSTRUCTED file %v\n", filename)
}

// --
// -- DEBUG MESSAGES
// --
//
// These are optional messages, not required in the assignment
// that might be used for debugging.

func DebugStopMongering(rumor *RumorMessage) {
	if !Verbose { return }
	fmt.Printf("STOP MONGERING rumor %v\n", rumor.Text)
}

func DebugTimeout(peer string) {
	if !Verbose { return }
	fmt.Printf("TIMEOUT from %v\n", peer)
}

func DebugSendStatus(status *StatusPacket, to string) {
	if !Verbose { return }
	fmt.Printf("SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}

func DebugForwardRumor(rumor *RumorMessage) {
	if !Verbose { return }
	fmt.Printf("FORWARD rumor %v\n", rumor.Text)
}

func DebugAskAndSendStatus(status *StatusPacket, to string) {
	if !Verbose { return }
	fmt.Printf("ASK AND SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		fmt.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	fmt.Printf("\n")
}

func DebugServerRequest(req *http.Request) {
	if !Verbose { return }
	fmt.Printf("%v %v\n", req.Method, req.URL)
}

func DebugSendRouteRumor(address string) {
	if !Verbose { return }
	fmt.Printf("SEND ROUTE RUMOR to %v\n", address)
}

func DebugReceiveRouteRumor(origin, address string) {
	if !Verbose { return }
	fmt.Printf("RECEIVE ROUTE RUMOR from %v at %v\n", origin, address)
}

func DebugUnknownDestination(destination string) {
	if !Verbose { return }
	fmt.Printf("UNKNOWN DESTINATION %v\n", destination)
}

func DebugScanChunk(chunkPosition int, hash []byte) {
	if !Verbose { return }
	fmt.Printf("SCAN CHUNK number %v hash %v...\n", chunkPosition, hex.EncodeToString(hash)[:8])
}

func DebugScanFile(filename string, size int, metahash []byte) {
	if !Verbose { return }
	fmt.Printf("SCAN FILE name %v size %v metahash %v\n", filename, size, hex.EncodeToString(metahash))
}

func DebugStartDownload(filename string, metahash []byte, source string) {
	if !Verbose { return }
	fmt.Printf("START DOWNLOADING file %v from %v metahash %v\n", filename, source, hex.EncodeToString(metahash))
}

func DebugDownloadTimeout(filename string, metahash []byte, source string) {
	if !Verbose { return }
	fmt.Printf("DOWNLOAD TIMEOUT file %v from %v metahash %v\n", filename, source, hex.EncodeToString(metahash))
}

func DebugDownloadCompleted(filename string, metahash []byte, source string) {
	if !Verbose { return }
	fmt.Printf("DOWNLOAD COMPLETED file %v from %v metahash %v\n", filename, source, hex.EncodeToString(metahash))
}

func DebugReceiveDataRequest(request *DataRequest) {
	if !Verbose { return }
	fmt.Printf("RECEIVE DATA REQUEST from %v to %v metahash %v\n", request.Origin, request.Destination, hex.EncodeToString(request.HashValue))
}

func DebugReceiveDataReply(reply *DataReply) {
	if !Verbose { return }
	fmt.Printf("RECEIVE DATA REPLY from %v to %v metahash %v\n", reply.Origin, reply.Destination, hex.EncodeToString(reply.HashValue))
}

func DebugForwardPointToPoint(destination, nextAddress string) {
	if !Verbose { return }
	fmt.Printf("ROUTE POINT-TO-POINT MESSAGE destination %v nextAddreess %v\n", destination, nextAddress)
}

func DebugHashNotFound(hash []byte, source string) {
	if !Verbose { return }
	fmt.Printf("NOT FOUND hash %v from %v\n", hex.EncodeToString(hash), source)
}

func DebugFileNotFound(file string) {
	if !Verbose { return }
	fmt.Printf("NOT FOUND file %v\n", file)
}

func DebugCorruptedDataReply(hash []byte, reply *DataReply) {
	if !Verbose { return }

	expected := hex.EncodeToString(hash)[:8]
	received := hex.EncodeToString(reply.HashValue)[:8]
	computedHash := sha256.Sum256(reply.Data)
	computed := hex.EncodeToString(computedHash[:])[:8]

	fmt.Printf("CORRUPTED DATA REPLY expected %v received %v computed %v\n", expected, received, computed)
}

func DebugSendNoDestination() {
	if !Verbose { return }
	fmt.Printf("WARNING attempt to send or forward to node with no destination\n")
}

func DebugSendNoOrigin() {
	if !Verbose { return }
	fmt.Printf("WARNING attempt to send or forward to node without specifying origin\n")
}

func DebugFileTooBig(name string) {
	if !Verbose { return }
	fmt.Printf("WARNING file %v is too big for Peerster (max. 2Mb)\n", name)
}

func DebugStartGossiper(clientAddress, gossipAddress, name string, peers []string, simple bool, rtimer time.Duration) {
	if !Verbose { return }
	fmt.Printf("START GOSSIPER\n")
	fmt.Printf("client address %v\n", clientAddress)
	fmt.Printf("gossip address %v\n", gossipAddress)
	fmt.Printf("name %v\n", name)
	fmt.Printf("peers %v\n", peers)
	fmt.Printf("simple %v\n", simple)
	fmt.Printf("rtimer %v\n", rtimer)
}

func DebugStopGossiper() {
	if !Verbose { return }
	fmt.Printf("STOP GOSSIPER")
}
