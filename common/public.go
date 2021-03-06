package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

//
//  DATA STRUCTURES
//

// A simple message. To be used only in simple mode.
type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type IRumorMessage interface {
    GetOrigin() string
    GetID() uint32
    Packed() *GossipPacket
}

// A rumor that should be sent to every node.
type RumorMessage struct {
	Origin 		string
	ID     		uint32
	Text   		string
}

// A packet to ask and download a chunk of a known file
type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

// A packet containing a chunk of a known file
type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

// A private message between two nodes
type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

// Packet containing all known nodes' information about missing rumors
type StatusPacket struct {
	Want 		[]PeerStatus
}

// A node's information about missing rumors
type PeerStatus struct {
	Identifier 	string
	NextID     	uint32
}

// Packet containing keywords for a search request
type SearchRequest struct {
	Origin   	string
	Budget   	uint64
	Keywords 	[]string
}

// Packet containing all the results of a search request
type SearchReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	Results     []*SearchResult
}

// A match for a given search request
type SearchResult struct {
	FileName     string
	MetafileHash []byte
	ChunkMap     []uint64
	ChunkCount   uint64
}

// A message announcing that a new transaction should be processed
type TxPublish struct {
	File      File
    User      User
	HopLimit  uint32
    ID        uint32
    Origin    string
}

// A message announcing that a new block was found
type BlockPublish struct {
	Block    Block
	HopLimit uint32
    ID       uint32
    Origin   string
}

// A file information
type File struct {
	Name          string
	Size          int64
	MetafileHash  []byte
}

// A registered user
type User struct {
	Name string
	PublicKey []byte
}

// A block on the blockchain
type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []TxPublish
}

type Signature struct {
	Origin       string // name of the sender
	Signature    []byte
}

type CypheredMessage struct {
    Destination string // name of the destination
    Payload     []byte // cyphered SignedMessage
    HopLimit    uint32
    IV          []byte // initialization vector for the AES cyphering
    Key         []byte // cyphered symmetric key
}

type OnionPacket struct {

	HopLimit	uint32
	Destination string
	Data 		[OnionHeaderSize + OnionPayloadSize]byte

}

type OnionSubHeader struct {
	PrevHop 	string				// 64B - Previous node in the route
	NextHop 	string				// 64B - Next node in the route
	// Signature	[]byte				// Signature of previous node
	Key			[]byte 				// 16B - Key for decryption
	IV			[]byte 				// 16B - Initialization vector for decryption
	Hash 		[]byte				// 32B - Hash of OnionMessage
}

// Aggregate of all other fields, should be used as top-level
// entity for external communication with other nodes.
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
    Signature     *Signature
    Cyphered      *CypheredMessage
	Onion		  *OnionPacket
}

//
//  CONSTRUCTORS
//

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

func NewCypheredMessage(destination string, payload []byte) *CypheredMessage {
    return &CypheredMessage{
        Destination: destination,
        Payload:     payload,
        HopLimit:    InitialHopLimit,
    }
}

// Copy the content of a search request but replace the budget.
func CopySearchRequest(request *SearchRequest, budget uint64) *SearchRequest {

	return &SearchRequest{
		Origin: request.Origin,
		Budget:	budget,
		Keywords: request.Keywords,
	}
}


//
//  PACKING METHODS
//

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

// Pack a TxPublish into a GossipPacket
func (publish *TxPublish) Packed() *GossipPacket {

	if publish == nil {
		panic("Cannot pack <nil> transaction into a GossipPacket")
	}

	return &GossipPacket{TxPublish: publish}
}

// Pack a BlockPublish into a GossipPacket
func (publish *BlockPublish) Packed() *GossipPacket {

	if publish == nil {
		panic("Cannot pack <nil> block into a GossipPacket")
	}

	return &GossipPacket{BlockPublish: publish}
}

// Pack a CypheredMessage into a GossipPacket
func (cyphered *CypheredMessage) Packed() *GossipPacket {

	if cyphered == nil {
		panic("Cannot pack <nil> cyphered message into a GossipPacket")
	}

	return &GossipPacket{Cyphered: cyphered}
}

// Pack an OnionPacket into a GossipPacket
func (onion *OnionPacket) Packed() *GossipPacket {

	if onion == nil {
		panic("Cannot pack <nil> onion packet into a GossipPacket")
	}

	return &GossipPacket{Onion: onion}
}

//
//  INTEGRITY CHECKS
//

