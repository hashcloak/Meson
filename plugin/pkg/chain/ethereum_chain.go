package chain

import (
	"encoding/json"
	"fmt"

	"github.com/hashcloak/Meson/plugin/pkg/command"
	"github.com/ugorji/go/codec"
)

// ETHChain is a struct for identifier blockchains and their forks
type ETHChain struct {
	chainID uint
	ticker  string
}

func (ec *ETHChain) WrapRequest(rpcURL string, cmd uint8, payload []byte) (*HttpData, error) {
	if len(rpcURL) == 0 {
		return nil, fmt.Errorf("non existent RPC URL for Ethereum chain")
	}

	var marshalledRequest []byte
	switch cmd {
	case command.PostTransaction:
		var req command.PostTransactionRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		err := dec.Decode(&req)
		if err != nil {
			return nil, err
		}
		marshalledRequest, err = json.Marshal(jsonrpcRequest{
			ID:      1,
			JSONRPC: "2.0",
			METHOD:  "eth_sendRawTransaction",
			Params:  []string{req.TxHex},
		})
		if err != nil {
			return nil, err
		}

	case command.EthQueryTransaction:
		var req command.EthQueryTransactionRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		err := dec.Decode(&req)
		if err != nil {
			return nil, err
		}
		blockNumberRequest := jsonrpcRequest{
			ID:      2,
			JSONRPC: "2.0",
			METHOD:  "eth_blockNumber",
		}
		receiptRequest := jsonrpcRequest{
			ID:      1,
			JSONRPC: "2.0",
			METHOD:  "eth_getTransactionReceipt",
			Params:  []string{req.TxHash},
		}
		marshalledRequest, err = json.Marshal([]jsonrpcRequest{
			blockNumberRequest,
			receiptRequest,
		})
		if err != nil {
			return nil, err
		}

	case command.EthQuery:
		var req command.EthQueryRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		err := dec.Decode(&req)
		if err != nil {
			return nil, err
		}
		nonceRequest := jsonrpcRequest{
			ID:      1,
			JSONRPC: "2.0",
			METHOD:  "eth_getTransactionCount",
			Params:  []string{req.From, "pending"},
		}
		gasPriceRequest := jsonrpcRequest{
			ID:      2,
			JSONRPC: "2.0",
			METHOD:  "eth_gasPrice",
		}
		param := map[string]interface{}{
			"from":  req.From,
			"to":    req.To,
			"value": fmt.Sprintf("0x%x", req.Value),
		}
		if req.Data != "" {
			param["data"] = req.Data
		}
		gasEstimateRequest := jsonrpcRequest{
			ID:      3,
			JSONRPC: "2.0",
			METHOD:  "eth_estimateGas",
			Params:  []interface{}{param},
		}
		ethCallRequest := jsonrpcRequest{
			ID:      3,
			JSONRPC: "2.0",
			METHOD:  "eth_call",
			Params:  []interface{}{param, "latest"},
		}
		marshalledRequest, err = json.Marshal([]jsonrpcRequest{
			nonceRequest,
			gasPriceRequest,
			gasEstimateRequest,
			ethCallRequest,
		})
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("invalid cmd %x for chain %d", cmd, ec.chainID)
	}

	if marshalledRequest == nil {
		return nil, fmt.Errorf("unexpected error when wrapping request")
	}
	return &HttpData{Method: "POST", URL: rpcURL, Body: marshalledRequest}, nil
}

func (ec *ETHChain) UnwrapResponse(cmd uint8, payload []RPCResponse) ([]byte, error) {
	// Check if response type is error
	for _, pl := range payload {
		if pl.Error != nil {
			return nil, errCodeAndMsg(pl.Error.Code, pl.Error.Message)
		}
	}

	// Command-wise processing
	switch cmd {
	case command.PostTransaction:
		if len(payload) != 1 {
			return nil, errNumResponse(1, len(payload))
		}
		return json.Marshal(command.PostTransactionResponse{
			TxHash: payload[0].Result,
		})
	case command.EthQueryTransaction:
		if len(payload) != 2 {
			return nil, errNumResponse(2, len(payload))
		}
		return json.Marshal(command.EthQueryTransactionResponse{
			BlockNumber: payload[0].Result,
			Tx:          payload[1].Result,
		})
	case command.EthQuery:
		if len(payload) != 4 {
			return nil, errNumResponse(4, len(payload))
		}
		return json.Marshal(command.EthQueryResponse{
			Nonce:      payload[0].Result,
			GasPrice:   payload[1].Result,
			GasLimit:   payload[2].Result,
			CallResult: payload[3].Result,
		})
	}
	return nil, fmt.Errorf("unexpected error when unwrapping response")
}
