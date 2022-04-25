package chain

import (
	"fmt"

	"github.com/ugorji/go/codec"
)

// HttpData is the common struct containing the body and url
// Tag is for specifying the response message
type HttpData struct {
	Method string
	URL    string
	Body   []byte
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RPCResponse struct {
	Version string    `json:"jsonrpc,omitempty"`
	ID      uint      `json:"id,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
	Result  string    `json:"result,omitempty"`
}

// IChain is an abstraction for a cryptocurrency
// It creates raw transactions
type IChain interface {
	WrapRequest(rpcURL string, cmd uint8, payload []byte) (*HttpData, error)
	UnwrapResponse(cmd uint8, payload []RPCResponse) ([]byte, error)
}

var jsonHandle codec.JsonHandle

func errNumResponse(expect, actual int) error {
	return fmt.Errorf("expect %d response, got %d", expect, actual)
}
func errCodeAndMsg(code int, msg string) error {
	return fmt.Errorf("error code: %d, msg: %s", code, msg)
}
