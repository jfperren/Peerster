package common

import (
	"bytes"
	"crypto/sha256"
)

// --
// -- DATA STRUCTURES
// --

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}

type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

type GossipPacket struct {
	Simple      	*SimpleMessage
	Rumor       	*RumorMessage
	Status      	*StatusPacket
	Private     	*PrivateMessage
	DataRequest 	*DataRequest
	DataReply   	*DataReply
	SearchRequest	*SearchRequest
	SearchReply   	*SearchReply
}

// --
// -- INITIALIZERS
// --

func NewSimpleMessage(origin, address, contents string) *SimpleMessage {
	return &SimpleMessage{
		origin,
		address,
		contents,
	}
}

func NewPrivateMessage(origin, destination, contents string) *PrivateMessage {

	return &PrivateMessage{
		Origin:      origin,
		ID:          0,
		Text:        contents,
		Destination: destination,
		HopLimit:    InitialHopLimit,
	}
}

func NewSearchReply(origin, destination string, results []*SearchResult) *SearchReply {

	return &SearchReply{
		Origin:      origin,
		Destination: destination,
		HopLimit:    InitialHopLimit,
		Results:	 results,
	}
}

func CopySearchRequest(request *SearchRequest, budget uint64) *SearchRequest {

	return &SearchRequest{
		Origin: request.Origin,
		Budget:	budget,
		Keywords: request.Keywords,
	}
}


// --
// -- CONVENIENCE METHODS
// --

// Pack a SimpleMessage into a GossipPacket
func (simple *SimpleMessage) Packed() *GossipPacket {

	if simple == nil {
		panic("Cannot pack <nil> Simple into a GossipPacket")
	}

	return &GossipPacket{simple, nil, nil, nil,
		nil, nil, nil, nil}
}

// Pack a RumorMessage into a GossipPacket
func (rumor *RumorMessage) Packed() *GossipPacket {

	if rumor == nil {
		panic("Cannot pack <nil> rumor into a GossipPacket")
	}

	return &GossipPacket{nil, rumor, nil, nil,
		nil, nil, nil, nil}
}

// Pack a StatusPacket into a GossipPacket
func (status *StatusPacket) Packed() *GossipPacket {

	if status == nil {
		panic("Cannot pack <nil> status into a GossipPacket")
	}

	return &GossipPacket{nil, nil, status, nil,
		nil, nil, nil, nil}
}

// Pack a PrivateMessage into a GossipPacket
func (private *PrivateMessage) Packed() *GossipPacket {

	if private == nil {
		panic("Cannot pack <nil> private message into a GossipPacket")
	}

	return &GossipPacket{nil, nil, nil, private,
		nil, nil, nil, nil}
}

// Pack a DataRequest into a GossipPacket
func (request *DataRequest) Packed() *GossipPacket {

	if request == nil {
		panic("Cannot pack <nil> data request into a GossipPacket")
	}

	return &GossipPacket{nil, nil, nil, nil,
		request, nil, nil, nil}
}

// Pack a DataReply into a GossipPacket
func (reply *DataReply) Packed() *GossipPacket {

	if reply == nil {
		panic("Cannot pack <nil> data reply into a GossipPacket")
	}

	return &GossipPacket{nil, nil, nil, nil,
		nil, reply, nil, nil}
}

// Pack a SearchRequest into a GossipPacket
func (request *SearchRequest) Packed() *GossipPacket {

	if request == nil {
		panic("Cannot pack <nil> search request into a GossipPacket")
	}

	return &GossipPacket{nil, nil, nil, nil,
		nil, nil, request, nil}
}

// Pack a SearchReply into a GossipPacket
func (reply *SearchReply) Packed() *GossipPacket {

	if reply == nil {
		panic("Cannot pack <nil> search reply into a GossipPacket")
	}

	return &GossipPacket{nil, nil, nil, nil,
		nil, nil, nil, reply}
}

// Checks if a given GossipPacket is valid. It is only valid if exactly one of its 6 fields is non-nil.
func (packet *GossipPacket) IsValid() bool {
	return boolCount(packet.Rumor != nil)+boolCount(packet.Simple != nil)+
		boolCount(packet.Status != nil)+boolCount(packet.Private != nil)+
		boolCount(packet.DataReply != nil)+boolCount(packet.DataRequest != nil)+
		boolCount(packet.SearchReply != nil)+boolCount(packet.SearchRequest != nil) == 1
}

func (reply *DataReply) VerifyHash(expected []byte) bool {

	computedHash := sha256.Sum256(reply.Data)
	dataIsConsistent := bytes.Compare(computedHash[:], reply.HashValue) == 0
	hashIsExpected := bytes.Compare(reply.HashValue, expected) == 0

	return dataIsConsistent && hashIsExpected
}

func (rumor *RumorMessage) IsRouteRumor() bool {
	return rumor.Text == ""
}
