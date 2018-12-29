package gossiper

import (
	"github.com/dedis/protobuf"
	"github.com/jfperren/Peerster/common"
	"log"
	"math/rand"
	"strings"
	"time"
)

// A Router is responsible for handling the neighbors (Peers) for gossip communication as well
// as a DSDV table (NextHop) for the more complex routing of private messages and downloads.
type Router struct {
	NextHop map[string]string // Routing Table
	Peers   []string          // List of known peer IP addresses
	Rtimer  time.Duration     // Interval for sending route rumors
}

func NewRouter(peers string, rtimer time.Duration) *Router {

	return &Router{
		NextHop: make(map[string]string),
		Peers:   strings.Split(peers, ","),
		Rtimer:  rtimer,
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

	currentAddress, found := router.NextHop[origin]

	// Only update if needed
	if !found || currentAddress != address {
		router.NextHop[origin] = address
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

    if gossiper.Crypto.Options == common.SignOnly {
        packet = &common.GossipPacket{
            Signed: gossiper.SignPacket(packet),
        }
    } else if gossiper.Crypto.Options == common.CypherIfPossible {
        destination, err := packet.GetDestination()
        if err != nil {
            packet = &common.GossipPacket{
                Signed: gossiper.SignPacket(packet),
            }
        } else {
            packet = &common.GossipPacket{
                Cyphered: gossiper.CypherPacket(gossiper.SignPacket(packet), destination),
            }
        }
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