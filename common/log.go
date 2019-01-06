package common

import (
	"crypto/sha256"
	"encoding/hex"
    "fmt"
	"log"
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
	log.Printf("CLIENT MESSAGE %v\n", message)
}

func LogSimpleMessage(message *SimpleMessage) {
	log.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
		message.OriginalName, message.RelayPeerAddr, message.Contents)
}

func LogRumor(rumor IRumorMessage, relayAddress string) {
    switch t := rumor.(type) {
    default:
        log.Printf("RUMOR origin %v from %v ID %v type %T\n", rumor.GetOrigin(), relayAddress, rumor.GetID(), t)
    case *RumorMessage:
        log.Printf("RUMOR origin %v from %v ID %v contents %v\n", rumor.GetOrigin(), relayAddress, rumor.GetID(), (rumor.(*RumorMessage)).Text)
    }
}

func LogMongering(peerAddress string) {
	log.Printf("MONGERING with %v\n", peerAddress)
}

func LogStatus(status *StatusPacket, relayAddress string) {
	log.Printf("STATUS from %v ", relayAddress)

	for _, peerStatus := range status.Want {
		log.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	log.Printf("\n")
}

func LogFlippedCoin(peerAddress string) {
	log.Printf("FLIPPED COIN sending rumor to %v\n", peerAddress)
}

func LogInSyncWith(peerAddress string) {
	log.Printf("IN SYNC WITH %v\n", peerAddress)
}

func LogPeers(peers []string) {
	log.Printf("PEERS %v\n", strings.Join(peers, ","))
}

func LogUpdateRoutingTable(origin, address string) {
	log.Printf("DSDV %v %v\n", origin, address)
}

func LogPrivate(private *PrivateMessage) {
	log.Printf("PRIVATE origin %v hop-limit %v contents %v\n", private.Origin, private.HopLimit, private.Text)
}

func LogDownloadingMetafile(filename string, seed string) {
	log.Printf("DOWNLOADING metafile of %v from %v\n", filename, seed)
}

func LogDownloadingChunk(filename string, n int, seed string) {
	log.Printf("DOWNLOADING %v chunk %v from %v\n", filename, n, seed)
}

func LogReconstructed(filename string) {
	log.Printf("RECONSTRUCTED file %v\n", filename)
}

func LogMatch(result SearchResult, origin string) {

	chunks := make([]string, 0)

	for _, chunkId := range result.ChunkMap {
		chunks = append(chunks, fmt.Sprintf("%v", chunkId))
	}

	log.Printf("FOUND match %v at %v metafile=%v chunks=%v\n",
		result.FileName,
		origin,
		hex.EncodeToString(result.MetafileHash[:]),
		strings.Join(chunks, ","),
	)
}

func LogSearchFinished() {
	log.Printf("SEARCH FINISHED\n")
}

func LogFoundBlock(hash [32]byte) {
	log.Printf("FOUND-BLOCK %v\n", hex.EncodeToString(hash[:]))
}

func LogChain(blocks []*Block) {

	blocksStr := make([]string, 0)

	for _, block := range blocks {
		blocksStr = append(blocksStr, block.Str())
	}

	log.Printf("CHAIN %v\n", strings.Join(blocksStr, " "))
}

func LogShorterFork(block *Block) {
	hash := block.Hash()
	log.Printf("FORK-SHORTER %v\n", hex.EncodeToString(hash[:]))
}

func LogForkLongerRewind(current []*Block) {
	log.Printf("FORK-LONGER rewind %v blocks\n", len(current))
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
        log.Printf("STOP MONGERING rumor of type %T\n", t)
        //log.Printf("unexpected type %T\n", t)     // %T prints whatever type t has
    case *RumorMessage:
        log.Printf("STOP MONGERING rumor %v\n", (rumor.(*RumorMessage)).Text)
    }
}

func DebugTimeout(peer string) {
	if !Verbose { return }
	log.Printf("TIMEOUT from %v\n", peer)
}

