package main


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
}

// --
// -- CONVENIENCE METHODS
// --

// Pack a SimpleMessage into a GossipPacket
func (simple *SimpleMessage) packed() *GossipPacket {

    if simple == nil {
        panic("Cannot pack <nil> Simple into a GossipPacket")
    }

    return &GossipPacket{simple,nil,nil}
}

// Pack a RumorMessage into a GossipPacket
func (rumor *RumorMessage) packed() *GossipPacket {

    if rumor == nil {
        panic("Cannot pack <nil> rumor into a GossipPacket")
    }

    return &GossipPacket{nil,rumor,nil}
}

// Pack a StatusPacket into a GossipPacket
func (status *StatusPacket) packed() *GossipPacket {

    if status == nil {
        panic("Cannot pack <nil> status into a GossipPacket")
    }

    return &GossipPacket{nil,nil,status}
}

// Checks if a given GossipPacket is valid. It is only valid if exactly one of its 3 fields is non-nil.
func (packet *GossipPacket) isValid() bool {
    return (packet.Simple == nil && packet.Rumor == nil && packet.Status != nil) || (packet.Simple == nil && packet.Rumor != nil && packet.Status == nil) || (packet.Simple != nil && packet.Rumor == nil && packet.Status == nil)
}