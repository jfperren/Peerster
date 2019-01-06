package gossiper

import (
    "bytes"
    "crypto/rand"
    "crypto/rsa"
    "errors"
    "github.com/jfperren/Peerster/common"
    mrand "math/rand"
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

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrNotEnoughMixerNodes = errors.New("could not create onion, needs at least 2 mixer nodes")

)

//
//  WRAP / UNWRAP
//

func (gossiper *Gossiper) GenerateOnion(gossipPacket *common.GossipPacket, route []string) (*common.OnionPacket, error) {


    return gossiper.Crypto.GenerateOnion(gossipPacket, route, gossiper.BlockChain.MixerNodes, gossiper.Name)
}

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
            // To signal that a node is the last one, we use an empty string
            next = common.NoNextHop
        } else {
            next = route[i+1]
        }

        RotateSubHeadersRight(onion)
        crypto.wrap(onion, *key, prev, next)
    }

    return onion, nil
}

func (gossiper *Gossiper) ProcessOnion(onion *common.OnionPacket) (*common.GossipPacket, *common.OnionSubHeader, error) {

    // First, unwrap the onion and get subHeader
    subHeader, err := gossiper.Crypto.unwrap(onion)
    if err != nil { return nil, subHeader, err }

    // If we are last, we should be able to get the data
    if isLast(subHeader) {

        // Try to decode payload into a gossip packet
        var gossipPacket common.GossipPacket
        err = Decode(onion.Data[common.OnionHeaderSize:], &gossipPacket)
        if err != nil { return nil, subHeader, err }

        return &gossipPacket, subHeader, nil

    } else {

        // Rotate subHeaders for the next node
        RotateSubHeadersLeft(onion)

        // Reset hop limit
        onion.HopLimit = common.InitialHopLimit

        return nil, subHeader, nil
    }
}

//
//  WRAP / UNWRAP
//

func (crypto *Crypto) wrap(onion *common.OnionPacket, key rsa.PublicKey, prev, next string) error {

    symmetricKey := NewCTRSecret()

    // Encrypt symmetric part
    otherData := onion.Data[common.OnionSubHeaderSize:]
    otherDataCipher, iv, err := CTRCipher(otherData, symmetricKey)
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
    otherData, err := CTRDecipher(otherDataCipher, subHeader.Key[:], subHeader.IV[:])

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
func RotateSubHeadersLeft(onion *common.OnionPacket) {

    for i := 0; i < common.OnionSubHeaderCount; i++ {

        j := i * common.OnionSubHeaderSize
        k := (i + 1) * common.OnionSubHeaderSize
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
func RotateSubHeadersRight(onion *common.OnionPacket) {

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

//
//  VERIFICATION
//

// Returns true if the onion this sub-header corresponds to has no next hop
func isLast(subHeader *common.OnionSubHeader) bool {
    return subHeader.NextHop == common.NoNextHop
}

// Checks that the data inside the onion is valid according to the sub-header
func isValid(onion *common.OnionPacket, subHeader *common.OnionSubHeader) bool {
    hash := onion.Hash()
    return bytes.Equal(hash[:], subHeader.Hash[:])
}

//
//  GOSSIPER FUNCTIONS
//

func (gossiper *Gossiper) wrapInOnionIfNeeded(gossipPacket *common.GossipPacket) (*common.GossipPacket, error) {

    if !gossiper.shouldWrapInOnion() {
        return gossipPacket, nil
    }

    route, err := gossiper.randomMixRoute()
    if err != nil { return nil, err }

    onion, err := gossiper.Crypto.GenerateOnion(gossipPacket, route, gossiper.BlockChain.MixerNodes, gossiper.Name)
    if err != nil { return nil, err }

    return onion.Packed(), nil
}

func (gossiper *Gossiper) shouldWrapInOnion() bool {
    return gossiper.MixLength > 0
}

// Return a random mixer node except the one given as parameter.
// Note - This is probably the ugliest piece of code I have ever written,
// but I am fairly convinced there is not a better way to do that.
func (gossiper *Gossiper) randomMixerNodeExcept(except string) string {

    for {

        index := mrand.Intn(len(gossiper.BlockChain.MixerNodes))
        i := 0

        for node, _ := range gossiper.BlockChain.MixerNodes {

            if i == index && node != except {
                return node
            }

            i++
        }
    }
}

func (gossiper *Gossiper) randomMixRoute() ([]string, error) {

    if len(gossiper.BlockChain.MixerNodes) < 2 {
        return nil, ErrNotEnoughMixerNodes
    }

    route := make([]string, 0)
    current := gossiper.Name

    for i := uint(0); i < gossiper.MixLength; i++ {
        node := gossiper.randomMixerNodeExcept(current)
        route = append(route, node)
        current = node
    }

    return route, nil
}