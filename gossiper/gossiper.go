package gossiper

import (
	"github.com/dedis/protobuf"
	"github.com/jfperren/Peerster/common"
	"log"
	"sync"
	"time"
)

// Root class of the Peerster program. It represents a "node" in the network. In this file, the more
// complex functions are implemented (rumor-mongering, downloading, handling packets,etc...).
//
// Lower-level functions are implemented in FileSystem (slicing and storing files & hashes), Dispatcher
// (dispatching packets to correct goroutines), Router (keep track of peers & routing table), RumorsDB
// (store rumors and compute vector clocks), on all of which Gossiper relies on.
type Gossiper struct {
	Name   			string // Name of this node
	Simple 			bool   // Stores if gossiper runs in simple mode.

	GossipSocket 	*common.UDPSocket // UDP Socket that connects to other nodes
	ClientSocket 	*common.UDPSocket // UDP Socket that connects to the client

	Rumors   		*RumorDatabase           // Database of known Rumors
	Messages 		[]*common.PrivateMessage // List of Private Messages Received

	FileSystem 		*FileSystem 	// Stores and serves shared files
	Dispatcher 		*Dispatcher 	// Dispatches incoming messages to expecting processes
	Router     		*Router     	// Handles routing to neighboring and non-neighboring nodes.
	SpamDetector 	*SpamDetector
	SearchEngine 	*SearchEngine 	//
	BlockChain		*BlockChain
}

const (
	ComparisonModeMissingOrNew = iota // Flag to be used when comparing two nodes' status packets
	ComparisonModeAllNew       = iota // Flag to be used when comparing a node status with the client status
)

// Create a new Gossiper using the given addresses.
//
//  - gossipAddress: Address on which the gossiper listens to other gossipers
//  - clientAddress: Address on which the gossiper listens for commands from the CLI
//  - name: Name of this gossiper (it will appear in this node's messages, should be unique).
//  - peers: list of IP addresses to which this gossiper is connected at start. More can be added later on.
//  - simple: Start this gossiper in simple mode (i.e. no gossip, only simple messages)
//  - rtimer: Time in seconds between route rumors. Set to 0 for not sending route rumors at all.
//  - separatefs: True if this gossiper uses its own subfolder for _Download and _SharedFiles.
//
// Note - Use gossiper.Start() to Start listening for messages.
//
func NewGossiper(gossipAddress, clientAddress, name string, peers string, simple bool, rtimer int, separatefs bool) *Gossiper {

	gossipSocket := common.NewUDPSocket(gossipAddress)
	var clientSocket *common.UDPSocket

	if clientAddress != "" {
		clientSocket = common.NewUDPSocket(clientAddress)
	}

	downloadPath := common.DownloadDir
	sharedPath := common.SharedFilesDir

	if separatefs {
		downloadPath = downloadPath + name + "/"
		sharedPath = sharedPath + name + "/"
	}

	return &Gossiper{
		Name:         name,
		Simple:       simple,
		GossipSocket: gossipSocket,
		ClientSocket: clientSocket,

		Rumors:     	NewRumorDatabase(),
		FileSystem: 	NewFileSystem(sharedPath, downloadPath),
		Dispatcher: 	NewDispatcher(),
		Router:     	NewRouter(peers, time.Duration(rtimer)*time.Second),
		SpamDetector:   NewSpamDetector(),
		SearchEngine: 	NewSearchEngine(),
		BlockChain:		NewBlockChain(),
	}
}

// --
// --  START & STOP
// --

// Start listening for UDP packets on Gossiper's clientAddress & gossipAddress
func (gossiper *Gossiper) Start() {

	go gossiper.receiveGossip()
	go gossiper.sendRouteRumors()

	if !gossiper.Simple {
		go gossiper.antiEntropy()
	}

	if gossiper.ClientSocket != nil {
		go gossiper.receiveClient()
	}

	// Allows the loops to run indefinitely after the main code is completed.
	wg := new(sync.WaitGroup)
	wg.Add(4)
	wg.Wait()
}

// Unbind from all ports, stop processes.
func (gossiper *Gossiper) Stop() {
	gossiper.ClientSocket.Unbind()
	gossiper.GossipSocket.Unbind()
}

