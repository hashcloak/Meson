package chain

import (
	"encoding/json"
	"fmt"

	"github.com/hashcloak/Meson/plugin/pkg/common"
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

func (ec *ETHChain) WrapPostRequest(rpcURL string, req *common.PostRequest) ([]HttpData, error) {
	if len(rpcURL) == 0 {
		return []HttpData{}, fmt.Errorf("non existent RPC URL for Ethereum chain")
	}
	marshalledRequest, err := json.Marshal(ethRequest{
		ID:      ec.chainID,
		JSONRPC: "2.0",
		METHOD:  "eth_sendRawTransaction",
		Params:  []string{req.TxHex},
	})
	return []HttpData{{URL: rpcURL, Body: marshalledRequest}}, err
}

func (ec *ETHChain) WrapQueryRequest(rpcURL string, req *common.QueryRequest) ([]HttpData, error) {
	if len(rpcURL) == 0 {
		return []HttpData{}, fmt.Errorf("non existent RPC URL for Ethereum chain")
	}
	nonceRequest, err := json.Marshal(ethRequest{
		ID:      ec.chainID,
		JSONRPC: "2.0",
		METHOD:  "eth_getTransactionCount",
		Params:  []string{req.From, "pending"},
	})
	if err != nil {
		return nil, err
	}
	gasPriceRequest, err := json.Marshal(ethRequest{
		ID:      ec.chainID,
		JSONRPC: "2.0",
		METHOD:  "eth_gasPrice",
	})
	if err != nil {
		return nil, err
	}
	param, err := json.Marshal(map[string]interface{}{
		"from":  req.From,
		"to":    req.To,
		"value": req.Value,
		"data":  req.Data,
	})
	if err != nil {
		return nil, err
	}
	gasEstimateRequest, err := json.Marshal(ethRequest{
		ID:      ec.chainID,
		JSONRPC: "2.0",
		METHOD:  "eth_estimateGas",
		Params:  []string{string(param)},
	})
	if err != nil {
		return nil, err
	}
	return []HttpData{
		{URL: rpcURL, Body: nonceRequest},
		{URL: rpcURL, Body: gasPriceRequest},
		{URL: rpcURL, Body: gasEstimateRequest},
	}, nil
}
