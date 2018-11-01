package gossiper

import (
	"encoding/json"
	"github.com/jfperren/Peerster/common"
	"net/http"
)

var g *Gossiper

type RumorsAndStatuses struct {
	Rumors []*common.RumorMessage
	Statuses []common.PeerStatus
}

type User struct {
	Name string
	Address string
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
	http.HandleFunc("/user", middleware(handleUser))

	go func () {
		err := http.ListenAndServe(":" + port, nil)
		if err == nil { panic(err) }
	}()
}

func middleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		common.DebugServerRequest(req)
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

		var theirStatuses []common.PeerStatus
		err = json.Unmarshal([]byte(req.Header.Get("x-statuses")), &theirStatuses)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Error decoding 'x-statuses' parameter")
			return
		}

		simpleMessage := &common.SimpleMessage{"", "", message}
		g.HandleClient(simpleMessage.Packed())

		_, rumors, myStatuses := g.CompareStatus(theirStatuses, ComparisonModeAllNew)
		body := &RumorsAndStatuses{rumors, myStatuses}

		//res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(body)

	case "GET":

		var theirStatuses []common.PeerStatus
		err := json.Unmarshal([]byte(req.Header.Get("x-statuses")), &theirStatuses)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		_, rumors, myStatuses := g.CompareStatus(theirStatuses, ComparisonModeAllNew)
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

		if common.Contains(g.Peers, peer) {
			json.NewEncoder(res).Encode("")
		} else {
			g.AddPeerIfNeeded(peer)
			json.NewEncoder(res).Encode(peer)
		}

	case "GET":
		json.NewEncoder(res).Encode(g.Peers)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleUser(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "GET":

		users := make([]*User, 0)

		for k, v := range(g.NextHop) {
			users = append(users, &User{k, v})
		}

		json.NewEncoder(res).Encode(users)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}
