package gossiper

import (
	"github.com/jfperren/Peerster/common"
	"strings"
	"math/rand"
)


type Router struct {
	NextHop		 	map[string]string	// Routing Table
	Peers         	[]string			// List of known peer IP addresses
}

func NewRouter(peers string) *Router {

	return &Router{
		NextHop:	   make(map[string]string),
		Peers:         strings.Split(peers, ","),
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