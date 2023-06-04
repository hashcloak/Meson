package command

import "math/big"

const (
	EthQuery            uint8 = 0x10
	EthQueryTransaction uint8 = 0x11
)

// Request Types
type EthQueryRequest struct {
	From  string
	To    string
	Value *big.Int
	Data  string
}

type EthQueryTransactionRequest struct {
	TxHash string
}

// Response Types
type EthQueryResponse struct {
	Nonce      string
	GasPrice   string
	GasLimit   string
	CallResult string
}

type EthQueryTransactionResponse struct {
	BlockNumber string
	Tx          string
}
