package common

import (
	"math/big"

	"github.com/ugorji/go/codec"
)

const (
	PostCommand   uint8 = 0
	QueryCommand  uint8 = 1
	TotalCommands uint8 = 2
)

type PostRequest struct {
	TxHex string
}
type QueryRequest struct {
	From  string
	To    string
	Value *big.Int
	Data  string
}

func PostRequestFromRaw(raw []byte) (*PostRequest, error) {
	var req PostRequest
	dec := codec.NewDecoderBytes(raw, &jsonHandle)
	if err := dec.Decode(&req); err != nil {
		return nil, errInvalidJson
	}
	return &req, nil
}

func QueryRequestFromRaw(raw []byte) (*QueryRequest, error) {
	var req QueryRequest
	dec := codec.NewDecoderBytes(raw, &jsonHandle)
	if err := dec.Decode(&req); err != nil {
		return nil, errInvalidJson
	}
	return &req, nil
}

func (req *PostRequest) ToRaw() []byte {
	var request []byte
	enc := codec.NewEncoderBytes(&request, &jsonHandle)
	_ = enc.Encode(req)
	return request
}

func (req *QueryRequest) ToRaw() []byte {
	var request []byte
	enc := codec.NewEncoderBytes(&request, &jsonHandle)
	_ = enc.Encode(req)
	return request
}
