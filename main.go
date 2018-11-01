package main

import (
	"flag"
	"fmt"
	"github.com/jfperren/Peerster/gossiper"
)

func main () {

  	// Define Flags

  	uiPort := flag.String("UIPort", "8080", "port for the UI client")
  	gossipAddr := flag.String("gossipAddr", "127.0.0.1:5000", "port for the gossiper")
  	name := flag.String("name", "REQUIRED", "name of the gossiper")
  	peers := flag.String("peers", "REQUIRED", "comma separated list of peers of the form ip:port")
  	simple := flag.Bool("simple", false, "runs gossiper in simple broadcast mode")
  	server := flag.Bool("server", false, "runs this node in server mode")
  	rtimer := flag.Int("rtimer", 0, "route rumors sending period in seconds, 0 to disable sending of route rumors.")

  	flag.Parse()

  	// Print Flags

  	fmt.Printf("port = %v\n", *uiPort)
  	fmt.Printf("gossipAddr = %v\n", *gossipAddr)
  	fmt.Printf("name = %v\n", *name)
  	fmt.Printf("peers = %v\n", *peers)
  	fmt.Printf("simple = %v\n", *simple)
	fmt.Printf("server = %v\n", *server)
  	fmt.Printf("rtimer = %v\n", *rtimer)

	var g *gossiper.Gossiper

  	if *server {
		g = gossiper.NewGossiper(*gossipAddr, "",  *name, *peers, *simple, *rtimer)
		gossiper.StartWebServer(g, *uiPort)
	} else {
		g = gossiper.NewGossiper(*gossipAddr, ":" + *uiPort,  *name, *peers, *simple, *rtimer)
	}

	g.Start()
}
