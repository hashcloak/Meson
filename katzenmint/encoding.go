package katzenmint

import (
	"encoding/hex"
	"fmt"

	"github.com/ugorji/go/codec"
)

var jsonHandle *codec.JsonHandle

func init() {
	jsonHandle = new(codec.JsonHandle)
	jsonHandle.Canonical = true
	jsonHandle.IntegerAsString = 'A'
	jsonHandle.MapKeyAsString = true
}

// DecodeHex return byte of the given hex string
// return nil if the src is not valid hex string
func DecodeHex(src string) (out []byte) {
	slen := len(src)
	if slen <= 0 {
		return
	}
	if (slen % 2) > 0 {
		src = fmt.Sprintf("0%s", src)
	}
	out, _ = hex.DecodeString(src)
	return
}

// EncodeHex return encoded hex string of the given
// bytes
func EncodeHex(src []byte) (out string) {
	out = hex.EncodeToString(src)
	return
}

func DecodeJson(data []byte, v interface{}) error {
	dec := codec.NewDecoderBytes(data, jsonHandle)
	err := dec.Decode(v)
	return err
}

func EncodeJson(v interface{}) (data []byte, err error) {
	err = codec.NewEncoderBytes(&data, jsonHandle).Encode(v)
	return
}
