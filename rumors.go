package main

import (
    "sync"
)

type RumorMessages struct {
    rumors  map[string]map[uint32]*RumorMessage
    IDs     map[string][]uint32
    origins map[string]bool
    mutex   *sync.Mutex
}

func makeRumors() *RumorMessages {
    return &RumorMessages{
        make(map[string]map[uint32]*RumorMessage),
        make(map[string][]uint32),
        make(map[string]bool),
        &sync.Mutex{},
    }
}

func (r *RumorMessages) get(origin string, ID uint32) *RumorMessage {

    rumorByID, found := r.rumors[origin]

    if !found {
      return nil
    }

    return rumorByID[ID]
}

func (r *RumorMessages) contains(rumor *RumorMessage) bool {

    return r.get(rumor.Origin, rumor.ID) != nil
}

/// Stores message in list of all messages
func (r *RumorMessages) put(rumor *RumorMessage) {

    r.mutex.Lock()

    if rumor == nil {
        panic("Should not try and store and <nil> rumor.")
    }

    //r.mutex.Lock()
    //defer r.mutex.Unlock()

    _, found := r.IDs[rumor.Origin]

    if !found {
        r.IDs[rumor.Origin] = make([]uint32, 0)
        r.rumors[rumor.Origin] = make(map[uint32]*RumorMessage)
    }

    r.rumors[rumor.Origin][rumor.ID] = rumor
    r.IDs[rumor.Origin] = append(r.IDs[rumor.Origin], rumor.ID)
    //sort.Sort(r.IDs[rumor.Origin])

    r.mutex.Unlock()
}

func (r *RumorMessages) nextIDFor(origin string) uint32 {

    r.mutex.Lock()

    counter := uint32(0)

    for _, id := range r.IDs[origin] {
        if id != counter {
            return counter
        }
        counter++
    }

    r.mutex.Unlock()

    return counter
}

func (r *RumorMessages) allOrigins() []string {

    r.mutex.Lock()

    res := make([]string, 0)

    for origin, _ := range r.IDs {
        res = append(res, origin)
    }

    r.mutex.Unlock()

    return res
}