// --
// --  EVENT LOOPS
// --

// Main loop for handling gossip packets from other nodes.
func (gossiper *Gossiper) receiveGossip() {
	for {
		var packet common.GossipPacket
		bytes, source, alive := gossiper.GossipSocket.Receive()

		if !alive {
			break
		}

		protobuf.Decode(bytes, &packet)

		if !packet.IsValid() {
			panic("Received invalid packet")
		}

		gossiper.Router.AddPeerIfNeeded(source)

		go gossiper.HandleGossip(&packet, source)
	}
}

// Main loop for pinging other nodes as part of the anti-entropy algorithm.
func (gossiper *Gossiper) antiEntropy() {
	for {
		peer, found := gossiper.Router.randomPeer()

		if found {
			packet := gossiper.GenerateStatusPacket().Packed()
			common.DebugAskAndSendStatus(packet.Status, peer)
			go gossiper.sendToNeighbor(peer, packet)
		}

		time.Sleep(common.AntiEntropyDT)
	}
}

// Main loop for sending route rumors
func (gossiper *Gossiper) sendRouteRumors() {

	if gossiper.Router.Rtimer == common.NoRouteRumor {
		return
	}

	for {
		peer, found := gossiper.Router.randomPeer()

		if found {
			routeRumor := gossiper.GenerateRouteRumor()
			common.DebugSendRouteRumor(peer)
			go gossiper.sendToNeighbor(peer, routeRumor.Packed())
		}

		time.Sleep(gossiper.Router.Rtimer)
	}
}

// Main loop for handling client packets.
func (gossiper *Gossiper) receiveClient() {

	for {

		var command common.Command
		bytes, _, alive := gossiper.ClientSocket.Receive()

		if !alive {
			break
		}

		protobuf.Decode(bytes, &command)
		go gossiper.HandleClient(&command)
	}
}

// --
// --  HANDLING NEW PACKETS
// --

//  Handle new packet from client
func (gossiper *Gossiper) HandleClient(command *common.Command) {

	if command == nil || !command.IsValid() {
		return // Fail gracefully
	}

	switch {

	case command.Message != nil:

		content := command.Message.Content
		common.LogClientMessage(content)

		if gossiper.Simple {

			message := common.NewSimpleMessage(gossiper.Name, gossiper.GossipSocket.Address, content)
			go gossiper.broadcastToNeighbors(message.Packed())

		} else {

			rumor := gossiper.generateRumor(content)

			gossiper.Rumors.Put(rumor)

			peer, found := gossiper.Router.randomPeer()

			if found {
				go gossiper.rumormonger(rumor, peer)
			}
		}

	case command.PrivateMessage != nil:

		destination := command.PrivateMessage.Destination
		content := command.PrivateMessage.Content

		private := common.NewPrivateMessage(gossiper.Name,destination, content)

		destined := gossiper.sendToNode(private.Packed(), destination, nil)
		gossiper.Messages = append(gossiper.Messages, private)

		if destined {
			common.LogPrivate(private)
		}

	case command.Download != nil:

		destination := command.Download.Destination
		filename := command.Download.FileName
		hash := command.Download.Hash

		go gossiper.StartDownload(filename, hash, destination, 0)

	case command.Upload != nil:

		metaFile, err := gossiper.FileSystem.ScanFile(command.Upload.FileName)

		if err != nil {

			transaction := NewTransaction(metaFile)

			if gossiper.BlockChain.tryAddFile(transaction) {
				gossiper.broadcastToNeighbors(transaction.Packed())
			}
		}

	case command.Search != nil:

		gossiper.RingSearch(command.Search.Keywords, command.Search.Budget)
	}
}

