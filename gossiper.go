package main

import (
	"fmt"
	"github.com/dedis/protobuf"
	"log"
	"math/rand"
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

	rumors          *RumorMessages
	NextID			uint32

	handlerLock		*sync.RWMutex
	handlers        map[string]chan*StatusPacket
	handlerCount	map[string]int
}

const (
	ComparisonModeMissingOrNew = iota
	ComparisonModeAllNew = iota
)


func NewGossiper(gossipAddress, clientAddress, name string, peers string, simple bool) *Gossiper {

	gossipSocket := NewUDPSocket(gossipAddress)
	var clientSocket *UDPSocket

	if clientAddress != "" {
		clientSocket = NewUDPSocket(clientAddress)
	}

	return &Gossiper{
		simple:         simple,
		gossipSocket:   gossipSocket,
		clientSocket:   clientSocket,
		gossipAddress:  gossipAddress,
		clientAddress:  clientAddress,
		Name:           name,
		peers:          strings.Split(peers, ","),
		rumors:         makeRumors(),
		NextID:			INITIAL_ID,
		handlerLock:	&sync.RWMutex{},
		handlers:       make(map[string]chan*StatusPacket),
		handlerCount:	make(map[string]int),
	}
}

func (gossiper *Gossiper) start() {

	go func() {

		for {
			var packet GossipPacket
			bytes, source, alive := gossiper.gossipSocket.Receive()

			if !alive { break }

			protobuf.Decode(bytes, &packet)

			if !packet.isValid() {
				panic("Received invalid packet")
			}

			go gossiper.handleGossip(&packet, source)
		}
	}()

	go func() {

		for {

			peer, found := gossiper.randomPeer()

			if found {
				packet := gossiper.generateStatusPacket().packed()
				debugAskAndSendStatus(packet.Status, peer)
				go gossiper.sendTo(peer, packet)
			}

			time.Sleep(ANTI_ENTROPY_DT)
		}
	}()

	if gossiper.clientSocket != nil {

		go func() {

			for {

				var packet GossipPacket
				bytes, _, alive := gossiper.clientSocket.Receive()

				if !alive { break }

				protobuf.Decode(bytes, &packet)

				if !packet.isValid() {
					panic("Received invalid packet")
				}

				go gossiper.handleClient(&packet)
			}
		}()
	}

	wg := new(sync.WaitGroup)
	wg.Add(3)
	wg.Wait()
}

func (gossiper *Gossiper) stop() {
	gossiper.clientSocket.Unbind()
	gossiper.gossipSocket.Unbind()
}

/// Sends to one peer
func (gossiper *Gossiper) sendTo(peerAddress string, packet *GossipPacket) {

	if !packet.isValid() {
		log.Panicf("Sending invalid packet: %v", packet)
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

	if packet == nil || packet.Simple == nil {
		return // Fail gracefully
	}

	logClientMessage(packet.Simple)
	logPeers(gossiper.peers)

	if gossiper.simple {
		go gossiper.relay(packet, true)
	} else {

		rumor := gossiper.generateRumor(packet.Simple.Contents)

		gossiper.rumors.put(rumor)

		peer, found := gossiper.randomPeer()

		if found {
			go gossiper.rumormonger(rumor, peer)
		}
	}
}

func (gossiper *Gossiper) handleGossip(packet *GossipPacket, source string) {

	if packet == nil || !packet.isValid() {
		return // Fail gracefully
	}

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

			gossiper.rumors.put(packet.Rumor)
			peer, found := gossiper.randomPeer()

			if found {
				//debugForwardRumor(packet.Rumor)
				go gossiper.rumormonger(packet.Rumor, peer)
			}
		}

		statusPacket := gossiper.generateStatusPacket()
		debugSendStatus(statusPacket, source)
		go gossiper.sendTo(source, statusPacket.packed())

	case packet.Status != nil:

		logStatus(packet.Status, source)
		logPeers(gossiper.peers)

		expected := gossiper.dispatchStatusPacket(source, packet.Status)
		if !expected {
			fmt.Print("UNEXPECTED STATUS\n")
			rumor, _, _ := gossiper.compareStatus(packet.Status.Want, ComparisonModeMissingOrNew)

			if rumor != nil {
				fmt.Print("SEND RUMOR AFTER UNEXPECTED STATUS %v\n", rumor)
				go gossiper.rumormonger(rumor, source)
			}
		}
	}
}

