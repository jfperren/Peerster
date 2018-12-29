package gossiper

import (
    "bytes"
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "encoding/binary"
    "fmt"
    "github.com/dedis/protobuf"
    "github.com/jfperren/Peerster/common"
)

//
//  ERRORS
//

const InvalidHeader = 1
const InvalidPayload = 2

// Errors thrown by the File System type
type OnionError struct {
    flag     int
}

//
//  WRAP / UNWRAP
//

// Wrap a regular GossipPacket into an onion that can be send on the mix network
func GenerateOnion(gossipPacket *common.GossipPacket, route []crypto.PublicKey) *common.OnionPacket {

    // TODO

    return nil
}

// // Unwraps one layer of an onion
// func (gossiper *Gossiper) UnwrapOnion(onion *common.OnionPacket) (*common.OnionPacket, *OnionError) {
//
//     data := gossiper.Crypto.Decypher(onion.Data[:])
//
//
//
//     common.OnionPacket
// }

func (crypto *Crypto) wrap(onion *common.OnionPacket, key rsa.PublicKey) {

    cipher := crypto.Cypher(onion.Data[:], key)

    var data [common.OnionHeaderSize + common.OnionPayloadSize]byte
    copy(data[:], cipher)

    onion.Data = data
}

func (crypto *Crypto) unwrap(onion *common.OnionPacket) {

    decipher := crypto.Decypher(onion.Data[:])

    var data [common.OnionHeaderSize + common.OnionPayloadSize]byte
    copy(data[:], decipher)

    onion.Data = data
}

func encode(object interface{}, blockSize int) ([]byte, error) {

    objSize := binary.Size(object)

    if objSize == -1 {
        // Err: Not encodable
    }

    totalSize := objSize + 8

    if totalSize > blockSize {
        // Err: Not enough space
    }

    buf := new(bytes.Buffer)

    err := binary.Write(buf, binary.LittleEndian, int64(objSize))

    if err != nil {
        return []byte{}, err
    }

    err = binary.Write(buf, binary.LittleEndian, object)

    if err != nil {
        return []byte{}, err
    }

    // Copy buffer bytes into a byte slice
    data := make([]byte, blockSize)
    copy(data, buf.Bytes())

    // Add padding at the end
    rand.Read(data[totalSize:])

    return data, nil
}

func decode(data []byte, structPtr interface{}) error {

    objSize, n := binary.Uvarint(data[:8])

    if n <= 0 {
        // Some error
    }

    buf := bytes.NewReader(data[8:objSize+8])
    err := binary.Read(buf, binary.LittleEndian, structPtr)

    return err
}


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

func rotateSubHeadersRight(onion *common.OnionPacket) {

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



// Unwraps one layer of an onion
func ExtractOnionSubHeader(onion *common.OnionPacket) (*common.OnionSubHeader, *OnionError) {

    // The subHeader should be at the beginning of the byte list
    subHeaderBytes := onion.Payload[:common.OnionSubHeaderSize]

    // Try and decode the bytes via protobuf
    var subHeader common.OnionSubHeader
    err := protobuf.Decode(subHeaderBytes, &subHeader)

    if err != nil {
        return nil, &OnionError{flag: InvalidHeader}
    }

    return &subHeader, nil
}


