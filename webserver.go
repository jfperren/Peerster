package main

import (
	"encoding/json"
	"net/http"
)

const WEB_SERVER_PORT = ":8080"

var g *Gossiper

func StartWebServer(gossiper *Gossiper) {

	// Stores gossiper
	g = gossiper

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

		var peer string

		json.NewDecoder(req.Body).Decode(&peer)

		if peer == "" {
			res.WriteHeader(http.StatusBadRequest)
		}

		g.addPeer(peer)
		json.NewEncoder(res).Encode(g.peers)

	case "GET":
		json.NewEncoder(res).Encode(g.peers)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleNode(res http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case "POST":

		var peer string

		json.NewDecoder(req.Body).Decode(&peer)

		if peer == "" {
			res.WriteHeader(http.StatusBadRequest)
		}

		g.addPeer(peer)
		json.NewEncoder(res).Encode(g.peers)

	case "GET":
		json.NewEncoder(res).Encode(g.peers)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}
