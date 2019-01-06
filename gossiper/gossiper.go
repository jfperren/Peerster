package gossiper

import (
	"github.com/dedis/protobuf"
	"github.com/jfperren/Peerster/common"
	"log"
	"sync"
	"time"
)

// Root class of the Peerster program. It represents a "node" in the network. In this file, the more
// complex functions are implemented (rumor-mongering, downloading, handling packets,etc...).
//
// Lower-level functions are implemented in FileSystem (slicing and storing files & hashes), Dispatcher
// (dispatching packets to correct goroutines), Router (keep track of peers & routing table), RumorsDB
// (store rumors and compute vector clocks), on all of which Gossiper relies on.
type Gossiper struct {
	Name   			string // Name of this node
	Simple 			bool   // Stores if gossiper runs in simple mode.
	MixLength       uint   // Number of hops messages should go through

	GossipSocket 	*common.UDPSocket // UDP Socket that connects to other nodes
	ClientSocket 	*common.UDPSocket // UDP Socket that connects to the client

	Rumors   		*RumorDatabase           // Database of known Rumors
	Messages 		[]*common.PrivateMessage // List of Private Messages Received

	FileSystem 		*FileSystem 	// Stores and serves shared files
	Dispatcher 		*Dispatcher 	// Dispatches incoming messages to expecting processes
	Router     		*Router     	// Handles routing to neighboring and non-neighboring nodes.
	SpamDetector 	*SpamDetector
	SearchEngine 	*SearchEngine 	//
	BlockChain		*BlockChain
    Crypto          *Crypto         // Stores the RSA keys, and handle the (de)cyphering and
                                    // signing/validating messages
    Mixer 			*Mixer // Stores pending packets to be forwarded through a mix-network
}

const (
	ComparisonModeMissingOrNew = iota // Flag to be used when comparing two nodes' status packets
	ComparisonModeAllNew       = iota // Flag to be used when comparing a node status with the client status
)

// Create a new Gossiper using the given addresses.
//
//  - gossipAddress: Address on which the gossiper listens to other gossipers
//  - clientAddress: Address on which the gossiper listens for commands from the CLI
//  - name: Name of this gossiper (it will appear in this node's messages, should be unique).
//  - peers: list of IP addresses to which this gossiper is connected at start. More can be added later on.
//  - simple: Start this gossiper in simple mode (i.e. no gossip, only simple messages)
//  - rtimer: Time in seconds between route rumors. Set to 0 for not sending route rumors at all.
//  - separatefs: True if this gossiper uses its own subfolder for _Download and _SharedFiles.
//
// Note - Use gossiper.Start() to Start listening for messages.
//
func NewGossiper(gossipAddress, clientAddress, name string, peers string, simple bool, rtimer int, separatefs bool, keySize, cryptoOpts int, mixLength uint) *Gossiper {

	gossipSocket := common.NewUDPSocket(gossipAddress)
	var clientSocket *common.UDPSocket

	if clientAddress != "" {
		clientSocket = common.NewUDPSocket(clientAddress)
	}

	downloadPath := common.DownloadDir
	sharedPath := common.SharedFilesDir

	if separatefs {
		downloadPath = downloadPath + name + "/"
		sharedPath = sharedPath + name + "/"
	}

	var mixer *Mixer
	if cryptoOpts != 0 {
		mixer = NewMixer()
	}

	return &Gossiper{
		Name:         	name,
		Simple:       	simple,
		GossipSocket: 	gossipSocket,
		ClientSocket: 	clientSocket,
		MixLength: 		mixLength,

		Rumors:     	NewRumorDatabase(),
		FileSystem: 	NewFileSystem(sharedPath, downloadPath),
		Dispatcher: 	NewDispatcher(),
		Router:     	NewRouter(peers, time.Duration(rtimer)*time.Second),
		SpamDetector:   NewSpamDetector(),
		SearchEngine: 	NewSearchEngine(),
		BlockChain:		NewBlockChain(),
        Crypto:         NewCrypto(keySize, cryptoOpts),
		Mixer:			mixer,
	}
}

// --
// --  START & STOP
// --


