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

// A Router is responsible for handling the neighbors (Peers) for gossip communication as well
// as a DSDV table (NextHop) for the more complex routing of private messages and downloads.
type Router struct {
	NextHop map[string]string // Routing Table
	Peers   []string          // List of known peer IP addresses
	Rtimer  time.Duration     // Interval for sending route rumors
    Mutex    *sync.RWMutex     // Read-write lock to access the routing table
}

func NewRouter(peers string, rtimer time.Duration) *Router {

	return &Router{
		NextHop: make(map[string]string),
		Peers:   strings.Split(peers, ","),
		Rtimer:  rtimer,
		Mutex: &sync.RWMutex{},
	}
}

// Add a new peer IP address to the list of known peers
func (router *Router) AddPeerIfNeeded(peer string) {

	if !common.Contains(router.Peers, peer) {
		router.Peers = append(router.Peers, peer)
	}
}

func (router *Router) randomPeer() (string, bool) {

	if len(router.Peers) == 0 {
		return "", false
	}

	return router.Peers[rand.Int()%len(router.Peers)], true
}

func (router *Router) randomPeerExcept(peer string) (string, bool) {

	if len(router.Peers) == 0 {
		return "", false
	}

	if len(router.Peers) == 1 && router.Peers[0] == peer {
		return "", false
	}

	for {
		potentialPeer, found := router.randomPeer()

		if !found {
			return potentialPeer, found
		}

		if potentialPeer != peer {
			return potentialPeer, true
		}
	}
}

func (router *Router) updateRoutingTable(origin, address string) {

	router.Mutex.RLock()
	currentAddress, found := router.NextHop[origin]
	router.Mutex.RUnlock()

	// Only update if needed
	if !found || currentAddress != address {
        router.Mutex.Lock()
		router.NextHop[origin] = address
        router.Mutex.Unlock()
		common.LogUpdateRoutingTable(origin, address)
	}
}

//
//  GOSSIPER FUNCTIONS
//

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

	gossiper.Router.Mutex.RLock()
	nextPeer, found := gossiper.Router.NextHop[destination]
	gossiper.Router.Mutex.RUnlock()

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

	if gossiper.ShouldAuthenticate() && !gossiper.IsAuthenticated() && packet.ShouldBeSigned() {

		if _, found := gossiper.BlockChain.Peers[gossiper.Name]; !found {
			common.DebugSkipSendNotAuthenticated()
			return
		}
	}

    if gossiper.Crypto.Options == common.SignOnly && packet.ShouldBeSigned() {

    	packet.Signature = gossiper.SignPacket(packet)

    } else if gossiper.Crypto.Options == common.CypherIfPossible {

		packet.Signature = gossiper.SignPacket(packet)

        if packet.ShouldBeCiphered() { // Cipher for destination

			packet = gossiper.CypherPacket(packet, *packet.GetDestination()).Packed()
        }
    }

	packet, err := gossiper.wrapInOnionIfNeeded(packet)
	if err != nil {
		panic(err)
	}

	bytes, err := protobuf.Encode(packet)
	if err != nil {
		panic(err)
	}

	gossiper.GossipSocket.Send(bytes, peerAddress)
}

// Broadcast a GossipPacket containing a Simple message to every neighboring node.
func (gossiper *Gossiper) broadcastToNeighborsExcept(packet *common.GossipPacket, except *[]string) {

	if !packet.IsEligibleForBroadcast() {
		log.Panicf("Cannot broadcast packet %v.", packet)
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