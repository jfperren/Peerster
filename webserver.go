package main

import (
	"encoding/json"
	"net/http"
)

var g *Gossiper

type RumorsAndStatuses struct {
	Rumors []*RumorMessage
	Statuses []PeerStatus
}

func StartWebServer(gossiper *Gossiper, port string) {

	// Stores gossiper
	g = gossiper

	// Files
	http.Handle("/", http.FileServer(http.Dir("./web")))

	// Routes
	http.HandleFunc("/id", middleware(handleId))
	http.HandleFunc("/message", middleware(handleMessage))
	http.HandleFunc("/node", middleware(handleNode))

	go func () {
		err := http.ListenAndServe(":" + port, nil)
		if err == nil { panic(err) }
	}()
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
		err := json.NewDecoder(req.Body).Decode(&message)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		if message == "" {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Message cannot be an empty string.")
			return
		}

		var theirStatuses []PeerStatus
		err = json.Unmarshal([]byte(req.Header.Get("x-statuses")), &theirStatuses)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Error decoding 'x-statuses' parameter")
			return
		}

		simpleMessage := &SimpleMessage{"", "", message}
		g.handleClient(simpleMessage.packed())

		_, rumors, myStatuses := g.compareStatus(theirStatuses, ComparisonModeAllNew)
		body := &RumorsAndStatuses{rumors, myStatuses}

		//res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(body)

	case "GET":

		var theirStatuses []PeerStatus
		err := json.Unmarshal([]byte(req.Header.Get("x-statuses")), &theirStatuses)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		_, rumors, myStatuses := g.compareStatus(theirStatuses, ComparisonModeAllNew)
		body := RumorsAndStatuses{rumors, myStatuses}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(body)

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

		if containsString(g.peers, peer) {
			json.NewEncoder(res).Encode("")
		} else {
			g.addPeerIfNeeded(peer)
			json.NewEncoder(res).Encode(peer)
		}

	case "GET":
		json.NewEncoder(res).Encode(g.peers)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}
