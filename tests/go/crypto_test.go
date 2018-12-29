package tests

import (
    "bytes"
    "testing"
    "github.com/jfperren/Peerster/gossiper"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "fmt"
    "log"
)

func TestSignatures(t *testing.T) {
    crypto := gossiper.NewCrypto(4096, 2)

    payload := []byte("hello")

    signature := crypto.Sign(payload)
    publicKey := crypto.PublicKey()

    if !crypto.Verify(payload, signature, publicKey) {
        t.Errorf("Wrong signature")
    }
}

func TestCypher(t *testing.T) {
    crypto := gossiper.NewCrypto(4096, 2)

    payload := []byte("hello")
    publicKey := crypto.PublicKey()

    cyphered := crypto.Cypher(payload, publicKey)

    decyphered := crypto.Decypher(cyphered)

    if !bytes.Equal(decyphered[:], payload[:]) {
        t.Errorf("Incorrect decyphering")
    }
}

func TestKeySharing(t *testing.T) {
    priv, _ := rsa.GenerateKey(rand.Reader, 512) // skipped error checking for brevity
    pub := priv.PublicKey
    bytes := x509.MarshalPKCS1PublicKey(&pub)
    pub2, err := x509.ParsePKCS1PublicKey(bytes)
    if err != nil {
        panic(err)
    }
    if pub.N.Cmp(pub2.N) != 0 || pub.E != pub2.E {
        log.Fatal("Public Keys at source and destination not equal")
    }
    fmt.Printf("OK - %#v\n", pub2)
}