// Handle packet from another node.
func (gossiper *Gossiper) HandleGossip(packet *common.GossipPacket, source string) {

	if packet == nil || !packet.IsValid() {
		common.DebugInvalidPacket(packet)
		return // Fail gracefully
	}

	switch {

	case packet.Simple != nil:

		common.LogSimpleMessage(packet.Simple)
		common.LogPeers(gossiper.Router.Peers)

		go gossiper.broadcastToNeighbors(packet)

	case packet.Rumor != nil:

		if packet.Rumor.IsRouteRumor() {
			common.DebugReceiveRouteRumor(packet.Rumor.Origin, source)
		} else {
			common.LogRumor(packet.Rumor, source)
		}

		common.LogPeers(gossiper.Router.Peers)

		// We only store & forward if the rumor is our next expected rumo
		// from the source.
		if gossiper.Rumors.Expects(packet.Rumor) {

			if gossiper.Name != packet.Rumor.Origin {
				gossiper.Router.updateRoutingTable(packet.Rumor.Origin, source)
			}

			gossiper.Rumors.Put(packet.Rumor)
			peer, found := gossiper.Router.randomPeerExcept(source)

			if found {
				common.DebugForwardRumor(packet.Rumor)
				go gossiper.rumormonger(packet.Rumor, peer)
			}
		}

		statusPacket := gossiper.GenerateStatusPacket()
		common.DebugSendStatus(statusPacket, source)
		go gossiper.sendToNeighbor(source, statusPacket.Packed())

	case packet.Status != nil:

		common.LogStatus(packet.Status, source)
		common.LogPeers(gossiper.Router.Peers)

		expected := gossiper.Dispatcher.dispatchStatusPacket(source, packet)

		if !expected {

			rumor, _, _ := gossiper.CompareStatus(packet.Status.Want, ComparisonModeMissingOrNew)

			if rumor != nil {
				go gossiper.rumormonger(rumor, source)
			}
		}

	case packet.Private != nil:

		destination := packet.Private.Destination
		hopLimit := &packet.Private.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {
			gossiper.Messages = append(gossiper.Messages, packet.Private)
			common.LogPrivate(packet.Private)
		}

	case packet.DataReply != nil:

		destination := packet.DataReply.Destination
		hopLimit := &packet.DataReply.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {

			common.DebugReceiveDataReply(packet.DataReply)

			gossiper.Dispatcher.dispatchDataReply(packet)
		}

	case packet.DataRequest != nil:

		destination := packet.DataRequest.Destination
		hopLimit := &packet.DataRequest.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {

			common.DebugReceiveDataRequest(packet.DataRequest)

			reply, ok := gossiper.GenerateDataReply(packet.DataRequest)

			if ok {
				gossiper.sendToNode(reply.Packed(), reply.Destination, nil)
			}
		}

	case packet.SearchRequest != nil:
		
		if packet.SearchRequest.Budget <= 0 {
			return
		}

		if !gossiper.SpamDetector.shouldProcessSearchRequest(packet.SearchRequest) {
			common.DebugIgnoreSpam(packet.SearchRequest.Origin, packet.SearchRequest.Keywords)
			return
		}

		if packet.SearchRequest.Origin == gossiper.Name {
			return
		}

		common.DebugProcessSearchRequest(packet.SearchRequest.Origin, packet.SearchRequest.Keywords)

		go gossiper.forwardSearchRequest(packet.SearchRequest, source)

		results := gossiper.FileSystem.Search(packet.SearchRequest.Keywords)
		reply := common.NewSearchReply(gossiper.Name, packet.SearchRequest.Origin, results)

		go gossiper.sendToNode(reply.Packed(), reply.Destination, &(reply.HopLimit))

	case packet.SearchReply != nil:

		destination := packet.SearchReply.Destination
		hopLimit := &packet.SearchReply.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {
			go gossiper.SearchEngine.StoreResults(packet.SearchReply.Results, packet.SearchReply.Origin)
		}

	case packet.TxPublish != nil:

		if gossiper.BlockChain.tryAddFile(packet.TxPublish) {

			packet.TxPublish.HopLimit--

			if packet.TxPublish.HopLimit > 0 {
				gossiper.broadcastToNeighborsExcept(packet.TxPublish.Packed(), &[]string{source})
			}
		}
	}
}

// --
// -- SEND PACKETS TO OTHER NODES
// --

