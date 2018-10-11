package main

import (
    // "fmt"
    //"os"
    "flag"
    "strings"
    "github.com/dedis/protobuf"
    "sync"
    "time"
    //"sort"
)

const ACK_TIMEOUT = 1

type Gossiper struct {
  simple          bool
  gossipSocket    *UDPSocket
  clientSocket    *UDPSocket
  gossipAddress   string
  clientAddress   string
  peers           []string
  Name            string

  // Registered handlers for status messages
  handlers        map[string]chan*StatusPacket
  rumors          *RumorMessages

  statusPacket    *StatusPacket
}


func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "port for the gossiper")
  name := flag.String("name", "REQUIRED", "name of the gossiper")
  peers := flag.String("peers", "REQUIRED", "comma separated list of peers of the form ip:port")
  simple := flag.Bool("simple", false, "runs gossiper in simple broadcast mode")

  flag.Parse()

  // Print Flags

  // fmt.Printf("port = %v\n", *uiPort)
  // fmt.Printf("gossipAddr = %v\n", *gossipAddr)
  // fmt.Printf("name = %v\n", *name)
  // fmt.Printf("peers = %v\n", *peers)
  // fmt.Printf("simple = %v\n", *simple)

  // Creates Gossiper

  Gossiper := NewGossiper(*gossipAddr, *uiPort,  *name, *peers, *simple)

  Gossiper.start()
}


func NewGossiper(gossipAddress, clientPort, name string, peers string, simple bool) *Gossiper {

  clientAddress := ":" + clientPort

  gossipSocket := NewUDPSocket(gossipAddress)
  clientSocket := NewUDPSocket(clientAddress)

  statusPacket := &StatusPacket{make([]PeerStatus, 5)}

  return &Gossiper{
    simple:         simple,
    gossipSocket:   gossipSocket,
    clientSocket:   clientSocket,
    gossipAddress:  gossipAddress,
    clientAddress:  clientAddress,
    Name:           name,
    peers:          strings.Split(peers, ","),
    handlers:       make(map[string]chan*StatusPacket),
    rumors:         makeRumors(),
    statusPacket:   statusPacket,
  }
}

func (gossiper *Gossiper) start() {

  go func() {
    var packet GossipPacket
    for {
      bytes, _ := gossiper.clientSocket.Receive()
      protobuf.Decode(bytes, &packet)
      gossiper.handleClient(&packet)
    }
  }()

  go func() {
    var packet GossipPacket
    for {
      bytes, source := gossiper.gossipSocket.Receive()
      protobuf.Decode(bytes, &packet)
      gossiper.handleGossip(&packet, source)
    }
  }()

  wg := new(sync.WaitGroup)
  wg.Add(2)
  wg.Wait()
}

/// Sends to one peer
func (gossiper *Gossiper) sendTo(peerAddress string, packet interface{}) {

  bytes, err := protobuf.Encode(packet)
  if err != nil { panic(err) }

  gossiper.gossipSocket.Send(bytes, peerAddress)
}

/// Sends to every peer
func (gossiper *Gossiper) relay(packet *GossipPacket, setName bool) {

  // TODO: Test type of message, panic if not Simple

  packet.Simple.RelayPeerAddr = gossiper.gossipAddress
  if setName { packet.Simple.OriginalName = gossiper.Name }

  for i := 0; i < len(gossiper.peers); i++ {
    if gossiper.peers[i] != packet.Simple.RelayPeerAddr {
      gossiper.sendTo(gossiper.peers[i], packet)
    }
  }
}


func (gossiper *Gossiper) handleClient(packet *GossipPacket) {

  // gossiper.logClientMessage(message)

  if gossiper.simple {
    go gossiper.relay(packet, true)
  } else {
    go gossiper.rumors.put(packet.Rumor)
    go gossiper.rumormonger(packet.Rumor, nil)
  }
}

func (gossiper *Gossiper) handleGossip(packet *GossipPacket, source string) {

  gossiper.addPeerIfNeeded(source)

  switch {

  case packet.Simple != nil:

    // gossiper.logPeerMessage(packet.Simple)
    // packet.Simple = gossiper.processPeerMessage(packet.Simple)
    go gossiper.relay(packet, false)

  case packet.Rumor != nil:

    if !gossiper.rumors.contains(packet.Rumor) {
      go gossiper.rumors.put(packet.Rumor)
      go gossiper.rumormonger(packet.Rumor, nil)
    }

    go gossiper.sendTo(source, gossiper.generateStatusPacket())

  case packet.Status != nil:

    expected := gossiper.dispatchStatusPacket(source, packet.Status)
    if !expected {
      rumor, _ := gossiper.compareStatus(packet.Status)
      go gossiper.rumormonger(rumor, &source)
    }
  }
}

