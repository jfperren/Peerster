package main

import (
	"flag"
	"github.com/jfperren/Peerster/common"
	"github.com/jfperren/Peerster/gossiper"
	"os"
	"os/signal"
	"syscall"
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
	verbose := flag.Bool("verbose", false, "display additional logs (useful for testing)")

  	flag.Parse()

	var g *gossiper.Gossiper

  	if *server {
		g = gossiper.NewGossiper(*gossipAddr, "",  *name, *peers, *simple, *rtimer)
		gossiper.StartWebServer(g, *uiPort)
		common.DebugStartGossiper("no_client_address", g.GossipSocket.Address, g.Name, g.Router.Peers, g.Simple, g.Router.Rtimer)
	} else {
		g = gossiper.NewGossiper(*gossipAddr, ":" + *uiPort,  *name, *peers, *simple, *rtimer)
		common.DebugStartGossiper(g.ClientSocket.Address, g.GossipSocket.Address, g.Name, g.Router.Peers, g.Simple, g.Router.Rtimer)
	}

	common.Verbose = *verbose

	g.Start()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		g.Stop()
		common.DebugStopGossiper()
		os.Exit(1)
	}()
}
