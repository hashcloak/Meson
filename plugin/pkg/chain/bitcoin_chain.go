package chain

import (
	"encoding/json"
	"fmt"

	"github.com/hashcloak/Meson/plugin/pkg/command"
	"github.com/ugorji/go/codec"
)

// BTCChain is a struct for identifier blockchains and their forks
type BTCChain struct {
	testnet bool
	ticker  string
}

func (ec *BTCChain) WrapRequest(rpcURL string, cmd uint8, payload []byte) (*HttpData, error) {
	if len(rpcURL) == 0 {
		return nil, fmt.Errorf("non existent RPC URL for Bitcoin chain")
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
			METHOD:  "sendrawtransaction",
			Params:  []string{req.TxHex},
		})
		if err != nil {
			return nil, err
		}

	case command.BtcQueryTransaction:
		var req command.BtcQueryTransactionRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		err := dec.Decode(&req)
		if err != nil {
			return nil, err
		}
		marshalledRequest, err = json.Marshal(jsonrpcRequest{
			ID:      1,
			JSONRPC: "2.0",
			METHOD:  "getrawtransaction",
			Params:  []string{req.TxHash},
		})
		if err != nil {
			return nil, err
		}

	case command.BtcQuery:
		var req command.BtcQueryRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		err := dec.Decode(&req)
		if err != nil {
			return nil, err
		}
		marshalledRequest, err = json.Marshal(jsonrpcRequest{
			ID:      1,
			JSONRPC: "2.0",
			METHOD:  "listunspent",
			Params:  []string{fmt.Sprintf("0x%x", req.Min), fmt.Sprintf("0x%x", req.Max), req.Target},
		})
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("invalid cmd %x for bitcoin chain", cmd)
	}

	if marshalledRequest == nil {
		return nil, fmt.Errorf("unexpected error when wrapping request")
	}
	return &HttpData{Method: "POST", URL: rpcURL, Body: marshalledRequest}, nil
}

func (ec *BTCChain) UnwrapResponse(cmd uint8, payload []RPCResponse) ([]byte, error) {
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
	case command.BtcQueryTransaction:
		if len(payload) != 1 {
			return nil, errNumResponse(1, len(payload))
		}
		return json.Marshal(command.BtcQueryTransactionResponse{
			Tx: payload[0].Result,
		})
	case command.BtcQuery:
		if len(payload) != 1 {
			return nil, errNumResponse(1, len(payload))
		}
		return json.Marshal(command.BtcQueryResponse{
			Utxo: payload[0].Result,
		})
	}
	return nil, fmt.Errorf("unexpected error when unwrapping response")
}
