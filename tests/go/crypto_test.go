package tests

import (
    "testing"
    "github.com/jfperren/Peerster/gossiper"
)

func TestSignatures(t *testing.T) {
    crypto := gossiper.NewCrypto()

    payload := []byte("hello")

    signature := crypto.Sign(payload)

    if !crypto.Verify(payload, signature, crypto.PublicKey()) {
        t.Errorf("Wrong signature")
    }
}

func TestCypher(t *testing.T) {
    crypto := gossiper.NewCrypto()

    payload := []byte("hello")

    cyphered := crypto.Cypher(payload, crypto.PublicKey())

    decyphered := crypto.Decypher(cyphered)

    if decyphered != payload {
        t.Errorf("Incorrect decyphering")
    }
}
