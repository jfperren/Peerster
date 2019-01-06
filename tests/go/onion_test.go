package tests

import (
    "bytes"
    "crypto/rand"
    "crypto/rsa"
    "github.com/jfperren/Peerster/common"
    "github.com/jfperren/Peerster/gossiper"
    "testing"
)

var (

    Alice = gossiper.NewGossiper(
        "127.0.0.1:9090",
        "",
        "Alice",
        "127.0.0.1:9091",
        false,
        0,
        true,
        common.CryptoKeySize,
        common.CypherIfPossible,
        1)

    Bob = gossiper.NewGossiper(
        "127.0.0.1:9091",
        "",
        "Bob",
        "127.0.0.1:9090",
        false,
        0,
        true,
        common.CryptoKeySize,
        common.CypherIfPossible,
        2)

    Charlie = gossiper.NewGossiper(
        "127.0.0.1:9092",
        "",
        "Charlie",
        "127.0.0.1:9090",
        false,
        0,
        true,
        common.CryptoKeySize,
        common.CypherIfPossible,
        1)

    Delta = gossiper.NewGossiper(
        "127.0.0.1:9093",
        "",
        "Delta",
        "127.0.0.1:9090",
        false,
        0,
        true,
        common.CryptoKeySize,
        common.CypherIfPossible,
        2)

    Keys = map[string]*rsa.PublicKey{
        Alice.Name:     &Alice.Crypto.PrivateKey.PublicKey,
        Bob.Name:       &Bob.Crypto.PrivateKey.PublicKey,
        Charlie.Name:   &Charlie.Crypto.PrivateKey.PublicKey,
        Delta.Name:     &Delta.Crypto.PrivateKey.PublicKey,
    }
)

func TestSubHeaderLeftRotation(t *testing.T) {

    onionData := [common.OnionSize]byte{}
    rand.Read(onionData[:])
    onion := common.OnionPacket{0, "", onionData}

    // Copy onion to keep a separate version
    onionDataCopy := [common.OnionSize]byte{}
    copy(onionDataCopy[:], onionData[:])

    gossiper.RotateSubHeadersLeft(&onion)

    for i := 0; i < common.OnionSubHeaderCount - 1; i++ {

        j := i * common.OnionSubHeaderSize
        k := (i + 1) * common.OnionSubHeaderSize
        l := (i + 2) * common.OnionSubHeaderSize

        if !bytes.Equal(onion.Data[j:k], onionDataCopy[k:l]) {
            t.Errorf("Block %v is not rotated correctly", i)
        }
    }

    if !bytes.Equal(onion.Data[common.OnionHeaderSize:], onionDataCopy[common.OnionHeaderSize:]) {
        t.Errorf("Left Rotation should not modify payload")
    }
}

func TestSubHeaderRightRotation(t *testing.T) {

    onionData := [common.OnionSize]byte{}
    rand.Read(onionData[:])
    onion := common.OnionPacket{0, "", onionData}

    // Copy onion to keep a separate version
    onionDataCopy := [common.OnionSize]byte{}
    copy(onionDataCopy[:], onionData[:])

    gossiper.RotateSubHeadersRight(&onion)

    for i := 0; i < common.OnionSubHeaderCount - 1; i++ {

        j := i * common.OnionSubHeaderSize
        k := (i + 1) * common.OnionSubHeaderSize
        l := (i + 2) * common.OnionSubHeaderSize

        if !bytes.Equal(onionDataCopy[j:k], onion.Data[k:l]) {
            t.Errorf("Block %v is not rotated correctly", i+1)
        }
    }

    if !bytes.Equal(make([]byte, common.OnionSubHeaderSize), onion.Data[0:common.OnionSubHeaderSize]) {
        t.Errorf("Block 0 is not rotated correctly")
    }

    if !bytes.Equal(onion.Data[common.OnionHeaderSize:], onionDataCopy[common.OnionHeaderSize:]) {
        t.Errorf("Right Rotation should not modify payload")
    }
}

