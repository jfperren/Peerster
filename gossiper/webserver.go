package gossiper

import (
	"encoding/hex"
	"encoding/json"
	"github.com/jfperren/Peerster/common"
	"net/http"
	"strconv"
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

type Message struct {
	Destination string
	Text string
}

type File struct {
	Name string
	Hash string
}

type FileRequest struct {
	Name string
	Destination string
	Hash string
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
	http.HandleFunc("/privateMessage", middleware(handlePrivateMessage))
	http.HandleFunc("/fileRequest", middleware(handleFileRequest))
	http.HandleFunc("/fileUpload", middleware(handleFileUpload))

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

		if common.Contains(g.Router.Peers, peer) {
			json.NewEncoder(res).Encode("")
		} else {
			g.Router.AddPeerIfNeeded(peer)
			json.NewEncoder(res).Encode(peer)
		}

	case "GET":
		json.NewEncoder(res).Encode(g.Router.Peers)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleUser(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "GET":

		users := make([]*User, 0)

		for k, v := range(g.Router.NextHop) {
			users = append(users, &User{k, v})
		}

		json.NewEncoder(res).Encode(users)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handlePrivateMessage(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "POST":

		var message Message
		err := json.NewDecoder(req.Body).Decode(&message)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		if message.Text == "" {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Message cannot be an empty string.")
			return
		}

		if message.Destination == "" {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Message destination be an empty string.")
			return
		}

		privateMessage := common.NewPrivateMessage("", message.Destination, message.Text)
		g.HandleClient(privateMessage.Packed())

		res.WriteHeader(http.StatusOK)

	case "GET":

		indexHeader := req.Header.Get("x-index")
		index, err := strconv.Atoi(indexHeader)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Error decoding 'x-index' parameter")
			return
		}

		if index < 0 {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Index should be bigger or equal to 0.")
		}

		var body []*common.PrivateMessage

		if  index >= len(g.Messages) {
			body = make([]*common.PrivateMessage, 0)
		} else {
			body = g.Messages[index:]
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(body)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleFileUpload(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "POST":

		var filename string
		err := json.NewDecoder(req.Body).Decode(&filename)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		if filename == "" {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Filename cannot be blank")
			return
		}

		metaFile, err := g.FileSystem.ScanFile(filename)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err.Error())
			return
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(File{
			metaFile.Name,
			hex.EncodeToString(metaFile.Hash),
		});

	case "GET":

		metaFiles := make([]File, 0)

		for _, metaFile := range(g.FileSystem.metaFiles) {
			metaFiles = append(metaFiles, File{
				metaFile.Name,
				hex.EncodeToString(metaFile.Hash),
			})
		}

		res.WriteHeader(http.StatusOK)
		json.NewEncoder(res).Encode(metaFiles)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleFileRequest(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "POST":

		var request FileRequest
		err := json.NewDecoder(req.Body).Decode(&request)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		if request.Name == "" || request.Hash == "" || request.Destination == "" {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Fields cannot be blank")
			return
		}

		hash, err := hex.DecodeString(request.Hash)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode(err)
			return
		}

		dataRequest := &common.DataRequest{
			request.Name,
			request.Destination,
			common.InitialHopLimit,
			hash,
		}

		g.HandleClient(dataRequest.Packed())

		res.WriteHeader(http.StatusOK)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}