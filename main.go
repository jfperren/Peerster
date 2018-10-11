package main

import (
    "fmt"
    //"os"
    "flag"
    "strings"
    "github.com/dedis/protobuf"
    "sync"
)

type Gossiper struct {
  gossipSocket    *UDPSocket
  clientSocket    *UDPSocket
  gossipAddress   string
  clientAddress   string
  peers           []string
  Name            string
}

type Message struct {
  Text string
}

func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "port for the gossiper")
  name := flag.String("name", "REQUIRED", "name of the gossiper")
  peers := flag.String("peers", "REQUIRED", "comma separated list of peers of the form ip:port")
  _ = flag.Bool("simple", false, "runs gossiper in simple broadcast mode")

  flag.Parse()

  // Print Flags

  // fmt.Printf("port = %v\n", *uiPort)
  // fmt.Printf("gossipAddr = %v\n", *gossipAddr)
  // fmt.Printf("name = %v\n", *name)
  // fmt.Printf("peers = %v\n", *peers)
  // fmt.Printf("simple = %v\n", *simple)

  // Creates Gossiper

  Gossiper := NewGossiper(*gossipAddr, *uiPort,  *name, *peers)

  Gossiper.start()
}


func NewGossiper(gossipAddress, clientPort, name string, peersList string) *Gossiper {

  return &Gossiper{
    gossipSocket:   NewUDPSocket(gossipAddress),
    clientSocket:   NewUDPSocket(":" + clientPort),
    gossipAddress:  gossipAddress,
    clientAddress:  ":" + clientPort,
    Name:           name,
    peers:          strings.Split(peersList, ","),
  }
}

func (gossiper *Gossiper) start() {

  go gossiper.listenClient()
  go gossiper.listenGossip()

  wg := new(sync.WaitGroup)
  wg.Add(2)
  wg.Wait()
}

func (gossiper *Gossiper) listenClient() {

  var message Message

  for {
    bytes := gossiper.clientSocket.Receive()
    protobuf.Decode(bytes, &message)
    gossiper.receiveClient(message)
  }
}

/// Sends to one peer
func (gossiper *Gossiper) send(peerAddress string, packet *GossipPacket) {

  bytes, err := protobuf.Encode(packet)
  if err != nil { panic(err) }

  gossiper.gossipSocket.Send(bytes, peerAddress)
}

/// Sends to every peer
func (gossiper *Gossiper) relay(packet *GossipPacket) {

  for i := 0; i < len(gossiper.peers); i++ {
    if gossiper.peers[i] != packet.Simple.RelayPeerAddr {
      gossiper.send(gossiper.peers[i], packet)
    }
  }
}

func (gossiper *Gossiper) listenGossip() {

  var packet GossipPacket

  for {
    bytes := gossiper.gossipSocket.Receive()
    protobuf.Decode(bytes, &packet)
    gossiper.receiveGossip(packet)
  }
}

func (gossiper *Gossiper) receiveGossip(packet GossipPacket) {

  if !contains(gossiper.peers, packet.Simple.RelayPeerAddr) {
    gossiper.peers = append(gossiper.peers, packet.Simple.RelayPeerAddr)
  }

  gossiper.logPeerMessage(packet.Simple)

  packet.Simple = gossiper.processPeerMessage(packet.Simple)

  gossiper.relay(&packet)
}

func (gossiper *Gossiper) receiveClient(message Message) {

  gossiper.logClientMessage(message)
  gossipPacket := gossiper.wrapClientMessage(message)
  gossiper.relay(gossipPacket)

}

func (gossiper *Gossiper) wrapClientMessage(message Message) *GossipPacket {

  simpleMessage := SimpleMessage {
    OriginalName:   gossiper.Name,
    RelayPeerAddr:  gossiper.gossipAddress,
    Contents:       message.Text,
  }

  return &GossipPacket { &simpleMessage, nil, nil }
}

func (gossiper *Gossiper) processPeerMessage(message *SimpleMessage) *SimpleMessage {
  // message.RelayPeerAddr = gossiper.gossipAddress.IP.String() + ":" + fmt.Sprint(gossiper.gossipAddress.Port)

  message.RelayPeerAddr = gossiper.gossipAddress

  return message
}

func (gossiper *Gossiper) logClientMessage(message Message) {
  fmt.Printf("CLIENT MESSAGE %v\n", message.Text)
}

func (gossiper *Gossiper) logPeerMessage(message *SimpleMessage) {
  fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
    message.OriginalName, message.RelayPeerAddr, message.Contents)
  fmt.Printf("%v\n", strings.Join(gossiper.peers, ","))
}

func contains(array []string, element string) bool {
	for _, o := range array {
		if o == element {
			return true
		}
	}
	return false
}