func TestOnionOneLayer(t *testing.T) {

    // Create dummy message
    message := common.NewSimpleMessage("origin", "127.0.0.1:8080","contents").Packed()

    // Wrap in onion
    onion, err := Alice.Crypto.GenerateOnion(message, []string{Bob.Name}, Keys, Alice.Name)

    if err != nil {
        t.Errorf("Error creating onion: %v", err)
    }

    if len(onion.Data) != common.OnionHeaderSize + common.OnionPayloadSize {
        t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
            common.OnionHeaderSize + common.OnionPayloadSize)
    }

    // Unwrap
    gossipPacket, subHeader, err := Bob.ProcessOnion(onion)

    if err != nil {
        t.Errorf("Error decoding onion: %v", err)
    }

    if len(onion.Data) != common.OnionHeaderSize + common.OnionPayloadSize {
        t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
            common.OnionHeaderSize + common.OnionPayloadSize)
    }

    if gossipPacket == nil {
        t.Errorf("Could not recover gossipPacket from Onion")
    }

    if gossipPacket.Simple == nil {
        t.Errorf("Could not recover simple message from Onion")
    }

    if gossipPacket.Simple.Contents != message.Simple.Contents ||
        gossipPacket.Simple.RelayPeerAddr != message.Simple.RelayPeerAddr ||
        gossipPacket.Simple.OriginalName != message.Simple.OriginalName {
        t.Errorf("Got a different message. Original is %v, received %v", message.Simple, gossipPacket.Simple)
    }

    if subHeader.NextHop != common.NoNextHop {
        t.Errorf("Onion should have NextHop equal to %v, instead has %v", common.NoNextHop, subHeader.NextHop)
    }

    if subHeader.PrevHop != Alice.Name {
        t.Errorf("Onion should have PrevHop equal to %v, instead has %v", Alice.Name, subHeader.PrevHop)
    }
}

func TestOnionTwoLayers(t *testing.T) {

    // Create dummy message
    message := common.NewSimpleMessage("origin", "127.0.0.1:8080","contents").Packed()

    // Wrap in onion
    onion, err := Alice.Crypto.GenerateOnion(message, []string{Bob.Name, Charlie.Name}, Keys, Alice.Name)

    if err != nil {
        t.Errorf("Error creating onion: %v", err)
    }

    if len(onion.Data) != common.OnionHeaderSize + common.OnionPayloadSize {
        t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
            common.OnionHeaderSize + common.OnionPayloadSize)
    }

    // Unwrap first layer
    gossipPacket, subHeader, err := Bob.ProcessOnion(onion)

    if err != nil {
        t.Errorf("Error decoding onion: %v", err)
    }

    if len(onion.Data) != common.OnionHeaderSize + common.OnionPayloadSize {
        t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
            common.OnionHeaderSize + common.OnionPayloadSize)
    }

    if gossipPacket != nil {
        t.Errorf("Bob should not be able to fully decode the Onion")
    }

    if err != nil {
        t.Errorf("Error decoding subHeader: %v", err)
    }

    if subHeader.NextHop != Charlie.Name {
        t.Errorf("Onion should have NextHop equal to %v, instead has %v", Charlie.Name, subHeader.NextHop)
    }

    if subHeader.PrevHop != Alice.Name {
        t.Errorf("Onion should have PrevHop equal to %v, instead has %v", Alice.Name, subHeader.PrevHop)
    }

    // Unwrap second layer
    gossipPacket, subHeader, err = Charlie.ProcessOnion(onion)

    if err != nil {
        t.Errorf("Error decoding onion: %v", err)
    }

    if len(onion.Data) != common.OnionHeaderSize + common.OnionPayloadSize {
        t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
            common.OnionHeaderSize + common.OnionPayloadSize)
    }

    if gossipPacket == nil {
        t.Errorf("Could not recover gossipPacket from Onion")
    }

    if gossipPacket.Simple == nil {
        t.Errorf("Could not recover simple message from Onion")
    }

    if gossipPacket.Simple.Contents != message.Simple.Contents ||
        gossipPacket.Simple.RelayPeerAddr != message.Simple.RelayPeerAddr ||
        gossipPacket.Simple.OriginalName != message.Simple.OriginalName {
        t.Errorf("Got a different message. Original is %v, received %v", message.Simple, gossipPacket.Simple)
    }

    if subHeader.NextHop != common.NoNextHop {
        t.Errorf("Onion should have NextHop equal to %v, instead has %v", common.NoNextHop, subHeader.NextHop)
    }

    if subHeader.PrevHop != Bob.Name {
        t.Errorf("Onion should have PrevHop equal to %v, instead has %v", Bob.Name, subHeader.PrevHop)
    }
}

