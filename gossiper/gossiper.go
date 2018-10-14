package gossiper

import (
	"github.com/dedis/protobuf"
	"github.com/jfperren/Peerster/common"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Gossiper struct {

	Simple        	bool								// Stores if gossiper runs in simple mode.
	GossipSocket  	*common.UDPSocket					// UDP Socket that connects to other nodes
	ClientSocket  	*common.UDPSocket					// UDP Socket that connects to the client
	GossipAddress 	string								// IP address used in GossipSocket
	ClientAddress 	string								// IP address used in ClientSocket
	Peers         	[]string							// List of known peer IP addresses
	Name          	string								// Name of this node
	Rumors 			*RumorDatabase						// Database of known Rumors
	NextID uint32										// NextID to be used for Rumors
	HandlerLock  *sync.RWMutex							// Lock for safely updating & reading handlers
	Handlers     map[string]chan*common.StatusPacket	// Channels waiting for StatusPackets
	HandlerCount map[string]int							// Count of rumormongering processes waiting on a node status
}

const (
	ComparisonModeMissingOrNew = iota					// Flag to be used when comparing two nodes' status packets
	ComparisonModeAllNew = iota							// Flag to be used when comparing a node status with the client status
)


// Create a new Gossiper using the given addresses. Use gossiper.Start()
// to Start listening for messages
func NewGossiper(gossipAddress, clientAddress, name string, peers string, simple bool) *Gossiper {

	gossipSocket := common.NewUDPSocket(gossipAddress)
	var clientSocket *common.UDPSocket

	if clientAddress != "" {
		clientSocket = common.NewUDPSocket(clientAddress)
	}

	return &Gossiper{
		Simple:        simple,
		GossipSocket:  gossipSocket,
		ClientSocket:  clientSocket,
		GossipAddress: gossipAddress,
		ClientAddress: clientAddress,
		Name:          name,
		Peers:         strings.Split(peers, ","),
		Rumors:        MakeRumorDatabase(),
		NextID:        common.InitialId,
		HandlerLock:   &sync.RWMutex{},
		Handlers:      make(map[string]chan*common.StatusPacket),
		HandlerCount:  make(map[string]int),
	}
}

// --
// --  START & STOP
// --

// Start listening for UDP packets on Gossiper's clientAddress & gossipAddress
func (gossiper *Gossiper) Start() {

	go gossiper.gossip()

	if !gossiper.Simple {
		go gossiper.antiEntropy()
	}

	if gossiper.ClientSocket != nil {
		go gossiper.client()
	}

	// Allows the loops to run indefinitely after the main code is completed.
	wg := new(sync.WaitGroup)
	wg.Add(3)
	wg.Wait()
}

// Main loop for handling gossip packets from other nodes.
func (gossiper *Gossiper) gossip() {
	for {
		var packet common.GossipPacket
		bytes, source, alive := gossiper.GossipSocket.Receive()

		if !alive { break }

		protobuf.Decode(bytes, &packet)

		if !packet.IsValid() {
			panic("Received invalid packet")
		}

		go gossiper.HandleGossip(&packet, source)
	}
}

// Main loop for pinging other nodes as part of the anti-entropy algorithm.
func (gossiper *Gossiper) antiEntropy() {
	for {
		peer, found := gossiper.randomPeer()

		if found {
			packet := gossiper.GenerateStatusPacket().Packed()
			common.DebugAskAndSendStatus(packet.Status, peer)
			go gossiper.sendTo(peer, packet)
		}

		time.Sleep(common.AntiEntropyDT)
	}
}

// Main loop for handling client packets.
func (gossiper *Gossiper) client(){
	for {

		var packet common.GossipPacket
		bytes, _, alive := gossiper.ClientSocket.Receive()

		if !alive { break }

		protobuf.Decode(bytes, &packet)

		if !packet.IsValid() {
			panic("Received invalid packet")
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

	if packet == nil || packet.Simple == nil {
		return // Fail gracefully
	}

	common.LogClientMessage(packet.Simple)
	common.LogPeers(gossiper.Peers)

	if gossiper.Simple {
		go gossiper.relay(packet, true)
	} else {

		rumor := gossiper.generateRumor(packet.Simple.Contents)

		gossiper.Rumors.Put(rumor)

		peer, found := gossiper.randomPeer()

		if found {
			go gossiper.rumormonger(rumor, peer)
		}
	}
}

// Handle packet from another node.
func (gossiper *Gossiper) HandleGossip(packet *common.GossipPacket, source string) {

	if packet == nil || !packet.IsValid() {
		return // Fail gracefully
	}

	gossiper.AddPeerIfNeeded(source)

	switch {

	case packet.Simple != nil:

		common.LogSimpleMessage(packet.Simple)
		common.LogPeers(gossiper.Peers)

		go gossiper.relay(packet, false)

	case packet.Rumor != nil:

		common.LogRumor(packet.Rumor, source)
		common.LogPeers(gossiper.Peers)

		// We only store & forward if the rumor is our next expected rumo
		// from the source.
		if gossiper.Rumors.Expects(packet.Rumor) {

			gossiper.Rumors.Put(packet.Rumor)
			peer, found := gossiper.randomPeer()

			if found {
				common.DebugForwardRumor(packet.Rumor)
				go gossiper.rumormonger(packet.Rumor, peer)
			}
		}

		statusPacket := gossiper.GenerateStatusPacket()
		common.DebugSendStatus(statusPacket, source)
		go gossiper.sendTo(source, statusPacket.Packed())

	case packet.Status != nil:

		common.LogStatus(packet.Status, source)
		common.LogPeers(gossiper.Peers)

		expected := gossiper.dispatchStatusPacket(source, packet.Status)

		if !expected {

			rumor, _, _ := gossiper.CompareStatus(packet.Status.Want, ComparisonModeMissingOrNew)

			if rumor != nil {
				go gossiper.rumormonger(rumor, source)
			}
		}
	}
}

// --
// -- SEND PACKETS TO OTHER NODES
// --

// Send a GossipPacket to a given node
func (gossiper *Gossiper) sendTo(peerAddress string, packet *common.GossipPacket) {

	if !packet.IsValid() {
		log.Panicf("Sending invalid packet: %v", packet)
	}

	bytes, err := protobuf.Encode(packet)
	if err != nil { panic(err) }

	gossiper.GossipSocket.Send(bytes, peerAddress)
}

// Relay a GossipPacket containing a Simple message to every known peer.
func (gossiper *Gossiper) relay(packet *common.GossipPacket, setName bool) {

	if packet.Simple == nil {
		panic("Cannot relay GossipPacker that does not contain SimpleMessage.")
	}

	packet.Simple.RelayPeerAddr = gossiper.GossipAddress
	if setName { packet.Simple.OriginalName = gossiper.Name }

	for i := 0; i < len(gossiper.Peers); i++ {
		if gossiper.Peers[i] != packet.Simple.RelayPeerAddr {
			gossiper.sendTo(gossiper.Peers[i], packet)
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
	go gossiper.sendTo(peer, rumor.Packed())

	// Start timer
	ticker := time.NewTicker(common.StatusTimeout)
	defer ticker.Stop()

	select {
	case statusPacket := <- gossiper.statusPacketsFrom(peer):

		// Compare status from peer with own messages
		otherRumor, _, statuses := gossiper.CompareStatus(statusPacket.Want, ComparisonModeMissingOrNew)

		switch  {
		case statuses != nil: // Peer has new messages
			statusPacket := &common.StatusPacket{statuses}
			go gossiper.sendTo(peer, statusPacket.Packed())
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

	gossiper.stopWaitingFrom(peer)

	newPeer, found := gossiper.randomPeer()

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

 // Dispatch a status packet to potential Handlers. Return true if the status packet was expected by a
 // rumormongering process.
func (gossiper *Gossiper) dispatchStatusPacket(source string, statusPacket *common.StatusPacket) bool {

	gossiper.HandlerLock.RLock()
	defer gossiper.HandlerLock.RUnlock()

	expected := gossiper.HandlerCount[source] > 0

	if expected {
		gossiper.Handlers[source] <- statusPacket
	}

	return expected
}

// Create / return a channel that will allow to receive status packets from a given node.
// Calling this function implicitly register that a new rumormongering process is waiting
// on a status packet. To indicate that this is no longer the case, call stopWaitingFrom(peer).
func (gossiper *Gossiper) statusPacketsFrom(peer string) chan *common.StatusPacket {

	gossiper.HandlerLock.Lock()
	defer gossiper.HandlerLock.Unlock()

	_, found := gossiper.Handlers[peer]

	if !found {
		gossiper.Handlers[peer] = make(chan *common.StatusPacket, common.StatusBufferSize)
	}

	gossiper.HandlerCount[peer] = gossiper.HandlerCount[peer] + 1

	return gossiper.Handlers[peer]
}

// Explicitly state that a given rumormongering process is no longer waiting for a status packet.
func (gossiper *Gossiper) stopWaitingFrom(peer string) {

	gossiper.HandlerLock.Lock()
	defer gossiper.HandlerLock.Unlock()

	count := gossiper.HandlerCount[peer]

	if count > 0 {
		count = count - 1
	} else {
		count = 0
	}

	gossiper.HandlerCount[peer] = count
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

func (gossiper *Gossiper) GenerateStatusPacket() *common.StatusPacket {

	peerStatuses := make([]common.PeerStatus, 0)

	for _, origin := range gossiper.Rumors.AllOrigins() {
		peerStatuses = append(peerStatuses, common.PeerStatus{origin, gossiper.Rumors.NextIDFor(origin)})
	}

	return &common.StatusPacket{peerStatuses}
}

// --
// -- PEERS
// --

// Add a new peer IP address to the list of known peers
func (gossiper *Gossiper) AddPeerIfNeeded(peer string) {

	if !common.Contains(gossiper.Peers, peer) {
		gossiper.Peers = append(gossiper.Peers, peer)
	}
}

func (gossiper *Gossiper) randomPeer() (string, bool) {

	if len(gossiper.Peers) == 0 {
		return "", false
	}

	return gossiper.Peers[rand.Int() % len(gossiper.Peers)], true
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
