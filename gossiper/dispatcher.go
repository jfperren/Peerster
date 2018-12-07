package gossiper

import (
	"encoding/hex"
	"github.com/jfperren/Peerster/common"
	"sync"
)

// A dispatcher is responsible for coordinating between multiple processes who are
// all waiting to receive some packets from a given source. Through the dispatcher,
// processes can explicitely state what sources they are waiting on and will be notified
// upon receipt of such packet
type Dispatcher struct {
	Handlers     map[string]chan *common.GossipPacket // Channels waiting for StatusPackets
	HandlerCount map[string]int                       // Count of rumormongering processes waiting on a node status
	HandlerLock  *sync.RWMutex                        // Lock for safely updating & reading handlers

}

func NewDispatcher() *Dispatcher {

	return &Dispatcher{
		HandlerLock:  &sync.RWMutex{},
		Handlers:     make(map[string]chan *common.GossipPacket),
		HandlerCount: make(map[string]int),
	}
}

//
//  CORE FUNCTIONS
//

// Dispatch a status packet to potential Handlers. Return true if the status packet was expected by a
// process.
func (dispatcher *Dispatcher) dispatchPacket(identifier string, packet *common.GossipPacket) bool {

	dispatcher.HandlerLock.RLock()
	defer dispatcher.HandlerLock.RUnlock()

	expected := dispatcher.HandlerCount[identifier] > 0

	if expected {
		dispatcher.Handlers[identifier] <- packet
	}

	return expected
}

// Create / return a channel that will allow to receive status packets from a given node.
// Calling this function implicitly register that a new rumormongering process is waiting
// on a status packet. To indicate that this is no longer the case, call stopWaitingFrom(peer).
func (dispatcher *Dispatcher) packets(identifier string) chan *common.GossipPacket {

	dispatcher.HandlerLock.Lock()
	defer dispatcher.HandlerLock.Unlock()

	_, found := dispatcher.Handlers[identifier]

	if !found {
		dispatcher.Handlers[identifier] = make(chan *common.GossipPacket, common.StatusBufferSize)
	}

	dispatcher.HandlerCount[identifier] = dispatcher.HandlerCount[identifier] + 1

	return dispatcher.Handlers[identifier]
}

// Explicitly state that a given rumormongering process is no longer waiting for a status packet.
func (dispatcher *Dispatcher) stopWaitingOn(identifier string) {

	dispatcher.HandlerLock.Lock()
	defer dispatcher.HandlerLock.Unlock()

	count := dispatcher.HandlerCount[identifier]

	if count > 0 {
		count = count - 1
	} else {
		count = 0
	}

	dispatcher.HandlerCount[identifier] = count
}

//
//  UNIQUE IDS
//

// Unique ID of a status message
func dispatchIdStatus(source string) string {
	return "status:" + source
}

// Unique ID of a data reply
func dispatchIdDataReply(hash []byte) string {
	return "data-reply:" + hex.EncodeToString(hash)
}

//
//  CONVENIENCE METHODS
//

func (dispatcher *Dispatcher) statusPackets(source string) chan *common.GossipPacket {
	return dispatcher.packets(dispatchIdStatus(source))
}

func (dispatcher *Dispatcher) dispatchStatusPacket(source string, packet *common.GossipPacket) bool {
	return dispatcher.dispatchPacket(dispatchIdStatus(source), packet)
}

func (dispatcher *Dispatcher) stopWaitingOnStatusPacket(source string) {
	dispatcher.stopWaitingOn(dispatchIdStatus(source))
}

func (dispatcher *Dispatcher) dataReplies(hash []byte) chan *common.GossipPacket {
	return dispatcher.packets(dispatchIdDataReply(hash))
}

func (dispatcher *Dispatcher) dispatchDataReply(packet *common.GossipPacket) bool {
	return dispatcher.dispatchPacket(dispatchIdDataReply(packet.DataReply.HashValue), packet)
}

func (dispatcher *Dispatcher) stopWaitingOnDataReply(hash []byte) {
	dispatcher.stopWaitingOn(dispatchIdDataReply(hash))
}
