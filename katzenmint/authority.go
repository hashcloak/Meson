package katzenmint

import (
	"github.com/ugorji/go/codec"
)

// Authority represents authority in katzenmint.
type Authority struct {
	// Auth is the prefix of the authority.
	Auth string

	// PubKey is the validator's public key.
	PubKey []byte

	// KeyType is the validator's key type.
	KeyType string

	// Power is the voting power of the authority.
	Power int64
}

func VerifyAndParseAuthority(payload []byte) (*Authority, error) {
	authority := new(Authority)
	dec := codec.NewDecoderBytes(payload, jsonHandle)
	if err := dec.Decode(authority); err != nil {
		return nil, err
	}
	// TODO: check authority
	return authority, nil
}
