package gossiper

import (
    "crypto"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "errors"
    "fmt"
    "github.com/jfperren/Peerster/common"
)

// Errors thrown by the Onion module
var (

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrPayloadIsNotMultipleOfBlockLength = errors.New("payload is not a multiple of block length")

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrCipherIsNotMultipleOfBlockLength = errors.New("cipher is not a multiple of block length")

)

type Crypto struct {
    PrivateKey *rsa.PrivateKey
    Options int
}

func NewCrypto(size, options int) *Crypto {
    c := Crypto{
        Options: options,
    }
    c.GenerateKey(size)
    return &c
}

func (c *Crypto) GenerateKey(size int) {
    privateKey, err := rsa.GenerateKey(rand.Reader, size)
    if err != nil {
        panic(err)
    }
    c.PrivateKey = privateKey
}

func (c *Crypto) PublicKey() rsa.PublicKey {
    publicKey := c.PrivateKey.Public().(*rsa.PublicKey)
    return *publicKey
}

func (c *Crypto) Decypher(payload []byte) []byte {
    decyphered, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.PrivateKey, payload, []byte(""))
    if err != nil {
        fmt.Printf("Error from decryption: %s\n", err)
        return []byte{}
    }
    return decyphered
}

func (c *Crypto) Cypher(payload []byte, publicKey rsa.PublicKey) []byte {
    cyphered, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &publicKey, payload, []byte(""))
    if err != nil {
        fmt.Printf("Error from encryption: %s\n", err)
        return []byte{}
    }
    return cyphered
}

func (c *Crypto) Sign(payload []byte) []byte {
    hashed := sha256.Sum256(payload)
    signature, err := rsa.SignPSS(rand.Reader, c.PrivateKey, crypto.SHA256, hashed[:], nil)
    if err != nil {
        fmt.Printf("Error from signing: %s\n", err)
        return []byte{}
    }
    return signature
}

func (c *Crypto) Verify(payload, signature []byte, publicKey rsa.PublicKey) bool {
    hashed := sha256.Sum256(payload)
    err := rsa.VerifyPSS(&publicKey, crypto.SHA256, hashed[:], signature, nil)
    if err != nil {
        fmt.Printf("Error from verification: %s\n", err)
        return false
    }
    return true
}

//
//  SYMMETRIC BLOCK CIPHER
//

func CBCCipher(payload []byte, key []byte) ([]byte, []byte, error) {

    if len(payload) % aes.BlockSize != 0 {
        return nil, nil, ErrPayloadIsNotMultipleOfBlockLength
    }

    // Create block cipher based on key
    block, err := aes.NewCipher(key)
    if err != nil { return nil, nil, err }

    // Create array for encrypted value
    encrypted:= make([]byte, len(payload))

    // Create random initialization vector
    iv := make([]byte, aes.BlockSize)

    // Create random iv at the beginning of the cipher text
    rand.Read(iv)

    // Encrypt payload
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(encrypted, payload)

    return encrypted, iv, nil
}

func CBCDecipher(encrypted, key, iv []byte) ([]byte, error) {

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    // CBC mode always works in whole blocks.
    if len(encrypted) % aes.BlockSize != 0 {
        return nil, ErrCipherIsNotMultipleOfBlockLength
    }

    mode := cipher.NewCBCDecrypter(block, iv)

    // Create buffer for plain text & decrypt
    plain := make([]byte, len(encrypted))
    mode.CryptBlocks(plain, encrypted)

    return plain, nil
}

func NewCBCSecret() []byte {
    key := make([]byte, common.CBCKeySize)
    rand.Read(key)
    return key
}