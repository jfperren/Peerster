package main

import (
	"encoding/json"
	"net/http"
)

const WEB_SERVER_PORT = ":8080"

var g *Gossiper
var status []PeerStatus

func StartWebServer(gossiper *Gossiper) {

	// Stores gossiper
	g = gossiper
	status = make([]PeerStatus, 0)

	// Files
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// Routes
	http.HandleFunc("/id", middleware(handleId))
	http.HandleFunc("/message", middleware(handleMessage))
	http.HandleFunc("/node", middleware(handleNode))

	err := http.ListenAndServe(WEB_SERVER_PORT, nil)
	if err == nil { panic(err) }
}

func middleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		debugServerRequest(req)
		handler(res, req)
	}
}

func handleId(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "GET":
		json.NewEncoder(res).Encode(g.Name)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleMessage(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":

		var message string

		json.NewDecoder(req.Body).Decode(&message)

		if message == "" {
			res.WriteHeader(http.StatusBadRequest)
		}

		simpleMessage := &SimpleMessage{"GUI", WEB_SERVER_PORT, message}
		g.handleClient(simpleMessage.packed())

		json.NewEncoder(res).Encode(getNewRumors())

	case "GET":

		json.NewEncoder(res).Encode(getNewRumors())

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getNewRumors() []*RumorMessage {
	rumors, hasChanged := g.rumors.newRumorsSince(status)

	if hasChanged {
		status = g.generateStatusPacket().Want
	}

	return rumors
}

func handleNode(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":

		var peer string

		json.NewDecoder(req.Body).Decode(&peer)

		if peer == "" {
			res.WriteHeader(http.StatusBadRequest)
		}

		json.NewEncoder(res).Encode(g.peers)

	case "GET":
		json.NewEncoder(res).Encode(g.peers)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}
