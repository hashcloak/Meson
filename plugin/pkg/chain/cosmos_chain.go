package chain

import (
	"encoding/json"
	"fmt"

	"github.com/hashcloak/Meson/plugin/pkg/command"
	"github.com/ugorji/go/codec"
)

// CosmosChain is a struct for identifier blockchains and their forks
type CosmosChain struct {
	ticker  string
	chainID int
}

func (ec *CosmosChain) WrapRequest(rpcURL string, cmd uint8, payload []byte) ([]HttpData, error) {
	if len(rpcURL) == 0 {
		return []HttpData{}, fmt.Errorf("no URL value for cosmos api")
	}
	switch cmd {
	case command.PostTransaction:
		var req command.PostTransactionRequest
		dec := codec.NewDecoderBytes(payload, &jsonHandle)
		if err := dec.Decode(&req); err != nil {
			return nil, err
		}
		URL := fmt.Sprintf("%s/broadcast_tx_async?tx=0x%s", rpcURL, req.TxHex)
		return []HttpData{{URL: URL}}, nil
	}
	return nil, fmt.Errorf("invalid cmd %x for chain %d", cmd, ec.chainID)
}

func (ec *CosmosChain) UnwrapResponse(cmd uint8, payload []string) ([]byte, error) {
	switch cmd {
	case command.PostTransaction:
		if len(payload) != 1 {
			return nil, fmt.Errorf("expect 1 response, got %d", len(payload))
		}
		return json.Marshal(command.PostTransactionResponse{
			TxHash: payload[0],
		})
	}
	return nil, fmt.Errorf("unexpected error")
}
