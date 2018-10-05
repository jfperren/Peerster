package main

import (
    "fmt"
    //"os"
    "flag"
    "net"
    "strings"
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
  address *net.UDPAddr
  conn *net.UDPConn
  peers []string
  Name string
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

  gossiper := NewGossiper(*gossipAddr, *name, *peers)
  message := SimpleMessage { "nodeB", "127.0.0.1:5002", "Bonjour" }
  gossiper.logClientMessage(&message)
  gossiper.logPeerMessage(&message)
}



func NewGossiper(address, name string, peersList string) *Gossiper {

  udpAddr, _ := net.ResolveUDPAddr("udp4", address)
  udpConn, _ := net.ListenUDP("udp4", udpAddr)

  return &Gossiper{
    address:  udpAddr,
    conn:     udpConn,
    Name:     name,
    peers:    strings.Split(peersList, ","),
  }
}

func (gossiper *Gossiper) processClientMessage(message *SimpleMessage) *SimpleMessage {
  message.OriginalName = gossiper.Name
  message.RelayPeerAddr = gossiper.address.IP.String()

  // Send

  return message
}

func (gossiper *Gossiper) processPeerMessage(message *SimpleMessage) *SimpleMessage {
  message.OriginalName = gossiper.Name
  message.RelayPeerAddr = gossiper.address.IP.String()

  // Send

  return message
}

func (gossiper *Gossiper) logClientMessage(message *SimpleMessage) {
  fmt.Printf("CLIENT MESSAGE %v\n", message.Contents)
}

func (gossiper *Gossiper) logPeerMessage(message *SimpleMessage) {
  fmt.Printf("SIMPLE MESSAGE origin %v from %v contents %v\n",
    message.OriginalName, message.RelayPeerAddr, message.Contents)
  fmt.Printf("PEERS %v\n", gossiper.peers)
}
