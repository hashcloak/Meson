package proxy

import (
	"encoding/json"
)

// An ethereum request abstraction. 
// Only need it for one method, though.
type ethRequest struct {
	// ChainId to indicate which Ethereum-based network
	ID int `json:"id"` 
	// Indicates which version of JSON RPC to use
	// Since all networks support JSON RPC 2.0,
	// this attribute is a constant
	JSONRPC string `json:"jsonrpc"`
	// Which method you want to call
	METHOD string `json:"method"`
	// Params for the method you want to call
	Params []string `json:"params"`
}

// Takes a chainId and signed transaction data as parameters
// Returns a JSON encoding of the request
func marshalRequest(id int, params []string) ([]byte, error) {
	request := ethRequest {
		ID: id,
		JSONRPC: "2.0",
		METHOD: "eth_sendRawTransaction",
		Params: params,
	}

	return json.Marshal(request)
}