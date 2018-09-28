package main

import (
    "fmt"
    //"os"
    "flag"
)


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
}
