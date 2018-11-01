package common

// --
// -- DATA STRUCTURES
// --

type SimpleMessage struct {
    OriginalName  string
    RelayPeerAddr string
    Contents      string
}

type RumorMessage struct {
    Origin        string
    ID            uint32
    Text          string
}

type PrivateMessage struct {
    Origin        string
    ID            uint32
    Text          string
    Destination   string
    HopLimit      uint32
}

type PeerStatus struct {
    Identifier    string
    NextID        uint32
}

type StatusPacket struct {
    Want          []PeerStatus
}

type GossipPacket struct {
    Simple        *SimpleMessage
    Rumor         *RumorMessage
    Status        *StatusPacket
    Private       *PrivateMessage
}

// --
// -- INITIALIZERS
// --

func NewPrivateMessage(origin, destination, contents string) *PrivateMessage {

    return &PrivateMessage{
        Origin: origin,
        ID: 0,
        Text: contents,
        Destination: destination,
        HopLimit: InitialHopLimit,
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

    return &GossipPacket{simple,nil,nil, nil}
}

// Pack a RumorMessage into a GossipPacket
func (rumor *RumorMessage) Packed() *GossipPacket {

    if rumor == nil {
        panic("Cannot pack <nil> rumor into a GossipPacket")
    }

    return &GossipPacket{nil,rumor,nil, nil}
}

// Pack a StatusPacket into a GossipPacket
func (status *StatusPacket) Packed() *GossipPacket {

    if status == nil {
        panic("Cannot pack <nil> status into a GossipPacket")
    }

    return &GossipPacket{nil,nil,status, nil}
}

// Pack a PrivateMessage into a GossipPacket
func (private *PrivateMessage) Packed() *GossipPacket {

    if private == nil {
        panic("Cannot pack <nil> status into a GossipPacket")
    }

    return &GossipPacket{nil,nil,nil, private}
}

// Checks if a given GossipPacket is valid. It is only valid if exactly one of its 4 fields is non-nil.
func (packet *GossipPacket) IsValid() bool {
    return boolCount(packet.Rumor != nil) + boolCount(packet.Simple != nil) +
        boolCount(packet.Status != nil) + boolCount(packet.Private != nil) == 1
}

func boolCount(b bool) int {
    if b { return 1 } else { return 0 }
}

func (rumor *RumorMessage) IsRouteRumor() bool {
    return rumor.Text == ""
}