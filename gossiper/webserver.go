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
	Rumors   []*common.IRumorMessage
	Statuses []common.PeerStatus
}

type User struct {
	Name    string
	Address string
	Secure  bool
}

type Message struct {
	Destination string
	Text        string
}

type File struct {
	Name string
	Hash string
}

type FileRequest struct {
	Name        string
	Destination string
	Hash        string
}

type SearchRequest struct {
	Keywords    string
	Budget      int
}

type SearchResult struct {
	Name    string
	Hash	string
	Full	bool
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
	http.HandleFunc("/fileDownload", middleware(handleFileDownload))
	http.HandleFunc("/fileUpload", middleware(handleFileUpload))
	http.HandleFunc("/fileSearch", middleware(handleFileSearch))

	go func() {
		err := http.ListenAndServe(":"+port, nil)
		if err == nil {
			panic(err)
		}
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
		if err != nil { handleErr(err, res); return }

		var theirStatuses []common.PeerStatus
		err = json.Unmarshal([]byte(req.Header.Get("x-statuses")), &theirStatuses)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(res).Encode("Error decoding 'x-statuses' parameter")
			return
		}

		command, err := common.NewMessageCommand(message)
		if err != nil { handleErr(err, res); return }

		g.HandleClient(command)

		_, rumors, myStatuses := g.CompareStatus(theirStatuses, ComparisonModeAllNew)
		body := &RumorsAndStatuses{rumors, myStatuses}

		json.NewEncoder(res).Encode(body)

	case "GET":

		var theirStatuses []common.PeerStatus
		err := json.Unmarshal([]byte(req.Header.Get("x-statuses")), &theirStatuses)
		if err != nil { handleErr(err, res); return }

		_, rumors, myStatuses := g.CompareStatus(theirStatuses, ComparisonModeAllNew)

		body := RumorsAndStatuses{rumors, myStatuses}

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

		for k, v := range g.Router.NextHop {

			_, found := g.BlockChain.Peers[k]

			users = append(users, &User{Name: k, Address:v, Secure:found})

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
		if err != nil { handleErr(err, res); return }

		command, err := common.NewPrivateMessageCommand(message.Text, message.Destination)
		if err != nil { handleErr(err, res); return }

		g.HandleClient(command)

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
			return
		}

		var body []*common.PrivateMessage

		if index >= len(g.Messages) {
			body = make([]*common.PrivateMessage, 0)
		} else {
			body = g.Messages[index:]
		}

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
		if err != nil { handleErr(err, res); return }

		command, err := common.NewUploadCommand(filename)
		if err != nil { handleErr(err, res); return }

		err = g.HandleClient(command)
		if err != nil { handleErr(err, res); return }

		metaFile := g.FileSystem.getFileWithName(filename)

		if metaFile == nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(res).Encode(File{
			metaFile.Name,
			hex.EncodeToString(metaFile.Hash),
		})

	case "GET":

		metaFiles := make([]File, 0)

		for _, metaFile := range g.FileSystem.metaFiles {
			metaFiles = append(metaFiles, File{
				metaFile.Name,
				hex.EncodeToString(metaFile.Hash),
			})
		}

		json.NewEncoder(res).Encode(metaFiles)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleFileDownload(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "POST":

		var request FileRequest
		err := json.NewDecoder(req.Body).Decode(&request)
		if err != nil { handleErr(err, res); return }

		command, err := common.NewDownloadCommand(request.Hash, request.Name, request.Destination)
		if err != nil { handleErr(err, res); return }

		g.HandleClient(command)

		res.WriteHeader(http.StatusOK)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleFileSearch(res http.ResponseWriter, req *http.Request) {

	switch req.Method {

	case "POST":

		var searchRequest SearchRequest
		err := json.NewDecoder(req.Body).Decode(&searchRequest)
		if err != nil { handleErr(err, res); return }

		command, err := common.NewSearchCommand(searchRequest.Keywords, uint64(searchRequest.Budget))
		if err != nil { handleErr(err, res); return }

		err = g.HandleClient(command)
		if err != nil { handleErr(err, res); return }

		res.WriteHeader(http.StatusOK)

	case "GET":

		results := make([]SearchResult, 0)

		for _, result := range g.SearchEngine.results {

			hash := hex.EncodeToString(result.MetafileHash)

			results = append(results, SearchResult{
				Name: result.FileName,
				Hash: hash,
				Full: g.SearchEngine.fileMaps[hash].isComplete(),
			})
		}

		json.NewEncoder(res).Encode(results)

	default:
		res.WriteHeader(http.StatusMethodNotAllowed)
	}
}


func handleErr(err error, res http.ResponseWriter) bool {

	res.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(res).Encode(err.Error())
	return true
}