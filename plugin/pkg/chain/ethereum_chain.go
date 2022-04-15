package chain

import (
	"encoding/json"
	"fmt"

	"github.com/hashcloak/Meson/plugin/pkg/command"
	"github.com/ugorji/go/codec"
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
	Params interface{} `json:"params"`
}

// ETHChain is a struct for identifier blockchains and their forks
type ETHChain struct {
	chainID uint
	ticker  string
}

func (ec *ETHChain) WrapRequest(rpcURL string, cmd uint8, payload []byte) ([]HttpData, error) {
	if len(rpcURL) == 0 {
		return []HttpData{}, fmt.Errorf("non existent RPC URL for Ethereum chain")
	}
	switch cmd {
	case command.PostTransaction:
		var req command.PostTransactionRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		if err := dec.Decode(&req); err != nil {
			return nil, err
		}
		marshalledRequest, err := json.Marshal(ethRequest{
			ID:      ec.chainID,
			JSONRPC: "2.0",
			METHOD:  "eth_sendRawTransaction",
			Params:  []string{req.TxHex},
		})
		if err != nil {
			return nil, err
		}
		return []HttpData{{URL: rpcURL, Body: marshalledRequest}}, nil

	case command.EthQuery:
		var req command.EthQueryRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		if err := dec.Decode(&req); err != nil {
			return nil, err
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
		param := map[string]interface{}{
			"to":    req.To,
			"value": fmt.Sprintf("0x%x", req.Value),
		}
		if req.Data != "" {
			param["data"] = req.Data
		}
		gasEstimateRequest, err := json.Marshal(ethRequest{
			ID:      ec.chainID,
			JSONRPC: "2.0",
			METHOD:  "eth_estimateGas",
			Params:  []interface{}{param},
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
	return nil, fmt.Errorf("invalid cmd %x for chain %d", cmd, ec.chainID)
}

func (ec *ETHChain) UnwrapResponse(cmd uint8, payload []string) ([]byte, error) {
	switch cmd {
	case command.PostTransaction:
		if len(payload) != 1 {
			return nil, fmt.Errorf("expect 1 response, got %d", len(payload))
		}
		return json.Marshal(command.PostTransactionResponse{
			TxHash: payload[0],
		})
	case command.EthQuery:
		if len(payload) != 3 {
			return nil, fmt.Errorf("expect 3 response, got %d", len(payload))
		}
		return json.Marshal(command.EthQueryResponse{
			Nonce:    payload[0],
			GasPrice: payload[1],
			GasLimit: payload[2],
		})
	}
	return nil, fmt.Errorf("unexpected error")
}
