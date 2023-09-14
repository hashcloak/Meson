package command

import "math/big"

const (
	BtcQuery            uint8 = 0x20
	BtcQueryTransaction uint8 = 0x21
)

// Request Types
type BtcQueryRequest struct {
	Min    *big.Int
	Max    *big.Int
	Target string
}

type BtcQueryTransactionRequest struct {
	TxHash string
}

// Response Types
type BtcQueryResponse struct {
	Utxo string
}

type BtcQueryTransactionResponse struct {
	Tx string
}