func (gossiper *Gossiper) rumormonger(rumor *RumorMessage, peer string) {

	if rumor == nil {
		panic("Cannot rumormonger with <nil> rumor!")
	}

	shouldContinue := false

	// Forward package to peer
	logMongering(peer)
	go gossiper.sendTo(peer, rumor.packed())

	// Start timer
	ticker := time.NewTicker(STATUS_TIMEOUT)
	defer ticker.Stop()

	select {
	case statusPacket := <- gossiper.statusPacketsFrom(peer):

		// Compare status from peer with own messages
		otherRumor, _, statuses := gossiper.compareStatus(statusPacket.Want, ComparisonModeMissingOrNew)

		switch  {
		case statuses != nil: // Peer has new messages
			statusPacket := &StatusPacket{statuses}
			go gossiper.sendTo(peer, statusPacket.packed())
			shouldContinue = true

		case otherRumor != nil: // Peer is missing messages

			go gossiper.rumormonger(otherRumor, peer)
			shouldContinue = true

		default:
			logInSyncWith(peer)
		}

	case <- ticker.C: // Timeout
		debugTimeout(peer)
	}

	gossiper.stopWaitingFrom(peer)

	newPeer, found := gossiper.randomPeer()

	if !found {
		return
	}

	if !shouldContinue && flipCoin() {
		logFlippedCoin(newPeer)
		shouldContinue = true
	}

	if shouldContinue {
		gossiper.rumormonger(rumor, newPeer)
	} else {
		debugStopMongering(rumor)
	}
}

/*
 * Sends a status packet to potential handlers.
 * @return `True` if the packet was consumed.
 */
func (gossiper *Gossiper) dispatchStatusPacket(source string, statusPacket *StatusPacket) bool {

	gossiper.handlerLock.RLock()
	defer gossiper.handlerLock.RUnlock()

	expected := gossiper.handlerCount[source] > 0

	if expected {
		gossiper.handlers[source] <- statusPacket
	}

	fmt.Printf("RESULT FROM DISPATCH: %v", expected)

	return expected
}

func (gossiper *Gossiper) statusPacketsFrom(peer string) chan *StatusPacket {

	gossiper.handlerLock.Lock()
	defer gossiper.handlerLock.Unlock()

	_, found := gossiper.handlers[peer]

	if !found {
		gossiper.handlers[peer] = make(chan *StatusPacket, STATUS_BUFFER_SIZE)
	}

	gossiper.handlerCount[peer] = gossiper.handlerCount[peer] + 1

	return gossiper.handlers[peer]
}

func (gossiper *Gossiper) stopWaitingFrom(peer string) {

	gossiper.handlerLock.Lock()
	defer gossiper.handlerLock.Unlock()

	count := gossiper.handlerCount[peer]

	if count > 0 {
		count = count - 1
	} else {
		count = 0
	}

	gossiper.handlerCount[peer] = count
}

func (gossiper *Gossiper) compareStatus(statuses []PeerStatus, mode int) (*RumorMessage, []*RumorMessage, []PeerStatus) {

	if mode > ComparisonModeAllNew || mode < ComparisonModeMissingOrNew {
		mode = ComparisonModeAllNew
	}

	// First, we generate a statusPacket based on our rumor list
	myStatuses := gossiper.generateStatusPacket().Want

	fmt.Printf("COMPARING %v and %v\n", myStatuses, statuses)

	// This map should store, for each node we know about, what is the nextID we want
	myNextIDs := make(map[string]uint32)

	// Should become true if during the process somewhere we saw a message that we do not yet have
	rumorsWanted := false

	//
	var allRumors []*RumorMessage

	if mode == ComparisonModeAllNew {
		allRumors = make([]*RumorMessage, 0)
	}

	for _, myStatus := range myStatuses {
		myNextIDs[myStatus.Identifier] = myStatus.NextID
	}

	for _, theirStatus := range statuses {

		theirNextID := theirStatus.NextID

		// In case someone sends something smaller than
		// possible, we fail gracefully
		if theirNextID < INITIAL_ID {
			return nil, nil, nil
		}

		myNextID, found := myNextIDs[theirStatus.Identifier]

		switch {

		case !found && theirNextID != INITIAL_ID:
			// They know about an origin node we don't know.
			// We make sure that they are not looking for the first message
			// because in this case they cannot send us anything.
			rumorsWanted = true

		case found && myNextID < theirNextID:
			// They have a message we don't
			rumorsWanted = true

		case found && myNextID > theirNextID:

			if mode == ComparisonModeMissingOrNew {
				return gossiper.rumors.get(theirStatus.Identifier, theirNextID), nil, nil
			} else {
				for i := theirNextID; i < myNextID; i++ {
					rumor := gossiper.rumors.get(theirStatus.Identifier, i)
					fmt.Printf("A - APPENDING %v\n", rumor)
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
		if myNextID == INITIAL_ID {
			continue
		}

		if mode == ComparisonModeMissingOrNew {
			// Get first rumor from this node
			return gossiper.rumors.get(identifier, INITIAL_ID), nil, nil
		} else {
			for i := INITIAL_ID; i < myNextID; i++ {
				rumor := gossiper.rumors.get(identifier, i)
				allRumors = append(allRumors, rumor)
			}
		}
	}

	// If we did not return already with a rumor to send, and we want rumors,
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

func (gossiper *Gossiper) randomPeer() (string, bool) {

	if len(gossiper.peers) == 0 {
		return "", false
	}

	return gossiper.peers[rand.Int() % len(gossiper.peers)], true
}