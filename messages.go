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

func (simple *SimpleMessage) packaged() *GossipPacket {
    return &GossipPacket{simple,nil,nil}
}

func (rumor *RumorMessage) packaged() *GossipPacket {
    return &GossipPacket{nil,rumor,nil}
}

func (status *StatusPacket) packed() *GossipPacket {
    return &GossipPacket{nil,nil,status}
}
