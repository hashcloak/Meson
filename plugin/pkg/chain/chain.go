package chain

import "github.com/ugorji/go/codec"

// HttpData is the common struct containing the body and url
// Tag is for specifying the response message
type HttpData struct {
	Body []byte
	URL  string
}

// IChain is an abstraction for a cryptocurrency
// It creates raw transactions
type IChain interface {
	WrapRequest(rpcURL string, cmd uint8, payload []byte) ([]HttpData, error)
	UnwrapResponse(cmd uint8, payload []string) ([]byte, error)
}

var jsonHandle codec.JsonHandle
