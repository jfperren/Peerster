package gossiper

import (
    "bytes"
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "encoding/binary"
    "errors"
    "github.com/jfperren/Peerster/common"
)

//
//  ERRORS
//

const InvalidHeader = 1
const InvalidPayload = 2

// Errors thrown by the File System type
var (
    ErrEncodingNotFixedSize =errors.New("Structure provided has variable size, cannot encode.")
    ErrEncodingNotEnoughSpace = errors.New("Not enough bytes to fit the data. Increase blockSize.")
    ErrDecodingNoSize = errors.New("Error getting size of bytes to decode")
)


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
        return []byte{}, ErrEncodingNotFixedSize
    }

    totalSize := objSize + 8

    if totalSize > blockSize {
        return []byte{}, ErrEncodingNotEnoughSpace
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
        return ErrDecodingNoSize
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

func rotateSubHeadersRight(onion *common.OnionPacket, insertion common.OnionSubHeader) {

    for i := common.OnionSubHeaderCount - 1; i >= 0; i-- {

        j := (i - 1) * common.OnionSubHeaderSize
        k := i  * common.OnionSubHeaderSize
        l := (i + 1) * common.OnionSubHeaderSize

        if i == 0 {
            bytes, _ := encode(insertion, common.OnionSubHeaderSize)
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

    err := decode(subHeaderBytes, &subHeader)

    if err != nil {
        return nil, err
    }

    return &subHeader, nil
}


