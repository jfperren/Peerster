package main

import (
    "fmt"
    //"os"
    "flag"
    "net"
    "strings"
    "github.com/dedis/protobuf"
    "sync"
)

type SimpleMessage struct {
  OriginalName string
  RelayPeerAddr string
  Contents string
}

type GossipPacket struct {
  Simple *SimpleMessage
}

type Gossiper struct {
  gossipAddress *net.UDPAddr
  gossipConn *net.UDPConn
  clientAddress *net.UDPAddr
  clientConn *net.UDPConn
  packetBuffer chan string
  peers []string
  Name string
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
  simple := flag.Bool("simple", false, "runs gossiper in simple broadcast mode")

  flag.Parse()

  // Print Flags

  fmt.Printf("port = %v\n", *uiPort)
  fmt.Printf("gossipAddr = %v\n", *gossipAddr)
  fmt.Printf("name = %v\n", *name)
  fmt.Printf("peers = %v\n", *peers)
  fmt.Printf("simple = %v\n", *simple)

  // Creates Gossiper

  Gossiper := NewGossiper(*gossipAddr, *uiPort,  *name, *peers)

  Gossiper.start()
}


func NewGossiper(gossipAddress, clientPort, name string, peersList string) *Gossiper {

  gossipAddr, gossipConn := connectUDP(gossipAddress)
  clientAddr, clientConn := connectUDP(":" + clientPort)

  return &Gossiper{
    gossipAddress:  gossipAddr,
    gossipConn:     gossipConn,
    clientAddress:  clientAddr,
    clientConn:     clientConn,
    packetBuffer:   make(chan string),
    Name:           name,
    peers:          strings.Split(peersList, ","),
  }
}

func connectUDP(address string) (*net.UDPAddr, *net.UDPConn) {

  udpAddr, addrErr := net.ResolveUDPAddr("udp4", address)
  if addrErr != nil { panic(addrErr) }

  udpConn, connErr := net.ListenUDP("udp4", udpAddr)
  if addrErr != nil { panic(connErr) }

  return udpAddr, udpConn
}

func (gossiper *Gossiper) start() {

  go gossiper.listenClient()
  go gossiper.listenGossip()

  wg := new(sync.WaitGroup)
  wg.Add(2)
  wg.Wait()
}

func (gossiper *Gossiper) listenClient() {

  buffer := make([]byte, 1024)
  var message Message

  for {
    _, _, err := gossiper.clientConn.ReadFromUDP(buffer)
    if err != nil { fmt.Println(err) }

    protobuf.Decode(buffer, &message)
    gossiper.receiveClient(message)
  }
}

/// Sends to one peer
func (gossiper *Gossiper) send(peerAddress string, packet *GossipPacket) {

  udpAddr, err := net.ResolveUDPAddr("udp4", peerAddress)
  if err != nil { panic(err) }

  // Encodes the message

  bytes, err := protobuf.Encode(packet)
  if err != nil { panic(err) }

  // Sends the message to the peer's UDP address via its Gossip connection

  _, err = gossiper.gossipConn.WriteToUDP(bytes, udpAddr)
  if err != nil { fmt.Println(err)  }

}

/// Sends to every peer
func (gossiper *Gossiper) relay(packet *GossipPacket) {
  for i := 0; i < len(gossiper.peers); i++ {
    gossiper.send(gossiper.peers[i], packet)
  }
}

func (gossiper *Gossiper) listenGossip() {

  buffer := make([]byte, 1024)
  var packet GossipPacket

  for {
    _, _, err := gossiper.gossipConn.ReadFromUDP(buffer)
    if err != nil { fmt.Println(err) }

    protobuf.Decode(buffer, packet)
    gossiper.receiveGossip(packet)
  }
}

func (gossiper *Gossiper) receiveGossip(packet GossipPacket) {
  gossiper.logPeerMessage(packet.Simple)
}

func (gossiper *Gossiper) receiveClient(message Message) {

  gossiper.logClientMessage(message)
  gossipPacket := gossiper.wrapClientMessage(message)
  gossiper.relay(gossipPacket)

}

func (gossiper *Gossiper) wrapClientMessage(message Message) *GossipPacket {

  simpleMessage := SimpleMessage {
    OriginalName:   gossiper.Name,
    RelayPeerAddr:  gossiper.gossipAddress.IP.String(),
    Contents:       message.Text,
  }

  return &GossipPacket { &simpleMessage }
}

func (gossiper *Gossiper) processPeerMessage(message *SimpleMessage) *SimpleMessage {
  message.OriginalName = gossiper.Name
  message.RelayPeerAddr = gossiper.gossipAddress.IP.String()

  // Send

  return message
}

func (gossiper *Gossiper) logClientMessage(message Message) {
  fmt.Printf("CLIENT MESSAGE %v\n", message.Text)
}

func (gossiper *Gossiper) logPeerMessage(message *SimpleMessage) {
  fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
    message.OriginalName, message.RelayPeerAddr, message.Contents)
  fmt.Printf("PEERS %v\n", gossiper.peers)
}