// Send a GossipPacket to any node on the network identified by name.
func (gossiper *Gossiper) sendToNode(packet *common.GossipPacket, destination string, hopLimit *uint32) bool {

	if destination == "" {
		common.DebugSendNoDestination()
	}

	if destination == gossiper.Name {

		// If it's for us, we don't need to do anything
		return true
	}

	if hopLimit != nil {

		// This is in forwarding mode, so we need to decrease the count and verify that we should still continue
		*hopLimit--

		// If it's non-zero, we forward according to the NextHop table
		if *hopLimit <= 0 {
			return false
		}
	}

	nextPeer, found := gossiper.Router.NextHop[destination]

	if found {
		common.DebugForwardPointToPoint(destination, nextPeer)
		go gossiper.sendToNeighbor(nextPeer, packet)
	} else {
		common.DebugUnknownDestination(destination)
	}

	return false
}

// Send a GossipPacket to a given neighboring node identified by IP address
func (gossiper *Gossiper) sendToNeighbor(peerAddress string, packet *common.GossipPacket) {

	if !packet.IsValid() {
		log.Panicf("Sending invalid packet: %v", packet)
	}

	bytes, err := protobuf.Encode(packet)
	if err != nil {
		panic(err)
	}

	gossiper.GossipSocket.Send(bytes, peerAddress)
}

// Broadcast a GossipPacket containing a Simple message to every neighboring node.
func (gossiper *Gossiper) broadcastToNeighborsExcept(packet *common.GossipPacket, except *[]string) {

	if packet.Simple == nil && packet.SearchRequest == nil {
		panic("Cannot broadcastToNeighbors GossipPacket that does not contain SimpleMessage or SearchRequest.")
	}

	for i := 0; i < len(gossiper.Router.Peers); i++ {

		peer := gossiper.Router.Peers[i]

		if except == nil || !common.Contains(*except, peer) {
			gossiper.sendToNeighbor(peer, packet)
		}
	}
}

func (gossiper *Gossiper) broadcastToNeighbors(packet *common.GossipPacket) {
	gossiper.broadcastToNeighborsExcept(packet, nil)
}

// --
// -- RUMORS, STATUS & DOWNLOAD
// --

// Start or continue to propagate a given rumor, by using the rumormongering algorithm
// and sending the given rumor to the given node.
//
// - If the node timeouts or sends a status that has the same IDs as us, we flip a coin
//   and continue with probability 50%.
//
// - If the node sends a status that says it is missing messages that we can send, we
//   continue the rumormongering process with the current rumor AND we start a new
//   rumormongering process with the missing rumor.
//
// - If the node sends a status that says we are missing messages, we
func (gossiper *Gossiper) rumormonger(rumor *common.RumorMessage, peer string) {

	if rumor == nil {
		panic("Cannot rumormonger with <nil> rumor!")
	}

	shouldContinue := false

	// Forward package to peer
	common.LogMongering(peer)
	go gossiper.sendToNeighbor(peer, rumor.Packed())

	// Start timer
	ticker := time.NewTicker(common.StatusTimeout)
	defer ticker.Stop()

	select {
	case packet := <-gossiper.Dispatcher.statusPackets(peer):

		statusPacket := packet.Status

		// Compare status from peer with own messages
		otherRumor, _, statuses := gossiper.CompareStatus(statusPacket.Want, ComparisonModeMissingOrNew)

		switch {
		case statuses != nil: // Peer has new messages
			statusPacket := &common.StatusPacket{statuses}
			go gossiper.sendToNeighbor(peer, statusPacket.Packed())
			shouldContinue = true

		case otherRumor != nil: // Peer is missing messages

			go gossiper.rumormonger(otherRumor, peer)
			shouldContinue = true

		default:
			common.LogInSyncWith(peer)
			shouldContinue = false
		}

	case <-ticker.C: // Timeout
		common.DebugTimeout(peer)
		shouldContinue = false
	}

	gossiper.Dispatcher.stopWaitingOnStatusPacket(peer)

	newPeer, found := gossiper.Router.randomPeer()

	if !found {
		return
	}

	if !shouldContinue && common.FlipCoin() {
		common.LogFlippedCoin(newPeer)
		shouldContinue = true
	}

	if shouldContinue {
		gossiper.rumormonger(rumor, newPeer)
	} else {
		common.DebugStopMongering(rumor)
	}
}