// Checks if a given GossipPacket is valid. It is only valid if exactly one of its 10 fields is non-nil.
func (packet *GossipPacket) IsValid() bool {
	return boolCount(packet.Rumor != nil)+boolCount(packet.Simple != nil)+
		boolCount(packet.Status != nil)+boolCount(packet.Private != nil)+
		boolCount(packet.DataReply != nil)+boolCount(packet.DataRequest != nil)+
		boolCount(packet.SearchReply != nil)+boolCount(packet.SearchRequest != nil)+
		boolCount(packet.TxPublish != nil)+boolCount(packet.BlockPublish != nil)+
		+boolCount(packet.Cyphered != nil) + boolCount(packet.Onion != nil) == 1
}

// Safety check that we only broadcast packets which are supposed to be broadcast.
func (packet *GossipPacket) IsEligibleForBroadcast() bool {
	return !(packet.Simple == nil && packet.SearchRequest == nil && packet.TxPublish == nil && packet.BlockPublish == nil)
}

func (packet *GossipPacket) GetDestination() *string {
    switch {
    case packet.Private != nil:
        return &packet.Private.Destination
    case packet.DataRequest != nil:
        return &packet.DataRequest.Destination
    case packet.DataReply != nil:
        return &packet.DataReply.Destination
    case packet.SearchReply != nil:
        return &packet.SearchReply.Destination
    default:
        return nil
    }
}

func (packet *GossipPacket) GetOrigin() *string {
	switch {
	case packet.Simple != nil:
		return &packet.Simple.OriginalName
	case packet.Rumor != nil:
		return &packet.Rumor.Origin
	case packet.Private != nil:
		return &packet.Private.Origin
	case packet.DataRequest != nil:
		return &packet.DataRequest.Origin
	case packet.DataReply != nil:
		return &packet.DataReply.Origin
	case packet.SearchRequest != nil:
		return &packet.SearchRequest.Origin
	case packet.SearchReply != nil:
		return &packet.SearchReply.Origin
	case packet.BlockPublish != nil:
		return &packet.BlockPublish.Origin
	case packet.TxPublish != nil:
		return &packet.TxPublish.Origin
	default:
		return nil
	}
}

// Verify that a DataReply has the correct data via computing and comparing the hash
func (reply *DataReply) VerifyHash(expected []byte) bool {

	computedHash := sha256.Sum256(reply.Data)
	dataIsConsistent := bytes.Compare(computedHash[:], reply.HashValue) == 0
	hashIsExpected := bytes.Compare(reply.HashValue, expected) == 0

	return dataIsConsistent && hashIsExpected
}

// If a route has an empty Text message, we consider it to be a routing rumor.
func (rumor *RumorMessage) IsRouteRumor() bool {
	return rumor.Text == ""
}

//
//  HASHING FUNCTIONS
//

// Hash of a whole block
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

// Hash of a transaction
func (t *TxPublish) Hash() (out [32]byte) {
	h := sha256.New()
    if t.File.Name != "" {
        binary.Write(h, binary.LittleEndian, uint32(len(t.File.Name)))
        h.Write([]byte(t.File.Name))
        h.Write(t.File.MetafileHash)
    } else {
        binary.Write(h, binary.LittleEndian, uint32(len(t.User.Name)))
        h.Write([]byte(t.User.Name))
        h.Write(t.User.PublicKey)
    }
    copy(out[:], h.Sum(nil))
    return
}

// Hash of an Onion
func (onion *OnionPacket) Hash() (out []byte) {
	hash := sha256.Sum256(onion.Data[OnionHeaderSize:])
	return hash[:]
}

//
//  STRING FUNCTIONS
//

func (block *Block) Str() string {

    hash := block.Hash()
    prev := block.PrevHash
    files := make([]string, 0)
	users := make([]string, 0)

    for _, transaction := range block.Transactions {
		if transaction.File.Name != "" {
			files = append(files, transaction.File.Name)
		}
    }

	for _, transaction := range block.Transactions {
		if transaction.User.Name != "" {
			users = append(users, transaction.User.Name)
		}
	}

	return fmt.Sprintf("%v:%v:%v:%v", hex.EncodeToString(hash[:]), hex.EncodeToString(prev[:]),
		strings.Join(files, FileNameSeparator), strings.Join(users, FileNameSeparator))
}


//
//  GETTERS
//

func (r *SimpleMessage) GetOrigin() string {
	return r.OriginalName
}

