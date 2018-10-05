package main

import (
  "fmt"
  "flag"
  "github.com/dedis/protobuf"
  "net"
)

type Message struct {
  Text string
}

func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  message := flag.String("msg", "Test message", "message to be sent")

  flag.Parse()

  // Print Flags

  // fmt.Printf("port = %v\n", *uiPort)
  // fmt.Printf("message = %v\n", *message)

  fmt.Printf("Sending message '%v' to port %v", *message, *uiPort)

  // Simply sends message via UDP. Note that we need to use a port different
  // from UIPort otherwise there will be some errors.

  udpAddr, addrErr := net.ResolveUDPAddr("udp4", ":5050")
  if addrErr != nil { panic(addrErr) }

  udpConn, connErr := net.ListenUDP("udp4", udpAddr)
  if connErr != nil { panic(connErr) }

  // Then resolves address for gossiper

  gossipAddr, gossipErr := net.ResolveUDPAddr("udp4", ":8080")
  if gossipErr != nil { panic(addrErr) }

  // Encodes the message

  bytes, encodeErr := protobuf.Encode(&Message{ *message })
  if encodeErr != nil { panic(addrErr) }

  // Sends message bytes to gossiper's UIPort via UDP

  _, err := udpConn.WriteToUDP(bytes, gossipAddr)
  if err != nil { fmt.Println("FML")  }

  // Close connection

  udpConn.Close()
}
