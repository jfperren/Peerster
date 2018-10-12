package main

import (
	"github.com/dedis/protobuf"
	"strings"
	"sync"
	"time"
)

type Gossiper struct {
	simple          bool
	gossipSocket    *UDPSocket
	clientSocket    *UDPSocket
	gossipAddress   string
	clientAddress   string
	peers           []string
	Name            string
	handlers        map[string]chan*StatusPacket
	rumors          *RumorMessages
	NextID			uint32
}


func NewGossiper(gossipAddress, clientPort, name string, peers string, simple bool) *Gossiper {

	clientAddress := ":" + clientPort

	gossipSocket := NewUDPSocket(gossipAddress)
	clientSocket := NewUDPSocket(clientAddress)

	return &Gossiper{
		simple:         simple,
		gossipSocket:   gossipSocket,
		clientSocket:   clientSocket,
		gossipAddress:  gossipAddress,
		clientAddress:  clientAddress,
		Name:           name,
		peers:          strings.Split(peers, ","),
		handlers:       make(map[string]chan*StatusPacket),
		rumors:         makeRumors(),
		NextID:			0,
	}
}

func (gossiper *Gossiper) start() {

	go func() {

		for {
			var packet GossipPacket
			bytes, _, alive := gossiper.clientSocket.Receive()

			if !alive { break }

			protobuf.Decode(bytes, &packet)

			if !packet.isValid() {
				panic("Received invalid packet")
			}

			gossiper.handleClient(&packet)
		}
	}()

	go func() {

		for {
			var packet GossipPacket
			bytes, source, alive := gossiper.gossipSocket.Receive()

			if !alive { break }

			protobuf.Decode(bytes, &packet)

			if !packet.isValid() {
				panic("Received invalid packet")
			}

			gossiper.handleGossip(&packet, source)
		}
	}()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	wg.Wait()
}

func (gossiper *Gossiper) stop() {
	gossiper.clientSocket.Unbind()
	gossiper.gossipSocket.Unbind()
}

/// Sends to one peer
func (gossiper *Gossiper) sendTo(peerAddress string, packet *GossipPacket) {

	if !packet.isValid() {
		panic("Sending invalid packet")
	}

	bytes, err := protobuf.Encode(packet)
	if err != nil { panic(err) }

	gossiper.gossipSocket.Send(bytes, peerAddress)
}

/// Sends to every peer
func (gossiper *Gossiper) relay(packet *GossipPacket, setName bool) {

	if packet.Simple == nil {
		panic("Cannot relay GossipPacker that does not contain SimpleMessage.")
	}

	packet.Simple.RelayPeerAddr = gossiper.gossipAddress
	if setName { packet.Simple.OriginalName = gossiper.Name }

	for i := 0; i < len(gossiper.peers); i++ {
		if gossiper.peers[i] != packet.Simple.RelayPeerAddr {
			gossiper.sendTo(gossiper.peers[i], packet)
		}
	}
}


func (gossiper *Gossiper) handleClient(packet *GossipPacket) {

	if packet.Simple == nil {
		panic("Client should send SimpleMessage only.")
	}

	logClientMessage(packet.Simple)
	logPeers(gossiper.peers)

	if gossiper.simple {
		go gossiper.relay(packet, true)
	} else {

		rumor := gossiper.generateRumor(packet.Simple.Contents)

		go gossiper.rumors.put(rumor)
		go gossiper.rumormonger(rumor, nil)
	}
}

func (gossiper *Gossiper) handleGossip(packet *GossipPacket, source string) {

	gossiper.addPeerIfNeeded(source)

	switch {

	case packet.Simple != nil:

		logSimpleMessage(packet.Simple)
		logPeers(gossiper.peers)

		go gossiper.relay(packet, false)

	case packet.Rumor != nil:

		logRumor(packet.Rumor, source)
		logPeers(gossiper.peers)

		if !gossiper.rumors.contains(packet.Rumor) {
			debugForwardRumor(packet.Rumor)
			gossiper.rumors.put(packet.Rumor)
			go gossiper.rumormonger(packet.Rumor, nil)
		}

		statusPacket := gossiper.generateStatusPacket()
		debugSendStatus(statusPacket, source)
		go gossiper.sendTo(source, statusPacket.packed())

	case packet.Status != nil:

		logStatus(packet.Status, source)
		logPeers(gossiper.peers)

		expected := gossiper.dispatchStatusPacket(source, packet.Status)
		if !expected {
			rumor, _ := gossiper.compareStatus(packet.Status)
			go gossiper.rumormonger(rumor, &source)
		}
	}
}

