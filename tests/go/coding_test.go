package tests

import (
    "github.com/jfperren/Peerster/gossiper"
    "testing"
)

type EncodableStruct struct {

    A int64
    B []int64
    C string

}

func TestEncodingDecodingNoPadding(t *testing.T) {

    blockSize := 28
    original := EncodableStruct{
        A: 12,
        B: []int64{1, 2, 3},
        C: "hello world",
    }

    encoded, err := gossiper.Encode(&original, blockSize)

    if err != nil {
        t.Errorf("Error encoding data: %v", err)
    }

    if len(encoded) != blockSize {
        t.Errorf("Encoded data does not have correct size. Was expecting %v, got %v", blockSize, len(encoded))
    }

    var decoded EncodableStruct

    err = gossiper.Decode(encoded, &decoded)

    if err != nil {
        t.Errorf("Error decoding data: %v", err)
    }

    if decoded.A != original.A {
        t.Errorf("Decoded struct is different than original. Field a was originally %v, now is %v", original.A, decoded.A)
    }

    if len(decoded.B) != len(original.B) {
        t.Errorf("Decoded struct is different than original. Field b was originally %v elements, now has %v", len(original.B), len(decoded.B))
    }

    if decoded.B[0] != original.B[0] {
        t.Errorf("Decoded struct is different than original. Field b[0] was originally %v, now is %v", original.B[0], decoded.B[0])
    }

    if decoded.B[1] != original.B[1] {
        t.Errorf("Decoded struct is different than original. Field b[1] was originally %v, now is %v", original.B[1], decoded.B[1])
    }

    if decoded.B[2] != original.B[2] {
        t.Errorf("Decoded struct is different than original. Field b[2] was originally %v, now is %v", original.B[2], decoded.B[2])
    }

    if decoded.C != original.C {
        t.Errorf("Decoded struct is different than original. Field a was originally %v, now is %v", original.C, decoded.C)
    }
}

func TestEncodingDecodingPadding(t *testing.T) {

    blockSize := 64
    original := EncodableStruct{
        A: 12,
        B: []int64{1, 2, 3},
        C: "hello world",
    }

    encoded, err := gossiper.Encode(&original, blockSize)

    if err != nil {
        t.Errorf("Error encoding data: %v", err)
    }

    if len(encoded) != blockSize {
        t.Errorf("Encoded data does not have correct size. Was expecting %v, got %v", blockSize, len(encoded))
    }

    var decoded EncodableStruct

    err = gossiper.Decode(encoded, &decoded)

    if err != nil {
        t.Errorf("Error decoding data: %v", err)
    }

    if decoded.A != original.A {
        t.Errorf("Decoded struct is different than original. Field a was originally %v, now is %v", original.A, decoded.A)
    }

    if len(decoded.B) != len(original.B) {
        t.Errorf("Decoded struct is different than original. Field b was originally %v elements, now has %v", len(original.B), len(decoded.B))
    }

    if decoded.B[0] != original.B[0] {
        t.Errorf("Decoded struct is different than original. Field b[0] was originally %v, now is %v", original.B[0], decoded.B[0])
    }

    if decoded.B[1] != original.B[1] {
        t.Errorf("Decoded struct is different than original. Field b[1] was originally %v, now is %v", original.B[1], decoded.B[1])
    }

    if decoded.B[2] != original.B[2] {
        t.Errorf("Decoded struct is different than original. Field b[2] was originally %v, now is %v", original.B[2], decoded.B[2])
    }

    if decoded.C != original.C {
        t.Errorf("Decoded struct is different than original. Field a was originally %v, now is %v", original.C, decoded.C)
    }
}

func TestEncodingDecodingNotEnoughSpace(t *testing.T) {

    blockSize := 27
    original := EncodableStruct{
        A: 12,
        B: []int64{1, 2, 3},
        C: "hello world",
    }

    _, err := gossiper.Encode(&original, blockSize)

    if err == nil {
        t.Errorf("Should throw an error when not enough space to encode")
    }
}