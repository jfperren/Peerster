package gossiper

import (
	"github.com/dedis/protobuf"
	"github.com/jfperren/Peerster/common"
	"log"
	"sync"
	"time"
)

type Gossiper struct {

	Simple        	bool								// Stores if gossiper runs in simple mode.
	GossipSocket  	*common.UDPSocket					// UDP Socket that connects to other nodes
	ClientSocket  	*common.UDPSocket					// UDP Socket that connects to the client
	GossipAddress 	string								// IP address used in GossipSocket
	ClientAddress 	string								// IP address used in ClientSocket

	Name          	string								// Name of this node
	Rumors 			*RumorDatabase						// Database of known Rumors
	NextID 			uint32								// NextID to be used for Rumors
	Rtimer			time.Duration						// Interval for sending route rumors

	FileSystem 		*FileSystem
	Dispatcher 		*Dispatcher
	Router			*Router
}

const (
	ComparisonModeMissingOrNew = iota					// Flag to be used when comparing two nodes' status packets
	ComparisonModeAllNew = iota							// Flag to be used when comparing a node status with the client status
)


// Create a new Gossiper using the given addresses. Use gossiper.Start()
// to Start listening for messages
func NewGossiper(gossipAddress, clientAddress, name string, peers string, simple bool, rtimer int) *Gossiper {

	gossipSocket := common.NewUDPSocket(gossipAddress)
	var clientSocket *common.UDPSocket

	if clientAddress != "" {
		clientSocket = common.NewUDPSocket(clientAddress)
	}

	return &Gossiper{
		Simple:        	simple,
		GossipSocket:  	gossipSocket,
		ClientSocket:  	clientSocket,
		GossipAddress: 	gossipAddress,
		ClientAddress: 	clientAddress,
		Name:          	name,

		Rumors:        	NewRumorDatabase(),
		NextID:        	common.InitialId,

		Rtimer:		   	time.Duration(rtimer) * time.Second,
		FileSystem:	   	NewFileSystem(
			common.SharedFilesDir + name + "/",
			common.DownloadDir + name + "/"),
		Dispatcher:		NewDispatcher(),
		Router:			NewRouter(peers),
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

// Main loop for handling gossip packets from other nodes.
func (gossiper *Gossiper) receiveGossip() {
	for {
		var packet common.GossipPacket
		bytes, source, alive := gossiper.GossipSocket.Receive()

		if !alive { break }

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

	if gossiper.Rtimer == common.DontSendRouteRumor {
		return
	}

	for {
		peer, found := gossiper.Router.randomPeer()

		if found {
			routeRumor := gossiper.GenerateRouteRumor()
			common.DebugSendRouteRumor(peer)
			go gossiper.sendToNeighbor(peer, routeRumor.Packed())
		}

		time.Sleep(gossiper.Rtimer)
	}
}

// Main loop for handling client packets.
func (gossiper *Gossiper) receiveClient(){

	for {

		var packet common.GossipPacket
		bytes, _, alive := gossiper.ClientSocket.Receive()

		if !alive { break }

		protobuf.Decode(bytes, &packet)

		if !packet.IsValid() {
			// Fail gracefully
			continue
		}

		go gossiper.HandleClient(&packet)
	}
}

// Unbind from all ports, stop processes.
func (gossiper *Gossiper) Stop() {
	gossiper.ClientSocket.Unbind()
	gossiper.GossipSocket.Unbind()
}

// --
// --  HANDLING NEW PACKETS
// --

//  Handle new packet from client
func (gossiper *Gossiper) HandleClient(packet *common.GossipPacket) {

	if packet == nil || !packet.IsValid() {
		return // Fail gracefully
	}

	switch {

	case packet.Simple != nil:

		common.LogClientMessage(packet.Simple)

		if gossiper.Simple {

			go gossiper.broadcastToNeighbors(packet, true)

		} else {

			rumor := gossiper.generateRumor(packet.Simple.Contents)

			gossiper.Rumors.Put(rumor)

			peer, found := gossiper.Router.randomPeer()

			if found {
				go gossiper.rumormonger(rumor, peer)
			}
		}

	case packet.Private != nil:

		// Replace origin with gossiper's name
		packet.Private.Origin = gossiper.Name

		destined := gossiper.sendToNode(packet, packet.Private.Destination, nil)

		if destined {
			common.LogPrivate(packet.Private)
		}

	case packet.DataRequest != nil:

		destination := packet.DataRequest.Destination
		filename := packet.DataRequest.Origin
		hash := packet.DataRequest.HashValue

		go gossiper.StartDownload(filename, hash, destination, 0)

	case packet.DataReply != nil:

		// By convention, we use DataReply objects with destination set as filename as a way for the client
		// to tell which file should be uploaded onto the network.
		filename := packet.DataReply.Destination

		gossiper.FileSystem.ScanFile(filename)
	}
}

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

	common.DebugStartDownload(name, nextHash, peer)

	go func() {

		ticker := time.NewTicker(common.DonwloadTimeout)
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

			stored := gossiper.FileSystem.processDataReply(name, metaHash, reply)

			if !stored {
				// There was an error,
				panic("Error storing data")
			}

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

// Handle packet from another node.
func (gossiper *Gossiper) HandleGossip(packet *common.GossipPacket, source string) {

	if packet == nil || !packet.IsValid() {
		return // Fail gracefully
	}

	switch {

	case packet.Simple != nil:

		common.LogSimpleMessage(packet.Simple)
		common.LogPeers(gossiper.Router.Peers)

		go gossiper.broadcastToNeighbors(packet, false)

	case packet.Rumor != nil:

		if packet.Rumor.IsRouteRumor() {
			common.DebugReceiveRouteRumor(packet.Rumor.Origin, source)
		} else {
			common.LogRumor(packet.Rumor, source)
		}

		common.LogPeers(gossiper.Router.Peers)

		// Update routing table
		gossiper.Router.updateRoutingTable(packet.Rumor.Origin, source)

		// We only store & forward if the rumor is our next expected rumo
		// from the source.
		if gossiper.Rumors.Expects(packet.Rumor) {

			gossiper.Rumors.Put(packet.Rumor)
			peer, found := gossiper.Router.randomPeer()

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
	if err != nil { panic(err) }

	gossiper.GossipSocket.Send(bytes, peerAddress)
}

// Broadcast a GossipPacket containing a Simple message to every neighboring node.
func (gossiper *Gossiper) broadcastToNeighbors(packet *common.GossipPacket, setName bool) {

	if packet.Simple == nil {
		panic("Cannot broadcastToNeighbors GossipPacker that does not contain SimpleMessage.")
	}

	packet.Simple.RelayPeerAddr = gossiper.GossipAddress
	if setName { packet.Simple.OriginalName = gossiper.Name }

	for i := 0; i < len(gossiper.Router.Peers); i++ {
		if gossiper.Router.Peers[i] != packet.Simple.RelayPeerAddr {
			gossiper.sendToNeighbor(gossiper.Router.Peers[i], packet)
		}
	}
}

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
	case packet := <- gossiper.Dispatcher.statusPackets(peer):

		statusPacket := packet.Status

		// Compare status from peer with own messages
		otherRumor, _, statuses := gossiper.CompareStatus(statusPacket.Want, ComparisonModeMissingOrNew)

		switch  {
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

	case <- ticker.C: // Timeout
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

// --
// -- HANDLING STATUS PACKETS AND VECTOR CLOCKS
// --

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

func (gossiper *Gossiper) GenerateStatusPacket() *common.StatusPacket {

	peerStatuses := make([]common.PeerStatus, 0)

	for _, origin := range gossiper.Rumors.AllOrigins() {
		peerStatuses = append(peerStatuses, common.PeerStatus{origin, gossiper.Rumors.NextIDFor(origin)})
	}

	return &common.StatusPacket{peerStatuses}
}

func (gossiper *Gossiper) GenerateRouteRumor() *common.RumorMessage {
	return &common.RumorMessage{gossiper.Name, gossiper.NextID, ""}
}

func (gossiper *Gossiper) GenerateDataReply(request *common.DataRequest) (*common.DataReply, bool) {

	var data []byte

	metaHash, found := gossiper.FileSystem.getMetaFile(request.HashValue)

	if found {

		data = metaHash.data

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

func (gossiper *Gossiper) GenerateDataRequest(destination string, hash []byte) *common.DataRequest {
	return &common.DataRequest{
		gossiper.Name,
		destination,
		common.InitialHopLimit,
		hash,
	}
}

// --
// -- RUMORS
// --

// Generate a new Rumor based on the string.
func (gossiper *Gossiper) generateRumor(message string) *common.RumorMessage {

	rumor := &common.RumorMessage{
		Origin: gossiper.Name,
		ID:     gossiper.NextID,
		Text:   message,
	}

	gossiper.NextID++

	return rumor
}