func (r *RumorMessage) GetOrigin() string {
    return r.Origin
}

func (r *RumorMessage) GetID() uint32 {
    return r.ID
}

func (r *TxPublish) GetOrigin() string {
    return r.Origin
}

func (r *TxPublish) GetID() uint32 {
    return r.ID
}

func (r *BlockPublish) GetOrigin() string {
    return r.Origin
}

func (r *BlockPublish) GetID() uint32 {
    return r.ID
}

//
//  MORE HASH
//

func (s *SimpleMessage) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(s.OriginalName))
	h.Write([]byte(s.Contents))
	copy(out[:], h.Sum(nil))
	return
}

func (r *RumorMessage) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.Origin))
	h.Write([]byte(r.Text))
	binary.Write(h, binary.LittleEndian, r.ID)
	copy(out[:], h.Sum(nil))
	return
}

func (s *PeerStatus) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(s.Identifier))
	binary.Write(h, binary.LittleEndian, s.NextID)
	copy(out[:], h.Sum(nil))
	return
}

func (s *StatusPacket) Hash() (out [32]byte) {
	h := sha256.New()
	binary.Write(h,binary.LittleEndian,
		uint32(len(s.Want)))
	for _, w := range s.Want {
		wh := w.Hash()
		h.Write(wh[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

func (p *PrivateMessage) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(p.Origin))
	h.Write([]byte(p.Destination))
	h.Write([]byte(p.Text))
	binary.Write(h, binary.LittleEndian, p.ID)
	copy(out[:], h.Sum(nil))
	return
}

func (r *DataRequest) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.Origin))
	h.Write([]byte(r.Destination))
	h.Write(r.HashValue)
	copy(out[:], h.Sum(nil))
	return
}

func (r *DataReply) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.Origin))
	h.Write([]byte(r.Destination))
	h.Write(r.HashValue)
	copy(out[:], h.Sum(nil))
	return
}

func (r *SearchRequest) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.Origin))
	binary.Write(h, binary.LittleEndian, r.Budget)
	binary.Write(h,binary.LittleEndian, uint32(len(r.Keywords)))
	for _, k := range r.Keywords {
		h.Write([]byte(k))
	}
	copy(out[:], h.Sum(nil))
	return
}

func (r *SearchResult) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.FileName))
	h.Write(r.MetafileHash)
	binary.Write(h,binary.LittleEndian, r.ChunkCount)
	binary.Write(h,binary.LittleEndian, uint32(len(r.ChunkMap)))
	for _, c := range r.ChunkMap {
		binary.Write(h,binary.LittleEndian, c)
	}
	copy(out[:], h.Sum(nil))
	return
}

func (r *SearchReply) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.Origin))
	h.Write([]byte(r.Destination))
	binary.Write(h,binary.LittleEndian, uint32(len(r.Results)))
	for _, sr := range r.Results {
		srh := sr.Hash()
		h.Write(srh[:])
	}
	copy(out[:], h.Sum(nil))
	return
}

func (r *BlockPublish) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write([]byte(r.Origin))
	binary.Write(h,binary.LittleEndian, r.ID)
	bh := r.Block.Hash()
	h.Write(bh[:])
	copy(out[:], h.Sum(nil))
	return
}

func (packet *GossipPacket) Hash() (out [32]byte) {

	switch {
	case packet.Simple != nil:
		return packet.Simple.Hash()
	case packet.Rumor != nil:
		return packet.Rumor.Hash()
	case packet.Status != nil:
		return packet.Status.Hash()
	case packet.Private != nil:
		return packet.Private.Hash()
	case packet.DataReply != nil:
		return packet.DataReply.Hash()
	case packet.DataRequest != nil:
		return packet.DataRequest.Hash()
	case packet.SearchRequest != nil:
		return packet.SearchRequest.Hash()
	case packet.SearchReply != nil:
		return packet.SearchReply.Hash()
	case packet.TxPublish != nil:
		return packet.TxPublish.Hash()
	case packet.BlockPublish != nil:
		return packet.BlockPublish.Hash()
	default:
		panic("Cannot hash")
	}
}

func (packet *GossipPacket) ShouldBeSigned() bool {
	return packet.TxPublish == nil && packet.BlockPublish == nil && packet.Cyphered == nil && packet.Onion == nil && packet.Signature == nil
}

func (packet *GossipPacket) ShouldBeCiphered() bool {
	return packet.GetDestination() != nil && packet.Cyphered == nil
}
