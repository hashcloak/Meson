package katzenmint

import (
	"bytes"
	"testing"
)

type decodeTest struct {
	Src string
	Out []byte
}

type encodeTest struct {
	Src []byte
	Out string
}

var (
	decodeHexTests = []decodeTest{
		{
			Src: "6b61747a656e6d696e74",
			Out: []byte{
				107, 97, 116, 122, 101, 110, 109, 105, 110, 116,
			},
		},
	}
	encodeHexTests = []encodeTest{
		{
			Src: []byte{
				107, 97, 116, 122, 101, 110, 109, 105, 110, 116,
			},
			Out: "6b61747a656e6d696e74",
		},
	}
)

func TestDecodeHex(t *testing.T) {
	for _, test := range decodeHexTests {
		b := DecodeHex(test.Src)
		if !bytes.Equal(test.Out, b) {
			t.Fatalf("decode hex results should be equal")
		}
	}
}

func TestEncodeeHex(t *testing.T) {
	for _, test := range encodeHexTests {
		b := EncodeHex(test.Src)
		if test.Out != b {
			t.Fatalf("encode hex results should be equal")
		}
	}
}
