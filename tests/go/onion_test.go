package tests

import (
    "crypto/rsa"
    "github.com/jfperren/Peerster/common"
    "github.com/jfperren/Peerster/gossiper"
    "testing"
)

var (

    Alice = gossiper.NewGossiper(
        "127.0.0.1:8080",
        "",
        "Alice",
        "127.0.0.1:8081",
        false,
        0,
        true,
        common.CryptoKeySize,
        common.CypherIfPossible,
        true)

    Bob = gossiper.NewGossiper(
        "127.0.0.1:8081",
        "",
        "Alice",
        "127.0.0.1:8080",
        false,
        0,
        true,
        common.CryptoKeySize,
        common.CypherIfPossible,
        true)

    Keys = map[string]*rsa.PublicKey{
        Alice.Name: &Alice.Crypto.PrivateKey.PublicKey,
        Bob.Name:   &Bob.Crypto.PrivateKey.PublicKey,
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
    gossipPacket, err := Bob.ProcessOnion(onion)

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
}