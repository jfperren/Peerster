package gossiper




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

func (gossiper *Gossiper) updateRoutingTable(origin, address string) {

	currentAddress, found := gossiper.NextHop[origin]

	// Only update if needed
	if !found || currentAddress != address {
		gossiper.NextHop[origin] = address
		common.LogUpdateRoutingTable(origin, address)
	}
}