package main

import (
  "encoding/hex"
  "flag"
  "github.com/dedis/protobuf"
  "github.com/jfperren/Peerster/common"
)


func main () {

  // Define Flags

  uiPort := flag.String("UIPort", "8080", "port for the UI client")
  message := flag.String("msg", "Test message", "message to be sent")
  dest := flag.String("dest", "", "destination for the private message")
  file := flag.String("file", "", "file to be indexed by the gossiper, or filename of the requested file")
  request := flag.String("request", "", "request a chunk or metafile of this hash")

  flag.Parse()

  // Packet to send
  var packet *common.GossipPacket

  switch {

  case *request != "":

    if *file == "" {
      panic("Cannot request a file without giving a name")
    }

    if *dest == "" {
      panic("Cannot request a file without specifying the destination node")
    }

    hash, err := hex.DecodeString(*request)

    if err != nil {
      panic("Error decoding hash specified in 'request'")
    }

    request := &common.DataRequest{*file, *dest, common.InitialHopLimit, hash}
    packet = request.Packed()

  case *file != "":

    reply := &common.DataReply{"client", *file, 0, make([]byte, 0), make([]byte, 0)}
    packet = reply.Packed()

  case *dest != "":

    private := &common.PrivateMessage{"client", 0, *message, *dest, common.InitialHopLimit}
    packet = private.Packed()

  case *message != "":

    simple := &common.SimpleMessage{"client", "", *message}
    packet = simple.Packed()

  default:

    panic("No message specified, unclear instructions.")
  }

  // Encodes message
  bytes, err := protobuf.Encode(packet)
  if err != nil { panic(err) }

  // Send
  socket := common.NewUDPSocket(":5050")
  socket.Send(bytes, ":" + *uiPort)
  socket.Unbind()

}