// Start a new download process. As long as there are missing chunks related to the metahash provided, it will
// continue to download the remaining ones. If there is a timeout or the response does not match the data requested,
// it will restart evert 5 seconds, up to a total of 10 tries before stopping.
//
//  - name: Name of the file as it will appear in the file system later on
//  - metaHash: Hash of the requested file
//  - peer: Name of the peer from which we want to download the file
//  - counter: Set to 0 for new downloads, it will increase up to 10 until timing out.
//
func (gossiper *Gossiper) StartDownload(name string, metaHash []byte, peer string, counter int) {

	if counter > common.MaxDownloadRequests {
		// Probably print some stuff
		common.DebugDownloadTimeout(name, metaHash, peer)
		return
	}

	nextHash, chunkId, completed := gossiper.FileSystem.downloadStatus(metaHash)

	if completed {
		common.DebugDownloadCompleted(name, metaHash, peer)
		return
	}

	if peer == "" {

		fileMap, found := gossiper.SearchEngine.FileMap(metaHash)

		if !found {
			common.DebugDownloadUnknownFile(metaHash)
			return
		}

		peer, found = fileMap.peerForChunk(uint64(chunkId), counter)

		if !found {
			common.DebugNoKnownOwnerForFile(metaHash)
			return
		}
	}

	common.DebugStartDownload(name, nextHash, peer)

	go func() {

		ticker := time.NewTicker(common.DownloadTimeout)
		defer ticker.Stop()

		select {
		case packet := <-gossiper.Dispatcher.dataReplies(nextHash):

			reply := packet.DataReply

			if reply == nil {
				// Error, we expect only a data reply from this
				return
			}

			if !reply.VerifyHash(nextHash) {
				// Unexpected hash or incorrect data, retry
				common.DebugCorruptedDataReply(nextHash, reply)
				go gossiper.StartDownload(name, metaHash, peer, counter+1)
				return
			}

			gossiper.FileSystem.processDataReply(name, metaHash, reply)

			// At this point, the download is successful, so we can log it

			if chunkId == common.MetaHashChunkId {
				common.LogDownloadingMetafile(name, packet.DataReply.Origin)
			} else {
				common.LogDownloadingChunk(name, chunkId, packet.DataReply.Origin)
			}

			// Then we start downloading the next chunk
			go gossiper.StartDownload(name, metaHash, peer, 0)

		case <-ticker.C: // Timeout
			go gossiper.StartDownload(name, metaHash, peer, counter+1)
		}

		gossiper.Dispatcher.stopWaitingOnDataReply(nextHash)
	}()

	request := gossiper.GenerateDataRequest(peer, nextHash)
	gossiper.sendToNode(request.Packed(), request.Destination, nil)

}

