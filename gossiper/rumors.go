package gossiper

import (
	"github.com/jfperren/Peerster/common"
	"sync"
	"time"
)

// Database-like object that stores / serves Rumors in a thread-safe way,
// while also offer higher-level functions related to those Rumors.
type RumorDatabase struct {
	nextID uint32                                     // NextID to be used for Rumors
	Rumors map[string]map[uint32]*common.RumorMessage // List of rumor Ids per node
	Mutex  *sync.RWMutex                              // Read-write lock to access the database
}

// Create an empty database of Rumors.
func NewRumorDatabase() *RumorDatabase {

	return &RumorDatabase{
		nextID: common.InitialId,
		Rumors: make(map[string]map[uint32]*common.RumorMessage),
		Mutex: &sync.RWMutex{},
	}
}

// Get the rumor with given ID and origin node name. If no such rumor
// exists, return nil.
func (r *RumorDatabase) Get(origin string, ID uint32) *common.RumorMessage {

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	rumorByID, found := r.Rumors[origin]

	if !found {
		return nil
	}

	return rumorByID[ID]
}

// Return true if the rumor is the next expected rumor from the origin node.
func (r *RumorDatabase) Expects(rumor *common.RumorMessage) bool {

	return r.NextIDFor(rumor.Origin) == rumor.ID
}

// Return true if a rumor is already contained in the database. Comparison
// is done using ID and origin node only and does not care about rumor text.
func (r *RumorDatabase) Put(rumor *common.RumorMessage) {

	if rumor == nil {
		panic("Should not try and store and <nil> rumor.")
	}

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	nextID := common.InitialId
	rumors, found := r.Rumors[rumor.Origin]

	if found {
		nextID += uint32(len(rumors))
	}

	if nextID != rumor.ID {
		return
	}

	if !found {
		r.Rumors[rumor.Origin] = make(map[uint32]*common.RumorMessage)
	}

	r.Rumors[rumor.Origin][rumor.ID] = rumor
}

// Returns, for a given origin node, the first message ID that is NOT
// contained in the database. If the origin is not known to the database,
// it will return InitialId.
func (r *RumorDatabase) NextIDFor(origin string) uint32 {

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	rumors, found := r.Rumors[origin]

	if !found {
		return common.InitialId
	}

	return uint32(len(rumors)) + common.InitialId
}

// Return a slice containing the name of each known node in the rumor
// database.
func (r *RumorDatabase) AllOrigins() []string {

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	res := make([]string, 0)

	for origin := range r.Rumors {
		res = append(res, origin)
	}

	return res
}

// Returns the next ID + increase it
func (r *RumorDatabase) ConsumeNextID() uint32 {

	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	nextID := r.nextID
	r.nextID = nextID + 1

	return nextID
}

//
//  GOSSIPER METHODS
//

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