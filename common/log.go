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

//
//  INFO MESSAGES
//  -------------
//
//  These are log messages to be used in the assignment.

func LogClientMessage(message string) {
	fmt.Printf("CLIENT MESSAGE %v\n", message)
}

func LogSimpleMessage(message *SimpleMessage) {
	fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
		message.OriginalName, message.RelayPeerAddr, message.Contents)
}

func LogRumor(rumor IRumorMessage, relayAddress string) {
    switch t := rumor.(type) {
    default:
        fmt.Printf("RUMOR origin %v from %v ID %v type %T\n", rumor.GetOrigin(), relayAddress, rumor.GetID(), t)
    case *RumorMessage:
        fmt.Printf("RUMOR origin %v from %v ID %v contents %v\n", rumor.GetOrigin(), relayAddress, rumor.GetID(), (rumor.(*RumorMessage)).Text)
    }
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

func LogMatch(result SearchResult, origin string) {

	chunks := make([]string, 0)

	for _, chunkId := range result.ChunkMap {
		chunks = append(chunks, fmt.Sprintf("%v", chunkId))
	}

	fmt.Printf("FOUND match %v at %v metafile=%v chunks=%v\n",
		result.FileName,
		origin,
		hex.EncodeToString(result.MetafileHash[:]),
		strings.Join(chunks, ","),
	)
}

func LogSearchFinished() {
	fmt.Printf("SEARCH FINISHED\n")
}

func LogFoundBlock(hash [32]byte) {
	fmt.Printf("FOUND-BLOCK %v\n", hex.EncodeToString(hash[:]))
}

func LogChain(blocks []*Block) {

	blocksStr := make([]string, 0)

	for _, block := range blocks {
		blocksStr = append(blocksStr, block.Str())
	}

	fmt.Printf("CHAIN %v\n", strings.Join(blocksStr, " "))
}

func LogShorterFork(block *Block) {
	hash := block.Hash()
	fmt.Printf("FORK-SHORTER %v\n", hex.EncodeToString(hash[:]))
}

func LogForkLongerRewind(current []*Block) {
	fmt.Printf("FORK-LONGER rewind %v blocks\n", len(current))
}

//
//  DEBUG MESSAGES
//  --------------
//
//  These are optional messages, not required in the assignment
//  that might be used for debugging.

func DebugStopMongering(rumor IRumorMessage) {
	if !Verbose { return }
    switch t := rumor.(type) {
    default:
        fmt.Printf("STOP MONGERING rumor of type %T\n", t)
        //fmt.Printf("unexpected type %T\n", t)     // %T prints whatever type t has
    case *RumorMessage:
        fmt.Printf("STOP MONGERING rumor %v\n", (rumor.(*RumorMessage)).Text)
    }
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

func DebugForwardRumor(rumor IRumorMessage) {
	if !Verbose { return }
    switch t := rumor.(type) {
    default:
        fmt.Printf("FORWARD rumor origin %v from %v ID %v type %T\n", rumor.GetOrigin(), rumor.GetID(), t)
    case *RumorMessage:
        fmt.Printf("FORWARD rumor %v\n", (rumor.(*RumorMessage)).Text)
    }
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
	fmt.Printf("STOP GOSSIPER\n")
}

func DebugStartSearch(keywords []string, budget uint64, increasing bool) {
	if !Verbose { return }
	fmt.Printf("START search %v budget %v increasing %v\n", strings.Join(keywords, SearchKeywordSeparator), budget, increasing)
}

func DebugSearchTimeout(keywords []string) {
	if !Verbose { return }
	fmt.Printf("TIMEOUT search %v\n", strings.Join(keywords, SearchKeywordSeparator))
}

func DebugSearchResults(keywords []string, results []*SearchResult) {
	if !Verbose { return }
	fmt.Printf("FOUND %v results for keywords %v\n", len(results), strings.Join(keywords, SearchKeywordSeparator))
}

func DebugInvalidPacket(packet *GossipPacket) {
	if !Verbose { return }
	fmt.Printf("WARNING received invalid packet %v\n", packet)
}

func DebugSearchStatus(count int, keywords []string) {
	if !Verbose { return }
	fmt.Printf("SEARCH STATUS %v full results for search %v\n", count, strings.Join(keywords, SearchKeywordSeparator))
}

func DebugProcessSearchRequest(origin string, keywords []string) {
	if !Verbose { return }
	fmt.Printf("PROCESS search request from %v keywords %v\n", origin, strings.Join(keywords, SearchKeywordSeparator))
}

func DebugIgnoreSpam(origin string, keywords []string) {
	if !Verbose { return }
	fmt.Printf("IGNORE SPAM from %v keywords %v\n", origin, strings.Join(keywords, SearchKeywordSeparator))
}

func DebugDownloadUnknownFile(hash []byte) {
	if !Verbose { return }
	fmt.Printf("WARNING cannot download unknown file %v\n", hex.EncodeToString(hash))
}

func DebugNoKnownOwnerForFile(hash []byte) {
	if !Verbose { return }
	fmt.Printf("WARNING cannot download file %v has no owner\n", hex.EncodeToString(hash))
}

func DebugForwardSearchRequest(request *SearchRequest, next string) {
	if !Verbose { return }
	fmt.Printf("FORWARD search request %v from %v to %v budget %v\n", strings.Join(request.Keywords, SearchKeywordSeparator), request.Origin, next, request.Budget)
}

func DebugIgnoreBlockIsNotValid(block *Block) {
	if !Verbose { return }
	hash := block.Hash()
	fmt.Printf("IGNORE block %v is invalid\n", hex.EncodeToString(hash[:]))
}

func DebugIgnoreBlockAlreadyPresent(block *Block) {
	if !Verbose { return }
	hash := block.Hash()
	fmt.Printf("IGNORE block %v is already in chain\n", hex.EncodeToString(hash[:]))
}

func DebugIgnoreBlockInconsistent(block *Block) {
	if !Verbose { return }
	hash := block.Hash()
	fmt.Printf("IGNORE block %v is inconsistent with current namespace\n", hex.EncodeToString(hash[:]))
}

func DebugIgnoreBlockPrevDoesntMatch(block *Block, prev [32]byte) {
	if !Verbose { return }
	hash := block.Hash()
	fmt.Printf("IGNORE block %v prev hash %v does not match chain end %v\n", hex.EncodeToString(hash[:]),
		hex.EncodeToString(block.PrevHash[:]), hex.EncodeToString(prev[:]))
}

func DebugIgnoreTransactionAlreadyInChain(transaction *TxPublish) {
	if !Verbose { return }
	fmt.Printf("IGNORE transaction %v already in chain\n", transaction.File.Name)
}

func DebugIgnoreTransactionAlreadyCandidate(transaction *TxPublish) {
	if !Verbose { return }
	fmt.Printf("IGNORE transaction %v already candidate\n", transaction.File.Name)
}

func DebugAddCandidateTransaction(transaction *TxPublish) {
	if !Verbose { return }
	fmt.Printf("CANDIDATE transaction %v successfully added\n", transaction.File.Name)
}

func DebugBroadcastTransaction(transaction *TxPublish) {
	if !Verbose { return }
	fmt.Printf("BROADCAST transaction %v\n", transaction.File.Name)
}

func DebugReceiveTransaction(transaction *TxPublish) {
	if !Verbose { return }
	fmt.Printf("RECEIVE transaction %v\n", transaction.File.Name)
}

func DebugSleep(duration time.Duration) {
	if !Verbose { return }
	fmt.Printf("SLEEP %v\n", duration)
}

func DebugChainLength(length int) {
	if !Verbose { return }
	fmt.Printf("CHAIN LENGTH %v\n", length)
}

func DebugBroadcastBlock(hash [32]byte) {
	if !Verbose { return }
	fmt.Printf("BROADCAST BLOCK %v\n", hex.EncodeToString(hash[:]))
}

func DebugServeSeachReply(reply *SearchReply) {
	if !Verbose { return }

	results := make([]string, 0)

	for _, result := range reply.Results {
		results = append(results, result.FileName)
	}

	fmt.Printf("SERVE SEARCH REPLY to %v results %v\n", reply.Destination, strings.Join(results, ","))
}

func DebugSkipSendNotAuthenticated() {
	if !Verbose { return }
	fmt.Printf("NOT AUTHENTICATED skip send packet\n")
}

func DebugDropUnauthenticatedOrigin(signed *Signature) {
	if !Verbose {
		return
	}
	fmt.Printf("DROP SIGNED MESSAGE no key ORIGIN %v\n", signed.Origin)
}

func DebugDropIncorrectSignature(signed *Signature) {
	if !Verbose {
		return
	}
	fmt.Printf("DROP SIGNED MESSAGE incorrect signature ORIGIN %v\n", signed.Origin)
}

func DebugDropWrongOrigin(signed *Signature) {
	if !Verbose { return }
	fmt.Printf("DROP SIGNED MESSAGE origin mismatch ORIGIN %v\n", signed.Origin)
}