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
    ErrOnionNextHopNotFound = errors.New("could not find name corresponding to public key in NextHop")

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrOnionHashDoNotMatch = errors.New("found hash that does not match computed hash")

)

//
//  WRAP / UNWRAP
//

// Wrap a regular GossipPacket into an onion that can be send on the mix network
func (crypto *Crypto) GenerateOnion(gossipPacket *common.GossipPacket, route []rsa.PublicKey) (*common.OnionPacket, error) {

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

        var next rsa.PublicKey
        var prev rsa.PublicKey

        if i == 0 {
            // Use our key as previous for the first node
            prev = crypto.PublicKey()
        } else {
            prev = route[i-1]
        }

        if i == len(route) - 1 {
            // To signal that a node is the last one, we use their key as nextHop
            next = route[i]
        } else {
            next = route[i+1]
        }

        subHeader := common.OnionSubHeader{
            PrevHop: prev,
            NextHop: next,
            Hash: onion.Hash(),
        }

        rotateSubHeadersRight(onion, subHeader)
        crypto.wrap(onion, node)
    }

    return onion, nil
}

func (gossiper *Gossiper) ProcessOnion(onion *common.OnionPacket) (*common.GossipPacket, error) {

    // First, unwrap the onion
    gossiper.Crypto.unwrap(onion)

    // Get first subHeader
    subHeader, err := ExtractOnionSubHeader(onion)
    if err != nil { return nil, err }

    // Check integrity of onion content
    if !isValid(onion, subHeader) {
        return nil, ErrOnionHashDoNotMatch
    }

    // If we are last, we should be able to get the data
    if isLast(subHeader) {

        // Try to decode payload into a gossip packet
        var gossipPacket common.GossipPacket
        err = Decode(onion.Data[common.OnionHeaderSize:], &gossipPacket)
        if err != nil { return nil, err }

        return &gossipPacket, nil

    } else {

        // Rotate subHeaders for the next node
        rotateSubHeadersLeft(onion)

        // Finds next destination based on keys

        found := false

        for name, key := range gossiper.BlockChain.MixerNodes {
            if key.E == subHeader.NextHop.E {
                onion.Destination = name
                onion.HopLimit = common.InitialHopLimit
                found = true
            }
        }

        if !found {
            return nil, ErrOnionNextHopNotFound
        }

        return nil, nil
    }
}

//
//  WRAP / UNWRAP
//

func (crypto *Crypto) wrap(onion *common.OnionPacket, key rsa.PublicKey) {

    cipher := crypto.Cypher(onion.Data[:], key)

    var data [common.OnionHeaderSize + common.OnionPayloadSize]byte
    copy(data[:], cipher)

    onion.Data = data
}

// Unwrap one layer of the onion using the node's public key
func (crypto *Crypto) unwrap(onion *common.OnionPacket) {

    decipher := crypto.Decypher(onion.Data[:])

    var data [common.OnionHeaderSize + common.OnionPayloadSize]byte
    copy(data[:], decipher)

    onion.Data = data
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
func rotateSubHeadersRight(onion *common.OnionPacket, insertion common.OnionSubHeader) {

    for i := common.OnionSubHeaderCount - 1; i >= 0; i-- {

        j := (i - 1) * common.OnionSubHeaderSize
        k := i  * common.OnionSubHeaderSize
        l := (i + 1) * common.OnionSubHeaderSize

        if i == 0 {
            bytes, _ := Encode(insertion, common.OnionSubHeaderSize)
            copy(onion.Data[k:l], bytes)
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