// Start listening for UDP packets on Gossiper's clientAddress & gossipAddress
func (gossiper *Gossiper) Start() {

	if gossiper.ShouldAuthenticate() {
		go gossiper.tryAuthenticate()
	}

	go gossiper.receiveGossip()
	go gossiper.sendRouteRumors()

	if !gossiper.Simple {
		go gossiper.antiEntropy()
	}

	if gossiper.ClientSocket != nil {
		go gossiper.receiveClient()
	}

	go gossiper.waitForNewBlocks()
	go gossiper.BlockChain.mine()

	if gossiper.Mixer != nil {
		go gossiper.ReleaseOnions()
	}

	// Allows the loops to run indefinitely after the main code is completed.
	wg := new(sync.WaitGroup)
	wg.Add(6)
	wg.Wait()
}

// Unbind from all ports, stop processes.
func (gossiper *Gossiper) Stop() {
	gossiper.ClientSocket.Unbind()
	gossiper.GossipSocket.Unbind()
}

//
//  EVENT LOOPS
//

// Main loop for handling gossip packets from other nodes.
func (gossiper *Gossiper) receiveGossip() {
	for {
		bytes, source, alive := gossiper.GossipSocket.Receive()

		if !alive {
			break
		}

        if gossiper.handleReceivedPacket(bytes, source) {
            continue
        }

		gossiper.Router.AddPeerIfNeeded(source)
	}
}

func (gossiper *Gossiper) handleReceivedPacket(bytes []byte, source string) bool {
    var packet common.GossipPacket
    protobuf.Decode(bytes, &packet)

    if !packet.IsValid() {
        common.DebugInvalidPacket(&packet)
        return true
    }

    go gossiper.HandleGossip(&packet, source)
    return false
}


// Main loop for handling client packets.
func (gossiper *Gossiper) receiveClient() {

	for {

		var command common.Command
		bytes, _, alive := gossiper.ClientSocket.Receive()

		if !alive {
			break
		}

		protobuf.Decode(bytes, &command)
		go gossiper.HandleClient(&command)
	}
}

//
//  HANDLING NEW PACKETS
//

//  Handle new packet from client
func (gossiper *Gossiper) HandleClient(command *common.Command) error {

	if command == nil || !command.IsValid() {
		return common.InvalidCommandError()
	}

	switch {

	case command.Message != nil:

		content := command.Message.Content
		common.LogClientMessage(content)

		if gossiper.Simple {

			message := common.NewSimpleMessage(gossiper.Name, gossiper.GossipSocket.Address, content)
			go gossiper.broadcastToNeighbors(message.Packed())

		} else {

			rumor := gossiper.GenerateRumor(content)

			gossiper.Rumors.Put(rumor)

			peer, found := gossiper.Router.randomPeer()

			if found {
				go gossiper.rumormonger(rumor, peer)
			}
		}

	case command.PrivateMessage != nil:

		destination := command.PrivateMessage.Destination
		content := command.PrivateMessage.Content

		private := common.NewPrivateMessage(gossiper.Name,destination, content)

		destined := gossiper.sendToNode(private.Packed(), destination, nil)
		gossiper.Messages = append(gossiper.Messages, private)

		if destined {
			common.LogPrivate(private)
		}

	case command.Download != nil:

		destination := command.Download.Destination
		filename := command.Download.FileName
		hash := command.Download.Hash

		go gossiper.StartDownload(filename, hash, destination, 0)

	case command.Upload != nil:

		metaFile, err := gossiper.FileSystem.ScanFile(command.Upload.FileName)
		if err != nil { return err }

		transaction := gossiper.NewTransaction(metaFile)

		if gossiper.BlockChain.TryAddFile(transaction) {
			gossiper.broadcastToNeighbors(transaction.Packed())
			common.DebugBroadcastTransaction(transaction)
		}

	case command.Search != nil:

		gossiper.RingSearch(command.Search.Keywords, command.Search.Budget)
	}

	return nil
}

