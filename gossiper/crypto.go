package gossiper

import (
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "fmt"
)

type Crypto struct {
    PrivateKey *rsa.PrivateKey
}

func NewCrypto() *Crypto {
    c := Crypto{}
    c.GenerateKey()
    return &c
}

func (c *Crypto) GenerateKey() {
    privateKey, err := rsa.GenerateKey(rand.Reader, 18000)
    if err != nil {
        panic(err)
    }
    c.PrivateKey = privateKey
}

func (c *Crypto) PublicKey() crypto.PublicKey {
    return c.PrivateKey.Public()
}

func (c *Crypto) Decypher(payload []byte) []byte {
    decyphered, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.PrivateKey, payload, []byte{})
    if err != nil {
        fmt.Println("Error from decryption: %s\n", err)
        return []byte{}
    }
    return decyphered
}

func (c *Crypto) Cypher(payload []byte, publicKey rsa.PublicKey) []byte {
    cyphered, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &publicKey, payload, []byte{})
    if err != nil {
        fmt.Println("Error from encryption: %s\n", err)
        return []byte{}
    }
    return cyphered
}

func (c *Crypto) Sign(payload []byte) []byte {
    hashed := sha256.Sum256(payload)
    signature, err := rsa.SignPSS(rand.Reader, c.PrivateKey, crypto.SHA256, hashed[:], nil)
    if err != nil {
        fmt.Println("Error from signing: %s\n", err)
        return []byte{}
    }
    return signature
}

func (c *Crypto) Verify(payload, signature []byte, publicKey rsa.PublicKey) bool {
    hashed := sha256.Sum256(payload)
    err := rsa.VerifyPSS(&publicKey, crypto.SHA256, hashed[:], signature, nil)
    if err != nil {
        fmt.Println("Error from verification: %s\n", err)
        return false
    }
    return true
}
