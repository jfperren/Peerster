package gossiper

import (
	"github.com/jfperren/Peerster/common"
	"strings"
	"math/rand"
	"time"
)


type Router struct {
	NextHop		 	map[string]string	// Routing Table
	Peers         	[]string			// List of known peer IP addresses
	Rtimer			time.Duration		// Interval for sending route rumors
}

func NewRouter(peers string, rtimer time.Duration) *Router {

	return &Router{
		NextHop:	make(map[string]string),
		Peers:      strings.Split(peers, ","),
		Rtimer:		rtimer,
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

	return router.Peers[rand.Int() % len(router.Peers)], true
}

func (router *Router) updateRoutingTable(origin, address string) {

	currentAddress, found := router.NextHop[origin]

	// Only update if needed
	if !found || currentAddress != address {
		router.NextHop[origin] = address
		common.LogUpdateRoutingTable(origin, address)
	}
}