func DebugSendStatus(status *StatusPacket, to string) {
	if !Verbose { return }
	log.Printf("SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		log.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	log.Printf("\n")
}

func DebugForwardRumor(rumor IRumorMessage) {
	if !Verbose { return }
    switch t := rumor.(type) {
    default:
        log.Printf("FORWARD rumor origin %v from %v ID %v type %T\n", rumor.GetOrigin(), rumor.GetID(), t)
    case *RumorMessage:
        log.Printf("FORWARD rumor %v\n", (rumor.(*RumorMessage)).Text)
    }
}

func DebugAskAndSendStatus(status *StatusPacket, to string) {
	if !Verbose { return }
	log.Printf("ASK AND SEND STATUS to %v ", to)

	for _, peerStatus := range status.Want {
		log.Printf("peer %v nextID %v ", peerStatus.Identifier, peerStatus.NextID)
	}

	log.Printf("\n")
}

func DebugServerRequest(req *http.Request) {
	if !Verbose { return }
	log.Printf("%v %v\n", req.Method, req.URL)
}

func DebugSendRouteRumor(address string) {
	if !Verbose { return }
	log.Printf("SEND ROUTE RUMOR to %v\n", address)
}

func DebugReceiveRouteRumor(origin, address string) {
	if !Verbose { return }
	log.Printf("RECEIVE ROUTE RUMOR from %v at %v\n", origin, address)
}

func DebugUnknownDestination(destination string) {
	if !Verbose { return }
	log.Printf("UNKNOWN DESTINATION %v\n", destination)
}

func DebugScanChunk(chunkPosition int, hash []byte) {
	if !Verbose { return }
	log.Printf("SCAN CHUNK number %v hash %v...\n", chunkPosition, hex.EncodeToString(hash)[:8])
}

func DebugScanFile(filename string, size int, metahash []byte) {
	if !Verbose { return }
	log.Printf("SCAN FILE name %v size %v metahash %v\n", filename, size, hex.EncodeToString(metahash))
}

func DebugStartDownload(filename string, metahash []byte, source string) {
	if !Verbose { return }
	log.Printf("START DOWNLOADING file %v from %v metahash %v\n", filename, source, hex.EncodeToString(metahash))
}

func DebugDownloadTimeout(filename string, metahash []byte, source string) {
	if !Verbose { return }
	log.Printf("DOWNLOAD TIMEOUT file %v from %v metahash %v\n", filename, source, hex.EncodeToString(metahash))
}

func DebugDownloadCompleted(filename string, metahash []byte, source string) {
	if !Verbose { return }
	log.Printf("DOWNLOAD COMPLETED file %v from %v metahash %v\n", filename, source, hex.EncodeToString(metahash))
}

func DebugReceiveDataRequest(request *DataRequest) {
	if !Verbose { return }
	log.Printf("RECEIVE DATA REQUEST from %v to %v metahash %v\n", request.Origin, request.Destination, hex.EncodeToString(request.HashValue))
}

func DebugReceiveDataReply(reply *DataReply) {
	if !Verbose { return }
	log.Printf("RECEIVE DATA REPLY from %v to %v metahash %v\n", reply.Origin, reply.Destination, hex.EncodeToString(reply.HashValue))
}

func DebugForwardPointToPoint(destination, nextAddress string) {
	if !Verbose { return }
	log.Printf("ROUTE POINT-TO-POINT MESSAGE destination %v nextAddreess %v\n", destination, nextAddress)
}

func DebugHashNotFound(hash []byte, source string) {
	if !Verbose { return }
	log.Printf("NOT FOUND hash %v from %v\n", hex.EncodeToString(hash), source)
}

func DebugFileNotFound(file string) {
	if !Verbose { return }
	log.Printf("NOT FOUND file %v\n", file)
}

func DebugCorruptedDataReply(hash []byte, reply *DataReply) {
	if !Verbose { return }

	expected := hex.EncodeToString(hash)[:8]
	received := hex.EncodeToString(reply.HashValue)[:8]
	computedHash := sha256.Sum256(reply.Data)
	computed := hex.EncodeToString(computedHash[:])[:8]

	log.Printf("CORRUPTED DATA REPLY expected %v received %v computed %v\n", expected, received, computed)
}

func DebugSendNoDestination() {
	if !Verbose { return }
	log.Printf("WARNING attempt to send or forward to node with no destination\n")
}

func DebugSendNoOrigin() {
	if !Verbose { return }
	log.Printf("WARNING attempt to send or forward to node without specifying origin\n")
}

func DebugFileTooBig(name string) {
	if !Verbose { return }
	log.Printf("WARNING file %v is too big for Peerster (max. 2Mb)\n", name)
}

func DebugStartGossiper(clientAddress, gossipAddress, name string, peers []string, simple bool, rtimer time.Duration) {
	if !Verbose { return }
	log.Printf("START GOSSIPER\n")
	log.Printf("client address %v\n", clientAddress)
	log.Printf("gossip address %v\n", gossipAddress)
	log.Printf("name %v\n", name)
	log.Printf("peers %v\n", peers)
	log.Printf("simple %v\n", simple)
	log.Printf("rtimer %v\n", rtimer)
}

func DebugStopGossiper() {
	if !Verbose { return }
	log.Printf("STOP GOSSIPER\n")
}

func DebugStartSearch(keywords []string, budget uint64, increasing bool) {
	if !Verbose { return }
	log.Printf("START search %v budget %v increasing %v\n", strings.Join(keywords, SearchKeywordSeparator), budget, increasing)
}

func DebugSearchTimeout(keywords []string) {
	if !Verbose { return }
	log.Printf("TIMEOUT search %v\n", strings.Join(keywords, SearchKeywordSeparator))
}

func DebugSearchResults(keywords []string, results []*SearchResult) {
	if !Verbose { return }
	log.Printf("FOUND %v results for keywords %v\n", len(results), strings.Join(keywords, SearchKeywordSeparator))
}

func DebugInvalidPacket(packet *GossipPacket) {
	if !Verbose { return }
	log.Printf("WARNING received invalid packet %v\n", packet)
}

func DebugSearchStatus(count int, keywords []string) {
	if !Verbose { return }
	log.Printf("SEARCH STATUS %v full results for search %v\n", count, strings.Join(keywords, SearchKeywordSeparator))
}

func DebugProcessSearchRequest(origin string, keywords []string) {
	if !Verbose { return }
	log.Printf("PROCESS search request from %v keywords %v\n", origin, strings.Join(keywords, SearchKeywordSeparator))
}

func DebugIgnoreSpam(origin string, keywords []string) {
	if !Verbose { return }
	log.Printf("IGNORE SPAM from %v keywords %v\n", origin, strings.Join(keywords, SearchKeywordSeparator))
}

func DebugDownloadUnknownFile(hash []byte) {
	if !Verbose { return }
	log.Printf("WARNING cannot download unknown file %v\n", hex.EncodeToString(hash))
}

func DebugNoKnownOwnerForFile(hash []byte) {
	if !Verbose { return }
	log.Printf("WARNING cannot download file %v has no owner\n", hex.EncodeToString(hash))
}

func DebugForwardSearchRequest(request *SearchRequest, next string) {
	if !Verbose { return }
	log.Printf("FORWARD search request %v from %v to %v budget %v\n", strings.Join(request.Keywords, SearchKeywordSeparator), request.Origin, next, request.Budget)
}

func DebugIgnoreBlockIsNotValid(block *Block) {
	if !Verbose { return }
	hash := block.Hash()
	log.Printf("IGNORE block %v is invalid\n", hex.EncodeToString(hash[:]))
}

func DebugIgnoreBlockAlreadyPresent(block *Block) {
	if !Verbose { return }
	hash := block.Hash()
	log.Printf("IGNORE block %v is already in chain\n", hex.EncodeToString(hash[:]))
}

func DebugIgnoreBlockInconsistent(block *Block) {
	if !Verbose { return }
	hash := block.Hash()
	log.Printf("IGNORE block %v is inconsistent with current namespace\n", hex.EncodeToString(hash[:]))
}

func DebugIgnoreBlockPrevDoesntMatch(block *Block, prev [32]byte) {
	if !Verbose { return }
	hash := block.Hash()
	log.Printf("IGNORE block %v prev hash %v does not match chain end %v\n", hex.EncodeToString(hash[:]),
		hex.EncodeToString(block.PrevHash[:]), hex.EncodeToString(prev[:]))
}

func DebugIgnoreTransactionAlreadyInChain(transaction *TxPublish) {
	if !Verbose { return }
	log.Printf("IGNORE transaction %v|%v already in chain\n", transaction.File.Name, transaction.User.Name)
}

func DebugIgnoreTransactionAlreadyCandidate(transaction *TxPublish) {
	if !Verbose { return }
	log.Printf("IGNORE transaction %v|%v already candidate\n", transaction.File.Name, transaction.User.Name)
}

func DebugAddCandidateTransaction(transaction *TxPublish) {
	if !Verbose { return }
	log.Printf("CANDIDATE transaction %v|%v successfully added\n", transaction.File.Name, transaction.User.Name)
}

func DebugBroadcastTransaction(transaction *TxPublish) {
	if !Verbose { return }
	log.Printf("BROADCAST transaction %v|%v\n", transaction.File.Name, transaction.User.Name)
}

func DebugReceiveTransaction(transaction *TxPublish) {
	if !Verbose { return }
	log.Printf("RECEIVE transaction %v|%v\n", transaction.File.Name, transaction.User.Name)
}

func DebugSleep(duration time.Duration) {
	if !Verbose { return }
	log.Printf("SLEEP %v\n", duration)
}

func DebugChainLength(length int) {
	if !Verbose { return }
	log.Printf("CHAIN LENGTH %v\n", length)
}

func DebugBroadcastBlock(hash [32]byte) {
	if !Verbose { return }
	log.Printf("BROADCAST BLOCK %v\n", hex.EncodeToString(hash[:]))
}

func DebugServeSeachReply(reply *SearchReply) {
	if !Verbose { return }

	results := make([]string, 0)

	for _, result := range reply.Results {
		results = append(results, result.FileName)
	}

	log.Printf("SERVE SEARCH REPLY to %v results %v\n", reply.Destination, strings.Join(results, ","))
}
