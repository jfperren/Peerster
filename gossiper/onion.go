package gossiper

import (
    "bytes"
    "crypto/rand"
    "crypto/rsa"
    "errors"
    "github.com/jfperren/Peerster/common"
)


// Errors thrown by the Onion module
var (

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrTooManyNodesOnRoute = errors.New("route is too long for size of header")

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrOnionHopKeyNotFound = errors.New("could not find name corresponding to public key in NextHop")

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrOnionHashDoNotMatch = errors.New("found hash that does not match computed hash")

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrOnionCouldNotDecipherSubHeader = errors.New("could not decipher subHeader")

)

//
//  WRAP / UNWRAP
//

// Wrap a regular GossipPacket into an onion that can be send on the mix network
func (crypto *Crypto) GenerateOnion(
    gossipPacket *common.GossipPacket,
    route []string,
    keys map[string]*rsa.PublicKey,
    name string,
) (*common.OnionPacket, error) {

    if len(route) > common.OnionSubHeaderCount {
        return nil, ErrTooManyNodesOnRoute
    }

    // Encrypt payload
    payload, err := Encode(gossipPacket, common.OnionPayloadSize)
    if err != nil { return nil, nil }

    // Create onion with random sub-headers
    data := [common.OnionHeaderSize + common.OnionPayloadSize]byte{}
    rand.Read(data[:common.OnionHeaderSize])
    copy(data[common.OnionHeaderSize:], payload)

    // Create onion with this data
    onion := &common.OnionPacket{Data: data}

    for i := len(route) - 1; i >= 0; i-- {

        node := route[i]

        key, found := keys[node]

        if !found {
            return nil, ErrOnionHopKeyNotFound
        }

        var next string
        var prev string

        if i == 0 {
            // Use our name as previous for the first node
            prev = name
        } else {
            prev = route[i-1]
        }

        if i == len(route) - 1 {
            // To signal that a node is the last one, we use their key as nextHop
            next = route[i]
        } else {
            next = route[i+1]
        }

        rotateSubHeadersRight(onion)
        crypto.wrap(onion, *key, prev, next)
    }

    return onion, nil
}

func (gossiper *Gossiper) ProcessOnion(onion *common.OnionPacket) (*common.GossipPacket, error) {

    // First, unwrap the onion and get subHeader
    subHeader, err := gossiper.Crypto.unwrap(onion)
    if err != nil { return nil, err }

    // If we are last, we should be able to get the data
    if isLast(subHeader) {

        // Try to decode payload into a gossip packet
        var gossipPacket common.GossipPacket
        err = Decode(onion.Data[common.OnionHeaderSize:], &gossipPacket)
        if err != nil { return nil, err }

        return &gossipPacket, nil

    } else {

        // Rotate subHeaders for the next node
        // rotateSubHeadersLeft(onion)

        // Reset hop limit
        onion.HopLimit = common.InitialHopLimit

        return nil, nil
    }
}

//
//  WRAP / UNWRAP
//

func (crypto *Crypto) wrap(onion *common.OnionPacket, key rsa.PublicKey, prev, next string) error {

    symmetricKey := crypto.NewCBCSecret()

    // Encrypt symmetric part
    otherData := onion.Data[common.OnionSubHeaderSize:]
    otherDataCipher, iv, err := crypto.CBCCipher(otherData, symmetricKey)
    if err != nil { return err }

    // Create subHeader with information
    subHeader := &common.OnionSubHeader{
        PrevHop: prev,
        NextHop: next,
        Key: symmetricKey,
        IV: iv,
        Hash: onion.Hash(),
    }

    // Encode subHeader
    subHeaderData, _ := Encode(subHeader, common.OnionSubHeaderPaddingSize)

    // Encrypt subHeader with asymmetric key
    cipherSubHeader := crypto.Cypher(subHeaderData, key)

    // Copy subHeader into onion
    copy(onion.Data[common.OnionSubHeaderSize:], otherDataCipher)
    copy(onion.Data[:common.OnionSubHeaderSize], cipherSubHeader)

    return nil
}

// Unwrap one layer of the onion using the node's public key
func (crypto *Crypto) unwrap(onion *common.OnionPacket) (*common.OnionSubHeader, error) {

    // First, extract the subHeader bytes
    subHeaderCipher := onion.Data[:common.OnionSubHeaderSize]

    // Decipher subHeader with private key
    subHeaderData := crypto.Decypher(subHeaderCipher)

    if len(subHeaderData) == 0 {
        return nil, ErrOnionCouldNotDecipherSubHeader
    }

    // Extract into subHeader structure
    var subHeader common.OnionSubHeader
    err := Decode(subHeaderData, &subHeader)
    if err != nil { return nil, err }

    // Extract rest of the data
    otherDataCipher := onion.Data[common.OnionSubHeaderSize:]
    otherData, err := crypto.CBCDecipher(otherDataCipher, subHeader.Key[:], subHeader.IV[:])

    // Copy it back into the onion
    copy(onion.Data[:common.OnionSubHeaderSize], subHeaderData)
    copy(onion.Data[common.OnionSubHeaderSize:], otherData)

    // Check integrity of onion content
    if !isValid(onion, &subHeader) {
        return nil, ErrOnionHashDoNotMatch
    }

    return &subHeader, nil
}

//
//  HEADER FUNCTIONS
//

// Rotate all subheaders by one chunk to the left and fill the last chunk with random bits
func rotateSubHeadersLeft(onion *common.OnionPacket) {

    for i := 0; i < common.OnionSubHeaderCount; i++ {

        j := i * common.OnionSubHeaderSize
        k := i + 1 * common.OnionSubHeaderSize
        l := (i + 2) * common.OnionSubHeaderSize

        if i == common.OnionSubHeaderCount - 1 {
            var padding [common.OnionSubHeaderSize]byte
            rand.Read(padding[:])
            copy(onion.Data[j:k], padding[:])
        } else {
            copy(onion.Data[j:k], onion.Data[k:l])
        }
    }
}

// Insert new sub-header at the start of the list and rotate all other subheaders by one chunk to the right
func rotateSubHeadersRight(onion *common.OnionPacket) {

    for i := common.OnionSubHeaderCount - 1; i >= 0; i-- {

        j := (i - 1) * common.OnionSubHeaderSize
        k := i  * common.OnionSubHeaderSize
        l := (i + 1) * common.OnionSubHeaderSize

        if i == 0 {
            copy(onion.Data[k:l], make([]byte, common.OnionSubHeaderSize))
        } else {
            copy(onion.Data[k:l], onion.Data[j:k])
        }
    }
}

// Unwraps one layer of an onion
func ExtractOnionSubHeader(onion *common.OnionPacket) (*common.OnionSubHeader, error) {

    subHeaderBytes := onion.Data[:common.OnionSubHeaderSize]

    var subHeader common.OnionSubHeader

    err := Decode(subHeaderBytes, &subHeader)

    if err != nil {
        return nil, err
    }

    return &subHeader, nil
}

//
//  VERIFICATION
//

// Returns true if the onion this sub-header corresponds to has no next hop
func isLast(subHeader *common.OnionSubHeader) bool {
    return subHeader.NextHop == subHeader.PrevHop
}

// Checks that the data inside the onion is valid according to the sub-header
func isValid(onion *common.OnionPacket, subHeader *common.OnionSubHeader) bool {
    hash := onion.Hash()
    return bytes.Equal(hash[:], subHeader.Hash[:])
}
