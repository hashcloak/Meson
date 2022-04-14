package chain

import (
	"fmt"

	"github.com/hashcloak/Meson/plugin/pkg/common"
)

// CosmosChain is a struct for identifier blockchains and their forks
type CosmosChain struct {
	ticker  string
	chainID int
}

func (ec *CosmosChain) WrapPostRequest(rpcURL string, req *common.PostRequest) ([]HttpData, error) {
	if len(rpcURL) == 0 {
		return []HttpData{}, fmt.Errorf("no URL value for cosmos api")
	}
	URL := fmt.Sprintf("%s/broadcast_tx_async?tx=0x%s", rpcURL, req.TxHex)
	return []HttpData{{URL: URL}}, nil
}

func (ec *CosmosChain) WrapQueryRequest(rpcURL string, req *common.QueryRequest) ([]HttpData, error) {
	// TODO
	return []HttpData{}, nil
}
