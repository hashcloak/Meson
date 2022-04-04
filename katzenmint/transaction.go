package katzenmint

import (
	"crypto/ed25519"
	"crypto/sha256"
	"fmt"

	"github.com/katzenpost/core/crypto/eddsa"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	"github.com/ugorji/go/codec"
)

// TODO: cache the hex decode result?
// signTransaction represents the transaction used to make transaction hash
type signTransaction struct {
	// version
	Version string

	// Epoch
	Epoch uint64

	// command
	Command Command

	// hex encoded ed25519 public key (should not be 0x prefxied)
	// TODO: is there better way to take PublicKey?
	PublicKey string

	// payload (mix descriptor/pki document/authority)
	Payload string
}

// TODO: find a better way to represent the Transaction
// maybe add nonce? switch to rlp encoding?
// Transaction represents a transaction used to make state change, eg:
// publish mix descriptor or add new authority
type Transaction struct {
	// version
	Version string

	// Epoch
	Epoch uint64

	// command
	Command Command

	// hex encoded ed25519 public key (should not be 0x prefxied)
	PublicKey string

	// hex encoded ed25519 signature (should not be 0x prefixed)
	Signature string

	// json encoded payload (eg. mix descriptor/authority)
	Payload string
}

// SerializeHash returns the serialize hash that user signed of the given transaction
func (tx *Transaction) SerializeHash() (txHash [32]byte) {
	signTx := new(signTransaction)
	signTx.Version = tx.Version
	signTx.Epoch = tx.Epoch
	signTx.Command = tx.Command
	signTx.PublicKey = tx.PublicKey
	signTx.Payload = tx.Payload
	serializedTx := make([]byte, 128)
	enc := codec.NewEncoderBytes(&serializedTx, jsonHandle)
	if err := enc.Encode(signTx); err != nil {
		return
	}
	txHash = sha256.Sum256(serializedTx)
	return
}

// IsVerified returns whether transaction was signed by the public key
func (tx *Transaction) IsVerified() (isVerified bool) {
	msgHash := tx.SerializeHash()
	if len(msgHash) <= 0 {
		return
	}
	pub := DecodeHex(tx.PublicKey)
	sig := DecodeHex(tx.Signature)
	pubKey := ed25519.PublicKey(pub)
	isVerified = ed25519.Verify(pubKey, msgHash[:], sig)
	return
}

// PublicKeyBytes returns public key bytes of the given transaction
func (tx *Transaction) PublicKeyBytes() (pk []byte) {
	pk = DecodeHex(tx.PublicKey)
	return
}

// PublicKeyByteArray returns public key bytes of the given transaction
func (tx *Transaction) PublicKeyByteArray() (pk [eddsa.PublicKeySize]byte) {
	pubkey := DecodeHex(tx.PublicKey)
	copy(pk[:], pubkey)
	return
}

// Address returns public address of the given transaction
func (tx *Transaction) Address() string {
	v := abcitypes.UpdateValidator(tx.PublicKeyBytes(), 0, "")
	pubkey, err := cryptoenc.PubKeyFromProto(v.PubKey)
	if err != nil {
		return ""
	}
	return string(pubkey.Address())
}

// FormTransaction returns the crafted transaction that can be posted
func FormTransaction(command Command, epoch uint64, payload string, privKey *eddsa.PrivateKey) ([]byte, error) {
	tx := &Transaction{
		Version:   protocolVersion,
		Epoch:     epoch,
		Command:   command,
		Payload:   payload,
		PublicKey: EncodeHex(privKey.PublicKey().Bytes()),
	}
	msgHash := tx.SerializeHash()
	sig := privKey.Sign(msgHash[:])
	tx.Signature = EncodeHex(sig)
	if !tx.IsVerified() {
		return nil, fmt.Errorf("unable to sign properly")
	}
	return EncodeJson(tx)
}
