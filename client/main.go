package main

import (
  "fmt"
  "flag"
  "github.com/dedis/protobuf"
  "github.com/jfperren/Peerster/common"
  "net"
)


func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  message := flag.String("msg", "Test message", "message to be sent")

  flag.Parse()

  // Print Flags

  // fmt.Printf("port = %v\n", *uiPort)
  // fmt.Printf("message = %v\n", *message)

  fmt.Printf("Sending message '%v' to port %v\n", *message, *uiPort)

  // Simply sends message via UDP. Note that we need to use a port different
  // from UIPort otherwise there will be some errors.

  // Resolves local address
  localAddr, err := net.ResolveUDPAddr("udp4", ":5050")
  if err != nil { panic(err) }

  // Bind local address
  conn, err := net.ListenUDP("udp4", localAddr)
  if err != nil { panic(err) }

  // Then resolves address for gossiper
  uiAddr, err := net.ResolveUDPAddr("udp4", ":" + *uiPort)
  if err != nil { panic(err) }

  // Creates packet
  packet := &common.GossipPacket{&common.SimpleMessage{"", "", *message},nil,nil}

  // Encodes message
  bytes, err := protobuf.Encode(packet)
  if err != nil { panic(err) }

  // Sends message bytes to gossiper via UDP
  _,  err = conn.WriteToUDP(bytes, uiAddr)
  if err != nil { panic(err) }

  // Close connection

  // conn.Close()
}
