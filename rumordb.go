package main

import "sync"


// Database-like object that stores / serves rumors in a thread-safe way,
// while also offer higher-level functions related to those rumors.
type RumorDatabase struct {
    rumors  map[string]map[uint32]*RumorMessage
    mutex   *sync.RWMutex
}

// Create an empty database of rumors.
func MakeRumorDatabase() *RumorDatabase {
    return &RumorDatabase{
        make(map[string]map[uint32]*RumorMessage),
        &sync.RWMutex{},
    }
}

// Get the rumor with given ID and origin node name. If no such rumor
// exists, return nil.
func (r *RumorDatabase) Get(origin string, ID uint32) *RumorMessage {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    rumorByID, found := r.rumors[origin]

    if !found {
      return nil
    }

    return rumorByID[ID]
}

// Return true if the rumor is the next expected rumor from the origin node.
func (r *RumorDatabase) Expects(rumor *RumorMessage) bool {

	r.mutex.RLock()
	defer r.mutex.RUnlock()

    return r.NextIDFor(rumor.Origin) == rumor.ID
}

// Return true if a rumor is already contained in the database. Comparison
// is done using ID and origin node only and does not care about rumor text.
func (r *RumorDatabase) Put(rumor *RumorMessage) {

    if rumor == nil {
        panic("Should not try and store and <nil> rumor.")
    }

	r.mutex.Lock()
	defer r.mutex.Unlock()

    // Does not process rumors which are not the next expected one.
    if rumor.ID != uint32(len(r.rumors[rumor.Origin]) + 1) {
    	return
	}

    _, found := r.rumors[rumor.Origin]

    if !found {
        r.rumors[rumor.Origin] = make(map[uint32]*RumorMessage)
    }

    r.rumors[rumor.Origin][rumor.ID] = rumor
}

// Returns, for a given origin node, the first message ID that is NOT
// contained in the database. If the origin is not known to the database,
// it will return InitialId.
func (r *RumorDatabase) NextIDFor(origin string) uint32 {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    rumors, found := r.rumors[origin]

    if !found {
    	return InitialId
	}

	return uint32(len(rumors)) + InitialId
}

// Return a slice containing the name of each known node in the rumor
// database.
func (r *RumorDatabase) AllOrigins() []string {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    res := make([]string, 0)

    for origin := range r.rumors {
        res = append(res, origin)
    }

    return res
}