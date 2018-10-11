package main

type RumorMessages struct {
  rumors          map[string]map[uint32]*RumorMessage
  IDs             map[string][]uint32
}

func makeRumors() *RumorMessages {
  return &RumorMessages{
    rumors: make(map[string]map[uint32]*RumorMessage),
    IDs: make(map[string][]uint32),
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

  _, found := r.IDs[rumor.Origin]

  if !found {
    r.IDs[rumor.Origin] = make([]uint32, 5)
  }

  _, found = r.rumors[rumor.Origin]

  if !found {
    r.rumors[rumor.Origin] = make(map[uint32]*RumorMessage)
  }

  r.rumors[rumor.Origin][rumor.ID] = rumor
  r.IDs[rumor.Origin] = append(r.IDs[rumor.Origin], rumor.ID)
  //sort.Sort(r.IDs[rumor.Origin])
}

func (r *RumorMessages) nextIDFor(origin string) uint32 {

  counter := uint32(0)

  for _, id := range r.IDs[origin] {
    if id != counter {
      return counter
    }
    counter++
  }

  return counter
}

func (r *RumorMessages) allOrigins() []string {

  res := make([]string, 5)

  for origin, _ := range r.IDs {
    res = append(res, origin)
  }

  return res
}