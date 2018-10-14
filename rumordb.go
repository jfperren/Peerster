package main

import "sync"


// Database-like object that stores / serves rumors in a thread-safe way,
// while also offer higher-level functions related to those rumors.
type RumorDatabase struct {
    rumors  map[string]map[uint32]*RumorMessage
    IDs     map[string][]uint32
    origins map[string]bool
    mutex   *sync.RWMutex
}

// Create an empty database of rumors.
func MakeRumorDatabase() *RumorDatabase {
    return &RumorDatabase{
        make(map[string]map[uint32]*RumorMessage),
        make(map[string][]uint32),
        make(map[string]bool),
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

// Return true if a rumor is already contained in the database. Comparison
// is done using ID and origin node only and does not care about rumor text.
func (r *RumorDatabase) Contains(rumor *RumorMessage) bool {

    if rumor == nil {
        panic("Cannot use Contains with a <nil> rumor.")
    }

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    return r.Get(rumor.Origin, rumor.ID) != nil
}

// Return true if a rumor is already contained in the database. Comparison
// is done using ID and origin node only and does not care about rumor text.
func (r *RumorDatabase) Put(rumor *RumorMessage) {

    if rumor == nil {
        panic("Should not try and store and <nil> rumor.")
    }

    r.mutex.Lock()
    defer r.mutex.Unlock()

    _, found := r.IDs[rumor.Origin]

    if !found {
        r.IDs[rumor.Origin] = make([]uint32, 0)
        r.rumors[rumor.Origin] = make(map[uint32]*RumorMessage)
    }

    r.rumors[rumor.Origin][rumor.ID] = rumor
    r.IDs[rumor.Origin] = insertSorted(r.IDs[rumor.Origin], rumor.ID)
}

// Returns, for a given origin node, the first message ID that is NOT
// contained in the database. If the origin is not known to the database,
// it will return InitialId.
func (r *RumorDatabase) NextIDFor(origin string) uint32 {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    idsFromOrigin, found := r.IDs[origin]

    if !found {
    	return InitialId
	}

	counter := uint32(InitialId)

    for _, id := range idsFromOrigin {

        if id != counter {
            return counter
        }
        counter++
    }

    return counter
}

// Return a slice containing the name of each known node in the rumor
// database.
func (r *RumorDatabase) AllOrigins() []string {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    res := make([]string, 0)

    for origin := range r.IDs {
        res = append(res, origin)
    }

    return res
}