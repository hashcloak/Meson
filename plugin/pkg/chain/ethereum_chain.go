package chain

import (
	"encoding/json"
	"fmt"
)

// An ethereum request abstraction.
// Only need it for one method, though.
type ethRequest struct {
	// ChainId to indicate which Ethereum-based network
	ID uint `json:"id"`
	// Indicates which version of JSON RPC to use
	// Since all networks support JSON RPC 2.0,
	// this attribute is a constant
	JSONRPC string `json:"jsonrpc"`
	// Which method you want to call
	METHOD string `json:"method"`
	// Params for the method you want to call
	Params []string `json:"params"`
}

// ETHChain is a struct for identifier blockchains and their forks
type ETHChain struct {
	chainID uint
	ticker  string
}

// NewRequest takes an RPC URL and a hexadecimal transaction.
// Returns PostRequest for ethereum nodes
func (ec *ETHChain) NewRequest(rpcURL string, txHex string) (PostRequest, error) {
	if len(rpcURL) == 0 {
		return PostRequest{}, fmt.Errorf("Non existent RPC URL for Ethereum chain")
	}
	er := ethRequest{
		ID:      ec.chainID,
		JSONRPC: "2.0",
		METHOD:  "eth_sendRawTransaction",
		Params:  []string{txHex},
	}
	marshalledRequest, err := json.Marshal(er)
	return PostRequest{URL: rpcURL, Body: marshalledRequest}, err
}
