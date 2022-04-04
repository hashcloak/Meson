package chain

import (
	"fmt"
)

// CosmosChain is a struct for identifier blockchains and their forks
type CosmosChain struct {
	ticker  string
	chainID int
}

// NewRequest takes an RPC URL and a hexadecimal transaction.
// Returns PostRequest for cosmos nodes
func (ec *CosmosChain) NewRequest(rpcURL string, txHex string) (PostRequest, error) {
	if len(rpcURL) == 0 {
		return PostRequest{}, fmt.Errorf("No URL value for cosmos api")
	}
	URL := fmt.Sprintf("%s/broadcast_tx_async?tx=0x%s", rpcURL, txHex)
	return PostRequest{URL: URL}, nil
}
