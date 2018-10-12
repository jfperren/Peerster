package main

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

func (simple *SimpleMessage) packed() *GossipPacket {

    if simple == nil {
        panic("Cannot pack <nil> simple into a GossipPacket")
    }

    return &GossipPacket{simple,nil,nil}
}

func (rumor *RumorMessage) packed() *GossipPacket {

    if rumor == nil {
        panic("Cannot pack <nil> rumor into a GossipPacket")
    }

    return &GossipPacket{nil,rumor,nil}
}

func (status *StatusPacket) packed() *GossipPacket {

    if status == nil {
        panic("Cannot pack <nil> status into a GossipPacket")
    }

    return &GossipPacket{nil,nil,status}
}

func (packet *GossipPacket) isValid() bool {
    return (packet.Simple == nil && packet.Rumor == nil && packet.Status != nil) || (packet.Simple == nil && packet.Rumor != nil && packet.Status == nil) || (packet.Simple != nil && packet.Rumor == nil && packet.Status == nil)
}