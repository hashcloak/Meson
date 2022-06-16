package katzenmint

import (
	"fmt"

	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
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

	// Credential is the credential for becoming an authority
	Credential string
}

// AuthorityChecked represents checked authority.
type AuthorityChecked struct {
	Auth       string
	Val        *abcitypes.ValidatorUpdate
	Credential string
}

func VerifyAndParseAuthority(payload []byte) (*AuthorityChecked, error) {
	authority := new(Authority)
	dec := codec.NewDecoderBytes(payload, jsonHandle)
	if err := dec.Decode(authority); err != nil {
		return nil, err
	}
	// Check authority
	if authority.Auth == "" {
		return nil, fmt.Errorf("no authority name provided")
	}
	if authority.Power < 0 {
		return nil, fmt.Errorf("negative voting power")
	}
	if authority.KeyType != ed25519.KeyType {
		return nil, fmt.Errorf("only support ed25519 key type")
	}
	pke := ed25519.PubKey(authority.PubKey)
	pkp, err := cryptoenc.PubKeyToProto(pke)
	if err != nil {
		return nil, err
	}
	valUpd := &abcitypes.ValidatorUpdate{
		PubKey: pkp,
		Power:  authority.Power,
	}
	checked := &AuthorityChecked{
		Auth:       authority.Auth,
		Val:        valUpd,
		Credential: authority.Credential,
	}
	return checked, nil
}