func (gossiper *Gossiper) rumormonger(rumor *RumorMessage, peer *string) {

	var shouldContinue bool

	if peer == nil {
		index := selectRandom(gossiper.peers)
		peer = &index
	}

	// Prepare channel to receive ack
	peerStatus := gossiper.awaitStatusPacket(*peer)

	// Forward package to peer
	logMongering(*peer)
	go gossiper.sendTo(*peer, rumor.packed())

	// Start timer
	ticker := time.NewTicker(ACK_TIMEOUT * time.Second)
	defer ticker.Stop()

	select {
	case statusPacket := <- peerStatus: // Received ACK

		// Compare status from peer with own messages
		otherRumor, status := gossiper.compareStatus(statusPacket)

		switch  {
		case status != nil: // Peer has new messages

			go gossiper.sendTo(*peer, status.packed())
			shouldContinue = true

		case otherRumor != nil: // Peer is missing messages

			go gossiper.rumormonger(otherRumor, peer)
			shouldContinue = true

		default:
			logInSyncWith(*peer)
			shouldContinue = flipCoin()
		}

	case <- ticker.C: // Timeout
		debugTimeout(*peer)
		shouldContinue = flipCoin()
	}

	if shouldContinue {
		index := selectRandom(gossiper.peers)
		peer = &index

		logFlippedCoin(*peer)
		go gossiper.rumormonger(rumor, peer)
	} else {
		debugStopMongering(rumor)
	}
}

/*
 * Sends a status packet to potential handlers.
 * @return `True` if the packet was consumed.
 */
func (gossiper *Gossiper) dispatchStatusPacket(source string, statusPacket *StatusPacket) bool {
	channel, found := gossiper.handlers[source]
	if found {
		channel <- statusPacket
	}
	return found
}

func (gossiper *Gossiper) awaitStatusPacket(peer string) chan *StatusPacket {
	channel := make(chan *StatusPacket)
	gossiper.handlers[peer] = channel
	return channel
}

func (gossiper *Gossiper) stopWaitForStatusPacket(peer string) {
	gossiper.handlers[peer] = nil
}

func (gossiper *Gossiper) compareStatus(statusPacket *StatusPacket) (*RumorMessage, *StatusPacket) {

	// TODO: Cover case where someone else's status packet contains node that we don't know but NextID is 0

	// First, we generate a statusPacket based on our rumor list
	myStatusPacket := gossiper.generateStatusPacket()

	// This map should store, for each node we know about, what is the nextID we want
	myNextIDs := make(map[string]uint32)

	// Should become true if during the process somewhere we saw a message that we do not yet have
	rumorsWanted := false

	for _, want := range myStatusPacket.Want {
		myNextIDs[want.Identifier] = want.NextID
	}

	for _, want := range statusPacket.Want {

		theirNextID := want.NextID
		myNextID, found := myNextIDs[want.Identifier]



		switch {

		case !found && theirNextID != 0:
			// They know about an origin node we don't know.
			// We make sure that they are not looking for the first message
			// because in this case they cannot send us anything.
			rumorsWanted = true

		case myNextID < theirNextID:
			// They have a message we don't
			rumorsWanted = true

		case myNextID > theirNextID:
			return gossiper.rumors.get(want.Identifier, theirNextID), nil
		}

		// We remove the ID from our NextIDs map to keep track of the fact
		// that we have seen this ID already.
		if found {
			delete(myNextIDs, want.Identifier)
		}
	}

	// After comparing with all their IDs, if there is still some value
	// in myNextIDs, it means that they don't know about such origin nodes
	for identifier, nextID := range myNextIDs {

		// If we are also waiting for the first message,
		// just skip this one, we cannot send anything.
		if nextID == 0 {
			continue
		}

		// Return the first rumor from this node
		return gossiper.rumors.get(identifier, 0), nil
	}

	// If we did not return already with a rumor to send, and we want rumors,
	// we simply return our status packet to notify.
	if rumorsWanted {
		return nil,  myStatusPacket
	}

	// If the two statusPackets are equivalent, we simply return nil
	return nil, nil
}

func (gossiper *Gossiper) addPeerIfNeeded(peer string) {

	if !containsString(gossiper.peers, peer) {
		gossiper.peers = append(gossiper.peers, peer)
	}
}

func (gossiper *Gossiper) generateStatusPacket() *StatusPacket {

	peerStatuses := make([]PeerStatus, 0)

	for _, origin := range gossiper.rumors.allOrigins() {
		peerStatuses = append(peerStatuses, PeerStatus{origin, gossiper.rumors.nextIDFor(origin)})
	}

	return &StatusPacket{peerStatuses}
}

func (gossiper *Gossiper) generateRumor(message string) *RumorMessage {

	rumor := &RumorMessage{
		gossiper.Name,
		gossiper.NextID,
		message,
	}

	gossiper.NextID++

	return rumor
}