package gossiper

import (
    "crypto/rand"
    "encoding/binary"
    "errors"
    "github.com/dedis/protobuf"
)

//
//  ERRORS
//

// Errors thrown by the Coding module
var (

    // Thrown when the block size given is smaller than the size of the structure to encode
    ErrEncodingNotEnoughSpace = errors.New("not enough bytes to fit the data. Increase blockSize")
)

// Encode a structure into a slice of bytes that contains its size and is padded with random bits
func Encode(structPtr interface{}, blockSize int) ([]byte, error) {

    objBytes, err := protobuf.Encode(structPtr)

    if err != nil {
        return []byte{}, err
    }

    objSize := len(objBytes)

    if err != nil {
        return []byte{}, err
    }

    totalSize := objSize + 8

    if totalSize > blockSize {
        return []byte{}, ErrEncodingNotEnoughSpace
    }

    // Copy buffer bytes into a byte slice
    data := make([]byte, blockSize)
    _ = binary.PutUvarint(data[:8], uint64(len(objBytes)))
    copy(data[8:totalSize], objBytes)

    // Add padding at the end
    rand.Read(data[totalSize:])

    return data, nil
}

// Decode bytes into a structure
func Decode(data []byte, structPtr interface{}) error {

    objSize, _ := binary.Uvarint(data[:8])

    return protobuf.Decode(data[8:8+objSize], structPtr)
}