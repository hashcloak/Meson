package katzenmint

import (
	"crypto/ed25519"
	"testing"

	"github.com/katzenpost/core/crypto/rand"
	"github.com/ugorji/go/codec"
)

type Payload struct {
	Text   string
	Number int
}

func TestTransaction(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("cannot generate key pair: %+v\n", err)
	}
	payload := make([]byte, 128)
	enc := codec.NewEncoderBytes(&payload, jsonHandle)
	if err := enc.Encode(Payload{
		Text:   "test",
		Number: 1,
	}); err != nil {
		t.Fatalf("cannot json marshal payload: %+v\n", err)
	}
	tx := new(Transaction)
	tx.Version = "1.0"
	tx.Epoch = 10
	tx.Command = 1
	tx.Payload = string(payload)
	tx.PublicKey = EncodeHex(pubKey[:])
	msgHash := tx.SerializeHash()
	sig := ed25519.Sign(privKey, msgHash[:])
	tx.Signature = EncodeHex(sig[:])
	if !tx.IsVerified() {
		t.Fatalf("transaction is not verified: %+v\n", tx)
	}
}
