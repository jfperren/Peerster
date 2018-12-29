package tests

import (
    "bytes"
    "testing"
    "github.com/jfperren/Peerster/gossiper"
)

func TestSignatures(t *testing.T) {
    crypto := gossiper.NewCrypto(4096)

    payload := []byte("hello")

    signature := crypto.Sign(payload)
    publicKey := crypto.PublicKey()

    if !crypto.Verify(payload, signature, publicKey) {
        t.Errorf("Wrong signature")
    }
}

func TestCypher(t *testing.T) {
    crypto := gossiper.NewCrypto(4096)

    payload := []byte("hello")
    publicKey := crypto.PublicKey()

    cyphered := crypto.Cypher(payload, publicKey)

    decyphered := crypto.Decypher(cyphered)

    if !bytes.Equal(decyphered[:], payload[:]) {
        t.Errorf("Incorrect decyphering")
    }
}