func TestOnionEightLayers(t *testing.T) {

    // Create dummy message
    message := common.NewSimpleMessage("origin", "127.0.0.1:8080","contents").Packed()

    // Create route
    nodes := []*gossiper.Gossiper{Bob, Delta, Charlie, Alice, Charlie, Bob, Delta, Charlie}
    route := []string{Bob.Name, Delta.Name, Charlie.Name, Alice.Name, Charlie.Name, Bob.Name, Delta.Name, Charlie.Name}

    // Wrap in onion
    onion, err := Alice.Crypto.GenerateOnion(message, route, Keys, Alice.Name)

    if err != nil {
        t.Errorf("Error creating onion: %v", err)
    }

    if len(onion.Data) != common.OnionHeaderSize + common.OnionPayloadSize {
        t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
            common.OnionHeaderSize + common.OnionPayloadSize)
    }

    for i, node := range nodes {

        // Unwrap one layer
        gossipPacket, subHeader, err := node.ProcessOnion(onion)

        if err != nil {
            t.Errorf("Error decoding onion: %v", err)
        }

        if len(onion.Data) != common.OnionHeaderSize+common.OnionPayloadSize {
            t.Errorf("Onion has wrong size, was expecting %v bytes, got %v", len(onion.Data),
                common.OnionHeaderSize+common.OnionPayloadSize)
        }

        if i == 0 {

            if gossipPacket != nil {
                t.Errorf("The first node should not be able to fully decode the Onion")
            }

            if subHeader.NextHop != route[i+1] {
                t.Errorf("Onion should have NextHop equal to %v, instead has %v", Charlie.Name, subHeader.NextHop)
            }

            if subHeader.PrevHop != Alice.Name {
                t.Errorf("Onion should have PrevHop equal to %v, instead has %v", Alice.Name, subHeader.PrevHop)
            }

        } else if i == common.OnionSubHeaderCount-1 {

            if gossipPacket == nil {
                t.Errorf("The last node should be able to fully decode the Onion")
            }

            if gossipPacket.Simple == nil {
                t.Errorf("Could not recover simple message from Onion")
            }

            if gossipPacket.Simple.Contents != message.Simple.Contents ||
                gossipPacket.Simple.RelayPeerAddr != message.Simple.RelayPeerAddr ||
                gossipPacket.Simple.OriginalName != message.Simple.OriginalName {
                t.Errorf("Got a different message. Original is %v, received %v", message.Simple, gossipPacket.Simple)
            }

            if subHeader.NextHop != common.NoNextHop {
                t.Errorf("Onion should have NextHop equal to %v, instead has %v", common.NoNextHop, subHeader.NextHop)
            }

            if subHeader.PrevHop != route[i-1] {
                t.Errorf("Onion should have PrevHop equal to %v, instead has %v", Alice.Name, subHeader.PrevHop)
            }

        } else {
            if gossipPacket != nil {
                t.Errorf("The first node should not be able to fully decode the Onion")
            }

            if subHeader.NextHop != route[i+1] {
                t.Errorf("Onion should have NextHop equal to %v, instead has %v", Charlie.Name, subHeader.NextHop)
            }

            if subHeader.PrevHop != route[i-1] {
                t.Errorf("Onion should have PrevHop equal to %v, instead has %v", Alice.Name, subHeader.PrevHop)
            }
        }
    }
}