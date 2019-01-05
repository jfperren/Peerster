package tests

import (
    "bytes"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "fmt"
    "github.com/jfperren/Peerster/common"
    "github.com/jfperren/Peerster/gossiper"
    "log"
    "testing"
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

func TestBCBCipher(t *testing.T) {

    payload := make([]byte, 256)
    rand.Read(payload)

    key := gossiper.NewCBCSecret()

    if len(key) != common.CBCKeySize {
        t.Errorf("Key size is expected to be %v bytes, but length is %v.", common.CBCKeySize, len(key))
    }

    cipher, iv, err := gossiper.CBCCipher(payload, key)

    if err != nil {
        t.Errorf("Error ciphering payload: %v", err)
    }

    if len(payload) != len(cipher) {
        t.Errorf("Payload and Cipher should have same size. Here, payload has % bytes and cipher has %v", len(payload), len(cipher))
    }

    plain, err := gossiper.CBCDecipher(cipher, key, iv)

    if err != nil {
        t.Errorf("Error deciphering payload: %v", err)
    }

    if !bytes.Equal(plain, payload) {
        t.Errorf("Initial message and deciphered one don't match")
    }
}