// func (gossiper *Gossiper) logClientMessage(message Message) {
//   fmt.Printf("CLIENT MESSAGE %v\n", message.Text)
// }
//
// func (gossiper *Gossiper) logPeerMessage(message *SimpleMessage) {
//   fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
//     message.OriginalName, message.RelayPeerAddr, message.Contents)
//   fmt.Printf("%v\n", strings.Join(gossiper.peers, ","))
// }

func (gossiper *Gossiper) rumormonger(rumor *RumorMessage, peer *string) {

  if peer == nil {
    index := selectRandom(gossiper.peers)
    peer = &index
  }

  // Send rumor to neighbor
  gossiper.sendTo(*peer, &GossipPacket{nil, rumor, nil})

  var shouldContinue bool

  ticker := time.NewTicker(ACK_TIMEOUT * time.Second)
  defer ticker.Stop()

  go gossiper.sendTo(*peer, rumor)

  select {
  case statusPacket := <- gossiper.awaitStatusPacket(*peer): // Received ACK

    // Compare status from peer with own messages
    otherRumor, status := gossiper.compareStatus(statusPacket)

    switch  {
    case status != nil: // Peer has unseen messages

      go gossiper.sendTo(*peer, status)
      shouldContinue = true

    case otherRumor != nil: // Peer is missing messages

      go gossiper.rumormonger(otherRumor, peer)
      shouldContinue = true

    default:
      shouldContinue = flipCoin()
    }

  case <- ticker.C: // Timeout
    shouldContinue = flipCoin()
  }

  if shouldContinue {
    go gossiper.rumormonger(rumor, nil)
  }
}

/*
 * Sends a status packet to potential handlers.
 * @return `True` if the packet was consumed.
 */
func (gossiper *Gossiper) dispatchStatusPacket(source string, statusPacket *StatusPacket) bool {
  channel, found := gossiper.handlers[source]
  if found {
    channel <- statusPacket
  }
  return found
}

func (gossiper *Gossiper) awaitStatusPacket(peer string) chan *StatusPacket {
  channel := make(chan *StatusPacket)
  gossiper.handlers[peer] = channel
  return channel
}

func (gossiper *Gossiper) stopWaitForStatusPacket(peer string) {
  gossiper.handlers[peer] = nil
}

func (gossiper *Gossiper) compareStatus(statusPacket *StatusPacket) (*RumorMessage, *StatusPacket) {

  thisStatuses := make(map[string]uint32)
  thisStatusPacket := gossiper.generateStatusPacket()

  gossiperIsMissingMessages := false

  for i := 0; i < len(thisStatusPacket.Want); i++ {
    status := thisStatusPacket.Want[i]
    thisStatuses[status.Identifier] = status.NextID
  }

  for i := 0; i < len(statusPacket.Want); i++ {

    peerStatus := statusPacket.Want[i]
    peerNextID := peerStatus.NextID
    thisNextID, found := thisStatuses[peerStatus.Identifier]

    switch {

    case !found: // They know about a node we don't know
      gossiperIsMissingMessages = true

    case thisNextID == peerNextID: // In sync, continue to next

    case thisNextID < peerNextID:
      gossiperIsMissingMessages = true

    case thisNextID > peerNextID:
      return gossiper.rumors.get(peerStatus.Identifier, peerNextID), nil
    }

    delete(thisStatuses, peerStatus.Identifier)
  }

  for identifier, nextID := range thisStatuses { // We know about a node they don't know
    return gossiper.rumors.get(identifier, nextID), nil
  }

  if gossiperIsMissingMessages {
    return nil,  thisStatusPacket
  }

  return nil, nil
}

func (gossiper *Gossiper) addPeerIfNeeded(peer string) {
  if !containsString(gossiper.peers, peer) {
    gossiper.peers = append(gossiper.peers, peer)
  }
}

func (gossiper *Gossiper) generateStatusPacket() *StatusPacket {

  peerStatuses := make([]PeerStatus, 5)

  for _, origin := range gossiper.rumors.allOrigins() {
    peerStatuses = append(peerStatuses, PeerStatus{origin, gossiper.rumors.nextIDFor(origin)})
  }

  return &StatusPacket{peerStatuses}
}