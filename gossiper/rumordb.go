package gossiper

import (
	"github.com/jfperren/Peerster/common"
	"sync"
)


// Database-like object that stores / serves Rumors in a thread-safe way,
// while also offer higher-level functions related to those Rumors.
type RumorDatabase struct {
	Rumors map[string]map[uint32]*common.RumorMessage
	Mutex  *sync.RWMutex
}

// Create an empty database of Rumors.
func NewRumorDatabase() *RumorDatabase {
	return &RumorDatabase{
		make(map[string]map[uint32]*common.RumorMessage),
		&sync.RWMutex{},
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

	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

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