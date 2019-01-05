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