package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
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

type TxPublish struct {
	File     File
	HopLimit uint32
}

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

type File struct {
	Name          string
	Size          int64
	MetafileHash  []byte
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

type GossipPacket struct {
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
	TxPublish     *TxPublish
	BlockPublish  *BlockPublish
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

	return &GossipPacket{Simple: simple}
}

// Pack a RumorMessage into a GossipPacket
func (rumor *RumorMessage) Packed() *GossipPacket {

	if rumor == nil {
		panic("Cannot pack <nil> rumor into a GossipPacket")
	}

	return &GossipPacket{Rumor: rumor}
}

// Pack a StatusPacket into a GossipPacket
func (status *StatusPacket) Packed() *GossipPacket {

	if status == nil {
		panic("Cannot pack <nil> status into a GossipPacket")
	}

	return &GossipPacket{Status: status}
}

// Pack a PrivateMessage into a GossipPacket
func (private *PrivateMessage) Packed() *GossipPacket {

	if private == nil {
		panic("Cannot pack <nil> private message into a GossipPacket")
	}

	return &GossipPacket{Private: private}
}

// Pack a DataRequest into a GossipPacket
func (request *DataRequest) Packed() *GossipPacket {

	if request == nil {
		panic("Cannot pack <nil> data request into a GossipPacket")
	}

	return &GossipPacket{DataRequest:request}
}

// Pack a DataReply into a GossipPacket
func (reply *DataReply) Packed() *GossipPacket {

	if reply == nil {
		panic("Cannot pack <nil> data reply into a GossipPacket")
	}

	return &GossipPacket{DataReply: reply}
}

// Pack a SearchRequest into a GossipPacket
func (request *SearchRequest) Packed() *GossipPacket {

	if request == nil {
		panic("Cannot pack <nil> search request into a GossipPacket")
	}

	return &GossipPacket{SearchRequest: request}
}

// Pack a SearchReply into a GossipPacket
func (reply *SearchReply) Packed() *GossipPacket {

	if reply == nil {
		panic("Cannot pack <nil> search reply into a GossipPacket")
	}

	return &GossipPacket{SearchReply: reply}
}

// Pack a SearchReply into a GossipPacket
func (publish *TxPublish) Packed() *GossipPacket {

	if publish == nil {
		panic("Cannot pack <nil> transaction into a GossipPacket")
	}

	return &GossipPacket{TxPublish: publish}
}

// Pack a SearchReply into a GossipPacket
func (publish *BlockPublish) Packed() *GossipPacket {

	if publish == nil {
		panic("Cannot pack <nil> block into a GossipPacket")
	}

	return &GossipPacket{BlockPublish: publish}
}

// Checks if a given GossipPacket is valid. It is only valid if exactly one of its 10 fields is non-nil.
func (packet *GossipPacket) IsValid() bool {
	return boolCount(packet.Rumor != nil)+boolCount(packet.Simple != nil)+
		boolCount(packet.Status != nil)+boolCount(packet.Private != nil)+
		boolCount(packet.DataReply != nil)+boolCount(packet.DataRequest != nil)+
		boolCount(packet.SearchReply != nil)+boolCount(packet.SearchRequest != nil)+
		boolCount(packet.TxPublish != nil)+boolCount(packet.BlockPublish != nil) == 1
}

func (packet *GossipPacket) IsElligibleForBroadcast() bool {
	return !(packet.Simple == nil && packet.SearchRequest == nil && packet.TxPublish == nil && packet.BlockPublish == nil)
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

//
//
//

func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])
	h.Write(b.Nonce[:])
	binary.Write(h,binary.LittleEndian,
		uint32(len(b.Transactions)))
	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h,binary.LittleEndian,
		uint32(len(t.File.Name)))
	h.Write([]byte(t.File.Name))
	h.Write(t.File.MetafileHash)
	copy(out[:], h.Sum(nil))
	return
}

func (block *Block) Str() string {

	hash := block.Hash()
	prev := block.PrevHash
	files := make([]string, 0)

	for _, transaction := range block.Transactions {
		files = append(files, transaction.File.Name)
	}

	return fmt.Sprintf("%v:%v:%v", hex.EncodeToString(hash[:]), hex.EncodeToString(prev[:]), strings.Join(files, FileNameSeparator))
}