package main

import (
    "sync"
)

type RumorMessages struct {
    rumors  map[string]map[uint32]*RumorMessage
    IDs     map[string][]uint32
    origins map[string]bool
    mutex   *sync.RWMutex
}

func makeRumors() *RumorMessages {
    return &RumorMessages{
        make(map[string]map[uint32]*RumorMessage),
        make(map[string][]uint32),
        make(map[string]bool),
        &sync.RWMutex{},
    }
}

func (r *RumorMessages) get(origin string, ID uint32) *RumorMessage {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    rumorByID, found := r.rumors[origin]

    if !found {
      return nil
    }

    return rumorByID[ID]
}

func (r *RumorMessages) contains(rumor *RumorMessage) bool {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    return r.get(rumor.Origin, rumor.ID) != nil
}



/// Stores message in list of all messages
func (r *RumorMessages) put(rumor *RumorMessage) {

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

func (r *RumorMessages) nextIDFor(origin string) uint32 {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    counter := uint32(INITIAL_ID)

    for _, id := range r.IDs[origin] {

        if id != counter {
            return counter
        }
        counter++
    }

    return counter
}

func (r *RumorMessages) newRumorsSince(statuses []PeerStatus) ([]*RumorMessage, bool) {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    newRumors := make([]*RumorMessage, 0)
    origins := make(map[string]uint32)
    hasChanged := false

    for _, origin := range r.allOrigins() {
        origins[origin] = r.nextIDFor(origin)
    }

    for _, status := range statuses {
        myNextID, found := origins[status.Identifier]
        theirNextID := status.NextID

        if !found {
            theirNextID = INITIAL_ID
        }

        for i := theirNextID; i < myNextID; i++ {

            hasChanged = true
            rumor := r.get(status.Identifier, i)
            newRumors = append(newRumors, rumor)
        }

        delete(origins, status.Identifier)
    }

    for origin, myNextID := range origins {

        for i := INITIAL_ID; i < myNextID; i++ {
            hasChanged = true
            rumor := r.get(origin, i)
            newRumors = append(newRumors, rumor)
        }
    }

    return newRumors, hasChanged
}

func (r *RumorMessages) allOrigins() []string {

    r.mutex.RLock()
    defer r.mutex.RUnlock()

    res := make([]string, 0)

    for origin, _ := range r.IDs {
        res = append(res, origin)
    }

    return res
}