// Handle packet from another node.
func (gossiper *Gossiper) HandleGossip(packet *common.GossipPacket, source string) {

	if packet == nil || !packet.IsValid() {
		common.DebugInvalidPacket(packet)
		return // Fail gracefully
	}

	destination := packet.GetDestination()

	if packet.Signature != nil && (destination == nil || *destination == gossiper.Name) {

		publicKey, exists := gossiper.BlockChain.GetPublicKey(packet.Signature.Origin)

		if !exists {
			common.DebugDropUnauthenticatedOrigin(packet.Signature)
			return
		}

		hash := packet.Hash()

		if !gossiper.Crypto.Verify(hash[:], packet.Signature.Signature, publicKey) {
			common.DebugDropIncorrectSignature(packet.Signature)
			return
		}
	}

	switch {

	case packet.Simple != nil:

		common.LogSimpleMessage(packet.Simple)
		common.LogPeers(gossiper.Router.Peers)

		go gossiper.broadcastToNeighbors(packet)

	case packet.Rumor != nil:

		if packet.Rumor.IsRouteRumor() {
			common.DebugReceiveRouteRumor(packet.Rumor.Origin, source)
		} else {
			common.LogRumor(packet.Rumor, source)
		}

		common.LogPeers(gossiper.Router.Peers)

		gossiper.handleRumor(packet.Rumor, source)

		statusPacket := gossiper.GenerateStatusPacket()
		common.DebugSendStatus(statusPacket, source)
		go gossiper.sendToNeighbor(source, statusPacket.Packed())

	case packet.Status != nil:

		common.LogStatus(packet.Status, source)
		common.LogPeers(gossiper.Router.Peers)

		expected := gossiper.Dispatcher.dispatchStatusPacket(source, packet)

		if !expected {

			rumor, _, _ := gossiper.CompareStatus(packet.Status.Want, ComparisonModeMissingOrNew)

			if rumor != nil {
				go gossiper.rumormonger(*rumor, source)
			}
		}

	case packet.Private != nil:

		destination := packet.Private.Destination
		hopLimit := &packet.Private.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {
			gossiper.Messages = append(gossiper.Messages, packet.Private)
			common.LogPrivate(packet.Private)
		}

	case packet.DataReply != nil:

		destination := packet.DataReply.Destination
		hopLimit := &packet.DataReply.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {

			common.DebugReceiveDataReply(packet.DataReply)

			gossiper.Dispatcher.dispatchDataReply(packet)
		}

	case packet.DataRequest != nil:

		destination := packet.DataRequest.Destination
		hopLimit := &packet.DataRequest.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {

			common.DebugReceiveDataRequest(packet.DataRequest)

			reply, ok := gossiper.GenerateDataReply(packet.DataRequest)

			if ok {
				gossiper.sendToNode(reply.Packed(), reply.Destination, nil)
			}
		}

	case packet.SearchRequest != nil:

		if packet.SearchRequest.Budget <= 0 {
			return
		}

		if !gossiper.SpamDetector.shouldProcessSearchRequest(packet.SearchRequest) {
			common.DebugIgnoreSpam(packet.SearchRequest.Origin, packet.SearchRequest.Keywords)
			return
		}

		if packet.SearchRequest.Origin == gossiper.Name {
			return
		}

		common.DebugProcessSearchRequest(packet.SearchRequest.Origin, packet.SearchRequest.Keywords)

		go gossiper.forwardSearchRequest(packet.SearchRequest, source)

		results := gossiper.FileSystem.Search(packet.SearchRequest.Keywords)
		reply := common.NewSearchReply(gossiper.Name, packet.SearchRequest.Origin, results)

		common.DebugServeSeachReply(reply)

		go gossiper.sendToNode(reply.Packed(), reply.Destination, nil)

	case packet.SearchReply != nil:

		destination := packet.SearchReply.Destination
		hopLimit := &packet.SearchReply.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {
			go gossiper.SearchEngine.StoreResults(packet.SearchReply.Results, packet.SearchReply.Origin)
		}

	case packet.TxPublish != nil:

        gossiper.handleRumor(packet.TxPublish, source)
		common.DebugReceiveTransaction(packet.TxPublish)

		if gossiper.BlockChain.TryAddTransaction(packet.TxPublish) {


			packet.TxPublish.HopLimit--

			if packet.TxPublish.HopLimit > 0 {
				gossiper.broadcastToNeighborsExcept(packet.TxPublish.Packed(), &[]string{source})
			}
		}

	case packet.BlockPublish != nil:

        gossiper.handleRumor(packet.BlockPublish, source)
		if gossiper.BlockChain.TryAddBlock(&packet.BlockPublish.Block) {

			packet.BlockPublish.HopLimit--

			if packet.BlockPublish.HopLimit > 0 {
				gossiper.broadcastToNeighborsExcept(packet.BlockPublish.Packed(), &[]string{source})
			}
		}

    case packet.Cyphered != nil:
        destination := packet.Cyphered.Destination
		hopLimit := &packet.Cyphered.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {
            // decypher symmetric key
            symmetricKey := gossiper.Crypto.Decypher(packet.Cyphered.Key)
            // decypher payload
            signedBytes, err := CTRDecipher(packet.Cyphered.Payload, symmetricKey, packet.Cyphered.IV)
            if err != nil {
                log.Println(err)
                return
            }
            var signed common.GossipPacket
            err = Decode(signedBytes, &signed)

            gossiper.HandleGossip(&signed, source)
        }

	case packet.Onion != nil:

		destination := packet.Onion.Destination
		hopLimit := &packet.Onion.HopLimit

		destined := gossiper.sendToNode(packet, destination, hopLimit)

		if destined {

			gossipPacket, _, err := gossiper.ProcessOnion(packet.Onion)

			if err != nil {
				// Drop the packet
			} else if gossipPacket != nil {
				// Process the packet as a normal packet
				gossiper.HandleGossip(gossipPacket, source)
			} else if gossiper.Mixer != nil {
				// Give it to the mixer logic to store and forward later on
				gossiper.Mixer.ForwardPacket(packet.Onion)
			}
		}
	}
}