// Compare the gossiper's vector clock with another one. There are two possible modes for this:
//
// - "Missing or New" mode:
//   This is the default mode and the one to use when comparing vector clocks of two gossiper nodes.
//
//    - If the two vector clocks are the same, return nil, nil, nil.
//    - If the other vector clock is missing message, send the first rumor only as first return value.
//    - If the other vector clock has all our messages, but we see that they have messages we don't have
//      return our vector clock as third return value.
//
// - "AllNew" mode:
//   This is the mode to use when comparing a gossiper's vector clock with a server / client one.
//
//    - If the other vector clock is missing messages, send ALL those messages as a second return value.
//    - Otherwise, return nil, nil, nil.
//
func (gossiper *Gossiper) CompareStatus(statuses []common.PeerStatus, mode int) (*common.RumorMessage, []*common.RumorMessage, []common.PeerStatus) {

	if mode > ComparisonModeAllNew || mode < ComparisonModeMissingOrNew {
		mode = ComparisonModeAllNew
	}

	// First, we generate a statusPacket based on our rumor list
	myStatuses := gossiper.GenerateStatusPacket().Want

	// This map should store, for each node we know about, what is the nextID we want
	myNextIDs := make(map[string]uint32)

	// Should become true if during the process somewhere we saw a message that we do not yet have
	rumorsWanted := false

	//
	var allRumors []*common.RumorMessage

	if mode == ComparisonModeAllNew {
		allRumors = make([]*common.RumorMessage, 0)
	}

	for _, myStatus := range myStatuses {
		myNextIDs[myStatus.Identifier] = myStatus.NextID
	}

	for _, theirStatus := range statuses {

		theirNextID := theirStatus.NextID

		// In case someone sends something smaller than
		// possible, we fail gracefully
		if theirNextID < common.InitialId {
			return nil, nil, nil
		}

		myNextID, found := myNextIDs[theirStatus.Identifier]

		switch {

		case !found && theirNextID != common.InitialId:
			// They know about an origin node we don't know.
			// We make sure that they are not looking for the first message
			// because in this case they cannot send us anything.
			rumorsWanted = true

		case found && myNextID < theirNextID:
			// They have a message we don't
			rumorsWanted = true

		case found && myNextID > theirNextID:

			if mode == ComparisonModeMissingOrNew {
				return gossiper.Rumors.Get(theirStatus.Identifier, theirNextID), nil, nil
			} else {
				for i := theirNextID; i < myNextID; i++ {
					rumor := gossiper.Rumors.Get(theirStatus.Identifier, i)
					allRumors = append(allRumors, rumor)
				}
			}
		}

		// We remove the ID from our NextIDs map to keep track of the fact
		// that we have seen this ID already.
		if found {
			delete(myNextIDs, theirStatus.Identifier)
		}
	}

	// After comparing with all their IDs, if there is still some value
	// in myNextIDs, it means that they don't know about such origin nodes
	for identifier, myNextID := range myNextIDs {

		// If we are also waiting for the first message,
		// just skip this one, we cannot send anything.
		if myNextID == common.InitialId {
			continue
		}

		if mode == ComparisonModeMissingOrNew {
			// Get first rumor from this node
			return gossiper.Rumors.Get(identifier, common.InitialId), nil, nil
		} else {
			for i := common.InitialId; i < myNextID; i++ {
				rumor := gossiper.Rumors.Get(identifier, i)
				allRumors = append(allRumors, rumor)
			}
		}
	}

	// If we did not return already with a rumor to send, and we want Rumors,
	// we simply return our status packet to notify.
	if rumorsWanted && mode == ComparisonModeMissingOrNew {
		return nil, nil, myStatuses
	}

	if mode == ComparisonModeMissingOrNew {
		// In "Missing or New" mode, we simply return all nil to indicate
		// that the two statusPackets are equivalent.
		return nil, nil, nil
	} else {
		// In "All New" mode, we return all new messages alongside our
		// status packet so that the caller can store it for next time
		return nil, allRumors, myStatuses
	}
}

// --
// -- GENERATING NEW PACKETS
// --

// Generate a status packet with the current vector clock.
func (gossiper *Gossiper) GenerateStatusPacket() *common.StatusPacket {

	peerStatuses := make([]common.PeerStatus, 0)

	for _, origin := range gossiper.Rumors.AllOrigins() {
		peerStatuses = append(peerStatuses, common.PeerStatus{origin, gossiper.Rumors.NextIDFor(origin)})
	}

	return &common.StatusPacket{peerStatuses}
}

// Generate a data reply to a given request
func (gossiper *Gossiper) GenerateDataReply(request *common.DataRequest) (*common.DataReply, bool) {

	var data []byte

	metaHash, found := gossiper.FileSystem.getMetaFile(request.HashValue)

	if found {

		data = metaHash.Data

	} else {

		chunk, found := gossiper.FileSystem.getChunk(request.HashValue)

		if !found {
			common.DebugHashNotFound(request.HashValue, request.Origin)
			return nil, false
		}

		data = chunk.data
	}

	return &common.DataReply{
		gossiper.Name,
		request.Origin,
		common.InitialHopLimit,
		request.HashValue,
		data,
	}, true
}

// Generate a data request to a given file
func (gossiper *Gossiper) GenerateDataRequest(destination string, hash []byte) *common.DataRequest {
	return &common.DataRequest{
		gossiper.Name,
		destination,
		common.InitialHopLimit,
		hash,
	}
}

// Generate a new Rumor based on the string.
func (gossiper *Gossiper) generateRumor(message string) *common.RumorMessage {

	rumor := &common.RumorMessage{
		Origin: gossiper.Name,
		ID:     gossiper.Rumors.ConsumeNextID(),
		Text:   message,
	}

	gossiper.Rumors.Put(rumor)

	return rumor
}

// Generate a route rumor
func (gossiper *Gossiper) GenerateRouteRumor() *common.RumorMessage {
	return gossiper.generateRumor("")
}
