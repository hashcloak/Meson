package command

import "math/big"

const (
	EthQuery uint8 = 0x10
)

// Request Types
type EthQueryRequest struct {
	From  string
	To    string
	Value *big.Int
	Data  string
}

// Response Types
type EthQueryResponse struct {
	Nonce    string
	GasPrice string
	GasLimit string
}