// Generate a status packet with the current vector clock.
func (gossiper *Gossiper) GenerateStatusPacket() *common.StatusPacket {

	peerStatuses := make([]common.PeerStatus, 0)

	for _, origin := range gossiper.Rumors.AllOrigins() {
		peerStatuses = append(peerStatuses, common.PeerStatus{origin, gossiper.Rumors.NextIDFor(origin)})
	}

	return &common.StatusPacket{peerStatuses}
}

// Generate a data reply to a given request
func (gossiper *Gossiper) GenerateDataReply(request *common.DataRequest) (*common.DataReply, bool) {

	var data []byte

	metaHash, found := gossiper.FileSystem.getMetaFile(request.HashValue)

	if found {

		data = metaHash.Data

	} else {

		chunk, found := gossiper.FileSystem.getChunk(request.HashValue)

		if !found {
			common.DebugHashNotFound(request.HashValue, request.Origin)
			return nil, false
		}

		data = chunk.data
	}

	return &common.DataReply{
		gossiper.Name,
		request.Origin,
		common.InitialHopLimit,
		request.HashValue,
		data,
	}, true
}

// Generate a data request to a given file
func (gossiper *Gossiper) GenerateDataRequest(destination string, hash []byte) *common.DataRequest {
	return &common.DataRequest{
		gossiper.Name,
		destination,
		common.InitialHopLimit,
		hash,
	}
}

// Generate a new Rumor based on the string.
func (gossiper *Gossiper) GenerateRumor(message string) *common.RumorMessage {

	rumor := &common.RumorMessage{
		Origin: gossiper.Name,
		ID:     gossiper.Rumors.ConsumeNextID(),
		Text:   message,
	}

	gossiper.Rumors.Put(rumor)

	return rumor
}

// Generate a route rumor
func (gossiper *Gossiper) GenerateRouteRumor() *common.RumorMessage {
	return gossiper.GenerateRumor("")
}

// Send ready onion packets
func (gossiper *Gossiper) ReleaseOnions() {
	for {
		packet := <- gossiper.Mixer.ToSend
		gossiper.sendToNode(packet.Packed(), packet.Destination, &packet.HopLimit)
	}
}

//////////////
//  CRYPTO  //
//////////////
func (gossiper *Gossiper) SignPacket(packet *common.GossipPacket) *common.Signature {

    if !packet.IsValid() {
		log.Printf("Sending invalid packet: %v\n", packet)
        return nil
	}

    hash := packet.Hash()

    return &common.Signature{
        Origin: gossiper.Name,
        Signature: gossiper.Crypto.Sign(hash[:]),
    }
}

func (gossiper *Gossiper) CypherPacket(packet *common.GossipPacket, destination string) *common.CypheredMessage {

    // get public key of destination
    publicKey, exists := gossiper.BlockChain.GetPublicKey(destination)

    if exists {
        bytes, err := EncodeBlock(packet, common.CTRKeySize)
        if err != nil {
            log.Println(err)
            return nil
        }

        symmetricKey := NewCTRSecret()
        cypheredPayload, iv, err := CTRCipher(bytes, symmetricKey)
        if err != nil {
            log.Println(err)
            return nil
        }

        cypheredKey := gossiper.Crypto.Cypher(symmetricKey, publicKey)

        return &common.CypheredMessage{
            Destination: destination,
            HopLimit: common.InitialHopLimit,
            Payload: cypheredPayload,
            IV: iv,
            Key: cypheredKey,
        }
    } else {
		common.DebugDropCannotCipher(packet)
        return nil
    }
}
